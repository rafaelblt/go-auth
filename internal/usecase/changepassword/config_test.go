package changepassword_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/usecase/changepassword"
	"github.com/stretchr/testify/assert"
)

func TestNewChangePassword(t *testing.T) {
	helper := NewTestHelper(t)
	valid := func() changepassword.Config {
		return changepassword.Config{
			UserReader:        helper.FakeUserReader,
			PasswordReader:    helper.FakePasswordReader,
			PasswordChecker:   helper.FakePasswordChecker,
			PasswordHasher:    helper.FakePasswordHasher,
			UnitOfWork:        helper.FakeUnitOfWork,
			Clock:             helper.FakeClock,
			DummyPasswordHash: helper.DummyPasswordHash,
		}
	}
	testCases := []struct {
		desc      string
		mutate    func(*changepassword.Config)
		expectErr bool
	}{
		{desc: "valid case", mutate: func(*changepassword.Config) {}, expectErr: false},
		{desc: "user reader nil", mutate: func(c *changepassword.Config) { c.UserReader = nil }, expectErr: true},
		{desc: "password reader nil", mutate: func(c *changepassword.Config) { c.PasswordReader = nil }, expectErr: true},
		{desc: "password checker nil", mutate: func(c *changepassword.Config) { c.PasswordChecker = nil }, expectErr: true},
		{desc: "password hasher nil", mutate: func(c *changepassword.Config) { c.PasswordHasher = nil }, expectErr: true},
		{desc: "uow nil", mutate: func(c *changepassword.Config) { c.UnitOfWork = nil }, expectErr: true},
		{desc: "clock nil", mutate: func(c *changepassword.Config) { c.Clock = nil }, expectErr: true},
		{desc: "dummy password hash zero", mutate: func(c *changepassword.Config) { c.DummyPasswordHash = password.Hashed{} }, expectErr: true},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			cfg := valid()
			tC.mutate(&cfg)

			uc, err := changepassword.New(cfg)

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
