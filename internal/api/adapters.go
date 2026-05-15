package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type response struct {
	StatusCode int
	Body       any
}

type handler interface {
	Handle(*http.Request) response
}

func adaptHandler(h handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := loggerFrom(r.Context())

		resp := h.Handle(r)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)

		if err := json.NewEncoder(w).Encode(resp.Body); err != nil {
			logger.Error("failed to encode http response body",
				"error", err,
				"body_type", fmt.Sprintf("%T", resp.Body))
		}
	}
}

func adaptMiddleware(m middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m(next, w, r)
		})
	}
}
