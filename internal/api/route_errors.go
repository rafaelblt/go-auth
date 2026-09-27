package api

import (
	"context"
	"net/http"
)

// See docs/architecture/http.md#unknown-routes.
func jsonRouteErrors(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, pattern := mux.Handler(r); pattern != "" {
			mux.ServeHTTP(w, r)
			return
		}
		mux.ServeHTTP(&routeErrorWriter{ResponseWriter: w, ctx: r.Context()}, r)
	})
}

type routeErrorWriter struct {
	http.ResponseWriter
	ctx      context.Context
	replaced bool
}

func (w *routeErrorWriter) WriteHeader(code int) {
	switch code {
	case http.StatusNotFound:
		writeJSON(w.ctx, w.ResponseWriter, routeNotFoundError())
		w.replaced = true
	case http.StatusMethodNotAllowed:
		writeJSON(w.ctx, w.ResponseWriter, methodNotAllowedError())
		w.replaced = true
	default:
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *routeErrorWriter) Write(b []byte) (int, error) {
	if w.replaced {
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}
