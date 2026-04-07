package usecase

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapUserToDTO(t *testing.T) {
	testCases := []struct {
		desc      string
		user      *domain.User
		expectErr bool
	}{
		{
			desc:      "default user",
			user:      testutil.DefaultUser(t),
			expectErr: false,
		},
		{
			desc:      "user zero",
			user:      &domain.User{},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			dto, err := MapUserToDTO(tC.user)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Nil(t, dto)
			} else {
				require.NoError(t, err)
				require.NotNil(t, dto)
				assert.Equal(t, tC.user.ID().Value().String(), dto.ID())
				assert.Equal(t, tC.user.Username().String(), dto.Username())
				assert.Equal(t, tC.user.Status().String(), dto.Status())
				assert.Equal(t, tC.user.CreatedAt(), dto.CreatedAt())
				assert.Equal(t, tC.user.UpdatedAt(), dto.UpdatedAt())
			}
		})
	}
}
