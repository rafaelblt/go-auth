package register_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/passwordtest"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/stretchr/testify/require"
)

type TestHelper struct {
	t                     *testing.T
	FakeUserExistsChecker *porttest.FakeUserExistsChecker
	FakeUnitOfWork        *porttest.FakeUnitOfWork
	FakePasswordHasher    *porttest.FakePasswordHasher
	FakeClock             *porttest.FakeClock
}

func NewTestHelper(t *testing.T) TestHelper {
	helper := TestHelper{
		t:                     t,
		FakeUserExistsChecker: shared.Ptr(porttest.NewFakeUserExistsChecker()),
		FakeUnitOfWork:        porttest.NewFakeUnitOfWork(),
		FakePasswordHasher:    shared.Ptr(porttest.NewFakePasswordHasher()),
		FakeClock:             porttest.NewFakeClock(),
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
	return usertest.MustUsername(helper.t, "username")
}

func (helper TestHelper) ValidPlainPassword() password.Plain {
	helper.t.Helper()
	return passwordtest.MustPlain(helper.t, "12345678")
}
