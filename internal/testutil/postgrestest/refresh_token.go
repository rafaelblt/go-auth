package postgrestest

import (
	"testing"

	"github.com/google/uuid"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/stretchr/testify/require"
)

func InsertRefreshToken(t *testing.T, db postgres.DB, token *session.RefreshToken) {
	t.Helper()

	require.NotNil(t, db, "db nil")
	require.NotNil(t, token, "refresh token nil")
	require.False(t, token.IsZero(), "refresh token zero")

	sql := `INSERT INTO refresh_tokens
			(id, session_id, parent_id, hash, expires_at, used_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := db.Exec(t.Context(), sql,
		token.ID().Value(),
		token.SessionID().Value(),
		refreshTokenParentID(token),
		token.Hash().Value(),
		token.ExpiresAt(),
		shared.PtrFromOk(token.UsedAt()),
		token.CreatedAt(),
		token.UpdatedAt(),
	)

	require.NoError(t, err, "refresh token insert failed")
}

func CheckRefreshTokenExists(t *testing.T, db postgres.DB, token *session.RefreshToken) bool {
	t.Helper()

	require.NotNil(t, db, "db nil")
	require.NotNil(t, token, "refresh token nil")
	require.False(t, token.IsZero(), "refresh token zero")

	sql := `SELECT EXISTS(
				SELECT 1 FROM refresh_tokens WHERE
				id=$1 AND
				session_id=$2 AND
				parent_id IS NOT DISTINCT FROM $3 AND
				hash=$4 AND
				expires_at=$5 AND
				used_at IS NOT DISTINCT FROM $6 AND
				created_at=$7 AND
				updated_at=$8
			)`

	var result bool
	err := db.QueryRow(t.Context(), sql,
		token.ID().Value(),
		token.SessionID().Value(),
		refreshTokenParentID(token),
		token.Hash().Value(),
		token.ExpiresAt(),
		shared.PtrFromOk(token.UsedAt()),
		token.CreatedAt(),
		token.UpdatedAt(),
	).Scan(&result)

	require.NoError(t, err, "refresh token exists query failed")

	return result
}

func refreshTokenParentID(token *session.RefreshToken) *uuid.UUID {
	parentID, ok := token.ParentID()
	if !ok {
		return nil
	}
	return shared.Ptr(parentID.Value())
}
