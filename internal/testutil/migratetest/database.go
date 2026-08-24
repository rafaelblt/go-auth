package migratetest

import (
	"context"
	"fmt"
	"testing"

	"github.com/rafaelblt/go-auth/internal/migrate"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/require"
)

func NewDatabaseWithMigrations(ctx context.Context) (*testutil.Database, error) {
	db, err := testutil.NewDatabase(ctx)
	if err != nil {
		return nil, err
	}

	if err := migrate.RunMigrations(db.ConnectionString()); err != nil {
		db.Close(ctx)
		return nil, fmt.Errorf("run migrations failed: %w", err)
	}

	return db, nil
}

func NewDatabaseWithMigrationsForTest(t *testing.T, ctx context.Context) *testutil.Database {
	db, err := NewDatabaseWithMigrations(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close(ctx)
	})

	return db
}
