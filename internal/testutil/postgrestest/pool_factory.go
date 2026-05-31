package postgrestest

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/testutil"
)

type PoolFactory struct {
	db   *testutil.Database
	pool *pgxpool.Pool
}

func NewPoolFactory(ctx context.Context) (*PoolFactory, error) {
	db, err := testutil.NewDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("database creation failed: %w", err)
	}
	pool, err := postgres.NewPool(ctx, db.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("pool creation failed: %w", err)
	}
	return &PoolFactory{db, pool}, nil
}

func (pf *PoolFactory) Acquire(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	ResetDB(ctx, pf.pool)
	return pf.pool
}

func (pf *PoolFactory) Close(ctx context.Context) error {
	pf.pool.Close()
	if err := pf.db.Close(ctx); err != nil {
		return fmt.Errorf("failed to close database: %w", err) 
	}
	return nil
}
