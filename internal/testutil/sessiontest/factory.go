package sessiontest

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/stretchr/testify/require"
)

func NewRefreshToken(t *testing.T, override func(params *session.RefreshTokenRestoreParams)) *session.RefreshToken {
	t.Helper()

    params := session.RefreshTokenRestoreParams{
        ID:        session.NewRefreshTokenID(),
        SessionID: session.NewSessionID(),
        Hash:      MustRefreshTokenHash(t, "default_hash"),
        ParentID:  nil,
        IssuedAt:  time.Now().UTC(),
        ExpiresAt: time.Now().UTC().Add(time.Hour),
        UsedAt:    nil,
    }

    if override != nil {
        override(&params)
    }

    entity, err := session.RestoreRefreshToken(params)
    require.NoError(t, err)
    return entity
}
