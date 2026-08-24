package testutil_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDatabaseForTest_StartsCleanDatabase(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDatabaseForTest(t, ctx)

	require.NotEmpty(t, db.ConnectionString())

	conn, err := pgx.Connect(ctx, db.ConnectionString())
	require.NoError(t, err, "pgx connect failed")
	t.Cleanup(func() { conn.Close(ctx) })

	var tables int
	err = conn.QueryRow(ctx, `
        SELECT count(*) FROM information_schema.tables
        WHERE table_schema = 'public'
    `).Scan(&tables)

	require.NoError(t, err)
	assert.Zero(t, tables, "expected database without tables")
}

func TestDatabase_Close(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewDatabase(ctx)
	require.NoError(t, err)

	err = db.Close(ctx)

	assert.NoError(t, err)
}
