package refreshtoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/session"
)

func generateToken() ([]byte, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return nil, fmt.Errorf("rand read failed: %w", err)
	}
	return token, nil
}

func encodeToken(token []byte) string {
	return base64.RawURLEncoding.EncodeToString(token)
}

func decodeToken(token string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("base64 url decode failed: %w", err)
	}
	return decoded, nil
}

func hashToken(token []byte) (session.RefreshTokenHash, error) {
	sum := sha256.Sum256(token)
	obj, err := session.NewRefreshTokenHash(sum[:])
	if err != nil {
		e := fmt.Errorf("refresh token hash creation failed: %w", err)
		return session.RefreshTokenHash{}, e
	}
	return obj, nil
}
