package jwt

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func service(t *testing.T) *AccessTokenService {
	return &AccessTokenService{
		signer:     edSigner(t),
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
				Signer:     edSigner(t),
				Clock:      porttest.NewFakeClock(),
				Expiration: 10 * time.Minute,
			},
			expectErr: false,
		},
		{
			desc: "signer nil",
			cfg: AccessTokenServiceConfig{
				Signer:     nil,
				Clock:      porttest.NewFakeClock(),
				Expiration: time.Minute,
			},
			expectErr: true,
		},
		{
			desc: "clock nil",
			cfg: AccessTokenServiceConfig{
				Signer:     edSigner(t),
				Clock:      nil,
				Expiration: time.Minute,
			},
			expectErr: true,
		},
		{
			desc: "expiration too short",
			cfg: AccessTokenServiceConfig{
				Signer:     edSigner(t),
				Clock:      porttest.NewFakeClock(),
				Expiration: minExpiration - 1,
			},
			expectErr: true,
		},
		{
			desc: "expiration too long",
			cfg: AccessTokenServiceConfig{
				Signer:     edSigner(t),
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
			assert.Equal(t, tC.cfg.Signer, service.signer)
			assert.Equal(t, tC.cfg.Clock, service.clock)
			assert.Equal(t, tC.cfg.Expiration, service.expiration)
		})
	}
}

func TestAccessTokenService_Issue_ReturnsToken(t *testing.T) {
	service := service(t)
	payload := port.AccessTokenPayload{
		UserID: user.NewID(),
	}

	issued, err := service.Issue(payload)

	require.NoError(t, err)
	require.NotZero(t, issued)
	assert.NotZero(t, issued.Token)
	assert.NotZero(t, issued.ExpiresAt)
}

func TestAccessTokenService_Issue_ReturnsError_WhenUserIDIsZero(t *testing.T) {
	service := service(t)
	payload := port.AccessTokenPayload{UserID: user.ID{}}

	issued, err := service.Issue(payload)

	assert.Error(t, err)
	assert.Zero(t, issued)
}

func TestAccessTokenService_Validate_ReturnsClaimsForValidToken(t *testing.T) {
	service := service(t)

	payload := port.AccessTokenPayload{UserID: user.NewID()}
	valid, err := service.Issue(payload)
	require.NoError(t, err)

	claims, err := service.Validate(valid.Token.Value())

	require.NoError(t, err)
	require.NotZero(t, claims)
	assert.Equal(t, payload.UserID, claims.UserID)
}

func TestAccessTokenService_Validate_ReturnsError_WhenTokenIsInvalid(t *testing.T) {
	service := service(t)

	claims, err := service.Validate("invalid")

	assert.ErrorIs(t, err, session.ErrTokenInvalid)
	assert.Zero(t, claims)
}

func TestAccessTokenService_Validate_ReturnsError_WhenTokenIsExpired(t *testing.T) {
	service := service(t)
	service.expiration = 1 * time.Microsecond

	payload := port.AccessTokenPayload{UserID: user.NewID()}
	issued, err := service.Issue(payload)
	require.NoError(t, err)

	claims, err := service.Validate(issued.Token.Value())

	assert.ErrorIs(t, err, session.ErrTokenExpired)
	assert.Zero(t, claims)
}
