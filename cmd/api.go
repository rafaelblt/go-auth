package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rafaelblt/go-auth/internal/bootstrap"
	"github.com/rafaelblt/go-auth/internal/testutil"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.Default()

	logger.Info("creating test database...")
	db, err := testutil.NewDatabase(ctx)
	if err != nil {
		logger.Error("failed to create test database", "error", err)
		return err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		db.Close(ctx)
	}()

	os.Setenv("DATABASE_URL", db.ConnectionString())
	os.Setenv("ADDRESS", "localhost:8080")
	logger.Info("loading config...")
	cfg, err := bootstrap.LoadConfig()
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
