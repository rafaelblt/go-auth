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
				Hash:      []byte("ao0kv-0opal"),
				ParentID:  shared.Ptr(NewRefreshTokenID()),
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now(),
			},
			expectErr: true,
		},
		{
			desc: "token hash nil",
			params: RefreshTokenCreationParams{
				SessionID: NewSessionID(),
				Hash:      nil,
				ParentID:  nil,
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now(),
			},
			expectErr: true,
		},
		{
			desc: "token hash empty",
			params: RefreshTokenCreationParams{
				SessionID: NewSessionID(),
				Hash:      []byte{},
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
				Hash:      []byte("ao0kv-0opal"),
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
				Hash:      []byte("ao0kv-0opal"),
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
				Hash:      []byte("ao0kv-0opal"),
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

func TestNewRefreshToken_CopyHash(t *testing.T) {
	providedHash := []byte("asdkofgklvaas")
	token, err := NewRefreshToken(RefreshTokenCreationParams{
		SessionID: NewSessionID(),
		Hash:      providedHash,
		ParentID:  shared.Ptr(NewRefreshTokenID()),
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now(),
	})
	require.NoError(t, err)
	
	assert.Equal(t, providedHash, token.Hash())
	providedHash[0] = 0xFF
	assert.NotEqual(t, providedHash, token.Hash())
}

func TestNewRefreshToken_CopyParentID(t *testing.T) {
	providedPID := shared.Ptr(NewRefreshTokenID())
	token, err := NewRefreshToken(RefreshTokenCreationParams{
		SessionID: NewSessionID(),
		Hash:      []byte("aspdokf0opas"),
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

// TODO: RefreshToken.Use()
// TODO: RefreshToken.HasParent()
