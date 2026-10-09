package changepassword_test

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
	"github.com/rafaelblt/go-auth/internal/usecase/changepassword"
	"github.com/rafaelblt/go-auth/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangePassword_ReturnsUser_WhenInputIsValid(t *testing.T) {
	helper := NewTestHelper(t)
	input, pwd := helper.ValidInput()

	output, err := helper.UseCase().Execute(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, pwd.UserID().String(), output.User.ID())
	assert.Equal(t, input.Username, output.User.Username())
}

func TestChangePassword_SavesHashOfNewPassword(t *testing.T) {
	helper := NewTestHelper(t)
	input, pwd := helper.ValidInput()

	_, err := helper.UseCase().Execute(context.Background(), input)
	require.NoError(t, err)

	update := testutil.Only(t, helper.FakeUnitOfWork.FakePasswordWriter.HashUpdates())
	newHash, ok := helper.FakePasswordHasher.GetHashByPassword(passwordtest.MustPlain(t, input.NewPassword))
	require.True(t, ok, "new password was not hashed")
	assert.Equal(t, pwd.ID(), update.Password.ID())
	assert.Equal(t, newHash, update.Password.Hash())
	assert.Equal(t, pwd.Hash(), update.Previous)
	assert.Equal(t, pwd.CreatedAt(), update.Password.CreatedAt())
	assert.Equal(t, helper.FakeClock.Now(), update.Password.UpdatedAt())
}

func TestChangePassword_RevokesEverySessionOfTheUser(t *testing.T) {
	helper := NewTestHelper(t)
	input, pwd := helper.ValidInput()

	_, err := helper.UseCase().Execute(context.Background(), input)
	require.NoError(t, err)

	revocation := testutil.Only(t, helper.FakeUnitOfWork.FakeSessionWriter.Revocations())
	assert.Equal(t, pwd.UserID(), revocation.UserID)
	assert.Equal(t, helper.FakeClock.Now(), revocation.RevokedAt)
}

func TestChangePassword_ReturnsMalformedError_WhenCredentialsAreInvalid(t *testing.T) {
	testCases := []struct {
		desc        string
		input       changepassword.Input
		expectedErr error
	}{
		{
			desc: "username too short",
			input: changepassword.Input{
				Username:        strings.Repeat("a", user.UsernameMinCodePoints-1),
				CurrentPassword: "12345678",
				NewPassword:     "87654321",
			},
			expectedErr: changepassword.ErrUsernameMalformed,
		},
		{
			desc: "current password too short",
			input: changepassword.Input{
				Username:        "username",
				CurrentPassword: strings.Repeat("a", password.PlainMinCodePoints-1),
				NewPassword:     "87654321",
			},
			expectedErr: changepassword.ErrPasswordMalformed,
		},
		{
			desc: "current password too long",
			input: changepassword.Input{
				Username:        "username",
				CurrentPassword: strings.Repeat("a", password.PlainMaxBytes+1),
				NewPassword:     "87654321",
			},
			expectedErr: changepassword.ErrPasswordMalformed,
		},
		{
			desc: "current password and new password too short",
			input: changepassword.Input{
				Username:        "username",
				CurrentPassword: strings.Repeat("a", password.PlainMinCodePoints-1),
				NewPassword:     strings.Repeat("a", password.PlainMinCodePoints-1),
			},
			expectedErr: changepassword.ErrPasswordMalformed,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper := NewTestHelper(t)
			output, err := helper.UseCase().Execute(context.Background(), tC.input)
			assert.Zero(t, output)
			assert.ErrorIs(t, err, tC.expectedErr)
			assert.Empty(t, helper.FakePasswordChecker.Calls())
		})
	}
}

func TestChangePassword_ReturnsValidationError_WhenNewPasswordIsInvalid(t *testing.T) {
	testCases := []struct {
		desc        string
		newPassword string
		issue       validation.Issue
	}{
		{
			desc:        "too short",
			newPassword: strings.Repeat("a", password.PlainMinCodePoints-1),
			issue:       validation.IssueTooShort(password.PlainMinCodePoints, validation.UnitCodePoint),
		},
		{
			desc:        "too long",
			newPassword: strings.Repeat("a", password.PlainMaxBytes+1),
			issue:       validation.IssueTooLong(password.PlainMaxBytes, validation.UnitByte),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper := NewTestHelper(t)
			input, _ := helper.ValidInput()
			input.NewPassword = tC.newPassword

			output, err := helper.UseCase().Execute(context.Background(), input)

			assert.Zero(t, output)
			var verr validation.ValidationError
			require.ErrorAs(t, err, &verr)
			expected := []validation.FieldError{
				validation.NewFieldError(changepassword.FieldNewPassword, tC.issue),
			}
			assert.ElementsMatch(t, expected, verr.Errors())
			assert.Empty(t, helper.FakePasswordChecker.Calls())
		})
	}
}

// Every rejection of the credentials answers the client with one code, the
// one login uses, so nothing tells an unknown user from a wrong password. Only
// the reason, which stays on the server, tells them apart.
func TestChangePassword_RejectionsShareOneCodeAndDifferInReason(t *testing.T) {
	rejections := []usecase.UseCaseError{
		changepassword.ErrUsernameMalformed,
		changepassword.ErrPasswordMalformed,
		changepassword.ErrUserNotFound,
		changepassword.ErrPasswordNotFound,
		changepassword.ErrPasswordMismatch,
		changepassword.ErrPasswordChanged,
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

func TestChangePassword_ReturnsUserNotFoundError_AfterVerifyingDummyHash(t *testing.T) {
	helper := NewTestHelper(t)
	input := changepassword.Input{
		Username:        usertest.MustUsername(t, "username").String(),
		CurrentPassword: "210-9i)S_D(Akfvfc12)",
		NewPassword:     "0k-nEw_pàsswörd",
	}

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	assert.ErrorIs(t, err, changepassword.ErrUserNotFound)
	call := testutil.Only(t, helper.FakePasswordChecker.Calls())
	assert.Equal(t, input.CurrentPassword, call.Plain.Value())
	assert.Equal(t, helper.DummyPasswordHash, call.Hash)
}

func TestChangePassword_ReturnsPasswordNotFoundError_AfterVerifyingDummyHash(t *testing.T) {
	helper := NewTestHelper(t)
	usr := usertest.NewUser(t, nil)
	helper.FakeUserReader.InsertUser(usr)
	input := changepassword.Input{
		Username:        usr.Username().String(),
		CurrentPassword: "210-9i)S_D(Akfvfc12)",
		NewPassword:     "0k-nEw_pàsswörd",
	}

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	assert.ErrorIs(t, err, changepassword.ErrPasswordNotFound)
	call := testutil.Only(t, helper.FakePasswordChecker.Calls())
	assert.Equal(t, input.CurrentPassword, call.Plain.Value())
	assert.Equal(t, helper.DummyPasswordHash, call.Hash)
}

func TestChangePassword_ReturnsError_WhenDummyVerificationFails(t *testing.T) {
	helper := NewTestHelper(t)
	expectedErr := errors.New("internal error")
	helper.FakePasswordChecker.SetError(expectedErr)

	output, err := helper.UseCase().Execute(context.Background(), changepassword.Input{
		Username:        usertest.MustUsername(t, "username").String(),
		CurrentPassword: "210-9i)S_D(Akfvfc12)",
		NewPassword:     "0k-nEw_pàsswörd",
	})

	assert.Zero(t, output)
	assert.ErrorIs(t, err, expectedErr)
	assert.NotErrorIs(t, err, changepassword.ErrUserNotFound)
}

func TestChangePassword_ReturnsPasswordMismatchError_AndSavesNothing_WhenCurrentPasswordIsIncorrect(t *testing.T) {
	helper := NewTestHelper(t)
	input, _ := helper.ValidInput()
	input.CurrentPassword += "INCORRECT"

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	assert.ErrorIs(t, err, changepassword.ErrPasswordMismatch)
	call := testutil.Only(t, helper.FakePasswordChecker.Calls())
	assert.NotEqual(t, helper.DummyPasswordHash, call.Hash)
	assert.Empty(t, helper.FakeUnitOfWork.FakePasswordWriter.HashUpdates())
	assert.Empty(t, helper.FakeUnitOfWork.FakeSessionWriter.Revocations())
}

func TestChangePassword_ReturnsPasswordChangedError_WhenAnotherChangeGotThereFirst(t *testing.T) {
	helper := NewTestHelper(t)
	input, _ := helper.ValidInput()
	helper.FakeUnitOfWork.FakePasswordWriter.SetError(password.ErrHashChanged)

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	assert.ErrorIs(t, err, changepassword.ErrPasswordChanged)
	assert.Empty(t, helper.FakeUnitOfWork.FakeSessionWriter.Revocations())
}

func TestChangePassword_ReturnsError_WhenADependencyFails(t *testing.T) {
	testCases := []struct {
		desc string
		fail func(TestHelper, error)
	}{
		{desc: "password verification", fail: func(h TestHelper, err error) { h.FakePasswordChecker.SetError(err) }},
		{desc: "password hasher", fail: func(h TestHelper, err error) { h.FakePasswordHasher.SetError(err) }},
		{desc: "unit of work", fail: func(h TestHelper, err error) { h.FakeUnitOfWork.SetError(err) }},
		{desc: "password writer", fail: func(h TestHelper, err error) { h.FakeUnitOfWork.FakePasswordWriter.SetError(err) }},
		{desc: "session writer", fail: func(h TestHelper, err error) { h.FakeUnitOfWork.FakeSessionWriter.SetError(err) }},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper := NewTestHelper(t)
			input, _ := helper.ValidInput()
			expectedErr := errors.New("internal error")
			tC.fail(helper, expectedErr)

			output, err := helper.UseCase().Execute(context.Background(), input)

			assert.Zero(t, output)
			assert.ErrorIs(t, err, expectedErr)
			var uerr usecase.UseCaseError
			assert.False(t, errors.As(err, &uerr), "an unexpected failure was answered as a rejection")
		})
	}
}
