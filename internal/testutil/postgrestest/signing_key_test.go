package postgrestest_test

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPERS

func selectSigningKeyRow(t *testing.T, db postgres.DB, generation int64) port.StoredSigningKey {
	t.Helper()

	row := port.StoredSigningKey{Generation: generation}
	err := db.QueryRow(t.Context(),
		`SELECT seed, active_at, created_at FROM signing_keys WHERE generation=$1`,
		generation,
	).Scan(&row.Seed, &row.ActiveAt, &row.CreatedAt)
	require.NoError(t, err, "signing key select failed")

	row.ActiveAt = row.ActiveAt.UTC()
	row.CreatedAt = row.CreatedAt.UTC()
	return row
}

func signingKey(generation int64) port.StoredSigningKey {
	seed := make([]byte, 32)
	rand.Read(seed)
	now := time.Now().UTC().Truncate(time.Microsecond)
	return port.StoredSigningKey{
		Generation: generation,
		Seed:       seed,
		ActiveAt:   now.Add(time.Hour),
		CreatedAt:  now,
	}
}

// TESTS

func TestInsertSigningKey(t *testing.T) {
	db := poolFactory.AcquireWithMigrations(t)
	key := signingKey(1)

	postgrestest.InsertSigningKey(t, db, key)

	assert.Equal(t, key, selectSigningKeyRow(t, db, key.Generation))
}

func TestCheckSigningKeyExists(t *testing.T) {
	db := poolFactory.AcquireWithMigrations(t)
	key := signingKey(1)
	postgrestest.InsertSigningKey(t, db, key)

	with := func(override func(k *port.StoredSigningKey)) port.StoredSigningKey {
		k := key
		override(&k)
		return k
	}

	tests := []struct {
		name     string
		key      port.StoredSigningKey
		expected bool
	}{
		{"matches the key", key, true},
		{"generation differs", with(func(k *port.StoredSigningKey) { k.Generation = 2 }), false},
		{"seed differs", with(func(k *port.StoredSigningKey) { k.Seed = signingKey(1).Seed }), false},
		{"active at differs", with(func(k *port.StoredSigningKey) { k.ActiveAt = k.ActiveAt.Add(time.Second) }), false},
		{"created at differs", with(func(k *port.StoredSigningKey) { k.CreatedAt = k.CreatedAt.Add(time.Second) }), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := postgrestest.CheckSigningKeyExists(t, db, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
