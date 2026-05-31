package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/api"
	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/infra/bcrypt"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
)

type App struct {
	handler http.Handler

	addr string

	pool   *pgxpool.Pool
	users  *postgres.UserRepo
	creds  *postgres.CredentialRepo
	uow    *postgres.UnitOfWork
	hasher *bcrypt.Hasher
	clock  *infra.SystemClock
}

type Config struct {
	Database string
}

func New(ctx context.Context, cfg Config) (*App, error) {
	app := App{}

	err := app.buildDeps(ctx, cfg)
	if err != nil {
		return nil, err
	}

	register, err := app.buildRegister()
	if err != nil {
		return nil, err
	}

	router, err := api.NewRouter(ctx, api.Config{
		Dependencies: api.Dependencies{
			Register: register,
		},
	})

	app.handler = router

	return &app, nil
}

func (app *App) Handler() http.Handler {
	return app.handler
}

func (app *App) Run(ctx context.Context) error {
	server := &http.Server{Addr: app.addr, Handler: app.handler}

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
	app.pool.Close()
	return nil
}

func (app *App) buildDeps(ctx context.Context, cfg Config) error {
	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("postgres pool creation failed: %w", err)
	}

	uow, err := postgres.NewUnitOfWork(pool)
	if err != nil {
		return fmt.Errorf("unit of work creation failed: %w", err)
	}

	userRepo, err := postgres.NewUserRepo(pool)
	if err != nil {
		return fmt.Errorf("user repo creation failed: %w", err)
	}

	credRepo, err := postgres.NewCredentialRepo(pool)
	if err != nil {
		return fmt.Errorf("credential repo creation failed: %w", err)
	}

	hasher, err := bcrypt.NewHasher(bcrypt.Config{Cost: 8})
	if err != nil {
		return fmt.Errorf("bcrypt hasher creation failed: %w", err)
	}

	app.pool = pool
	app.uow = uow
	app.users = &userRepo
	app.creds = &credRepo
	app.hasher = &hasher
	return nil
}

func (app *App) buildRegister() (register.Register, error) {
	uc, err := register.New(register.Config{
		UserExistsChecker: app.users,
		UnitOfWork:        app.uow,
		PasswordHasher:    app.hasher,
		Clock:             app.clock,
	})
	if err != nil {
		return register.Register{}, fmt.Errorf("register creation failed: %w", err)
	}
	return uc, nil
}
