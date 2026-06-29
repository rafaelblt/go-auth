package refresh_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/stretchr/testify/assert"
)

func TestNewRefresh(t *testing.T) {
	helper := NewTestHelper(t)
	testCases := []struct {
		desc      string
		config    refresh.Config
		expectErr bool
	}{
		{
			desc: "valid case",
			config: refresh.Config{
				SessionReader:         helper.FakeSessionReader,
				AccessTokenIssuer:     helper.FakeAccessTokenIssuer,
				RefreshTokenResolver:  helper.FakeRefreshTokenResolver,
				RefreshTokenGenerator: helper.FakeRefreshTokenGenerator,
				UnitOfWork:            helper.FakeUnitOfWork,
				Clock:                 helper.FakeClock,
				RefreshTokenTTL:       helper.RefreshTokenTTL,
			},
			expectErr: false,
		},
		{
			desc: "session reader nil",
			config: refresh.Config{
				SessionReader:         nil,
				AccessTokenIssuer:     helper.FakeAccessTokenIssuer,
				RefreshTokenResolver:  helper.FakeRefreshTokenResolver,
				RefreshTokenGenerator: helper.FakeRefreshTokenGenerator,
				UnitOfWork:            helper.FakeUnitOfWork,
				Clock:                 helper.FakeClock,
				RefreshTokenTTL:       helper.RefreshTokenTTL,
			},
			expectErr: true,
		},
		{
			desc: "access token issuer nil",
			config: refresh.Config{
				SessionReader:         helper.FakeSessionReader,
				AccessTokenIssuer:     nil,
				RefreshTokenResolver:  helper.FakeRefreshTokenResolver,
				RefreshTokenGenerator: helper.FakeRefreshTokenGenerator,
				UnitOfWork:            helper.FakeUnitOfWork,
				Clock:                 helper.FakeClock,
				RefreshTokenTTL:       helper.RefreshTokenTTL,
			},
			expectErr: true,
		},
		{
			desc: "refresh token resolver nil",
			config: refresh.Config{
				SessionReader:         helper.FakeSessionReader,
				AccessTokenIssuer:     helper.FakeAccessTokenIssuer,
				RefreshTokenResolver:  nil,
				RefreshTokenGenerator: helper.FakeRefreshTokenGenerator,
				UnitOfWork:            helper.FakeUnitOfWork,
				Clock:                 helper.FakeClock,
				RefreshTokenTTL:       helper.RefreshTokenTTL,
			},
			expectErr: true,
		},
		{
			desc: "refresh token generator nil",
			config: refresh.Config{
				SessionReader:         helper.FakeSessionReader,
				AccessTokenIssuer:     helper.FakeAccessTokenIssuer,
				RefreshTokenResolver:  helper.FakeRefreshTokenResolver,
				RefreshTokenGenerator: nil,
				UnitOfWork:            helper.FakeUnitOfWork,
				Clock:                 helper.FakeClock,
				RefreshTokenTTL:       helper.RefreshTokenTTL,
			},
			expectErr: true,
		},
		{
			desc: "uow nil",
			config: refresh.Config{
				SessionReader:         helper.FakeSessionReader,
				AccessTokenIssuer:     helper.FakeAccessTokenIssuer,
				RefreshTokenResolver:  helper.FakeRefreshTokenResolver,
				RefreshTokenGenerator: helper.FakeRefreshTokenGenerator,
				UnitOfWork:            nil,
				Clock:                 helper.FakeClock,
				RefreshTokenTTL:       helper.RefreshTokenTTL,
			},
			expectErr: true,
		},
		{
			desc: "clock nil",
			config: refresh.Config{
				SessionReader:         helper.FakeSessionReader,
				AccessTokenIssuer:     helper.FakeAccessTokenIssuer,
				RefreshTokenResolver:  helper.FakeRefreshTokenResolver,
				RefreshTokenGenerator: helper.FakeRefreshTokenGenerator,
				UnitOfWork:            helper.FakeUnitOfWork,
				Clock:                 nil,
				RefreshTokenTTL:       helper.RefreshTokenTTL,
			},
			expectErr: true,
		},
		{
			desc: "refresh token ttl zero",
			config: refresh.Config{
				SessionReader:         helper.FakeSessionReader,
				AccessTokenIssuer:     helper.FakeAccessTokenIssuer,
				RefreshTokenResolver:  helper.FakeRefreshTokenResolver,
				RefreshTokenGenerator: helper.FakeRefreshTokenGenerator,
				UnitOfWork:            helper.FakeUnitOfWork,
				Clock:                 helper.FakeClock,
				RefreshTokenTTL:       0,
			},
			expectErr: true,
		},
		{
			desc: "refresh token ttl negative",
			config: refresh.Config{
				SessionReader:         helper.FakeSessionReader,
				AccessTokenIssuer:     helper.FakeAccessTokenIssuer,
				RefreshTokenResolver:  helper.FakeRefreshTokenResolver,
				RefreshTokenGenerator: helper.FakeRefreshTokenGenerator,
				UnitOfWork:            helper.FakeUnitOfWork,
				Clock:                 helper.FakeClock,
				RefreshTokenTTL:       -1,
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			uc, err := refresh.New(tC.config)
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
