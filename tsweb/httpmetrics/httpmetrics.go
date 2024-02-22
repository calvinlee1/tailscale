package httpmetrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"tailscale.com/util/ctxkey"
)

type labels struct {
	handler  string
	endpoint string
}

var LabelsKey ctxkey.Key[*labels]

var counter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "tsweb_http_requests_total",
	Help: "http request counter",
}, []string{"handler", "endpoint", "method"})
var duration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "tsweb_http_duration_seconds",
	Help: "duration of http requests",
}, []string{"handler", "endpoint"})
var inFlight = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "tsweb_http_requests_in_flight",
	Help: "http request currently in flight",
}, []string{"handler"})

func Instrument(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := &labels{}
		r = r.WithContext(LabelsKey.WithValue(r.Context(), l))
		start := time.Now()
		inFlight.With(prometheus.Labels{
			"handler": l.handler,
		}).Inc()
		h.ServeHTTP(w, r)
		counter.With(prometheus.Labels{
			"handler":  l.handler,
			"endpoint": l.endpoint,
			"method":   r.Method,
		}).Inc()
		inFlight.With(prometheus.Labels{
			"handler": l.handler,
		}).Dec()
		duration.With(prometheus.Labels{
			"handler":  l.handler,
			"endpoint": l.endpoint,
		}).Observe(time.Since(start).Seconds())
	})
}

func SetHandler(r *http.Request, h string) {
	l, ok := LabelsKey.ValueOk(r.Context())
	if ok {
		l.handler = h
	}
}

func SetEndpoint(r *http.Request, e string) {
	l, ok := LabelsKey.ValueOk(r.Context())
	if ok {
		l.endpoint = e
	}
}
