package domain_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCredentialProvider(t *testing.T) {
	testCases := []struct {
		desc      string
		value     string
		expectErr bool
	}{
		{desc: "empty value", value: "", expectErr: true},
		{desc: "reserved provider", value: "local", expectErr: true},
		{desc: "valid value", value: "github", expectErr: false},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			provider, err := domain.NewCredentialProvider(tC.value)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, provider)
			} else {
				require.NoError(t, err)
				require.NotZero(t, provider)
				assert.Equal(t, provider.String(), tC.value)
			}
		})
	}
}