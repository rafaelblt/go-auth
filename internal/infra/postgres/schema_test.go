package postgres_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func beginTxForTest(t *testing.T) pgx.Tx {
	t.Helper()
	pool := poolFactory.Acquire(t)
	tx, err := pool.Begin(context.Background())
	require.NoError(t, err, "tx begin failed")
	t.Cleanup(func() { tx.Rollback(context.Background()) })
	return tx
}

func TestSchemaVersion_WithDBNil(t *testing.T) {
	version, err := postgres.SchemaVersion(t.Context(), nil)

	assert.Error(t, err)
	assert.Zero(t, version)
}

func TestSchemaVersion_WithMigratedDB(t *testing.T) {
	pool := poolFactory.Acquire(t)
	latest, err := migrations.Latest()
	require.NoError(t, err)

	version, err := postgres.SchemaVersion(t.Context(), pool)

	assert.NoError(t, err)
	assert.Equal(t, latest, version)
}

func TestSchemaVersion_WithoutSchemaMigrationsTable(t *testing.T) {
	tx := beginTxForTest(t)
	_, err := tx.Exec(t.Context(), "DROP TABLE schema_migrations")
	require.NoError(t, err, "drop schema_migrations failed")

	version, err := postgres.SchemaVersion(t.Context(), tx)

	assert.ErrorIs(t, err, postgres.ErrSchemaNotInitialized)
	assert.Zero(t, version)
}

func TestSchemaVersion_WithEmptySchemaMigrations(t *testing.T) {
	tx := beginTxForTest(t)
	_, err := tx.Exec(t.Context(), "DELETE FROM schema_migrations")
	require.NoError(t, err, "delete schema_migrations rows failed")

	version, err := postgres.SchemaVersion(t.Context(), tx)

	assert.ErrorIs(t, err, postgres.ErrSchemaNotInitialized)
	assert.Zero(t, version)
}

func TestSchemaVersion_WithNilVersionRecorded(t *testing.T) {
	tx := beginTxForTest(t)
	_, err := tx.Exec(t.Context(), "UPDATE schema_migrations SET version = -1, dirty = false")
	require.NoError(t, err, "update schema_migrations failed")

	version, err := postgres.SchemaVersion(t.Context(), tx)

	assert.ErrorIs(t, err, postgres.ErrSchemaNotInitialized)
	assert.Zero(t, version)
}

func TestSchemaVersion_WithDirtySchema(t *testing.T) {
	tx := beginTxForTest(t)
	_, err := tx.Exec(t.Context(), "UPDATE schema_migrations SET dirty = true")
	require.NoError(t, err, "update schema_migrations failed")

	version, err := postgres.SchemaVersion(t.Context(), tx)

	assert.ErrorIs(t, err, postgres.ErrSchemaDirty)
	assert.Zero(t, version)
}

func TestSchemaVersion_WithCanceledContext(t *testing.T) {
	pool := poolFactory.Acquire(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	version, err := postgres.SchemaVersion(ctx, pool)

	assert.Error(t, err)
	assert.NotErrorIs(t, err, postgres.ErrSchemaNotInitialized)
	assert.NotErrorIs(t, err, postgres.ErrSchemaDirty)
	assert.Zero(t, version)
}
