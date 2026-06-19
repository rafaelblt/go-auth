package sessiontest

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/require"
)

func NewSession(t *testing.T, override func(p *session.SessionRestoreParams)) *session.Session {
	t.Helper()

	params := session.SessionRestoreParams{
		ID:        session.NewSessionID(),
		UserID:    user.NewID(),
		IssuedAt:  time.Date(1969, 7, 20, 0, 0, 0, 0, time.UTC),
		RevokedAt: nil,
	}

	if override != nil {
		override(&params)
	}

	entity, err := session.RestoreSession(params)
	require.NoError(t, err)
	return entity
}

func NewRefreshToken(t *testing.T, override func(params *session.RefreshTokenRestoreParams)) *session.RefreshToken {
	t.Helper()

	params := session.RefreshTokenRestoreParams{
		ID:        session.NewRefreshTokenID(),
		SessionID: session.NewSessionID(),
		Hash:      MustRefreshTokenHash(t, "default_hash"),
		ParentID:  nil,
		IssuedAt:  time.Date(1957, 10, 4, 0, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 6, 17, 23, 40, 0, 0, time.UTC),
		UsedAt:    nil,
	}

	if override != nil {
		override(&params)
	}

	entity, err := session.RestoreRefreshToken(params)
	require.NoError(t, err)
	return entity
}
