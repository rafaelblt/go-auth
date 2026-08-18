package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rafaelblt/go-auth/internal/bootstrap"
	"github.com/rafaelblt/go-auth/internal/config"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	logger := slog.Default()
	logger.Info("starting app setup...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("loading config...")
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("config load failed", "error", err)
		return err
	}

	logger.Info("building app...")
	app, err := bootstrap.NewApp(ctx, cfg)
	if err != nil {
		logger.Error("app build failed", "error", err)
		return err
	}
	defer app.Close()

	logger.Info("running app...")
	err = app.Run(ctx)
	if err != nil {
		logger.Error("stopping app...", "error", err)
		return err
	}
	logger.Info("stopping app...")
	return nil
}
