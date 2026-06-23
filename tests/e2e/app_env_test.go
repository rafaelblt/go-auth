package e2e

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/bootstrap"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/require"
)

type TestApp struct {
	cli *testutil.HTTPClient
	fxt *Fixtures
}

type TestEnv struct {
	Client   *testutil.HTTPClient
	Fixtures *Fixtures
}

func NewTestApp(ctx context.Context, cfg bootstrap.Config) *TestApp {
	cli, err := testutil.NewHTTPClient("http://" + cfg.Address)
	if err != nil {
		log.Fatalf("new http client failed: %s", err)
	}
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("new pgxpool failed: %s", err)
	}
	fxt := &Fixtures{pool: pool, cfg: cfg}
	testApp := TestApp{
		cli: cli,
		fxt: fxt,
	}
	return &testApp
}

func (ta *TestApp) NewEnv(t *testing.T) TestEnv {
	err := ta.ResetDB(t.Context(), ta.fxt.pool)
	require.NoError(t, err, "reset db failed: %w", err)
	env := TestEnv{
		Client:   ta.cli,
		Fixtures: ta.fxt,
	}
	return env
}

func (ta *TestApp) ResetDB(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
        TRUNCATE TABLE users, credentials
        RESTART IDENTITY CASCADE
    `)
	if err != nil {
		return fmt.Errorf("truncate tables failed: %w", err)
	}
	return nil
}
