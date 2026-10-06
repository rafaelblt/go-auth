package e2e

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/bootstrap"
	"github.com/rafaelblt/go-auth/internal/config"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
	"github.com/stretchr/testify/assert"
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

// startSecondApp starts another app with params, on its own address and over
// the suite's database unless params names another, and stops it when the test
// ends. The shared app keeps the defaults every other test relies on.
func startSecondApp(t *testing.T, params config.ConfigParams) *testutil.HTTPClient {
	t.Helper()

	params.Address = "localhost:8081"
	if params.DatabaseURL == "" {
		params.DatabaseURL = testApp.cfg.DatabaseURL()
	}
	params.BcryptCost = shared.Ptr(6)
	cfg, err := config.NewConfig(params)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	app, err := bootstrap.NewApp(ctx, bootstrap.AppParams{Config: cfg})
	require.NoError(t, err)

	var runErr error
	stopped := make(chan struct{})
	go func() {
		runErr = app.Run(ctx)
		close(stopped)
	}()
	t.Cleanup(func() {
		cancel()
		<-stopped
		app.Close()
		assert.NoError(t, runErr)
	})

	baseURL := "http://" + cfg.Address()
	deadline := time.After(5 * time.Second)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		resp, err := http.Get(baseURL + JWKSPath)
		if err == nil {
			resp.Body.Close()
			break
		}
		select {
		case <-stopped:
			t.Fatalf("second app stopped before serving: %v", runErr)
		case <-deadline:
			t.Fatalf("second app not serving after 5s: %v", err)
		case <-ticker.C:
		}
	}

	client, err := testutil.NewHTTPClient(baseURL)
	require.NoError(t, err)
	return client
}
