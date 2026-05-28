package register_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/usecasetest"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/require"
)

type TestHelper struct {
	t                     *testing.T
	FakeUserExistsChecker *usecasetest.FakeUserExistsChecker
	FakeUnitOfWork        *usecasetest.FakeUnitOfWork
	FakePasswordHasher    *usecasetest.FakePasswordHasher
	FakeClock             *usecasetest.FakeClock
}

func NewTestHelper(t *testing.T) TestHelper {
	helper := TestHelper{
		t:                     t,
		FakeUserExistsChecker: shared.Ptr(usecasetest.NewFakeUserExistsChecker()),
		FakeUnitOfWork:        usecasetest.NewFakeUnitOfWork(),
		FakePasswordHasher:    shared.Ptr(usecasetest.NewFakePasswordHasher()),
		FakeClock: usecasetest.NewFakeClock(),
	}
	return helper
}

func (helper TestHelper) UseCase() register.Register {
	helper.t.Helper()
	uc, err := register.New(register.Config{
		UserExistsChecker: helper.FakeUserExistsChecker,
		UnitOfWork:        helper.FakeUnitOfWork,
		PasswordHasher:    helper.FakePasswordHasher,
		Clock:             helper.FakeClock,
	})
	require.NoError(helper.t, err)
	return uc
}

func (helper TestHelper) ValidInput() register.Input {
	helper.t.Helper()
	return register.Input{
		Username: helper.ValidUsername().String(),
		Password: helper.ValidPlainPassword().Value(),
	}
}

func (helper TestHelper) ValidUsername() user.Username {
	helper.t.Helper()
	username, err := user.NewUsername("username")
	require.NoError(helper.t, err)
	return username
}

func (helper TestHelper) ValidPlainPassword() credential.PlainPassword {
	helper.t.Helper()
	pwd, err := credential.NewPlainPassword("12345678")
	require.NoError(helper.t, err)
	return pwd
}
