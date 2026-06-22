package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type App struct {
	router http.Handler
	addr   string
	deps   infraDeps
}

func NewApp(ctx context.Context, cfg Config) (*App, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("config invalid: %w", err)
	}

	deps, err := newInfra(ctx, cfg)
	if err != nil {
		return nil, err
	}

	uc, err := newUsecases(cfg, deps)
	if err != nil {
		deps.pool.Close()
		return nil, err
	}

	router, err := newRouter(ctx, uc)
	if err != nil {
		deps.pool.Close()
		return nil, err
	}

	app := App{
		router: router,
		addr:   cfg.Address,
		deps:   deps,
	}
	return &app, nil
}

func (app *App) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:              app.addr,
		Handler:           app.router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}

func (app *App) Close() error {
	app.deps.pool.Close()
	return nil
}
