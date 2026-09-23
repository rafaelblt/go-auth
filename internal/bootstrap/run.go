package bootstrap

import (
	"context"
	"log/slog"

	"github.com/rafaelblt/go-auth/internal/config"
)

// Run is the whole process: it loads the configuration from the environment,
// builds the app and serves until ctx is cancelled. cmd/api.go only turns the
// error it returns into an exit code.
func Run(ctx context.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		// The configured logger cannot exist yet, so this one line goes to
		// standard error in slog's default format.
		slog.Error("config load failed", "error", err)
		return err
	}

	logger := newLogger(cfg.LogFormat())
	// Nothing reads the default logger. Setting it only keeps stray slog calls,
	// from a library or from code with no logger in hand, in the same format.
	slog.SetDefault(logger)

	logger.Info("building app...")
	app, err := NewApp(ctx, AppParams{Config: cfg, Logger: logger})
	if err != nil {
		logger.Error("app build failed", "error", err)
		return err
	}
	defer app.Close()

	logger.Info("running app...")
	if err := app.Run(ctx); err != nil {
		logger.Error("app run failed", "error", err)
		return err
	}

	logger.Info("stopping app...")
	return nil
}
