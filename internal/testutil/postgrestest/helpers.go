package postgrestest

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/password"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/require"
)

func InsertUser(t *testing.T, db postgres.DB, usr *user.User) {
	t.Helper()

	require.NotNil(t, db, "db nil")
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

func InsertCredential(t *testing.T, db postgres.DB, cred *password.Credential) {
	t.Helper()

	require.NotNil(t, db, "db nil")
	require.NotNil(t, cred, "credential nil")
	require.NotZero(t, cred, "credential zero")

	sql := `INSERT INTO password_credentials
			(id, user_id, hash, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)`

	_, err := db.Exec(t.Context(), sql,
		cred.ID().Value(),
		cred.UserID().Value(),
		cred.Hash().Value(),
		cred.CreatedAt(),
		cred.UpdatedAt(),
	)

	require.NoError(t, err, "insert credential failed")
}

func InsertSession(t *testing.T, db postgres.DB, sess *session.Session) {
	t.Helper()

	require.NotNil(t, db, "db nil")
	require.NotNil(t, sess, "session nil")
	require.NotZero(t, sess, "session zero")

	sql := `INSERT INTO sessions
			(id, user_id, revoked_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)`

	_, err := db.Exec(t.Context(), sql,
		sess.ID().Value(),
		sess.UserID().String(),
		shared.PtrFromOk(sess.RevokedAt()),
		sess.CreatedAt(),
		sess.UpdatedAt(),
	)

	require.NoError(t, err, "session insert failed")
}

func InsertRefreshToken(t *testing.T, db postgres.DB, token *session.RefreshToken) {
	t.Helper()

	require.NotNil(t, db, "db nil")
	require.NotNil(t, token, "refresh token nil")
	require.NotZero(t, token, "refresh token zero")

	var parentID *string
	pID, ok := token.ParentID()
	if ok {
		parentID = shared.Ptr(pID.String())
	}

	sql := `INSERT INTO refresh_tokens
			(id, session_id, parent_id, hash, expires_at, used_at, created_at, updated_at)
			VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := db.Exec(t.Context(), sql,
		token.ID().String(),
		token.SessionID().String(),
		parentID,
		token.Hash().Value(),
		token.ExpiresAt(),
		shared.PtrFromOk(token.UsedAt()),
		token.CreatedAt(),
		token.UpdatedAt(),
	)

	require.NoError(t, err, "refresh token insert failed")
}

func CheckSessionExists(t *testing.T, db postgres.DB, sess *session.Session) bool {
	t.Helper()

	require.NotNil(t, db, "db nil")
	require.NotNil(t, sess, "session nil")
	require.NotZero(t, sess, "session zero")

	var result bool

	err := db.QueryRow(t.Context(),
		`SELECT EXISTS(
			SELECT 1 FROM sessions WHERE
			id=$1 AND
			user_id=$2 AND
			created_at=$3 AND
			revoked_at IS NOT DISTINCT FROM $4
		)`,
		sess.ID().Value(),
		sess.UserID().Value(),
		sess.CreatedAt(),
		shared.PtrFromOk(sess.RevokedAt()),
	).Scan(&result)
	require.NoError(t, err)

	return result
}

func CheckRefreshTokenExists(t *testing.T, db postgres.DB, token *session.RefreshToken) bool {
	t.Helper()

	require.NotNil(t, db, "db nil")
	require.NotNil(t, token, "refresh token nil")
	require.NotZero(t, token, "refresh token zero")

	var result bool

	var parentID *string
	pID, ok := token.ParentID()
	if ok {
		parentID = shared.Ptr(pID.String())
	}

	err := db.QueryRow(t.Context(),
		`SELECT EXISTS(
			SELECT 1 FROM refresh_tokens WHERE
			id=$1 AND
			session_id=$2 AND
			parent_id IS NOT DISTINCT FROM $3 AND
			hash=$4 AND
			expires_at=$5 AND
			used_at IS NOT DISTINCT FROM $6 AND
			created_at=$7 AND
			updated_at=$8
		)`,
		token.ID().Value(),
		token.SessionID().Value(),
		parentID,
		token.Hash().Value(),
		token.ExpiresAt(),
		shared.PtrFromOk(token.UsedAt()),
		token.CreatedAt(),
		token.UpdatedAt(),
	).Scan(&result)
	require.NoError(t, err, "db query row failed")

	return result
}
