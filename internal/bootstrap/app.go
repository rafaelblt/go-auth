package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type App struct {
	router http.Handler
	cfg    Config
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
	stopBackground := app.startBackground(ctx)
	defer stopBackground()

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
		shutdownCtx, cancel := context.WithTimeout(
			context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}

func (app *App) startBackground(ctx context.Context) (stop func()) {
	ctx, cancel := context.WithCancel(ctx)

	tasks := []periodicTask{
		{
			name:     "jwt_keyring_rotation",
			interval: 7 * 24 * time.Hour,
			timeout:  3 * time.Second,
			run:      app.deps.Ed25519Keyring.Rotate,
		},
	}

	var wg sync.WaitGroup
	for _, t := range tasks {
		wg.Go(func() {
			runPeriodic(ctx, slog.Default(), t)
		})
	}

	return func() {
		cancel()
		wg.Wait()
	}
}

func (app *App) Close() error {
	app.deps.pool.Close()
	return nil
}
