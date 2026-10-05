package e2e

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/config"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
	"github.com/stretchr/testify/require"
)

type TestApp struct {
	cfg      config.Config
	client   *testutil.HTTPClient
	fixtures *Fixtures
	asserts  *Asserts
}

type TestEnv struct {
	Client   *testutil.HTTPClient
	Fixtures *Fixtures
	Asserts  *Asserts
}

func NewTestApp(ctx context.Context, cfg config.Config) *TestApp {
	cli, err := testutil.NewHTTPClient("http://" + cfg.Address())
	if err != nil {
		log.Fatalf("new http client failed: %s", err)
	}
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("new pgxpool failed: %s", err)
	}
	fixtures := &Fixtures{pool: pool, cfg: cfg}
	asserts := &Asserts{pool: pool}
	testApp := TestApp{
		cfg:      cfg,
		client:   cli,
		fixtures: fixtures,
		asserts:  asserts,
	}
	return &testApp
}

func (ta *TestApp) NewEnv(t *testing.T) TestEnv {
	err := ta.ResetDB(t.Context(), ta.fixtures.pool)
	require.NoError(t, err, "reset db failed: %w", err)
	env := TestEnv{
		Client:   ta.client,
		Fixtures: ta.fixtures,
		Asserts:  ta.asserts,
	}
	return env
}

func (ta *TestApp) ResetDB(ctx context.Context, pool *pgxpool.Pool) error {
	err := postgrestest.TruncateTables(ctx, pool)
	if err != nil {
		return fmt.Errorf("truncate tables failed: %w", err)
	}
	return nil
}
