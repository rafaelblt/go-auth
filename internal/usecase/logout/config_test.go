package logout_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/usecase/logout"
	"github.com/stretchr/testify/assert"
)

func TestNewLogout(t *testing.T) {
	helper := NewTestHelper(t)
	testCases := []struct {
		desc      string
		config    logout.Config
		expectErr bool
	}{
		{
			desc: "valid case",
			config: logout.Config{
				SessionReader:      helper.FakeSessionReader,
				RefreshTokenReader: helper.FakeRefreshTokenReader,
				UnitOfWork:         helper.FakeUnitOfWork,
				Clock:              helper.FakeClock,
			},
			expectErr: false,
		},
		{
			desc: "session reader nil",
			config: logout.Config{
				SessionReader:      nil,
				RefreshTokenReader: helper.FakeRefreshTokenReader,
				UnitOfWork:         helper.FakeUnitOfWork,
				Clock:              helper.FakeClock,
			},
			expectErr: true,
		},
		{
			desc: "refresh token reader nil",
			config: logout.Config{
				SessionReader:      helper.FakeSessionReader,
				RefreshTokenReader: nil,
				UnitOfWork:         helper.FakeUnitOfWork,
				Clock:              helper.FakeClock,
			},
			expectErr: true,
		},
		{
			desc: "uow nil",
			config: logout.Config{
				SessionReader:      helper.FakeSessionReader,
				RefreshTokenReader: helper.FakeRefreshTokenReader,
				UnitOfWork:         nil,
				Clock:              helper.FakeClock,
			},
			expectErr: true,
		},
		{
			desc: "clock nil",
			config: logout.Config{
				SessionReader:      helper.FakeSessionReader,
				RefreshTokenReader: helper.FakeRefreshTokenReader,
				UnitOfWork:         helper.FakeUnitOfWork,
				Clock:              nil,
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			uc, err := logout.New(tC.config)
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
