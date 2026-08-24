package migrate_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/migrate"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func connectForTest(t *testing.T, connString string) *pgx.Conn {
	t.Helper()
	conn, err := pgx.Connect(t.Context(), connString)
	require.NoError(t, err, "pgx connect failed")
	t.Cleanup(func() { conn.Close(t.Context()) })
	return conn
}

func assertDatabaseSchemaVersion(t *testing.T, dbURL string, expectedVersion uint) {
	t.Helper()
	conn := connectForTest(t, dbURL)

	var version int64
	var dirty bool
	err := conn.QueryRow(t.Context(),
		"SELECT version, dirty FROM schema_migrations LIMIT 1",
	).Scan(&version, &dirty)
	require.NoError(t, err, "query row to get current schema version failed")

	assert.EqualValues(t, expectedVersion, version, "current version different from expected")
	assert.False(t, dirty, "dirty is true")
}

func assertTableExistsInDatabase(t *testing.T, dbURL string, expectedTable string) {
	t.Helper()
	conn := connectForTest(t, dbURL)

	sql := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = '%s'
		)
	`, expectedTable)

	var tableExists bool
	err := conn.QueryRow(t.Context(), sql).Scan(&tableExists)
	require.NoError(t, err, "query row to check expected table failed")

	assert.True(t, tableExists, "expected table not exists")
}

func TestRunMigrations_WithCleanDatabase(t *testing.T) {
	db := testutil.NewDatabaseForTest(t, t.Context())

	err := migrate.RunMigrations(db.ConnectionString())

	require.NoError(t, err)

	latest, err := migrations.Latest()
	require.NoError(t, err)
	assertDatabaseSchemaVersion(t, db.ConnectionString(), latest)
	assertTableExistsInDatabase(t, db.ConnectionString(), "users")
}

func TestRunMigrations_WhenAlreadyMigrated(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDatabaseForTest(t, ctx)
	require.NoError(t, migrate.RunMigrations(db.ConnectionString()))

	err := migrate.RunMigrations(db.ConnectionString())

	assert.NoError(t, err)
}

func TestRunMigrations_WithInvalidURL(t *testing.T) {
	err := migrate.RunMigrations("not-a-valid-url")

	assert.Error(t, err)
}

func TestRunMigrations_WithUnreachableDatabase(t *testing.T) {
	dbURL := "postgres://user:pass@localhost:1/db?sslmode=disable&connect_timeout=1"

	err := migrate.RunMigrations(dbURL)

	assert.Error(t, err)
}
