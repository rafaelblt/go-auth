package usecase_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func TestMapUserToDTO(t *testing.T) {
	testCases := []struct {
		desc  string
		user  *user.User
		panic bool
	}{
		{
			desc:  "default user",
			user:  usertest.NewUser(t, nil),
			panic: false,
		},
		{
			desc:  "user zero",
			user:  &user.User{},
			panic: true,
		},
		{
			desc:  "user nil",
			user:  nil,
			panic: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			if tC.panic {
				assert.Panics(t, func() {
					usecase.MapUserToDTO(tC.user)
				})
			} else {
				dto := usecase.MapUserToDTO(tC.user)
				assert.Equal(t, tC.user.ID().Value().String(), dto.ID())
				assert.Equal(t, tC.user.Username().String(), dto.Username())
			}
		})
	}
}

func TestMapAccessTokenIssuedToDTO(t *testing.T) {
	issued := port.AccessTokenIssued{
		Token:     sessiontest.MustAccessToken(t, "access"),
		IssuedAt:  time.Date(2026, 6, 17, 23, 10, 0, 400_000_000, time.UTC),
		ExpiresAt: time.Date(2026, 6, 17, 23, 40, 0, 0, time.UTC),
	}

	dto := usecase.MapAccessTokenIssuedToDTO(issued)

	assert.Equal(t, issued.Token.Value(), dto.Value())
	assert.Equal(t, issued.ExpiresAt, dto.ExpiresAt())
	assert.Equal(t, 29*time.Minute+59*time.Second, dto.ExpiresIn())
}

func TestMapAccessTokenIssuedToDTO_PanicsWithZeroToken(t *testing.T) {
	issued := port.AccessTokenIssued{
		ExpiresAt: time.Date(2026, 6, 17, 23, 40, 0, 0, time.UTC),
	}
	assert.Panics(t, func() { usecase.MapAccessTokenIssuedToDTO(issued) })
}

func TestMapRefreshTokenToDTO(t *testing.T) {
	token := sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.CreatedAt = time.Date(2026, 6, 17, 23, 39, 58, 700_000_000, time.UTC)
		p.ExpiresAt = time.Date(2026, 6, 17, 23, 40, 0, 0, time.UTC)
	})
	secret := sessiontest.NewRefreshTokenSecret(t)

	dto := usecase.MapRefreshTokenToDTO(token, secret)

	assert.Equal(t, secret.Value(), dto.Value())
	assert.Equal(t, token.ExpiresAt(), dto.ExpiresAt())
	assert.Equal(t, time.Second, dto.ExpiresIn())
}

func TestMapRefreshTokenToDTO_Panics(t *testing.T) {
	testCases := []struct {
		desc   string
		token  *session.RefreshToken
		secret session.RefreshTokenSecret
	}{
		{
			desc:   "token nil",
			token:  nil,
			secret: sessiontest.NewRefreshTokenSecret(t),
		},
		{
			desc:   "token zero",
			token:  &session.RefreshToken{},
			secret: sessiontest.NewRefreshTokenSecret(t),
		},
		{
			desc:   "secret zero",
			token:  sessiontest.NewRefreshToken(t, nil),
			secret: session.RefreshTokenSecret{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Panics(t, func() {
				usecase.MapRefreshTokenToDTO(tC.token, tC.secret)
			})
		})
	}
}
