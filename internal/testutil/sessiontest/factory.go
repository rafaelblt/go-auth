package sessiontest

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/stretchr/testify/require"
)

func NewSession(t *testing.T, override func(p *session.SessionRestoreParams)) *session.Session {
	t.Helper()

	params := session.SessionRestoreParams{
		ID:        session.NewSessionID(),
		UserID:    user.NewID(),
		RevokedAt: nil,
		CreatedAt: time.Date(1961, 4, 12, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(1969, 7, 20, 0, 0, 0, 0, time.UTC),
	}

	if override != nil {
		override(&params)
	}

	entity, err := session.RestoreSession(params)
	require.NoError(t, err)
	return entity
}

func NewRefreshToken(t *testing.T, override func(p *session.RefreshTokenRestoreParams)) *session.RefreshToken {
	t.Helper()

	params := session.RefreshTokenRestoreParams{
		ID:        session.NewRefreshTokenID(),
		SessionID: session.NewSessionID(),
		Hash:      MustRefreshTokenHash(t, []byte{0, 1, 0, 7, 2, 0, 2, 6}),
		ParentID:  nil,
		ExpiresAt: time.Date(2026, 6, 17, 23, 40, 0, 0, time.UTC),
		UsedAt:    nil,
		CreatedAt: time.Date(1957, 10, 4, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(1983, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	if override != nil {
		override(&params)
	}

	entity, err := session.RestoreRefreshToken(params)
	require.NoError(t, err)
	return entity
}
