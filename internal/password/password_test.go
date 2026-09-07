package password

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPassword(t *testing.T) {
	hash, err := NewHashed("hash")
	require.NoError(t, err)
	testCases := []struct {
		desc      string
		params    CreationParams
		expectErr bool
	}{
		{
			desc:      "all params zero",
			params:    CreationParams{},
			expectErr: true,
		},
		{
			desc: "user id zero",
			params: CreationParams{
				UserID:    user.ID{},
				Hash:      hash,
				CreatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "hash zero",
			params: CreationParams{
				UserID:    user.NewID(),
				Hash:      Hashed{},
				CreatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "valid case",
			params: CreationParams{
				UserID:    user.NewID(),
				Hash:      hash,
				CreatedAt: time.Now().UTC(),
			},
			expectErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			pwd, err := NewPassword(tC.params)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, pwd)
				return
			}
			require.NoError(t, err)
			require.NotZero(t, pwd)
			assert.NotZero(t, pwd.ID())
			assert.Equal(t, pwd.UserID(), tC.params.UserID)
			assert.Equal(t, pwd.Hash(), tC.params.Hash)
			assert.Equal(t, pwd.CreatedAt(), tC.params.CreatedAt)
			assert.Equal(t, pwd.UpdatedAt(), tC.params.CreatedAt)
		})
	}
}

func TestRestorePassword(t *testing.T) {
	hash, err := NewHashed("hash")
	require.NoError(t, err)
	testCases := []struct {
		desc      string
		params    RestoreParams
		expectErr bool
	}{
		{
			desc:      "all params zero",
			params:    RestoreParams{},
			expectErr: true,
		},
		{
			desc: "id zero",
			params: RestoreParams{
				ID:        ID{},
				UserID:    user.NewID(),
				Hash:      hash,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "user id zero",
			params: RestoreParams{
				ID:        NewID(),
				UserID:    user.ID{},
				Hash:      hash,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "hash zero",
			params: RestoreParams{
				ID:        NewID(),
				UserID:    user.NewID(),
				Hash:      Hashed{},
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "created at zero",
			params: RestoreParams{
				ID:        NewID(),
				UserID:    user.NewID(),
				Hash:      hash,
				CreatedAt: time.Time{},
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "updated at zero",
			params: RestoreParams{
				ID:        NewID(),
				UserID:    user.NewID(),
				Hash:      hash,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Time{},
			},
			expectErr: true,
		},
		{
			desc: "valid case",
			params: RestoreParams{
				ID:        NewID(),
				UserID:    user.NewID(),
				Hash:      hash,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			pwd, err := RestorePassword(tC.params)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, pwd)
			} else {
				require.NoError(t, err)
				require.NotZero(t, pwd)
				assert.Equal(t, pwd.ID(), tC.params.ID)
				assert.Equal(t, pwd.UserID(), tC.params.UserID)
				assert.Equal(t, pwd.Hash(), tC.params.Hash)
				assert.Equal(t, pwd.CreatedAt(), tC.params.CreatedAt)
				assert.Equal(t, pwd.UpdatedAt(), tC.params.UpdatedAt)
			}
		})
	}
}
