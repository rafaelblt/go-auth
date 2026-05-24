package domain

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestNewRefreshToken(t *testing.T) {
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
				Hash:      "ao0kv-0opal",
				ParentID:  shared.Ptr(NewRefreshTokenID()),
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now(),
				UsedAt:    shared.Ptr(time.Now()),
			},
			expectErr: true,
		},
		{
			desc: "token hash empty",
			params: RefreshTokenCreationParams{
				SessionID: NewSessionID(),
				Hash:      "",
				ParentID:  nil,
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now(),
				UsedAt:    nil,
			},
			expectErr: true,
		},
		{
			desc: "parent id zero",
			params: RefreshTokenCreationParams{
				SessionID: NewSessionID(),
				Hash:      "ao0kv-0opal",
				ParentID:  shared.Ptr(RefreshTokenID{}),
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now(),
				UsedAt:    shared.Ptr(time.Now()),
			},
			expectErr: true,
		},
		{
			desc: "valid case",
			params: RefreshTokenCreationParams{
				SessionID: NewSessionID(),
				Hash:      "ao0kv-0opal",
				ParentID:  shared.Ptr(NewRefreshTokenID()),
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now(),
				UsedAt:    shared.Ptr(time.Now()),
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
			assert.Equal(t, tC.params.SessionID, token.SessionID())
			assert.Equal(t, tC.params.Hash, token.Hash())
			assert.Equal(t, tC.params.ParentID, token.ParentID())
			assert.Equal(t, tC.params.IssuedAt, token.IssuedAt())
			assert.Equal(t, tC.params.ExpiresAt, token.ExpiresAt())
			assert.Equal(t, tC.params.UsedAt, token.UsedAt())
		})
	}
}
