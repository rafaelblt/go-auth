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

func makeValidationErrorLogFields(errs map[string]ValidationErrors) slog.Attr {
    attrs := make([]any, 0, len(errs))

	for field, verrs := range errs {
		errCodes := make([]string, 0, len(verrs))
		for _, err := range verrs {
			errCodes = append(errCodes, err.Code)
		}
		attrs = append(attrs, slog.Any(field, errCodes))
	}

	return slog.Group("fields", attrs...)
}