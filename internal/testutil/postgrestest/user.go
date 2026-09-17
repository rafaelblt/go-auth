package postgrestest

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/stretchr/testify/require"
)

func InsertUser(t *testing.T, db postgres.DB, usr *user.User) {
	t.Helper()

	require.NotNil(t, db, "db nil")
	require.NotNil(t, usr, "user nil")
	require.False(t, usr.IsZero(), "user zero")

	sql := `INSERT INTO users
			(id, username, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)`

	_, err := db.Exec(t.Context(), sql,
		usr.ID().Value(),
		usr.Username().String(),
		usr.Status().String(),
		usr.CreatedAt(),
		usr.UpdatedAt(),
	)

	require.NoError(t, err, "user insert failed")
}
