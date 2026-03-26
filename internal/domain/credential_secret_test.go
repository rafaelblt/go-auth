package domain_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCredentialSecret(t *testing.T) {
	testCases := []struct {
		desc      string
		value     string
		expectErr bool
	}{
		{desc: "empty value", value: "", expectErr: true},
		{desc: "valid value", value: "secret", expectErr: false},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			secret, err := domain.NewCredentialSecret(tC.value)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, secret)
			} else {
				require.NoError(t, err)
				require.NotZero(t, secret)
				assert.Equal(t, secret.Value(), tC.value)
			}
		})
	}
}
