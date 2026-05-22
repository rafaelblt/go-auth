package infra

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/usecase"
)

const (
	minSecretBytes = 32
	minExpiration = 1 * time.Minute
	maxExpiration = 24 * time.Hour
	maxIssuerLength = 256
)

type AccessTokenService struct {
	secret     []byte
	issuer     string
	method     jwt.SigningMethod
	clock      usecase.Clock
	expiration time.Duration
}

type AccessTokenServiceConfig struct {
	Secret     []byte
	Issuer     string
	Method     jwt.SigningMethod
	Clock      usecase.Clock
	Expiration time.Duration
}

func NewAccessTokenService(cfg AccessTokenServiceConfig) (*AccessTokenService, error) {
	secret := cfg.Secret
	issuer := strings.TrimSpace(cfg.Issuer)
	method := cfg.Method
	clock := cfg.Clock
	expiration := cfg.Expiration

	if len(secret) < minSecretBytes {
		return nil, errors.New("secret must be at least 32 bytes")
	}
	if len(issuer) == 0 {
		return nil, errors.New("issuer must be a non-empty value")
	}
	if len(issuer) > maxIssuerLength {
		return nil, errors.New("issuer must be at most 256 chars.")
	}
	if method == nil {
		return nil, errors.New("signing method cannot be nil")
	}
	if _, ok := method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("signing method must be HMAC-SHA family")
	}
	if clock == nil {
		return nil, errors.New("clock cannot be nil")
	}
	if expiration < minExpiration || expiration > maxExpiration {
		return nil, errors.New("expiration must be between 1m and 24h")
	}

	service := AccessTokenService{
		secret:     secret,
		issuer:     issuer,
		method:     method,
		clock:      clock,
		expiration: expiration,
	}
	return &service, nil
}

func (s *AccessTokenService) Issue(payload usecase.AccessTokenPayload) (usecase.AccessToken, error) {
	if payload.UserID.IsZero() {
		return usecase.AccessToken{}, errors.New("user id cannot be zero")
	}

	now := s.clock.UtcNow()
	exp := now.Add(s.expiration)

	claims := jwt.RegisteredClaims{
		Issuer:    s.issuer,
		Subject:   payload.UserID.Value().String(),
		ExpiresAt: jwt.NewNumericDate(exp),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(s.method, claims)

	raw, err := token.SignedString(s.secret)
	if err != nil {
		return usecase.AccessToken{}, err
	}

	accessToken := usecase.AccessToken{
		Raw:       raw,
		ExpiresAt: exp,
	}

	return accessToken, nil
}

func (e *AccessTokenService) Validate(raw string) (usecase.AccessTokenClaims, error) {
	token, err := jwt.Parse(raw, e.keyfunc)
	if err != nil {
		return usecase.AccessTokenClaims{}, err
	}

	sub, err := token.Claims.GetSubject()
	if err != nil {
		return usecase.AccessTokenClaims{}, err
	}

	userID, err := domain.ParseUserID(sub)
	if err != nil {
		return usecase.AccessTokenClaims{}, err
	}

	result := usecase.AccessTokenClaims{
		UserID: userID,
	}
	return result, nil
}

func (e *AccessTokenService) keyfunc(token *jwt.Token) (any, error) {
	alg := token.Method.Alg()
	if alg != e.method.Alg() {
		return nil, fmt.Errorf("invalid algorithm: %v", alg)
	}
	return e.secret, nil
}
