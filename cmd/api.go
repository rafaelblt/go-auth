package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rafaelblt/go-auth/internal/app"
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

	logger.Info("building app...")
	application, err := app.New(ctx, app.Config{
		Database: db.ConnectionString(),
	})
	if err != nil {
		logger.Error("app build failed", "error", err)
		return err
	}
	defer application.Close()

	logger.Info("running app...")
	return application.Run(ctx)
}
