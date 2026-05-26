package session

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/assert"
)

func TestNewSession(t *testing.T) {
	testCases := []struct {
		desc       string
		params     SessionCreationParams
		normalized string
		expectErr  bool
	}{
		{
			desc: "user id zero",
			params: SessionCreationParams{
				UserID:   user.ID{},
				IssuedAt: time.Now(),
			},
			expectErr: true,
		},
		{
			desc: "valid case",
			params: SessionCreationParams{
				UserID:   user.NewID(),
				IssuedAt: time.Now(),
			},
			expectErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			token, err := NewSession(tC.params)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, token)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tC.params.UserID, token.UserID())
			assert.Equal(t, tC.params.IssuedAt, token.IssuedAt())
		})
	}
}
