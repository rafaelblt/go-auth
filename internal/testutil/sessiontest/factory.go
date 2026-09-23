// Package sessiontest holds the Session and RefreshToken factories and musts.
//
// See docs/development/testing.md#factories-and-musts.
package sessiontest

import (
	"crypto/sha256"
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
		Hash:      NewRefreshTokenHash(t, "default"),
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

// NewRefreshTokenSecret returns a fresh secret. It goes through
// session.NewRefreshToken, the only way the domain hands one out, and drops
// the token.
func NewRefreshTokenSecret(t *testing.T) session.RefreshTokenSecret {
	t.Helper()

	now := time.Now()
	_, secret, err := session.NewRefreshToken(session.RefreshTokenCreationParams{
		SessionID: session.NewSessionID(),
		CreatedAt: now,
		ExpiresAt: now,
	})
	require.NoError(t, err)
	return secret
}

// NewRefreshTokenHash returns a valid hash derived from seed, so equal seeds
// give equal hashes and different seeds give different ones.
func NewRefreshTokenHash(t *testing.T, seed string) session.RefreshTokenHash {
	t.Helper()

	sum := sha256.Sum256([]byte(seed))
	hash, err := session.NewRefreshTokenHash(sum[:])
	require.NoError(t, err)
	return hash
}
