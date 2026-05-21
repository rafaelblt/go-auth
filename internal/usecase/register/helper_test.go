package register_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/usecasetest"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/stretchr/testify/require"
)

type TestHelper struct {
	t                     *testing.T
	FakeUserExistsChecker *usecasetest.FakeUserExistsChecker
	FakeUnitOfWork        *usecasetest.FakeUnitOfWork
	FakeUserWriter        *usecasetest.FakeUserWriter
	FakeCredentialWriter  *usecasetest.FakeCredentialWriter
	FakePasswordHasher    *usecasetest.FakePasswordHasher
	FakeClock             *usecasetest.FakeClock
}

func NewTestHelper(t *testing.T) TestHelper {
	helper := TestHelper{
		t:                     t,
		FakeUserExistsChecker: shared.Ptr(usecasetest.NewFakeUserExistsChecker()),
		FakeUserWriter:        shared.Ptr(usecasetest.NewFakeUserWriter()),
		FakeCredentialWriter:  shared.Ptr(usecasetest.NewFakeCredentialWriter()),
		FakePasswordHasher:    shared.Ptr(usecasetest.NewFakePasswordHasher()),
		FakeClock:             shared.Ptr(usecasetest.NewFakeClock(time.Now().UTC())),
	}
	uowDeps := usecase.UowDeps{
		UserWriter:       helper.FakeUserWriter,
		CredentialWriter: helper.FakeCredentialWriter,
	}
	helper.FakeUnitOfWork = shared.Ptr(usecasetest.NewFakeUnitOfWork(uowDeps))
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

func (helper TestHelper) ValidUsername() domain.Username {
	helper.t.Helper()
	username, err := domain.NewUsername("username")
	require.NoError(helper.t, err)
	return username
}

func (helper TestHelper) ValidPlainPassword() domain.PlainPassword {
	helper.t.Helper()
	pwd, err := domain.NewPlainPassword("12345678")
	require.NoError(helper.t, err)
	return pwd
}
