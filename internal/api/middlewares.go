package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type middleware = func(http.Handler, http.ResponseWriter, *http.Request)

func logging(next http.Handler, w http.ResponseWriter, r *http.Request) {
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
}
