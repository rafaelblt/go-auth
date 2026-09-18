package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const refreshTokenSecretSize = 32

// RefreshTokenSecret is the credential handed to the client. Only its hash is
// ever stored.
type RefreshTokenSecret struct {
	value []byte
}

func newRefreshTokenSecret() (RefreshTokenSecret, error) {
	value := make([]byte, refreshTokenSecretSize)
	// Unreachable since Go 1.24: crypto/rand.Read never returns an error and
	// crashes the program instead. Kept so a failure is never swallowed.
	_, err := rand.Read(value)
	if err != nil {
		return RefreshTokenSecret{}, fmt.Errorf("rand read failed: %w", err)
	}
	return RefreshTokenSecret{value: value}, nil
}

func ParseRefreshTokenSecret(raw string) (RefreshTokenSecret, error) {
	value, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return RefreshTokenSecret{}, fmt.Errorf("base64 url decode failed: %w", err)
	}
	if len(value) != refreshTokenSecretSize {
		e := fmt.Errorf("value length %d, want %d", len(value), refreshTokenSecretSize)
		return RefreshTokenSecret{}, e
	}
	return RefreshTokenSecret{value: value}, nil
}

func (s RefreshTokenSecret) Value() string {
	return base64.RawURLEncoding.EncodeToString(s.value)
}

func (s RefreshTokenSecret) Hash() RefreshTokenHash {
	if s.IsZero() {
		return RefreshTokenHash{}
	}
	sum := sha256.Sum256(s.value)
	return RefreshTokenHash{value: sum[:]}
}

func (s RefreshTokenSecret) IsZero() bool {
	return s.value == nil
}
