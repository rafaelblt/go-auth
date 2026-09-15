package bootstrap

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/rafaelblt/go-auth/internal/config"
)

func NewLogger(format config.LogFormat) *slog.Logger {
	var handler slog.Handler
	switch format {
	case config.LogFormatJSON:
		handler = slog.NewJSONHandler(os.Stdout, nil)
	case config.LogFormatText:
		handler = slog.NewTextHandler(os.Stdout, nil)
	default:
		panic(fmt.Sprintf("log format %q has no handler", format))
	}
	return slog.New(handler)
}
