package session_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSession(t *testing.T) {
	testCases := []struct {
		desc      string
		params    session.SessionCreationParams
		expectErr bool
	}{
		{
			desc: "user id zero",
			params: session.SessionCreationParams{
				UserID:    user.ID{},
				CreatedAt: time.Now(),
			},
			expectErr: true,
		},
		{
			desc: "valid case",
			params: session.SessionCreationParams{
				UserID:    user.NewID(),
				CreatedAt: time.Now(),
			},
			expectErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			sess, err := session.NewSession(tC.params)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Nil(t, sess)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tC.params.UserID, sess.UserID())
			assert.Nil(t, shared.PtrFromOk(sess.RevokedAt()))
			assert.Equal(t, tC.params.CreatedAt, sess.CreatedAt())
			assert.Equal(t, tC.params.CreatedAt, sess.UpdatedAt())
		})
	}
}

func TestRestoreSession(t *testing.T) {
	testCases := []struct {
		desc      string
		params    session.SessionRestoreParams
		expectErr bool
	}{
		{
			desc: "session id zero",
			params: session.SessionRestoreParams{
				ID:        session.SessionID{},
				UserID:    user.NewID(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectErr: true,
		},
		{
			desc: "user id zero",
			params: session.SessionRestoreParams{
				ID:        session.NewSessionID(),
				UserID:    user.ID{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectErr: true,
		},
		{
			desc: "session without revoked at",
			params: session.SessionRestoreParams{
				ID:        session.NewSessionID(),
				UserID:    user.NewID(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		{
			desc: "session with revoked at",
			params: session.SessionRestoreParams{
				ID:        session.NewSessionID(),
				UserID:    user.NewID(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				RevokedAt: shared.Ptr(time.Now()),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			sess, err := session.RestoreSession(tC.params)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Nil(t, sess)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tC.params.ID, sess.ID())
			assert.Equal(t, tC.params.UserID, sess.UserID())
			assert.Equal(t, tC.params.RevokedAt, shared.PtrFromOk(sess.RevokedAt()))
			assert.Equal(t, tC.params.CreatedAt, sess.CreatedAt())
			assert.Equal(t, tC.params.UpdatedAt, sess.UpdatedAt())
		})
	}
}

func TestRestoreSession_ClonesRevokedAt(t *testing.T) {
	provided := shared.Ptr(time.Now())
	sess, err := session.RestoreSession(session.SessionRestoreParams{
		ID:        session.NewSessionID(),
		UserID:    user.NewID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		RevokedAt: provided,
	})
	require.NoError(t, err)

	*provided = time.Now().Add(time.Hour)

	retrieved, ok := sess.RevokedAt()
	assert.True(t, ok)
	assert.NotEqual(t, *provided, retrieved)
}

func TestSession_Revoke_UpdateSession(t *testing.T) {
	sess := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		p.RevokedAt = nil
	})

	now := time.Now().UTC()
	sess.Revoke(now)

	revokedAt, isRevoked := sess.RevokedAt()
	assert.True(t, isRevoked)
	assert.Equal(t, now, revokedAt)
	assert.Equal(t, now, sess.UpdatedAt())
}

func TestSession_Revoke_NotUpdateSession_WhenIsAlreadyRevoked(t *testing.T) {
	alreadyRevoked := time.Now()
	sess := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		p.RevokedAt = &alreadyRevoked
		p.UpdatedAt = alreadyRevoked
	})

	now := time.Now().UTC().Add(time.Hour)
	sess.Revoke(now)

	revokedAt, isRevoked := sess.RevokedAt()
	assert.True(t, isRevoked)
	assert.Equal(t, alreadyRevoked, revokedAt)
	assert.Equal(t, alreadyRevoked, sess.UpdatedAt())
}

func TestSession_IsRevoked_ReturnsTrue_WhenSessionIsRevoked(t *testing.T) {
	sess := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		p.RevokedAt = shared.Ptr(time.Now())
	})
	assert.True(t, sess.IsRevoked())
}

func TestSession_IsRevoked_ReturnsFalse_WhenSessionIsNotRevoked(t *testing.T) {
	sess := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		p.RevokedAt = nil
	})
	assert.False(t, sess.IsRevoked())
}

func TestSession_IsZero_ReturnsFalse_WhenSessionIsZero(t *testing.T) {
	sess := session.Session{}
	assert.True(t, sess.IsZero())
}

func TestSession_IsZero_ReturnsFalse_WhenSessionIsNotZero(t *testing.T) {
	sess := sessiontest.NewSession(t, nil)
	assert.False(t, sess.IsZero())
}
