// Package bootstrap is the composition root: it wires every dependency, checks
// that the database schema matches the binary, and runs the HTTP server and the
// background tasks. It is the only package that knows how the whole application
// fits together.
//
// See docs/architecture/startup.md.
package bootstrap

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/rafaelblt/go-auth/internal/config"
)

type dependencies struct {
	Logger   *slog.Logger
	Infra    infraDeps
	UseCases usecases
}

type App struct {
	router http.Handler
	addr   string
	deps   dependencies
}

type AppParams struct {
	Config config.Config

	// Logger is optional: a nil one is built from Config's log format, which
	// already decides what the logger would be.
	Logger *slog.Logger
}

func NewApp(ctx context.Context, params AppParams) (_ *App, err error) {
	cfg := params.Config
	if cfg.IsZero() {
		return nil, errors.New("config invalid: not built")
	}

	logger := params.Logger
	if logger == nil {
		logger = newLogger(cfg.LogFormat())
	}

	infra, err := newInfra(ctx, cfg)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			infra.close()
		}
	}()

	if cfg.AutoMigrate() {
		if err = autoMigrate(cfg.DatabaseURL()); err != nil {
			return nil, err
		}
	}

	err = verifySchema(ctx, infra.pool)
	if err != nil {
		return nil, err
	}

	uc, err := newUsecases(cfg, infra)
	if err != nil {
		return nil, err
	}

	deps := dependencies{Logger: logger, Infra: infra, UseCases: uc}

	router, err := newRouter(cfg, deps)
	if err != nil {
		return nil, err
	}

	app := App{
		router: router,
		addr:   cfg.Address(),
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
			run:      app.deps.Infra.Ed25519Keyring.Rotate,
		},
	}

	var wg sync.WaitGroup
	for _, t := range tasks {
		wg.Go(func() {
			runPeriodic(ctx, app.deps.Logger, t)
		})
	}

	return func() {
		cancel()
		wg.Wait()
	}
}

func (app *App) Close() error {
	app.deps.Infra.close()
	return nil
}
