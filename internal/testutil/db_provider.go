package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseProvider struct {
	db *Database
}

func NewDatabaseProvider(ctx context.Context) (*DatabaseProvider, error) {
	db, err := NewDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("database creation failed: %w", err)
	}
	return &DatabaseProvider{db}, nil
}

func (p *DatabaseProvider) NewPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	t.Cleanup(func() {
        p.reset(ctx)
    })
	return p.db.Pool()
}

func (p *DatabaseProvider) reset(ctx context.Context) error {
    _, err := p.db.Pool().Exec(ctx, `
        TRUNCATE TABLE users, credentials
        RESTART IDENTITY CASCADE
    `)
    return err
}

func (p *DatabaseProvider) Close(ctx context.Context) error {
	return p.db.Close(ctx)
}
