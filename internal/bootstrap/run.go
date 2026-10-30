package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/rafaelblt/go-auth/internal/config"
)

// Run is the whole process: it loads the configuration from the environment,
// builds the app and serves until ctx is cancelled. cmd/api.go only turns the
// error it returns into an exit code.
func Run(ctx context.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		// The configured logger cannot exist yet, and the error is a report of
		// one line per setting, which a log line would quote onto one line: it
		// goes to standard error as it is.
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	logger := newLogger(cfg.LogFormat())
	// The service passes its logger by hand. Setting the default keeps stray
	// slog calls, and the standard log package that net/http writes through, in
	// the same format.
	slog.SetDefault(logger)

	logger.Info("building app...")
	app, err := NewApp(ctx, AppParams{Config: cfg, Logger: logger})
	if err != nil {
		logger.Error("app build failed", "error", err)
		return err
	}
	defer app.Close()

	logRateLimit(logger, cfg)
	logSigningKeyEncryption(logger, cfg)
	logger.Info("running app...")
	if err := app.Run(ctx); err != nil {
		logger.Error("app run failed", "error", err)
		return err
	}

	logger.Info("stopping app...")
	return nil
}
