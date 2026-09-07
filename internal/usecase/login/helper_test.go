package login_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/password"
	"github.com/rafaelblt/go-auth/internal/testutil/passwordtest"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/require"
)

type TestHelper struct {
	t                         *testing.T
	FakeUserReader            *porttest.FakeUserReader
	FakePasswordReader        *porttest.FakePasswordReader
	FakePasswordChecker       *porttest.FakePasswordChecker
	FakeAccessTokenIssuer     *porttest.FakeAccessTokenIssuer
	FakeRefreshTokenGenerator *porttest.FakeRefreshTokenGenerator
	FakeUnitOfWork            *porttest.FakeUnitOfWork
	FakeClock                 *porttest.FakeClock
	RefreshTokenTTL           time.Duration
}

func NewTestHelper(t *testing.T) TestHelper {
	helper := TestHelper{
		t:                         t,
		FakeUserReader:            porttest.NewFakeUserReader(),
		FakePasswordReader:        porttest.NewFakePasswordReader(),
		FakePasswordChecker:       porttest.NewFakePasswordChecker(),
		FakeAccessTokenIssuer:     porttest.NewFakeAccessTokenIssuer(),
		FakeRefreshTokenGenerator: porttest.NewFakeRefreshTokenGenerator(),
		FakeUnitOfWork:            porttest.NewFakeUnitOfWork(),
		FakeClock:                 porttest.NewFakeClock(),
		RefreshTokenTTL:           24 * time.Hour,
	}
	return helper
}

func (helper TestHelper) UseCase() login.Login {
	helper.t.Helper()
	uc, err := login.New(login.Config{
		UserReader:            helper.FakeUserReader,
		PasswordReader:        helper.FakePasswordReader,
		PasswordChecker:       helper.FakePasswordChecker,
		AccessTokenIssuer:     helper.FakeAccessTokenIssuer,
		RefreshTokenGenerator: helper.FakeRefreshTokenGenerator,
		UnitOfWork:            helper.FakeUnitOfWork,
		Clock:                 helper.FakeClock,
		RefreshTokenTTL:       helper.RefreshTokenTTL,
	})
	require.NoError(helper.t, err)
	return uc
}

func (helper TestHelper) ValidInput() login.Input {
	helper.t.Helper()
	usr, pwd := helper.GetUserAndPassword()
	return login.Input{
		Username: usr.Username().String(),
		Password: pwd.Value(),
	}
}

func (helper TestHelper) GetUserAndPassword() (*user.User, password.Plain) {
	helper.t.Helper()

	usr := usertest.NewUser(helper.t, nil)
	pwd := passwordtest.MustPlain(helper.t, "al-=-vçd1çf1-04ktsx")

	helper.AddUserAndPassword(usr, pwd)
	return usr, pwd
}

func (helper TestHelper) AddUserAndPassword(user *user.User, plain password.Plain) {
	helper.FakeUserReader.InsertUser(user)
	hash := passwordtest.MustHashed(helper.t, plain.Value())
	pwd := passwordtest.NewPassword(helper.t, func(params *password.RestoreParams) {
		params.UserID = user.ID()
		params.Hash = hash
	})
	helper.FakePasswordReader.InsertPassword(pwd)
	helper.FakePasswordChecker.SetPair(plain, hash)
}
