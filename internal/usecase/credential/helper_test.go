package credential_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/testutil/passwordtest"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/usecase/credential"
	"github.com/stretchr/testify/require"
)

type TestHelper struct {
	t                   *testing.T
	FakeUserReader      *porttest.FakeUserReader
	FakePasswordReader  *porttest.FakePasswordReader
	FakePasswordChecker *porttest.FakePasswordChecker
	DummyPasswordHash   password.Hashed
}

func NewTestHelper(t *testing.T) TestHelper {
	return TestHelper{
		t:                   t,
		FakeUserReader:      porttest.NewFakeUserReader(),
		FakePasswordReader:  porttest.NewFakePasswordReader(),
		FakePasswordChecker: porttest.NewFakePasswordChecker(),
		DummyPasswordHash:   passwordtest.MustHashed(t, "dummy-hash"),
	}
}

func (helper TestHelper) Config() credential.Config {
	return credential.Config{
		UserReader:        helper.FakeUserReader,
		PasswordReader:    helper.FakePasswordReader,
		PasswordChecker:   helper.FakePasswordChecker,
		DummyPasswordHash: helper.DummyPasswordHash,
	}
}

func (helper TestHelper) Checker() credential.Checker {
	helper.t.Helper()
	checker, err := credential.NewChecker(helper.Config())
	require.NoError(helper.t, err)
	return checker
}

// AddAccount adds a user whose password is plain, and returns both.
func (helper TestHelper) AddAccount(plain password.Plain) (*user.User, *password.Password) {
	helper.t.Helper()
	usr := usertest.NewUser(helper.t, nil)
	helper.FakeUserReader.InsertUser(usr)
	hash := passwordtest.MustHashed(helper.t, plain.Value())
	pwd := passwordtest.NewPassword(helper.t, func(params *password.RestoreParams) {
		params.UserID = usr.ID()
		params.Hash = hash
	})
	helper.FakePasswordReader.InsertPassword(pwd)
	helper.FakePasswordChecker.SetPair(plain, hash)
	return usr, pwd
}
