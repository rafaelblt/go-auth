package postgrestest

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/stretchr/testify/require"
)

func InsertSigningKey(t *testing.T, db postgres.DB, key port.StoredSigningKey) {
	t.Helper()

	require.NotNil(t, db, "db nil")

	sql := `INSERT INTO signing_keys
			(generation, seed, active_at, created_at)
			VALUES ($1, $2, $3, $4)`

	_, err := db.Exec(t.Context(), sql,
		key.Generation,
		key.Seed,
		key.ActiveAt,
		key.CreatedAt,
	)

	require.NoError(t, err, "signing key insert failed")
}

func CheckSigningKeyExists(t *testing.T, db postgres.DB, key port.StoredSigningKey) bool {
	t.Helper()

	require.NotNil(t, db, "db nil")

	sql := `SELECT EXISTS(
				SELECT 1 FROM signing_keys WHERE
				generation=$1 AND
				seed=$2 AND
				active_at=$3 AND
				created_at=$4
			)`

	var result bool
	err := db.QueryRow(t.Context(), sql,
		key.Generation,
		key.Seed,
		key.ActiveAt,
		key.CreatedAt,
	).Scan(&result)

	require.NoError(t, err, "signing key exists query failed")

	return result
}
