package infra

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClock struct{ now time.Time }

func (c fakeClock) UtcNow() time.Time { return c.now }

func accessTokenService() *AccessTokenService {
	return &AccessTokenService{
		secret:     []byte("sa-dfo-0aSLfṕ,1.azd"),
		issuer:     "issuer",
		method:     jwt.SigningMethodHS256,
		clock:      fakeClock{time.Now().UTC()},
		expiration: time.Minute,
	}
}

func TestAccessTokenService_Issue_ReturnsValidToken(t *testing.T) {
	service := accessTokenService()
	payload := usecase.AccessTokenPayload{
		UserID: domain.NewUserID(),
	}

	token, err := service.Issue(payload)

	require.NoError(t, err)
	require.NotZero(t, token)
	assert.NotZero(t, token.Raw)
	assert.NotZero(t, token.ExpiresAt)
}

func TestAccessTokenService_Issue_ReturnsTokenWithExpiration(t *testing.T) {
	service := accessTokenService()
	payload := usecase.AccessTokenPayload{
		UserID: domain.NewUserID(),
	}

	token, err := service.Issue(payload)

	require.NoError(t, err)
	expectedExp := service.clock.UtcNow().Add(service.expiration)
	assert.Equal(t, expectedExp, token.ExpiresAt)
}

func TestAccessTokenService_Issue_ReturnsErrorWithZeroUserID(t *testing.T) {
	service := accessTokenService()
	payload := usecase.AccessTokenPayload{UserID: domain.UserID{}}

	token, err := service.Issue(payload)

	require.Error(t, err)
	require.Zero(t, token)
}

func TestAccessTokenService_Validate_ReturnsErrorForRandomToken(t *testing.T) {
	service := accessTokenService()

	claims, err := service.Validate("random")

	assert.Error(t, err)
	assert.Zero(t, claims)
}

func TestAccessTokenService_Validate_ReturnsClaimsForValidToken(t *testing.T) {
	service := accessTokenService()

	payload := usecase.AccessTokenPayload{UserID: domain.NewUserID()}
	validToken, err := service.Issue(payload)
	require.NoError(t, err)

	claims, err := service.Validate(validToken.Raw)

	require.NoError(t, err)
	require.NotZero(t, claims)
	assert.Equal(t, payload.UserID, claims.UserID)
}

func TestAccessTokenService_Validate_ReturnsErrorForTokenWithNoClaims(t *testing.T) {
	service := accessTokenService()
	token := jwt.New(service.method)

	claims, err := service.Validate(token.Raw)

	require.Error(t, err)
	require.Zero(t, claims)
}

func TestAccessTokenService_Validate_ReturnsErrorForTokenWithDifferentSecret(t *testing.T) {
	service := accessTokenService()
	service.secret = []byte("bla-bla-bla-bla-67-3.14")

	payload := usecase.AccessTokenPayload{UserID: domain.NewUserID()}
	token, err := service.Issue(payload)
	require.NoError(t, err)

	claims, err := service.Validate(token.Raw)

	require.Error(t, err)
	require.Zero(t, claims)
}

func TestAccessTokenService_Validate_ReturnsErrorForTokenWithDifferentMethod(t *testing.T) {
	serviceHS256 := accessTokenService()

	serviceHS512 := accessTokenService()
	serviceHS512.method = jwt.SigningMethodHS512

	payload := usecase.AccessTokenPayload{UserID: domain.NewUserID()}
	token, err := serviceHS512.Issue(payload)
	require.NoError(t, err)

	claims, err := serviceHS256.Validate(token.Raw)

	require.Error(t, err)
	require.Zero(t, claims)
}

func TestAccessTokenService_Validate_ReturnsErrorForTokenWithInvalidUserID(t *testing.T) {
	service := accessTokenService()

	token := jwt.NewWithClaims(service.method, jwt.RegisteredClaims{
		Issuer:    service.issuer,
		Subject:   "invalid user id",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	claims, err := service.Validate(token.Raw)

	require.Error(t, err)
	require.Zero(t, claims)
}
