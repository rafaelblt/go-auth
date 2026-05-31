package jwt

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func service() *AccessTokenService {
	return &AccessTokenService{
		secret:     []byte("sa-dfo-0aSLf2q90kasdplg[=as-hgoq3ṕ,1.azd"),
		issuer:     "issuer",
		method:     jwt.SigningMethodHS256,
		clock:      porttest.NewFakeClock(),
		expiration: time.Minute,
	}
}

func TestNewAccessTokenService(t *testing.T) {
	testCases := []struct {
		desc      string
		cfg       AccessTokenServiceConfig
		expectErr bool
	}{
		{
			desc: "config valid",
			cfg: AccessTokenServiceConfig{
				Secret:     []byte("089sua)DFK!W0OLKADF-LPDÇŹXA,d0pfoiajgm9p81uh98sajdf1qw"),
				Issuer:     "issuer",
				Method:     jwt.SigningMethodHS256,
				Clock:      porttest.NewFakeClock(),
				Expiration: 10*time.Minute,
			},
			expectErr: false,
		},
		{
			desc: "secret zero",
			cfg: AccessTokenServiceConfig{
				Secret:     []byte{},
				Issuer:     "issuer",
				Method:     jwt.SigningMethodHS256,
				Clock:      porttest.NewFakeClock(),
				Expiration: time.Minute,
			},
			expectErr: true,
		},
		{
			desc: "secret weak",
			cfg: AccessTokenServiceConfig{
				Secret:     []byte("abc"),
				Issuer:     "issuer",
				Method:     jwt.SigningMethodHS256,
				Clock:      porttest.NewFakeClock(),
				Expiration: time.Minute,
			},
			expectErr: true,
		},
		{
			desc: "issuer empty",
			cfg: AccessTokenServiceConfig{
				Secret:     []byte("secret"),
				Issuer:     "",
				Method:     jwt.SigningMethodHS256,
				Clock:      porttest.NewFakeClock(),
				Expiration: time.Minute,
			},
			expectErr: true,
		},
		{
			desc: "issuer too long",
			cfg: AccessTokenServiceConfig{
				Secret:     []byte("secret"),
				Issuer:     strings.Repeat("a", maxIssuerLength+1),
				Method:     jwt.SigningMethodHS256,
				Clock:      porttest.NewFakeClock(),
				Expiration: time.Minute,
			},
			expectErr: true,
		},
		{
			desc: "signing method nil",
			cfg: AccessTokenServiceConfig{
				Secret:     []byte("secret"),
				Issuer:     "issuer",
				Method:     nil,
				Clock:      porttest.NewFakeClock(),
				Expiration: time.Minute,
			},
			expectErr: true,
		},
		{
			desc: "signing method none",
			cfg: AccessTokenServiceConfig{
				Secret:     []byte("secret"),
				Issuer:     "issuer",
				Method:     jwt.SigningMethodNone,
				Clock:      porttest.NewFakeClock(),
				Expiration: time.Minute,
			},
			expectErr: true,
		},
		{
			desc: "signing method not hmac",
			cfg: AccessTokenServiceConfig{
				Secret:     []byte("secret"),
				Issuer:     "issuer",
				Method:     jwt.SigningMethodPS256,
				Clock:      porttest.NewFakeClock(),
				Expiration: time.Minute,
			},
			expectErr: true,
		},
		{
			desc: "clock nil",
			cfg: AccessTokenServiceConfig{
				Secret:     []byte("secret"),
				Issuer:     "issuer",
				Method:     jwt.SigningMethodHS256,
				Clock:      nil,
				Expiration: time.Minute,
			},
			expectErr: true,
		},
		{
			desc: "expiration too short",
			cfg: AccessTokenServiceConfig{
				Secret:     []byte("secret"),
				Issuer:     "issuer",
				Method:     jwt.SigningMethodHS256,
				Clock:      porttest.NewFakeClock(),
				Expiration: minExpiration - 1,
			},
			expectErr: true,
		},
		{
			desc: "expiration too long",
			cfg: AccessTokenServiceConfig{
				Secret:     []byte("secret"),
				Issuer:     "issuer",
				Method:     jwt.SigningMethodHS256,
				Clock:      porttest.NewFakeClock(),
				Expiration: maxExpiration + 1,
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			service, err := NewAccessTokenService(tC.cfg)
			if tC.expectErr {
				require.Error(t, err)
				require.Zero(t, service)
				return
			}
			require.NoError(t, err)
			require.NotZero(t, service)
			assert.Equal(t, tC.cfg.Secret, service.secret)
			assert.Equal(t, tC.cfg.Issuer, service.issuer)
			assert.Equal(t, tC.cfg.Method, service.method)
			assert.Equal(t, tC.cfg.Clock, service.clock)
			assert.Equal(t, tC.cfg.Expiration, service.expiration)
		})
	}
}

func TestAccessTokenService_Issue_ReturnsValidToken(t *testing.T) {
	service := service()
	payload := port.AccessTokenPayload{
		UserID: user.NewID(),
	}

	issued, err := service.Issue(payload)

	require.NoError(t, err)
	require.NotZero(t, issued)
	assert.NotZero(t, issued.Token)
	assert.NotZero(t, issued.ExpiresAt)
}

func TestAccessTokenService_Issue_ReturnsTokenWithExpiration(t *testing.T) {
	service := service()
	payload := port.AccessTokenPayload{
		UserID: user.NewID(),
	}

	issued, err := service.Issue(payload)

	require.NoError(t, err)
	expectedExp := service.clock.Now().Add(service.expiration)
	assert.Equal(t, expectedExp, issued.ExpiresAt)
}

func TestAccessTokenService_Issue_ReturnsErrorWithZeroUserID(t *testing.T) {
	service := service()
	payload := port.AccessTokenPayload{UserID: user.ID{}}

	issued, err := service.Issue(payload)

	require.Error(t, err)
	require.Zero(t, issued)
}

func TestAccessTokenService_Validate_ReturnsErrorForRandomToken(t *testing.T) {
	service := service()

	claims, err := service.Validate("random")

	assert.Error(t, err)
	assert.Zero(t, claims)
}

func TestAccessTokenService_Validate_ReturnsClaimsForValidToken(t *testing.T) {
	service := service()

	payload := port.AccessTokenPayload{UserID: user.NewID()}
	valid, err := service.Issue(payload)
	require.NoError(t, err)

	claims, err := service.Validate(valid.Token.Value())

	require.NoError(t, err)
	require.NotZero(t, claims)
	assert.Equal(t, payload.UserID, claims.UserID)
}

func TestAccessTokenService_Validate_ReturnsErrorForTokenWithNoClaims(t *testing.T) {
	service := service()
	token := jwt.New(service.method)

	claims, err := service.Validate(token.Raw)

	require.Error(t, err)
	require.Zero(t, claims)
}

func TestAccessTokenService_Validate_ReturnsErrorForTokenWithDifferentSecret(t *testing.T) {
	service := service()

	payload := port.AccessTokenPayload{UserID: user.NewID()}
	issued, err := service.Issue(payload)
	require.NoError(t, err)

	service.secret = []byte("bla-bla-bla-bla-67-3.14")

	claims, err := service.Validate(issued.Token.Value())

	require.Error(t, err)
	require.Zero(t, claims)
}

func TestAccessTokenService_Validate_ReturnsErrorForTokenWithDifferentMethod(t *testing.T) {
	serviceHS256 := service()

	serviceHS512 := service()
	serviceHS512.method = jwt.SigningMethodHS512

	payload := port.AccessTokenPayload{UserID: user.NewID()}
	issued, err := serviceHS512.Issue(payload)
	require.NoError(t, err)

	claims, err := serviceHS256.Validate(issued.Token.Value())

	require.Error(t, err)
	require.Zero(t, claims)
}

func TestAccessTokenService_Validate_ReturnsErrorForTokenWithInvalidUserID(t *testing.T) {
	service := service()

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
