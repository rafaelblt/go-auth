package changepassword_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/passwordtest"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/usecase/changepassword"
	"github.com/stretchr/testify/require"
)

type TestHelper struct {
	t                   *testing.T
	FakeUserReader      *porttest.FakeUserReader
	FakePasswordReader  *porttest.FakePasswordReader
	FakePasswordChecker *porttest.FakePasswordChecker
	FakePasswordHasher  *porttest.FakePasswordHasher
	FakeUnitOfWork      *porttest.FakeUnitOfWork
	FakeClock           *porttest.FakeClock
	DummyPasswordHash   password.Hashed
}

func NewTestHelper(t *testing.T) TestHelper {
	helper := TestHelper{
		t:                   t,
		FakeUserReader:      porttest.NewFakeUserReader(),
		FakePasswordReader:  porttest.NewFakePasswordReader(),
		FakePasswordChecker: porttest.NewFakePasswordChecker(),
		FakePasswordHasher:  shared.Ptr(porttest.NewFakePasswordHasher()),
		FakeUnitOfWork:      porttest.NewFakeUnitOfWork(),
		FakeClock:           porttest.NewFakeClock(),
		DummyPasswordHash:   passwordtest.MustHashed(t, "dummy-hash"),
	}
	return helper
}

func (helper TestHelper) UseCase() changepassword.ChangePassword {
	helper.t.Helper()
	uc, err := changepassword.New(changepassword.Config{
		UserReader:        helper.FakeUserReader,
		PasswordReader:    helper.FakePasswordReader,
		PasswordChecker:   helper.FakePasswordChecker,
		PasswordHasher:    helper.FakePasswordHasher,
		UnitOfWork:        helper.FakeUnitOfWork,
		Clock:             helper.FakeClock,
		DummyPasswordHash: helper.DummyPasswordHash,
	})
	require.NoError(helper.t, err)
	return uc
}

// ValidInput changes the password of a user it adds, and returns the stored
// password as it was before the change.
func (helper TestHelper) ValidInput() (changepassword.Input, *password.Password) {
	helper.t.Helper()
	usr := usertest.NewUser(helper.t, nil)
	current := passwordtest.MustPlain(helper.t, "al-=-vçd1çf1-04ktsx")
	pwd := helper.AddUserAndPassword(usr, current)
	input := changepassword.Input{
		Username:        usr.Username().String(),
		CurrentPassword: current.Value(),
		NewPassword:     "0k-nEw_pàsswörd",
	}
	return input, pwd
}

func (helper TestHelper) AddUserAndPassword(usr *user.User, plain password.Plain) *password.Password {
	helper.t.Helper()
	helper.FakeUserReader.InsertUser(usr)
	hash := passwordtest.MustHashed(helper.t, plain.Value())
	pwd := passwordtest.NewPassword(helper.t, func(params *password.RestoreParams) {
		params.UserID = usr.ID()
		params.Hash = hash
	})
	helper.FakePasswordReader.InsertPassword(pwd)
	helper.FakePasswordChecker.SetPair(plain, hash)
	return shared.ClonePtr(pwd)
}
