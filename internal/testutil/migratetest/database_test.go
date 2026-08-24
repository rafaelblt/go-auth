package migratetest_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/testutil/migratetest"
	"github.com/rafaelblt/go-auth/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDatabaseWithMigrationsForTest(t *testing.T) {
	ctx := context.Background()
	db := migratetest.NewDatabaseWithMigrationsForTest(t, ctx)

	conn, err := pgx.Connect(ctx, db.ConnectionString())
	require.NoError(t, err, "pgx connect failed")
	t.Cleanup(func() { conn.Close(ctx) })

	latest, err := migrations.Latest()
	require.NoError(t, err)

	var version int64
	var dirty bool
	err = conn.QueryRow(ctx,
		"SELECT version, dirty FROM schema_migrations LIMIT 1",
	).Scan(&version, &dirty)

	require.NoError(t, err)
	assert.EqualValues(t, latest, version)
	assert.False(t, dirty)
}
