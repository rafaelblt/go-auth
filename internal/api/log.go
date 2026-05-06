package api

import (
	"context"
	"log/slog"
)

const loggerKey = "logger"

func loggerFrom(ctx context.Context) *slog.Logger {
    if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
        return l
    }
    return slog.Default()
}
