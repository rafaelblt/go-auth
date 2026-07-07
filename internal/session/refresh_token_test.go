package session_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRefreshToken(t *testing.T) {
	testCases := []struct {
		desc      string
		params    session.RefreshTokenCreationParams
		expectErr bool
	}{
		{
			desc: "session id zero",
			params: session.RefreshTokenCreationParams{
				SessionID: session.SessionID{},
				Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
				ParentID:  shared.Ptr(session.NewRefreshTokenID()),
				CreatedAt: time.Now(),
				ExpiresAt: time.Now(),
			},
			expectErr: true,
		},
		{
			desc: "token hash zero",
			params: session.RefreshTokenCreationParams{
				SessionID: session.NewSessionID(),
				Hash:      session.RefreshTokenHash{},
				ParentID:  nil,
				CreatedAt: time.Now(),
				ExpiresAt: time.Now(),
			},
			expectErr: true,
		},
		{
			desc: "parent id zero",
			params: session.RefreshTokenCreationParams{
				SessionID: session.NewSessionID(),
				Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
				ParentID:  shared.Ptr(session.RefreshTokenID{}),
				CreatedAt: time.Now(),
				ExpiresAt: time.Now(),
			},
			expectErr: true,
		},
		{
			desc: "created at after expires at",
			params: session.RefreshTokenCreationParams{
				SessionID: session.NewSessionID(),
				Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
				ParentID:  shared.Ptr(session.NewRefreshTokenID()),
				CreatedAt: time.Now().Add(time.Minute),
				ExpiresAt: time.Now(),
			},
			expectErr: true,
		},
		{
			desc: "valid case",
			params: session.RefreshTokenCreationParams{
				SessionID: session.NewSessionID(),
				Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
				ParentID:  shared.Ptr(session.NewRefreshTokenID()),
				CreatedAt: time.Now(),
				ExpiresAt: time.Now(),
			},
			expectErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			token, err := session.NewRefreshToken(tC.params)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, token)
				return
			}
			require.NoError(t, err)
			require.NotZero(t, token)
			assert.NotZero(t, token.ID())
			assert.Equal(t, tC.params.SessionID, token.SessionID())
			assert.Equal(t, tC.params.Hash, token.Hash())
			assert.Equal(t, tC.params.ParentID, shared.PtrFromOk(token.ParentID()))
			assert.Equal(t, tC.params.ExpiresAt, token.ExpiresAt())
			assert.Nil(t, shared.PtrFromOk(token.UsedAt()))
			assert.Equal(t, tC.params.CreatedAt, token.CreatedAt())
			assert.Equal(t, tC.params.CreatedAt, token.UpdatedAt())
		})
	}
}

func TestNewRefreshToken_ClonesParentID(t *testing.T) {
	provided := shared.Ptr(session.NewRefreshTokenID())
	token, err := session.NewRefreshToken(session.RefreshTokenCreationParams{
		SessionID: session.NewSessionID(),
		Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
		ParentID:  provided,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now(),
	})
	require.NoError(t, err)

	*provided = session.NewRefreshTokenID()

	retrieved, ok := token.ParentID()
	assert.True(t, ok)
	assert.NotEqual(t, provided.Value(), retrieved.Value())
}

func TestRestoreRefreshToken(t *testing.T) {
	testCases := []struct {
		desc       string
		params     session.RefreshTokenRestoreParams
		normalized string
		expectErr  bool
	}{
		{
			desc: "id zero",
			params: session.RefreshTokenRestoreParams{
				ID:        session.RefreshTokenID{},
				SessionID: session.NewSessionID(),
				Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
				ParentID:  shared.Ptr(session.NewRefreshTokenID()),
				CreatedAt: time.Now(),
				ExpiresAt: time.Now(),
				UsedAt:    shared.Ptr(time.Now()),
			},
			expectErr: true,
		},
		{
			desc: "session id zero",
			params: session.RefreshTokenRestoreParams{
				ID:        session.NewRefreshTokenID(),
				SessionID: session.SessionID{},
				Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
				ParentID:  shared.Ptr(session.NewRefreshTokenID()),
				CreatedAt: time.Now(),
				ExpiresAt: time.Now(),
				UsedAt:    shared.Ptr(time.Now()),
			},
			expectErr: true,
		},
		{
			desc: "token hash zero",
			params: session.RefreshTokenRestoreParams{
				ID:        session.NewRefreshTokenID(),
				SessionID: session.NewSessionID(),
				Hash:      session.RefreshTokenHash{},
				ParentID:  nil,
				CreatedAt: time.Now(),
				ExpiresAt: time.Now(),
				UsedAt:    shared.Ptr(time.Now()),
			},
			expectErr: true,
		},
		{
			desc: "parent id zero",
			params: session.RefreshTokenRestoreParams{
				ID:        session.NewRefreshTokenID(),
				SessionID: session.NewSessionID(),
				Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
				ParentID:  shared.Ptr(session.RefreshTokenID{}),
				CreatedAt: time.Now(),
				ExpiresAt: time.Now(),
				UsedAt:    shared.Ptr(time.Now()),
			},
			expectErr: true,
		},
		{
			desc: "only required fields",
			params: session.RefreshTokenRestoreParams{
				ID:        session.NewRefreshTokenID(),
				SessionID: session.NewSessionID(),
				Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
				ParentID:  nil,
				CreatedAt: time.Now(),
				ExpiresAt: time.Now(),
				UsedAt:    nil,
			},
		},
		{
			desc: "with all fields",
			params: session.RefreshTokenRestoreParams{
				ID:        session.NewRefreshTokenID(),
				SessionID: session.NewSessionID(),
				Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
				ParentID:  shared.Ptr(session.NewRefreshTokenID()),
				CreatedAt: time.Now(),
				ExpiresAt: time.Now(),
				UsedAt:    shared.Ptr(time.Now()),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			token, err := session.RestoreRefreshToken(tC.params)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, token)
				return
			}
			require.NoError(t, err)
			require.NotZero(t, token)
			assert.Equal(t, tC.params.ID, token.ID())
			assert.Equal(t, tC.params.SessionID, token.SessionID())
			assert.Equal(t, tC.params.Hash, token.Hash())
			assert.Equal(t, tC.params.ParentID, shared.PtrFromOk(token.ParentID()))
			assert.Equal(t, tC.params.ExpiresAt, token.ExpiresAt())
			assert.Equal(t, tC.params.UsedAt, shared.PtrFromOk(token.UsedAt()))
			assert.Equal(t, tC.params.CreatedAt, token.CreatedAt())
			assert.Equal(t, tC.params.UpdatedAt, token.UpdatedAt())
		})
	}
}

func TestRestoreRefreshToken_ClonesParentID(t *testing.T) {
	provided := shared.Ptr(session.NewRefreshTokenID())
	token, err := session.RestoreRefreshToken(session.RefreshTokenRestoreParams{
		ID:        session.NewRefreshTokenID(),
		SessionID: session.NewSessionID(),
		Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
		ParentID:  provided,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now(),
	})
	require.NoError(t, err)

	*provided = session.NewRefreshTokenID()

	retrieved, ok := token.ParentID()
	assert.True(t, ok)
	assert.NotEqual(t, provided.Value(), retrieved.Value())
}

func TestRestoreRefreshToken_ClonesUsedAt(t *testing.T) {
	provided := shared.Ptr(time.Now())
	token, err := session.RestoreRefreshToken(session.RefreshTokenRestoreParams{
		ID:        session.NewRefreshTokenID(),
		SessionID: session.NewSessionID(),
		Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now(),
		UsedAt:    provided,
	})
	require.NoError(t, err)

	*provided = time.Now().Add(time.Hour)

	retrieved, ok := token.UsedAt()
	assert.True(t, ok)
	assert.False(t, provided.Equal(retrieved))
}

func TestRefreshToken_Use(t *testing.T) {
	usedAt := time.Now()
	testCases := []struct {
		desc        string
		token       *session.RefreshToken
		expectedErr error
	}{
		{
			desc: "ExpiresAt one hour after UsedAt",
			token: sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
				p.ExpiresAt = usedAt.Add(time.Hour)
			}),
		},
		{
			desc: "ExpiresAt and UsedAt equal",
			token: sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
				p.ExpiresAt = usedAt
			}),
		},
		{
			desc: "UsedAt after ExpiresAt",
			token: sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
				p.ExpiresAt = usedAt.Add(-1)
			}),
			expectedErr: session.ErrTokenExpired,
		},
		{
			desc: "token already used",
			token: sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
				p.ExpiresAt = usedAt.Add(-2)
				p.UsedAt = shared.Ptr(usedAt.Add(-1))
			}),
			expectedErr: session.ErrTokenAlreadyUsed,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := tC.token.Use(usedAt)
			assert.ErrorIs(t, err, tC.expectedErr)
			if tC.expectedErr == nil {
				retrievedUsedAt, isUsed := tC.token.UsedAt()
				assert.True(t, isUsed)
				assert.Equal(t, usedAt, retrievedUsedAt)
				assert.Equal(t, usedAt, tC.token.UpdatedAt())
			}
		})
	}
}

func TestRefreshToken_IsUsed_ReturnsTrue_WhenTokenIsUsed(t *testing.T) {
	token := sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.UsedAt = shared.Ptr(time.Now())
	})
	assert.True(t, token.IsUsed())
}

func TestRefreshToken_IsUsed_ReturnsFalse_WhenTokenIsNotUsed(t *testing.T) {
	token := sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.UsedAt = nil
	})
	assert.False(t, token.IsUsed())
}

func TestRefreshToken_HasParent_ReturnsTrue_WhenTokenHasParent(t *testing.T) {
	token := sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.ParentID = shared.Ptr(session.NewRefreshTokenID())
	})
	assert.True(t, token.HasParent())
}

func TestRefreshToken_HasParent_ReturnsFalse_WhenTokenHasNotParent(t *testing.T) {
	token := sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.ParentID = nil
	})
	assert.False(t, token.HasParent())
}

func TestRefreshToken_IsZero_ReturnsFalse_WhenTokenIsZero(t *testing.T) {
	token := session.RefreshToken{}
	assert.True(t, token.IsZero())
}

func TestRefreshToken_IsZero_ReturnsFalse_WhenTokenIsNotZero(t *testing.T) {
	token := sessiontest.NewRefreshToken(t, nil)
	assert.False(t, token.IsZero())
}
