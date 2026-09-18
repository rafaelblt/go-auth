package login_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/stretchr/testify/assert"
)

func TestNewLogin(t *testing.T) {
	helper := NewTestHelper(t)
	testCases := []struct {
		desc      string
		config    login.Config
		expectErr bool
	}{
		{
			desc: "valid case",
			config: login.Config{
				UserReader:        helper.FakeUserReader,
				PasswordReader:    helper.FakePasswordReader,
				PasswordChecker:   helper.FakePasswordChecker,
				AccessTokenIssuer: helper.FakeAccessTokenIssuer,
				UnitOfWork:        helper.FakeUnitOfWork,
				Clock:             helper.FakeClock,
				RefreshTokenTTL:   helper.RefreshTokenTTL,
				DummyPasswordHash: helper.DummyPasswordHash,
			},
			expectErr: false,
		},
		{
			desc: "user reader nil",
			config: login.Config{
				UserReader:        nil,
				PasswordReader:    helper.FakePasswordReader,
				PasswordChecker:   helper.FakePasswordChecker,
				AccessTokenIssuer: helper.FakeAccessTokenIssuer,
				UnitOfWork:        helper.FakeUnitOfWork,
				Clock:             helper.FakeClock,
				RefreshTokenTTL:   helper.RefreshTokenTTL,
				DummyPasswordHash: helper.DummyPasswordHash,
			},
			expectErr: true,
		},
		{
			desc: "password reader nil",
			config: login.Config{
				UserReader:        helper.FakeUserReader,
				PasswordReader:    nil,
				PasswordChecker:   helper.FakePasswordChecker,
				AccessTokenIssuer: helper.FakeAccessTokenIssuer,
				UnitOfWork:        helper.FakeUnitOfWork,
				Clock:             helper.FakeClock,
				RefreshTokenTTL:   helper.RefreshTokenTTL,
				DummyPasswordHash: helper.DummyPasswordHash,
			},
			expectErr: true,
		},
		{
			desc: "password checker nil",
			config: login.Config{
				UserReader:        helper.FakeUserReader,
				PasswordReader:    helper.FakePasswordReader,
				PasswordChecker:   nil,
				AccessTokenIssuer: helper.FakeAccessTokenIssuer,
				UnitOfWork:        helper.FakeUnitOfWork,
				Clock:             helper.FakeClock,
				RefreshTokenTTL:   helper.RefreshTokenTTL,
				DummyPasswordHash: helper.DummyPasswordHash,
			},
			expectErr: true,
		},
		{
			desc: "access token issuer nil",
			config: login.Config{
				UserReader:        helper.FakeUserReader,
				PasswordReader:    helper.FakePasswordReader,
				PasswordChecker:   helper.FakePasswordChecker,
				AccessTokenIssuer: nil,
				UnitOfWork:        helper.FakeUnitOfWork,
				Clock:             helper.FakeClock,
				RefreshTokenTTL:   helper.RefreshTokenTTL,
				DummyPasswordHash: helper.DummyPasswordHash,
			},
			expectErr: true,
		},
		{
			desc: "uow nil",
			config: login.Config{
				UserReader:        helper.FakeUserReader,
				PasswordReader:    helper.FakePasswordReader,
				PasswordChecker:   helper.FakePasswordChecker,
				AccessTokenIssuer: helper.FakeAccessTokenIssuer,
				UnitOfWork:        nil,
				Clock:             helper.FakeClock,
				RefreshTokenTTL:   helper.RefreshTokenTTL,
				DummyPasswordHash: helper.DummyPasswordHash,
			},
			expectErr: true,
		},
		{
			desc: "clock nil",
			config: login.Config{
				UserReader:        helper.FakeUserReader,
				PasswordReader:    helper.FakePasswordReader,
				PasswordChecker:   helper.FakePasswordChecker,
				AccessTokenIssuer: helper.FakeAccessTokenIssuer,
				UnitOfWork:        helper.FakeUnitOfWork,
				Clock:             nil,
				RefreshTokenTTL:   helper.RefreshTokenTTL,
				DummyPasswordHash: helper.DummyPasswordHash,
			},
			expectErr: true,
		},
		{
			desc: "refresh token ttl zero",
			config: login.Config{
				UserReader:        helper.FakeUserReader,
				PasswordReader:    helper.FakePasswordReader,
				PasswordChecker:   helper.FakePasswordChecker,
				AccessTokenIssuer: helper.FakeAccessTokenIssuer,
				UnitOfWork:        helper.FakeUnitOfWork,
				Clock:             helper.FakeClock,
				RefreshTokenTTL:   0,
				DummyPasswordHash: helper.DummyPasswordHash,
			},
			expectErr: true,
		},
		{
			desc: "refresh token ttl negative",
			config: login.Config{
				UserReader:        helper.FakeUserReader,
				PasswordReader:    helper.FakePasswordReader,
				PasswordChecker:   helper.FakePasswordChecker,
				AccessTokenIssuer: helper.FakeAccessTokenIssuer,
				UnitOfWork:        helper.FakeUnitOfWork,
				Clock:             helper.FakeClock,
				RefreshTokenTTL:   -1,
				DummyPasswordHash: helper.DummyPasswordHash,
			},
			expectErr: true,
		},
		{
			desc: "dummy password hash zero",
			config: login.Config{
				UserReader:        helper.FakeUserReader,
				PasswordReader:    helper.FakePasswordReader,
				PasswordChecker:   helper.FakePasswordChecker,
				AccessTokenIssuer: helper.FakeAccessTokenIssuer,
				UnitOfWork:        helper.FakeUnitOfWork,
				Clock:             helper.FakeClock,
				RefreshTokenTTL:   helper.RefreshTokenTTL,
				DummyPasswordHash: password.Hashed{},
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			uc, err := login.New(tC.config)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, uc)
				return
			}
			assert.NoError(t, err)
			assert.NotZero(t, uc)
		})
	}
}
