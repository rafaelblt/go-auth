package postgrestest

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/require"
)

func InsertUser(t *testing.T, db postgres.DB, usr *user.User) {
	t.Helper()

	require.NotNil(t, usr, "db nil")
	require.NotNil(t, usr, "user nil")
	require.NotZero(t, usr, "user zero")

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

func InsertSession(t *testing.T, db postgres.DB, sess *session.Session) {
	t.Helper()

	require.NotNil(t, sess, "db nil")
	require.NotNil(t, sess, "session nil")
	require.NotZero(t, sess, "session zero")

	sql := `INSERT INTO sessions
			(id, user_id, issued_at, revoked_at)
			VALUES ($1, $2, $3, $4)`

	var revokedAt *time.Time
	r, ok := sess.RevokedAt()
	if ok {
		revokedAt = &r
	}
	_, err := db.Exec(t.Context(), sql,
		sess.ID().Value(),
		sess.UserID().String(),
		sess.IssuedAt(),
		revokedAt,
	)

	require.NoError(t, err, "session insert failed")
}
