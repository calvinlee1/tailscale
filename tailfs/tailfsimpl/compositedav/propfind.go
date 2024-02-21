// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

package compositedav

import (
	"bytes"
	"math"
	"net/http"

	"tailscale.com/tailfs/tailfsimpl/shared"
)

func (h *Handler) handlePROPFIND(w http.ResponseWriter, r *http.Request) {
	pathComponents := shared.CleanAndSplit(r.URL.Path)
	mpl := h.maxPathLength(r)
	if !shared.IsRoot(r.URL.Path) && len(pathComponents)+getDepth(r) > mpl {
		// Delegate to a Child
		depth := getDepth(r)
		if h.StatCache != nil {
			cached := h.StatCache.get(r.URL.Path, depth)
			if cached != nil {
				w.WriteHeader(207)
				w.Write(cached)
				return
			}

			bw := &bufferingResponseWriter{ResponseWriter: w}
			h.delegate(pathComponents[mpl-1:], bw, r)
			b := bw.buf.Bytes()
			if bw.status == 207 {
				h.StatCache.set(r.URL.Path, depth, b)
			}
			w.WriteHeader(bw.status)
			w.Write(b)
			return
		}

		// no caching, no need to buffer response
		h.delegate(pathComponents[mpl-1:], w, r)

		return
	}

	h.handle(w, r)
}

func getDepth(r *http.Request) int {
	switch r.Header.Get("Depth") {
	case "0":
		return 0
	case "1":
		return 1
	case "infinity":
		return math.MaxInt
	}
	return 0
}

type bufferingResponseWriter struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
}

func (bw *bufferingResponseWriter) WriteHeader(statusCode int) {
	bw.status = statusCode
}

func (bw *bufferingResponseWriter) Write(p []byte) (int, error) {
	return bw.buf.Write(p)
}
