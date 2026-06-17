package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/infra/bcrypt"
	"github.com/rafaelblt/go-auth/internal/infra/jwt"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/infra/refreshtoken"
)

type App struct {
	router http.Handler
	addr   string
	deps   dependencies
}

type dependencies struct {
	Pool                  *pgxpool.Pool
	UserRepo              *postgres.UserRepo
	CredentialRepo        *postgres.CredentialRepo
	UnitOfWork            *postgres.UnitOfWork
	PasswordHasher        *bcrypt.Hasher
	Clock                 *infra.SystemClock
	AccessTokenService    *jwt.AccessTokenService
	RefreshTokenGenerator *refreshtoken.Generator
}

func NewApp(ctx context.Context, cfg Config) (*App, error) {
	app := App{}

	deps, err := newInfra(ctx, cfg)
	if err != nil {
		return nil, err
	}

	uc, err := newUsecases(cfg, deps)
	if err != nil {
		return nil, err
	}

	router, err := newRouter(ctx, uc)
	if err != nil {
		return nil, err
	}

	app.router = router
	return &app, nil
}

func (app *App) Router() http.Handler {
	return app.router
}

func (app *App) Run(ctx context.Context) error {
	server := &http.Server{Addr: app.addr, Handler: app.router}

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
	app.deps.Pool.Close()
	return nil
}
