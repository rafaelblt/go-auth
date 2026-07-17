package credential

import (
	"testing"

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
			secret, err := NewSecret(tC.value)
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
		secret Secret
		expect bool
	}{
		{desc: "not zero", secret: Secret{"value"}, expect: false},
		{desc: "zero", secret: Secret{}, expect: true},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expect, tC.secret.IsZero())
		})
	}
}
