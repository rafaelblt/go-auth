package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// loggerKeyType is unexported so that no other package can build a key that
// compares equal to loggerKey and overwrite the request logger.
type loggerKeyType struct{}

var loggerKey = loggerKeyType{}

type writerRecorder struct {
	http.ResponseWriter
	status int
}

func (wr *writerRecorder) WriteHeader(code int) {
	wr.status = code
	wr.ResponseWriter.WriteHeader(code)
}

func loggerFrom(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

func logging(next http.Handler, w http.ResponseWriter, r *http.Request) {
	requestID := uuid.NewString()
	start := time.Now().UTC()

	logger := slog.With(
		slog.String("request_id", requestID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("ip", r.RemoteAddr),
	)

	logger.Info("request received")

	ctx := context.WithValue(r.Context(), loggerKey, logger)
	wr := writerRecorder{w, http.StatusOK}
	next.ServeHTTP(&wr, r.WithContext(ctx))

	end := time.Now().UTC()
	logger.Info("request finished",
		slog.Int("status", wr.status),
		slog.Duration("duration", end.Sub(start)))
}
