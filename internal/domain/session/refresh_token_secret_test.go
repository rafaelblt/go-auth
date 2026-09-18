package session

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRefreshTokenSecret(t *testing.T) {
	secret, err := newRefreshTokenSecret()

	require.NoError(t, err)
	assert.Len(t, secret.value, refreshTokenSecretSize)
}

func TestParseRefreshTokenSecret(t *testing.T) {
	valid := bytes.Repeat([]byte{1}, refreshTokenSecretSize)
	testCases := []struct {
		desc      string
		input     string
		expectErr bool
	}{
		{
			desc:      "valid case",
			input:     base64.RawURLEncoding.EncodeToString(valid),
			expectErr: false,
		},
		{
			desc:      "empty",
			input:     "",
			expectErr: true,
		},
		{
			desc:      "not base64 url",
			input:     "!!!!",
			expectErr: true,
		},
		{
			desc:      "padded base64 url",
			input:     base64.URLEncoding.EncodeToString(valid),
			expectErr: true,
		},
		{
			desc:      "shorter than secret size",
			input:     base64.RawURLEncoding.EncodeToString(valid[1:]),
			expectErr: true,
		},
		{
			desc:      "longer than secret size",
			input:     base64.RawURLEncoding.EncodeToString(append(valid, 1)),
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			secret, err := ParseRefreshTokenSecret(tC.input)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, secret)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, valid, secret.value)
		})
	}
}

func TestRefreshTokenSecret_Value_RoundTripsThroughParse(t *testing.T) {
	secret, err := newRefreshTokenSecret()
	require.NoError(t, err)

	parsed, err := ParseRefreshTokenSecret(secret.Value())

	require.NoError(t, err)
	assert.Equal(t, secret.value, parsed.value)
}

func TestRefreshTokenSecret_Hash_IsSHA256OfValue(t *testing.T) {
	secret, err := newRefreshTokenSecret()
	require.NoError(t, err)

	sum := sha256.Sum256(secret.value)

	assert.Equal(t, sum[:], secret.Hash().Value())
}

func TestRefreshTokenSecret_Hash_ZeroWhenSecretZero(t *testing.T) {
	assert.True(t, RefreshTokenSecret{}.Hash().IsZero())
}
