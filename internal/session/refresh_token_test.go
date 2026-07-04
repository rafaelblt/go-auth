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

func TestNewRefreshToken_ValidateParams(t *testing.T) {
	testCases := []struct {
		desc       string
		params     session.RefreshTokenCreationParams
		normalized string
		expectErr  bool
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
			assert.NoError(t, err)
			assert.NotZero(t, token)
		})
	}
}

func TestNewRefreshToken_ClonesParentID(t *testing.T) {
	providedPID := shared.Ptr(session.NewRefreshTokenID())
	token, err := session.NewRefreshToken(session.RefreshTokenCreationParams{
		SessionID: session.NewSessionID(),
		Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
		ParentID:  providedPID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now(),
	})
	require.NoError(t, err)

	*providedPID = session.NewRefreshTokenID()
	retrieved, ok := token.ParentID()
	assert.True(t, ok)
	assert.NotEqual(t, providedPID.Value(), retrieved.Value())
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
			assert.Equal(t, tC.params.CreatedAt, token.CreatedAt())
			assert.Equal(t, tC.params.ExpiresAt, token.ExpiresAt())
			parentID, ok := token.ParentID()
			if tC.params.ParentID == nil {
				assert.False(t, ok)
				assert.Zero(t, parentID)
			} else {
				assert.True(t, ok)
				assert.Equal(t, *tC.params.ParentID, parentID)
			}
			usedAt, ok := token.UsedAt()
			if tC.params.UsedAt == nil {
				assert.False(t, ok)
				assert.Zero(t, usedAt)
			} else {
				assert.True(t, ok)
				assert.Equal(t, *tC.params.UsedAt, usedAt)
			}
		})
	}
}

func TestRestoreRefreshToken_ClonesParentID(t *testing.T) {
	providedPID := shared.Ptr(session.NewRefreshTokenID())
	token, err := session.RestoreRefreshToken(session.RefreshTokenRestoreParams{
		ID:        session.NewRefreshTokenID(),
		SessionID: session.NewSessionID(),
		Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
		ParentID:  providedPID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now(),
	})
	require.NoError(t, err)

	*providedPID = session.NewRefreshTokenID()
	retrieved, ok := token.ParentID()
	assert.True(t, ok)
	assert.NotEqual(t, providedPID.Value(), retrieved.Value())
}

func TestRestoreRefreshToken_ClonesUsedAt(t *testing.T) {
	providedUsedAt := shared.Ptr(time.Now())
	token, err := session.RestoreRefreshToken(session.RefreshTokenRestoreParams{
		ID:        session.NewRefreshTokenID(),
		SessionID: session.NewSessionID(),
		Hash:      sessiontest.MustRefreshTokenHash(t, []byte{1}),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now(),
		UsedAt:    providedUsedAt,
	})
	require.NoError(t, err)

	*providedUsedAt = time.Now().Add(time.Hour)
	retrieved, ok := token.UsedAt()
	assert.True(t, ok)
	assert.False(t, providedUsedAt.Equal(retrieved))
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
		})
	}
}

func TestRefreshToken_IsUsed(t *testing.T) {
	testCases := []struct {
		desc   string
		token  *session.RefreshToken
		isUsed bool
	}{
		{
			desc: "token already used",
			token: sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
				p.UsedAt = shared.Ptr(time.Now())
			}),
			isUsed: true,
		},
		{
			desc: "token not used",
			token: sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
				p.UsedAt = nil
			}),
			isUsed: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.token.IsUsed(), tC.isUsed)
		})
	}
}

func TestRefreshToken_HasParent(t *testing.T) {
	testCases := []struct {
		desc      string
		token     *session.RefreshToken
		hasParent bool
	}{
		{
			desc: "token has parent",
			token: sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
				p.ParentID = shared.Ptr(session.NewRefreshTokenID())
			}),
			hasParent: true,
		},
		{
			desc: "token has not parent",
			token: sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
				p.ParentID = nil
			}),
			hasParent: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.token.HasParent(), tC.hasParent)
		})
	}
}
