package session

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRefreshTokenHash(t *testing.T) {
	testCases := []struct {
		desc      string
		input     []byte
		expectErr bool
	}{
		{
			desc:      "valid case",
			input:     bytes.Repeat([]byte{1}, sha256.Size),
			expectErr: false,
		},
		{
			desc:      "empty",
			input:     []byte{},
			expectErr: true,
		},
		{
			desc:      "shorter than sha256 size",
			input:     bytes.Repeat([]byte{1}, sha256.Size-1),
			expectErr: true,
		},
		{
			desc:      "longer than sha256 size",
			input:     bytes.Repeat([]byte{1}, sha256.Size+1),
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			hash, err := NewRefreshTokenHash(tC.input)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, hash)
				return
			}
			assert.NoError(t, err)
			assert.NotZero(t, hash)
		})
	}
}

func TestNewRefreshTokenHash_CloneBytes(t *testing.T) {
	v := bytes.Repeat([]byte{1}, sha256.Size)

	hash, err := NewRefreshTokenHash(v)
	require.NoError(t, err)

	v[0] = 4
	assert.False(t, bytes.Equal(v, hash.value), "byte slice is not cloned")
}

func TestRefreshTokenHash_Value_ReturnsClone(t *testing.T) {
	hash, err := NewRefreshTokenHash(bytes.Repeat([]byte{1}, sha256.Size))
	require.NoError(t, err)

	value := hash.Value()
	value[0] = 9

	assert.Equal(t, bytes.Repeat([]byte{1}, sha256.Size), hash.value, "byte slice is not cloned")
}
