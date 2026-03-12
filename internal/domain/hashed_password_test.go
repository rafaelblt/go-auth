package domain_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHashedPassword(t *testing.T) {
	testCases := []struct {
		desc        string
		hash        string
		isValid     bool
		expectedErr error
	}{
		{
			desc:    "valid hash",
			hash:    "masd9of8j)!@Pkjfsda[fl[!@+_)ief[sdapfç]asd=",
			isValid: true,
		},
		{
			desc:        "empty hash value",
			hash:        "",
			isValid:     false,
			expectedErr: domain.ErrHashedPasswordEmpty,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			hashedpwd, err := domain.NewHashedPassword(tC.hash)
			if tC.isValid {
				require.NoError(t, err)
				assert.Equal(t, tC.hash, hashedpwd.Value())
			} else {
				require.Error(t, err)
				assert.Zero(t, hashedpwd)
				assert.ErrorIs(t, err, tC.expectedErr)
			}
		})
	}
}

func TestHashedPasswordIsZero_ShouldReturnTrue_WhenIsZero(t *testing.T) {
	zeroHash := domain.HashedPassword{}
	assert.True(t, zeroHash.IsZero())
}

func TestHashedPasswordIsZero_ShouldReturnFalse_WhenIsNotZero(t *testing.T) {
	hash, _ := domain.NewHashedPassword("hash")
	assert.False(t, hash.IsZero())
}
