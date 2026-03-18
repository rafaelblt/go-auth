package infra

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapUserDomainToModel(t *testing.T) {
	testCases := []struct {
		desc      string
		user      *domain.User
		expectErr bool
	}{
		{
			desc:      "default user",
			user:      testutil.DefaultUser(t).Entity,
			expectErr: false,
		},
		{
			desc:      "user nil",
			user:      nil,
			expectErr: true,
		},
		{
			desc:      "user zero",
			user:      &domain.User{},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			model, err := MapUserDomainToModel(tC.user)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, model)
			} else {
				require.NoError(t, err)
				idBytes, err := tC.user.ID().Value().MarshalBinary()
				require.NoError(t, err)
				assert.Equal(t, idBytes, model.ID)
				assert.Equal(t, tC.user.Username().String(), model.Username)
				assert.Equal(t, tC.user.Status().String(), model.Status)
				assert.Equal(t, tC.user.CreatedAt(), model.CreatedAt)
				assert.Equal(t, tC.user.UpdatedAt(), model.UpdatedAt)
			}
		})
	}
}
