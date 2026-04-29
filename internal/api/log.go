package api

import (
	"context"
	"log/slog"
)

func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	traceID, _ := ctx.Value("trace_id").(string)
    args = append([]any{"trace_id", traceID}, args...)
    slog.Log(ctx, level, msg, args...)
}
