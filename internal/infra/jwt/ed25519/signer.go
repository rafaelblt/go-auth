package ed25519

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rafaelblt/go-auth/internal/port"
)

type Signer struct {
	keyring *Keyring
	clock   port.Clock
}

type SignerConfig struct {
	Keyring *Keyring
	Clock   port.Clock
}

func NewSigner(cfg SignerConfig) (*Signer, error) {
	if cfg.Keyring == nil {
		return nil, errors.New("keyring nil")
	}
	if cfg.Clock == nil {
		return nil, errors.New("clock nil")
	}

	signer := Signer{
		keyring: cfg.Keyring,
		clock:   cfg.Clock,
	}
	return &signer, nil
}

func (s *Signer) Sign(claims jwt.RegisteredClaims) (string, error) {
	if err := s.validateClaims(claims); err != nil {
		return "", err
	}

	key := s.keyring.SigningKey()

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = key.id

	signed, err := token.SignedString(key.private)
	if err != nil {
		return "", fmt.Errorf("jwt sign failed: %w", err)
	}

	return signed, nil
}

func (s *Signer) Parse(raw string) (jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(raw, &jwt.RegisteredClaims{}, s.keyfunc,
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(s.clock.Now),
		jwt.WithValidMethods([]string{keyAlgorithm}),
	)

	switch {
	case err == nil:
		// continue
	case errors.Is(err, jwt.ErrTokenExpired):
		return jwt.RegisteredClaims{}, ErrTokenExpired
	case errors.Is(err, jwt.ErrTokenMalformed),
		errors.Is(err, jwt.ErrTokenUnverifiable),
		errors.Is(err, jwt.ErrTokenSignatureInvalid),
		errors.Is(err, jwt.ErrTokenNotValidYet),
		errors.Is(err, jwt.ErrTokenInvalidClaims),
		errors.Is(err, errKidMissing),
		errors.Is(err, errKidUnknown):
		return jwt.RegisteredClaims{}, ErrTokenInvalid
	default:
		return jwt.RegisteredClaims{}, fmt.Errorf("jwt parse failed: %w", err)
	}

	claims := token.Claims.(*jwt.RegisteredClaims)
	return *claims, nil
}

func (s *Signer) keyfunc(token *jwt.Token) (any, error) {
	kid, _ := token.Header["kid"].(string)
	if kid == "" {
		return nil, errKidMissing
	}

	pub, err := s.keyring.PublicKeyByID(kid)
	if err != nil {
		return nil, err
	}

	return pub, nil
}

func (s *Signer) validateClaims(claims jwt.RegisteredClaims) error {
	if claims.ExpiresAt == nil || claims.ExpiresAt.IsZero() {
		return errors.New("claim ExpiresAt zero")
	}
	return nil
}
