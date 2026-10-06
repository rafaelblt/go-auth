package ed25519

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"fmt"
)

// encryptionKeySize is the size of an AES-256 key.
const encryptionKeySize = 32

// sealedSeedSize is a seed sealed with AES-256-GCM: a 12-byte nonce, the
// 32-byte seed, then a 16-byte tag.
const sealedSeedSize = ed25519.SeedSize + 12 + 16

// seedCipher seals the seeds of the keys the keyring adds, and opens the
// stored ones. Its zero value has no encryption key.
type seedCipher struct {
	aead cipher.AEAD // nil when no encryption key is set
}

func newSeedCipher(encryptionKey []byte) (seedCipher, error) {
	if len(encryptionKey) == 0 {
		return seedCipher{}, nil
	}
	if len(encryptionKey) != encryptionKeySize {
		return seedCipher{}, fmt.Errorf("encryption key size %d, not %d", len(encryptionKey), encryptionKeySize)
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return seedCipher{}, fmt.Errorf("aes cipher creation failed: %w", err)
	}
	aead, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return seedCipher{}, fmt.Errorf("gcm creation failed: %w", err)
	}
	return seedCipher{aead: aead}, nil
}

func (c seedCipher) seal(generation int64, seed []byte) []byte {
	if c.aead == nil {
		return seed
	}
	return c.aead.Seal(nil, nil, seed, additionalData(generation))
}

// open returns the seed of a stored key. A plaintext seed opens whether or not
// a key is set, so that setting one needs no migration.
//
// See docs/development/decisions/0054-signing-key-seeds-are-sealed-when-a-key-is-set.md.
func (c seedCipher) open(generation int64, stored []byte) ([]byte, error) {
	switch len(stored) {
	case ed25519.SeedSize:
		return stored, nil
	case sealedSeedSize:
		if c.aead == nil {
			return nil, errors.New("seed sealed, but no encryption key set")
		}
		seed, err := c.aead.Open(nil, nil, stored, additionalData(generation))
		if err != nil {
			return nil, fmt.Errorf("seed open failed: %w", err)
		}
		return seed, nil
	default:
		return nil, fmt.Errorf("seed size %d, not %d or %d", len(stored), ed25519.SeedSize, sealedSeedSize)
	}
}

// additionalData binds a sealed seed to its generation, so that it opens only
// in the row it was sealed for.
func additionalData(generation int64) []byte {
	return binary.BigEndian.AppendUint64(nil, uint64(generation))
}
