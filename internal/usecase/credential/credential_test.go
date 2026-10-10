package credential_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/testutil/passwordtest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/credential"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChecker(t *testing.T) {
	testCases := []struct {
		desc  string
		unset func(*credential.Config)
	}{
		{desc: "user reader nil", unset: func(c *credential.Config) { c.UserReader = nil }},
		{desc: "password reader nil", unset: func(c *credential.Config) { c.PasswordReader = nil }},
		{desc: "password checker nil", unset: func(c *credential.Config) { c.PasswordChecker = nil }},
		{desc: "dummy password hash zero", unset: func(c *credential.Config) { c.DummyPasswordHash = password.Hashed{} }},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			cfg := NewTestHelper(t).Config()
			tC.unset(&cfg)

			_, err := credential.NewChecker(cfg)

			assert.Error(t, err)
		})
	}
	t.Run("valid", func(t *testing.T) {
		_, err := credential.NewChecker(NewTestHelper(t).Config())
		assert.NoError(t, err)
	})
}

func TestParse_ReturnsUsernameAndPassword(t *testing.T) {
	username, plain, err := credential.Parse("username", "a-v4lid-pàssword")

	require.NoError(t, err)
	assert.Equal(t, usertest.MustUsername(t, "username"), username)
	assert.Equal(t, "a-v4lid-pàssword", plain.Value())
}

func TestParse_ReturnsMalformedError_WhenAValueBreaksAFormatRule(t *testing.T) {
	testCases := []struct {
		desc        string
		username    string
		password    string
		expectedErr error
	}{
		{
			desc:        "username too short",
			username:    strings.Repeat("a", user.UsernameMinCodePoints-1),
			password:    "a-v4lid-pàssword",
			expectedErr: credential.ErrUsernameMalformed,
		},
		{
			desc:        "password too short",
			username:    "username",
			password:    strings.Repeat("a", password.PlainMinCodePoints-1),
			expectedErr: credential.ErrPasswordMalformed,
		},
		{
			desc:        "password too long",
			username:    "username",
			password:    strings.Repeat("a", password.PlainMaxBytes+1),
			expectedErr: credential.ErrPasswordMalformed,
		},
		{
			desc:        "both",
			username:    strings.Repeat("a", user.UsernameMinCodePoints-1),
			password:    strings.Repeat("a", password.PlainMinCodePoints-1),
			expectedErr: credential.ErrUsernameMalformed,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			username, plain, err := credential.Parse(tC.username, tC.password)

			assert.ErrorIs(t, err, tC.expectedErr)
			assert.Zero(t, username)
			assert.Zero(t, plain)
		})
	}
}

// Every rejection answers the client with one code, so nothing tells a
// malformed username from an unknown user or a wrong password. Only the
// reason, which stays on the server, tells them apart.
func TestRejectionsShareOneCodeAndDifferInReason(t *testing.T) {
	rejections := []usecase.UseCaseError{
		credential.ErrUsernameMalformed,
		credential.ErrPasswordMalformed,
		credential.ErrUserNotFound,
		credential.ErrPasswordNotFound,
		credential.ErrPasswordMismatch,
		credential.ErrPasswordChanged,
	}

	reasons := map[string]bool{}
	for _, rejection := range rejections {
		assert.Equal(t, "invalid_credentials", rejection.Code())
		assert.Equal(t, usecase.ErrorKindUnauthorized, rejection.Kind())
		assert.NotEmpty(t, rejection.Reason())
		reasons[rejection.Reason()] = true
	}
	assert.Len(t, reasons, len(rejections))
}

func TestChecker_Check_ReturnsUserAndPassword_WhenPasswordIsCorrect(t *testing.T) {
	helper := NewTestHelper(t)
	plain := passwordtest.MustPlain(t, "al-=-vçd1çf1-04ktsx")
	usr, pwd := helper.AddAccount(plain)

	gotUser, gotPassword, err := helper.Checker().Check(context.Background(), usr.Username(), plain)

	require.NoError(t, err)
	assert.Equal(t, usr.ID(), gotUser.ID())
	assert.Equal(t, pwd.ID(), gotPassword.ID())
	call := testutil.Only(t, helper.FakePasswordChecker.Calls())
	assert.Equal(t, pwd.Hash(), call.Hash)
}

func TestChecker_Check_ReturnsUserNotFoundError_AfterVerifyingDummyHash(t *testing.T) {
	helper := NewTestHelper(t)
	plain := passwordtest.MustPlain(t, "210-9i)S_D(Akfvfc12)")

	usr, pwd, err := helper.Checker().Check(context.Background(), usertest.MustUsername(t, "username"), plain)

	assert.ErrorIs(t, err, credential.ErrUserNotFound)
	assert.Nil(t, usr)
	assert.Nil(t, pwd)
	call := testutil.Only(t, helper.FakePasswordChecker.Calls())
	assert.Equal(t, plain, call.Plain)
	assert.Equal(t, helper.DummyPasswordHash, call.Hash)
}

func TestChecker_Check_ReturnsPasswordNotFoundError_AfterVerifyingDummyHash(t *testing.T) {
	helper := NewTestHelper(t)
	usr := usertest.NewUser(t, nil)
	helper.FakeUserReader.InsertUser(usr)
	plain := passwordtest.MustPlain(t, "210-9i)S_D(Akfvfc12)")

	gotUser, pwd, err := helper.Checker().Check(context.Background(), usr.Username(), plain)

	assert.ErrorIs(t, err, credential.ErrPasswordNotFound)
	assert.Nil(t, gotUser)
	assert.Nil(t, pwd)
	call := testutil.Only(t, helper.FakePasswordChecker.Calls())
	assert.Equal(t, plain, call.Plain)
	assert.Equal(t, helper.DummyPasswordHash, call.Hash)
}

func TestChecker_Check_ReturnsPasswordMismatchError_AfterVerifyingOnlyTheStoredHash(t *testing.T) {
	helper := NewTestHelper(t)
	usr, pwd := helper.AddAccount(passwordtest.MustPlain(t, "al-=-vçd1çf1-04ktsx"))
	wrong := passwordtest.MustPlain(t, "al-=-vçd1çf1-04ktsx-INCORRECT")

	gotUser, gotPassword, err := helper.Checker().Check(context.Background(), usr.Username(), wrong)

	assert.ErrorIs(t, err, credential.ErrPasswordMismatch)
	assert.Nil(t, gotUser)
	assert.Nil(t, gotPassword)
	call := testutil.Only(t, helper.FakePasswordChecker.Calls())
	assert.Equal(t, pwd.Hash(), call.Hash)
}

func TestChecker_Check_ReturnsError_WhenVerificationFails(t *testing.T) {
	testCases := []struct {
		desc       string
		addAccount bool
	}{
		{desc: "dummy verification", addAccount: false},
		{desc: "verification", addAccount: true},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper := NewTestHelper(t)
			plain := passwordtest.MustPlain(t, "al-=-vçd1çf1-04ktsx")
			username := usertest.MustUsername(t, "username")
			if tC.addAccount {
				usr, _ := helper.AddAccount(plain)
				username = usr.Username()
			}
			expectedErr := errors.New("internal error")
			helper.FakePasswordChecker.SetError(expectedErr)

			usr, pwd, err := helper.Checker().Check(context.Background(), username, plain)

			assert.ErrorIs(t, err, expectedErr)
			var uerr usecase.UseCaseError
			assert.False(t, errors.As(err, &uerr), "an unexpected failure was answered as a rejection")
			assert.Nil(t, usr)
			assert.Nil(t, pwd)
		})
	}
}
