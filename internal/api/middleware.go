package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/rafaelblt/go-auth/internal/port"
)

// loggerKeyType is unexported so that no other package can build a key that
// compares equal to loggerKey and overwrite the request logger.
type loggerKeyType struct{}

var loggerKey = loggerKeyType{}

func loggerFrom(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

func logging(base *slog.Logger, clock port.Clock) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := uuid.NewString()
			start := clock.Now()

			logger := base.With(
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("ip", r.RemoteAddr),
			)

			logger.Info("request received")

			ctx := context.WithValue(r.Context(), loggerKey, logger)
			wr := writerRecorder{w, http.StatusOK}
			next.ServeHTTP(&wr, r.WithContext(ctx))

			end := clock.Now()
			logger.Info("request finished",
				slog.Int("status", wr.status),
				slog.Duration("duration", end.Sub(start)))
		})
	}
}

type writerRecorder struct {
	http.ResponseWriter
	status int
}

func (wr *writerRecorder) WriteHeader(code int) {
	wr.status = code
	wr.ResponseWriter.WriteHeader(code)
}

func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if rec == http.ErrAbortHandler {
					panic(rec)
				}

				loggerFrom(r.Context()).Error("panic recovered", "panic", rec)
				writeJSON(r.Context(), w, internalServerError())
			}
		}()
		next.ServeHTTP(w, r)
	})
}

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
