package credential_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/testutil/credentialtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecret(t *testing.T) {
	testCases := []struct {
		desc    string
		value   string
		wantErr bool
	}{
		{desc: "empty value", value: "", wantErr: true},
		{desc: "valid value", value: "secret", wantErr: false},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			secret, err := credential.NewSecret(tC.value)
			if tC.wantErr {
				assert.Error(t, err)
				assert.Zero(t, secret)
				return
			}
			require.NoError(t, err)
			require.NotZero(t, secret)
			assert.Equal(t, secret.Value(), tC.value)
		})
	}
}

func TestSecret_IsZero(t *testing.T) {
	testCases := []struct {
		desc   string
		secret credential.Secret
		expect bool
	}{
		{
			desc:   "not zero",
			secret: credentialtest.MustSecret(t, "secret"),
			expect: false,
		},
		{
			desc:   "zero",
			secret: credential.Secret{},
			expect: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expect, tC.secret.IsZero())
		})
	}
}
