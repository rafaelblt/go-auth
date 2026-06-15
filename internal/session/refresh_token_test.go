package session

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRefreshToken_ValidateParams(t *testing.T) {
	testCases := []struct {
		desc       string
		params     RefreshTokenCreationParams
		normalized string
		expectErr  bool
	}{
		{
			desc: "session id zero",
			params: RefreshTokenCreationParams{
				SessionID: SessionID{},
				Hash:      RefreshTokenHash{value: "hash"},
				ParentID:  shared.Ptr(NewRefreshTokenID()),
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now(),
			},
			expectErr: true,
		},
		{
			desc: "token hash zero",
			params: RefreshTokenCreationParams{
				SessionID: NewSessionID(),
				Hash:      RefreshTokenHash{},
				ParentID:  nil,
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now(),
			},
			expectErr: true,
		},
		{
			desc: "parent id zero",
			params: RefreshTokenCreationParams{
				SessionID: NewSessionID(),
				Hash:      RefreshTokenHash{value: "hash"},
				ParentID:  shared.Ptr(RefreshTokenID{}),
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now(),
			},
			expectErr: true,
		},
		{
			desc: "issued at after expires at",
			params: RefreshTokenCreationParams{
				SessionID: NewSessionID(),
				Hash:      RefreshTokenHash{value: "hash"},
				ParentID:  shared.Ptr(NewRefreshTokenID()),
				IssuedAt:  time.Now().Add(time.Minute),
				ExpiresAt: time.Now(),
			},
			expectErr: true,
		},
		{
			desc: "valid case",
			params: RefreshTokenCreationParams{
				SessionID: NewSessionID(),
				Hash:      RefreshTokenHash{value: "hash"},
				ParentID:  shared.Ptr(NewRefreshTokenID()),
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now(),
			},
			expectErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			token, err := NewRefreshToken(tC.params)
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

func TestNewRefreshToken_CopyParentID(t *testing.T) {
	providedPID := shared.Ptr(NewRefreshTokenID())
	token, err := NewRefreshToken(RefreshTokenCreationParams{
		SessionID: NewSessionID(),
		Hash:      RefreshTokenHash{value: "hash"},
		ParentID:  providedPID,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now(),
	})
	require.NoError(t, err)
	
	pid, ok := token.ParentID()
	require.True(t, ok)
	assert.Equal(t, *providedPID, pid)

	providedPID = nil
	pid, ok = token.ParentID()
	assert.True(t, ok)
	assert.NotEqual(t, providedPID, pid)
}

// TODO: RestoreRefreshToken()
// TODO: RefreshToken.Use()
// TODO: RefreshToken.HasParent()
