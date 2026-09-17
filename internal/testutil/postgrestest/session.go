package postgrestest

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/stretchr/testify/require"
)

func InsertSession(t *testing.T, db postgres.DB, sess *session.Session) {
	t.Helper()

	require.NotNil(t, db, "db nil")
	require.NotNil(t, sess, "session nil")
	require.False(t, sess.IsZero(), "session zero")

	sql := `INSERT INTO sessions
			(id, user_id, revoked_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)`

	_, err := db.Exec(t.Context(), sql,
		sess.ID().Value(),
		sess.UserID().Value(),
		shared.PtrFromOk(sess.RevokedAt()),
		sess.CreatedAt(),
		sess.UpdatedAt(),
	)

	require.NoError(t, err, "session insert failed")
}

func CheckSessionExists(t *testing.T, db postgres.DB, sess *session.Session) bool {
	t.Helper()

	require.NotNil(t, db, "db nil")
	require.NotNil(t, sess, "session nil")
	require.False(t, sess.IsZero(), "session zero")

	sql := `SELECT EXISTS(
				SELECT 1 FROM sessions WHERE
				id=$1 AND
				user_id=$2 AND
				revoked_at IS NOT DISTINCT FROM $3 AND
				created_at=$4 AND
				updated_at=$5
			)`

	var result bool
	err := db.QueryRow(t.Context(), sql,
		sess.ID().Value(),
		sess.UserID().Value(),
		shared.PtrFromOk(sess.RevokedAt()),
		sess.CreatedAt(),
		sess.UpdatedAt(),
	).Scan(&result)

	require.NoError(t, err, "session exists query failed")

	return result
}
