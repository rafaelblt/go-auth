package ed25519

import (
	"crypto/ed25519"
	"encoding/hex"
	"regexp"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var thumbprintFormat = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

func TestRestoreKey_MakesTheDTO(t *testing.T) {
	key, err := restoreKey(storedKeyForTest(t, 1, time.Now().UTC()))

	require.NoError(t, err)
	require.NotEmpty(t, key.public)
	assert.Equal(t, []byte(key.public), key.dto.Key)
	assert.Equal(t, thumbprint(key.public), key.dto.ID)
	assert.Equal(t, keyAlgorithm, key.dto.Algorithm)
	assert.Equal(t, keyCurve, key.dto.Curve)
	assert.Equal(t, keyType, key.dto.Type)
}

func TestRestoreKey_DerivesTheKeyFromTheSeed(t *testing.T) {
	stored := storedKeyForTest(t, 3, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))

	key, err := restoreKey(stored)

	require.NoError(t, err)
	private := ed25519.NewKeyFromSeed(stored.Seed)
	assert.Equal(t, private, key.private)
	assert.Equal(t, private.Public(), key.public)
	assert.Equal(t, thumbprint(key.public), key.id)
	assert.Equal(t, stored.Generation, key.generation)
	assert.Equal(t, stored.ActiveAt, key.activeAt)
}

func TestRestoreKey_ReturnsError(t *testing.T) {
	valid := storedKeyForTest(t, 1, time.Now().UTC())
	with := func(override func(s *port.StoredSigningKey)) port.StoredSigningKey {
		s := valid
		override(&s)
		return s
	}

	testCases := []struct {
		desc   string
		stored port.StoredSigningKey
	}{
		{"generation zero", with(func(s *port.StoredSigningKey) { s.Generation = 0 })},
		{"generation negative", with(func(s *port.StoredSigningKey) { s.Generation = -1 })},
		{"seed nil", with(func(s *port.StoredSigningKey) { s.Seed = nil })},
		{"seed of 31 bytes", with(func(s *port.StoredSigningKey) { s.Seed = s.Seed[:31] })},
		{"seed of 64 bytes", with(func(s *port.StoredSigningKey) { s.Seed = append(s.Seed, s.Seed...) })},
		{"active at zero", with(func(s *port.StoredSigningKey) { s.ActiveAt = time.Time{} })},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var key *key
			var err error
			require.NotPanics(t, func() { key, err = restoreKey(tC.stored) })

			assert.Nil(t, key)
			assert.Error(t, err)
		})
	}
}

func TestNewSeed_ReturnsADifferentSeedOfSeedSizeEachTime(t *testing.T) {
	first, err := newSeed()
	require.NoError(t, err)
	second, err := newSeed()
	require.NoError(t, err)

	assert.Len(t, first, ed25519.SeedSize)
	assert.Len(t, second, ed25519.SeedSize)
	assert.NotEqual(t, first, second)
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
