package ed25519

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
)

type key struct {
	id         string
	generation int64
	activeAt   time.Time
	public     ed25519.PublicKey
	private    ed25519.PrivateKey
	dto        port.PublicKey
}

func restoreKey(stored port.StoredSigningKey) (*key, error) {
	if stored.Generation < 1 {
		return nil, errors.New("generation not positive")
	}
	// ed25519.NewKeyFromSeed panics on any other size.
	if len(stored.Seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("seed size %d, not %d", len(stored.Seed), ed25519.SeedSize)
	}
	if stored.ActiveAt.IsZero() {
		return nil, errors.New("active at zero")
	}

	prv := ed25519.NewKeyFromSeed(stored.Seed)
	pub := prv.Public().(ed25519.PublicKey)

	kid := thumbprint(pub)

	dto := port.PublicKey{
		ID:        kid,
		Type:      keyType,
		Algorithm: keyAlgorithm,
		Curve:     keyCurve,
		Key:       pub,
	}

	key := key{
		id:         kid,
		generation: stored.Generation,
		activeAt:   stored.ActiveAt,
		public:     pub,
		private:    prv,
		dto:        dto,
	}

	return &key, nil
}

func newSeed() ([]byte, error) {
	_, prv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("ed25519 generate key failed: %w", err)
	}
	return prv.Seed(), nil
}

func thumbprint(key ed25519.PublicKey) string {
	canonical := fmt.Sprintf(`{"crv":%q,"kty":%q,"x":%q}`,
		keyCurve,
		keyType,
		base64.RawURLEncoding.EncodeToString(key))
	sum := sha256.Sum256([]byte(canonical))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
