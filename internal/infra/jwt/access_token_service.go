package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/user"
)

const (
	minExpiration = 1 * time.Minute
	maxExpiration = 24 * time.Hour
)

type AccessTokenService struct {
	signer     Signer
	expiration time.Duration
	clock      port.Clock
}

type Signer interface {
	sign(claims) (string, error)
	parse(string) (claims, error)
}

type AccessTokenServiceConfig struct {
	Signer     Signer
	Clock      port.Clock
	Expiration time.Duration
}

func NewAccessTokenService(cfg AccessTokenServiceConfig) (*AccessTokenService, error) {
	if cfg.Signer == nil {
		return nil, errors.New("signer nil")
	}
	if cfg.Clock == nil {
		return nil, errors.New("clock nil")
	}
	if cfg.Expiration < minExpiration || cfg.Expiration > maxExpiration {
		return nil, errors.New("expiration must be between 1m and 24h")
	}

	service := AccessTokenService{
		signer:     cfg.Signer,
		clock:      cfg.Clock,
		expiration: cfg.Expiration,
	}
	return &service, nil
}

func (s *AccessTokenService) Issue(payload port.AccessTokenPayload) (port.AccessTokenIssued, error) {
	if payload.UserID.IsZero() {
		return port.AccessTokenIssued{}, errors.New("user id zero")
	}

	expiresAt := s.clock.Now().Add(s.expiration)

	token, err := s.signer.sign(claims{
		Subject:   payload.UserID.Value().String(),
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return port.AccessTokenIssued{}, fmt.Errorf("signer sign failed: %w", err)
	}

	accessToken, err := session.NewAccessToken(token)
	if err != nil {
		return port.AccessTokenIssued{}, err
	}

	issued := port.AccessTokenIssued{
		Token:     accessToken,
		ExpiresAt: expiresAt,
	}
	return issued, nil
}

func (s *AccessTokenService) Validate(raw string) (port.AccessTokenClaims, error) {
	claims, err := s.signer.parse(raw)
	switch {
	case err == nil:
		// continue
	case errors.Is(err, errTokenInvalid):
		return port.AccessTokenClaims{}, session.ErrTokenInvalid
	case errors.Is(err, errTokenExpired):
		return port.AccessTokenClaims{}, session.ErrTokenExpired
	}

	userID, err := user.ParseID(claims.Subject)
	if err != nil {
		return port.AccessTokenClaims{}, fmt.Errorf("user id parse failed: %w", err)
	}

	result := port.AccessTokenClaims{
		UserID: userID,
	}
	return result, nil
}
