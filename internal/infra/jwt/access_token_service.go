package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/port"
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
	Sign(jwt.RegisteredClaims) (string, error)
	Parse(string) (jwt.RegisteredClaims, error)
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

	claims := jwt.RegisteredClaims{
		Subject:   payload.UserID.Value().String(),
		ExpiresAt: jwt.NewNumericDate(s.clock.Now().Add(s.expiration)),
	}

	token, err := s.signer.Sign(claims)
	if err != nil {
		return port.AccessTokenIssued{}, fmt.Errorf("signer sign failed: %w", err)
	}

	accessToken, err := session.NewAccessToken(token)
	if err != nil {
		return port.AccessTokenIssued{}, err
	}

	issued := port.AccessTokenIssued{
		Token:     accessToken,
		ExpiresAt: claims.ExpiresAt.Time,
	}
	return issued, nil
}

func (s *AccessTokenService) Validate(raw string) (port.AccessTokenClaims, error) {
	claims, err := s.signer.Parse(raw)
	switch {
	case err == nil:
		// continue
	case errors.Is(err, ErrTokenInvalid):
		return port.AccessTokenClaims{}, session.ErrTokenInvalid
	case errors.Is(err, ErrTokenExpired):
		return port.AccessTokenClaims{}, session.ErrTokenExpired
	default:
		e := fmt.Errorf("unexpected error from signer parse: %w", err)
		return port.AccessTokenClaims{}, e
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
