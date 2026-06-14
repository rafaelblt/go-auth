package infra

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/session"
)

type RefreshTokenService struct {
	clock      port.Clock
	expiration time.Duration
}

type payload = port.RefreshTokenPayload
type issued = port.RefreshTokenIssued

func (rts *RefreshTokenService) Issue(payload payload) (issued, error) {
	if payload.SessionID.IsZero() {
		return issued{}, errors.New("session id zero")
	}

	tokenBytes, err := rts.generateToken()
	if err != nil {
		return issued{}, err
	}
	tokenString := rts.encodeToken(tokenBytes)
	tokenHash := rts.hashToken(tokenBytes)

	now := rts.clock.Now()
	expiresAt := now.Add(rts.expiration)

	refreshToken, err := session.NewRefreshToken(session.RefreshTokenCreationParams{
		SessionID: payload.SessionID,
		Hash:      tokenHash,
		ParentID:  nil,
		IssuedAt:  rts.clock.Now(),
		ExpiresAt: expiresAt,
	})
	if err != nil {
		e := fmt.Errorf("refresh token creation failed: %w", err)
		return issued{}, e
	}

	issued := issued{
		Token:    refreshToken,
		RawValue: tokenString,
	}
	return issued, nil
}

func (rts *RefreshTokenService) generateToken() ([]byte, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return nil, fmt.Errorf("rand read failed: %w", err)
	}
	return token, nil
}

func (rts *RefreshTokenService) encodeToken(token []byte) string {
	return base64.RawURLEncoding.EncodeToString(token)
}

func (rts *RefreshTokenService) decodeToken(token string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("base64 url decode failed: %w", err)
	}
	return decoded, nil
}

func (rts *RefreshTokenService) hashToken(token []byte) []byte {
	sum := sha256.Sum256(token)
	return sum[:]
}
