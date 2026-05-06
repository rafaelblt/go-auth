package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

var Middlewares = []func(http.Handler) http.Handler {
	JSONContentType,
	Logging,
}

func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewString()
		traceID := uuid.NewString()

		logger := slog.With(
			slog.String("request_id", requestID),
			slog.String("trace_id", traceID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("ip", r.RemoteAddr),
		)

		ctx := context.WithValue(r.Context(), loggerKey, logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
