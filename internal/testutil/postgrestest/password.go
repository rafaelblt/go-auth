package postgrestest

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/password"
	"github.com/stretchr/testify/require"
)

func InsertPassword(t *testing.T, db postgres.DB, pwd *password.Password) {
	t.Helper()

	require.NotNil(t, db, "db nil")
	require.NotNil(t, pwd, "password nil")
	require.False(t, pwd.IsZero(), "password zero")

	sql := `INSERT INTO passwords
			(id, user_id, hash, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)`

	_, err := db.Exec(t.Context(), sql,
		pwd.ID().Value(),
		pwd.UserID().Value(),
		pwd.Hash().Value(),
		pwd.CreatedAt(),
		pwd.UpdatedAt(),
	)

	require.NoError(t, err, "password insert failed")
}
