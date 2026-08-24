package postgres_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaVersion_WithDBNil(t *testing.T) {
	version, err := postgres.SchemaVersion(t.Context(), nil)

	assert.Error(t, err)
	assert.Zero(t, version)
}

func TestSchemaVersion_WithMigratedDB(t *testing.T) {
	pool := poolFactory.AcquireWithMigrations(t)
	latest, err := migrations.Latest()
	require.NoError(t, err)

	version, err := postgres.SchemaVersion(t.Context(), pool)

	assert.NoError(t, err)
	assert.Equal(t, latest, version)
}

func TestSchemaVersion_WithoutSchemaMigrationsTable(t *testing.T) {
	pool := poolFactory.Acquire(t)

	version, err := postgres.SchemaVersion(t.Context(), pool)

	assert.ErrorIs(t, err, postgres.ErrSchemaNotInitialized)
	assert.Zero(t, version)
}

func TestSchemaVersion_WithEmptySchemaMigrations(t *testing.T) {
	pool := poolFactory.AcquireWithMigrations(t)
	_, err := pool.Exec(t.Context(), "DELETE FROM schema_migrations")
	require.NoError(t, err, "delete schema_migrations rows failed")

	version, err := postgres.SchemaVersion(t.Context(), pool)

	assert.ErrorIs(t, err, postgres.ErrSchemaNotInitialized)
	assert.Zero(t, version)
}

func TestSchemaVersion_WithNegativeVersionRecorded(t *testing.T) {
	pool := poolFactory.AcquireWithMigrations(t)
	_, err := pool.Exec(t.Context(), "UPDATE schema_migrations SET version = -1, dirty = false")
	require.NoError(t, err, "update schema_migrations failed")

	version, err := postgres.SchemaVersion(t.Context(), pool)

	assert.ErrorIs(t, err, postgres.ErrSchemaNotInitialized)
	assert.Zero(t, version)
}

func TestSchemaVersion_WithDirtySchema(t *testing.T) {
	pool := poolFactory.AcquireWithMigrations(t)
	_, err := pool.Exec(t.Context(), "UPDATE schema_migrations SET dirty = true")
	require.NoError(t, err, "update schema_migrations failed")

	version, err := postgres.SchemaVersion(t.Context(), pool)

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
