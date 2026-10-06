package ed25519

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPERS

func encryptionKeyForTest(t *testing.T) []byte {
	t.Helper()

	key := make([]byte, encryptionKeySize)
	_, err := rand.Read(key)
	require.NoError(t, err)

	return key
}

func seedCipherForTest(t *testing.T, encryptionKey []byte) seedCipher {
	t.Helper()

	seeds, err := newSeedCipher(encryptionKey)
	require.NoError(t, err)

	return seeds
}

func seedForTest(t *testing.T) []byte {
	t.Helper()

	seed, err := newSeed()
	require.NoError(t, err)

	return seed
}

// TESTS

func TestNewSeedCipher_ReturnsError_WhenTheKeyIsNot32Bytes(t *testing.T) {
	testCases := []struct {
		desc string
		size int
	}{
		{"16 bytes", 16},
		{"24 bytes", 24},
		{"31 bytes", 31},
		{"33 bytes", 33},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var err error
			require.NotPanics(t, func() { _, err = newSeedCipher(make([]byte, tC.size)) })

			assert.Error(t, err)
		})
	}
}

func TestSeedCipher_KeepsTheSeedAsItIs_WhenNoKeyIsSet(t *testing.T) {
	seeds := seedCipherForTest(t, nil)
	seed := seedForTest(t)

	sealed := seeds.seal(1, seed)
	opened, err := seeds.open(1, seed)

	assert.Equal(t, seed, sealed)
	require.NoError(t, err)
	assert.Equal(t, seed, opened)
}

func TestSeedCipher_SealsTheSeed_AndOpensIt_WhenAKeyIsSet(t *testing.T) {
	seeds := seedCipherForTest(t, encryptionKeyForTest(t))
	seed := seedForTest(t)

	sealed := seeds.seal(1, seed)

	assert.Len(t, sealed, sealedSeedSize)
	assert.False(t, bytes.Contains(sealed, seed), "sealed seed holds the seed")
	opened, err := seeds.open(1, sealed)
	require.NoError(t, err)
	assert.Equal(t, seed, opened)
}

func TestSeedCipher_Seal_UsesAFreshNonceEachTime(t *testing.T) {
	seeds := seedCipherForTest(t, encryptionKeyForTest(t))
	seed := seedForTest(t)

	first := seeds.seal(1, seed)
	second := seeds.seal(1, seed)

	assert.NotEqual(t, first, second)
}

func TestSeedCipher_Open_ReturnsAPlaintextSeed_WhenAKeyIsSet(t *testing.T) {
	seeds := seedCipherForTest(t, encryptionKeyForTest(t))
	seed := seedForTest(t)

	opened, err := seeds.open(1, seed)

	require.NoError(t, err)
	assert.Equal(t, seed, opened)
}

func TestSeedCipher_Open_ReturnsError(t *testing.T) {
	seeds := seedCipherForTest(t, encryptionKeyForTest(t))
	sealed := seeds.seal(1, seedForTest(t))
	flipped := bytes.Clone(sealed)
	flipped[len(flipped)/2] ^= 1

	testCases := []struct {
		desc       string
		seeds      seedCipher
		generation int64
		stored     []byte
	}{
		{"sealed, opened with no key", seedCipher{}, 1, sealed},
		{"sealed under another key", seedCipherForTest(t, encryptionKeyForTest(t)), 1, sealed},
		{"sealed for generation 1, opened as generation 2", seeds, 2, sealed},
		{"one sealed byte flipped", seeds, 1, flipped},
		{"31 bytes", seeds, 1, make([]byte, 31)},
		{"61 bytes", seeds, 1, make([]byte, 61)},
		{"nil", seeds, 1, nil},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var seed []byte
			var err error
			require.NotPanics(t, func() { seed, err = tC.seeds.open(tC.generation, tC.stored) })

			assert.Nil(t, seed)
			assert.Error(t, err)
		})
	}
}
