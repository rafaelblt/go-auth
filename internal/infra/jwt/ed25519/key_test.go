package ed25519

import (
	"crypto/ed25519"
	"encoding/hex"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var thumbprintFormat = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

func TestNewKey_MakesTheDTO(t *testing.T) {
	key, err := newKey()

	require.NoError(t, err)
	require.NotEmpty(t, key.public)
	assert.Equal(t, []byte(key.public), key.dto.Key)
	assert.Equal(t, thumbprint(key.public), key.dto.ID)
	assert.Equal(t, keyAlgorithm, key.dto.Algorithm)
	assert.Equal(t, keyCurve, key.dto.Curve)
	assert.Equal(t, keyType, key.dto.Type)
}

func TestThumbprint_RFC8037Vector(t *testing.T) {
	expected := "kPrK_qmxVWaYVA9wwBF6Iuo3vVzz7TxHCTwXBygrS4k"
	raw, err := hex.DecodeString(
		"d75a980182b10ab7d54bfed3c964073a0ee172f3daa62325af021a68f707511a")
	require.NoError(t, err)

	retrieved := thumbprint(ed25519.PublicKey(raw))

	assert.Equal(t, expected, retrieved)
}

func FuzzThumbprint(f *testing.F) {
	f.Add(make([]byte, ed25519.PublicKeySize))
	f.Fuzz(func(t *testing.T, key []byte) {
		retrieved := thumbprint(key)
		assert.Regexp(t, thumbprintFormat, retrieved)
	})
}
