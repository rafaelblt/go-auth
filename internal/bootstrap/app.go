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
	Infra    infraDeps
	UseCases usecases
}

type App struct {
	router http.Handler
	cfg    config.Config
	addr   string
	deps   dependencies
}

func NewApp(ctx context.Context, cfg config.Config) (_ *App, err error) {
	if cfg.IsZero() {
		return nil, errors.New("config invalid: not built")
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

	deps := dependencies{Infra: infra, UseCases: uc}

	router, err := newRouter(ctx, deps)
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
			runPeriodic(ctx, slog.Default(), t)
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
