package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/rafaelblt/go-auth/internal/api"
	"github.com/rafaelblt/go-auth/internal/testutil"
)

func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Aplica de trás pra frente pra manter a ordem correta
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func main() {
	logger := slog.Default()

	logger.Info("starting app...")

	logger.Info("starting test database...")
	db, err := testutil.NewDatabase(context.Background())
	if err != nil {
		logger.Error("failed to create test database", "error", err)
		os.Exit(1)
	}
	logger.Info("test database created successfully")

	logger.Info("creating api...")
	handler, err := api.NewAPI(context.Background(), api.APIConfig{DBConnection: db.ConnectionString()})
	if err != nil {
		logger.Error("failed to create api", "error", err)
		os.Exit(1)
	}
	logger.Info("api ready")

	logger.Info("listening...")
	http.ListenAndServe(":8080", handler)
}
