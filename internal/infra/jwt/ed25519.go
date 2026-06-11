package jwt

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rafaelblt/go-auth/internal/port"
)

type Ed25519Signer struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	clock      port.Clock
}

type Ed25519Config struct {
	PrivateKey ed25519.PrivateKey
	Clock      port.Clock
}

type claims struct {
	Subject   string
	ExpiresAt time.Time
}

var (
	errTokenExpired = errors.New("token expired")
	errTokenInvalid = errors.New("token invalid")
)

func NewEd25519(cfg Ed25519Config) (*Ed25519Signer, error) {
	if len(cfg.PrivateKey) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid ed25519 private key size")
	}
	if cfg.Clock == nil {
		return nil, errors.New("clock nil")
	}
	publicKey := cfg.PrivateKey.Public().(ed25519.PublicKey)
	signer := Ed25519Signer{
		privateKey: cfg.PrivateKey,
		publicKey:  publicKey,
		clock:      cfg.Clock,
	}
	return &signer, nil
}

func (s *Ed25519Signer) sign(claims claims) (string, error) {
	if err := s.validateClaims(claims); err != nil {
		return "", err
	}

	registered := jwt.RegisteredClaims{
		Subject:   claims.Subject,
		ExpiresAt: jwt.NewNumericDate(claims.ExpiresAt),
	}

	t, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, registered).SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("jwt sign failed: %w", err)
	}

	return t, nil
}

func (s *Ed25519Signer) parse(raw string) (claims, error) {
	token, err := jwt.ParseWithClaims(raw, &jwt.RegisteredClaims{}, s.keyfunc,
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(s.clock.Now),
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
	)

	switch {
	case err == nil:
		// continue
	case errors.Is(err, jwt.ErrTokenExpired):
		return claims{}, errTokenExpired
	case errors.Is(err, jwt.ErrTokenMalformed),
		errors.Is(err, jwt.ErrTokenSignatureInvalid),
		errors.Is(err, jwt.ErrTokenNotValidYet),
		errors.Is(err, jwt.ErrTokenInvalidClaims):
		return claims{}, errTokenInvalid
	default:
		return claims{}, fmt.Errorf("jwt parse failed: %w", err)
	}

	registered := token.Claims.(*jwt.RegisteredClaims)
	claims := claims{
		Subject:   registered.Subject,
		ExpiresAt: registered.ExpiresAt.Time,
	}

	return claims, nil
}

func (s *Ed25519Signer) keyfunc(token *jwt.Token) (any, error) {
	return s.publicKey, nil
}

func (s *Ed25519Signer) validateClaims(claims claims) error {
	if claims.ExpiresAt.IsZero() {
		return errors.New("claim ExpiresAt zero")
	}

	if claims.Subject == "" {
		return errors.New("claim Subject empty")
	}

	return nil
}
