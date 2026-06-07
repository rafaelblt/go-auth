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

type middleware = func(http.Handler, http.ResponseWriter, *http.Request)

func adaptHandler(h handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := loggerFrom(r.Context())

		resp := h.Handle(r)

		var status int
		var buf []byte

		buf, err := json.Marshal(resp.Body)
		if err != nil {
			msg := "failed to encode http response body"
			logger.Error(msg, "error", err, "body_type", fmt.Sprintf("%T", resp.Body))
			buf, _ = json.Marshal(internalServerErrorBody)
			status = http.StatusInternalServerError
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(buf)
	}
}

func adaptMiddleware(m middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m(next, w, r)
		})
	}
}
