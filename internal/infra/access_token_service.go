package infra

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/usecase"
)

type AccessTokenService struct {
	secret     []byte
	issuer     string
	method     jwt.SigningMethod
	clock      usecase.Clock
	expiration time.Duration
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
