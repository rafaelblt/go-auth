package login_test

import (
	"testing"

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
				UserReader:         helper.FakeUserReader,
				CredentialReader:   helper.FakeCredentialReader,
				PasswordChecker:    helper.FakePasswordChecker,
				AccessTokenIssuer:  helper.FakeAccessTokenIssuer,
				RefreshTokenIssuer: helper.FakeRefreshTokenIssuer,
				UnitOfWork:         helper.FakeUnitOfWork,
				Clock:              helper.FakeClock,
			},
			expectErr: false,
		},
		{
			desc: "user reader nil",
			config: login.Config{
				UserReader:         nil,
				CredentialReader:   helper.FakeCredentialReader,
				PasswordChecker:    helper.FakePasswordChecker,
				AccessTokenIssuer:  helper.FakeAccessTokenIssuer,
				RefreshTokenIssuer: helper.FakeRefreshTokenIssuer,
				UnitOfWork:         helper.FakeUnitOfWork,
				Clock:              helper.FakeClock,
			},
			expectErr: true,
		},
		{
			desc: "credential reader nil",
			config: login.Config{
				UserReader:         helper.FakeUserReader,
				CredentialReader:   nil,
				PasswordChecker:    helper.FakePasswordChecker,
				AccessTokenIssuer:  helper.FakeAccessTokenIssuer,
				RefreshTokenIssuer: helper.FakeRefreshTokenIssuer,
				UnitOfWork:         helper.FakeUnitOfWork,
				Clock:              helper.FakeClock,
			},
			expectErr: true,
		},
		{
			desc: "password checker nil",
			config: login.Config{
				UserReader:         helper.FakeUserReader,
				CredentialReader:   helper.FakeCredentialReader,
				PasswordChecker:    nil,
				AccessTokenIssuer:  helper.FakeAccessTokenIssuer,
				RefreshTokenIssuer: helper.FakeRefreshTokenIssuer,
				UnitOfWork:         helper.FakeUnitOfWork,
				Clock:              helper.FakeClock,
			},
			expectErr: true,
		},
		{
			desc: "access token issuer nil",
			config: login.Config{
				UserReader:         helper.FakeUserReader,
				CredentialReader:   helper.FakeCredentialReader,
				PasswordChecker:    helper.FakePasswordChecker,
				AccessTokenIssuer:  nil,
				RefreshTokenIssuer: helper.FakeRefreshTokenIssuer,
				UnitOfWork:         helper.FakeUnitOfWork,
				Clock:              helper.FakeClock,
			},
			expectErr: true,
		},
		{
			desc: "refresh token issuer nil",
			config: login.Config{
				UserReader:         helper.FakeUserReader,
				CredentialReader:   helper.FakeCredentialReader,
				PasswordChecker:    helper.FakePasswordChecker,
				AccessTokenIssuer:  helper.FakeAccessTokenIssuer,
				RefreshTokenIssuer: nil,
				UnitOfWork:         helper.FakeUnitOfWork,
				Clock:              helper.FakeClock,
			},
			expectErr: true,
		},
		{
			desc: "uow nil",
			config: login.Config{
				UserReader:         helper.FakeUserReader,
				CredentialReader:   helper.FakeCredentialReader,
				PasswordChecker:    helper.FakePasswordChecker,
				AccessTokenIssuer:  helper.FakeAccessTokenIssuer,
				RefreshTokenIssuer: helper.FakeRefreshTokenIssuer,
				UnitOfWork:         nil,
				Clock:              helper.FakeClock,
			},
			expectErr: true,
		},
		{
			desc: "clock nil",
			config: login.Config{
				UserReader:         helper.FakeUserReader,
				CredentialReader:   helper.FakeCredentialReader,
				PasswordChecker:    helper.FakePasswordChecker,
				AccessTokenIssuer:  helper.FakeAccessTokenIssuer,
				RefreshTokenIssuer: helper.FakeRefreshTokenIssuer,
				UnitOfWork:         helper.FakeUnitOfWork,
				Clock:              nil,
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
