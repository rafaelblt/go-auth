package ed25519

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/port"
)

type key struct {
	id      string
	public  ed25519.PublicKey
	private ed25519.PrivateKey
	dto     port.PublicKey
}

func newKey() (*key, error) {
	pub, prv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("ed25519 generate key failed: %w", err)
	}

	kid := thumbprint(pub)

	dto := port.PublicKey{
		ID:        kid,
		Type:      keyType,
		Algorithm: keyAlgorithm,
		Curve:     keyCurve,
		Key:       pub,
	}

	key := key{
		id:      kid,
		public:  pub,
		private: prv,
		dto:     dto,
	}

	return &key, nil
}

func thumbprint(key ed25519.PublicKey) string {
	canonical := fmt.Sprintf(`{"crv":%q,"kty":%q,"x":%q}`,
		keyCurve,
		keyType,
		base64.RawURLEncoding.EncodeToString(key))
	sum := sha256.Sum256([]byte(canonical))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
