package register_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegister(t *testing.T) {
	helper := NewTestHelper(t)
	testCases := []struct {
		desc      string
		config    register.Config
		expectErr bool
	}{
		{
			desc: "valid case",
			config: register.Config{
				UserExistsChecker: helper.FakeUserExistsChecker,
				UnitOfWork:        helper.FakeUnitOfWork,
				PasswordHasher:    helper.FakePasswordHasher,
				Clock:             helper.FakeClock,
			},
			expectErr: false,
		},
		{
			desc: "user exists checker nil",
			config: register.Config{
				UserExistsChecker: nil,
				UnitOfWork:        helper.FakeUnitOfWork,
				PasswordHasher:    helper.FakePasswordHasher,
				Clock:             helper.FakeClock,
			},
			expectErr: true,
		},
		{
			desc: "uow nil",
			config: register.Config{
				UserExistsChecker: helper.FakeUserExistsChecker,
				UnitOfWork:        nil,
				PasswordHasher:    helper.FakePasswordHasher,
				Clock:             helper.FakeClock,
			},
			expectErr: true,
		},
		{
			desc: "password hasher nil",
			config: register.Config{
				UserExistsChecker: helper.FakeUserExistsChecker,
				UnitOfWork:        helper.FakeUnitOfWork,
				PasswordHasher:    nil,
				Clock:             helper.FakeClock,
			},
			expectErr: true,
		},
		{
			desc: "clock nil",
			config: register.Config{
				UserExistsChecker: helper.FakeUserExistsChecker,
				UnitOfWork:        helper.FakeUnitOfWork,
				PasswordHasher:    helper.FakePasswordHasher,
				Clock:             nil,
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			uc, err := register.New(tC.config)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, uc)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, uc)
			}
		})
	}
}

func TestRegister_ReturnsOutput(t *testing.T) {
	helper := NewTestHelper(t)
	input := helper.ValidInput()

	output, err := helper.UseCase().Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotZero(t, output)
	assert.Equal(t, input.Username, output.User.Username)
}

func TestRegister_ReturnsValidationError_WhenInputIsInvalid(t *testing.T) {
	testCases := []struct {
		desc     string
		input    register.Input
		expected []usecase.FieldError
	}{
		{
			desc: "username too short",
			input: register.Input{
				Username: "x",
				Password: "12345678",
			},
			expected: []usecase.FieldError{
				usecase.NewFieldError(register.UsernameField, domain.ErrUsernameTooShort),
			},
		},
		{
			desc: "username too long",
			input: register.Input{
				Username: strings.Repeat("a", domain.UsernameMaxLen+1),
				Password: "12345678",
			},
			expected: []usecase.FieldError{
				usecase.NewFieldError(register.UsernameField, domain.ErrUsernameTooLong),
			},
		},
		{
			desc: "password too short",
			input: register.Input{
				Username: "username",
				Password: strings.Repeat("a", domain.PlainPasswordMinLen-1),
			},
			expected: []usecase.FieldError{
				usecase.NewFieldError(register.PasswordField, domain.ErrPlainPasswordTooShort),
			},
		},
		{
			desc: "password too long",
			input: register.Input{
				Username: "username",
				Password: strings.Repeat("a", domain.PlainPasswordMaxLen+1),
			},
			expected: []usecase.FieldError{
				usecase.NewFieldError(register.PasswordField, domain.ErrPlainPasswordTooLong),
			},
		},
		{
			desc: "username and password too short",
			input: register.Input{
				Username: strings.Repeat("a", domain.UsernameMinLen-1),
				Password: strings.Repeat("a", domain.PlainPasswordMinLen-1),
			},
			expected: []usecase.FieldError{
				usecase.NewFieldError(register.UsernameField, domain.ErrUsernameTooShort),
				usecase.NewFieldError(register.PasswordField, domain.ErrPlainPasswordTooShort),
			},
		},
		{
			desc: "username and password too long",
			input: register.Input{
				Username: strings.Repeat("a", domain.UsernameMaxLen+1),
				Password: strings.Repeat("a", domain.PlainPasswordMaxLen+1),
			},
			expected: []usecase.FieldError{
				usecase.NewFieldError(register.UsernameField, domain.ErrUsernameTooLong),
				usecase.NewFieldError(register.PasswordField, domain.ErrPlainPasswordTooLong),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper := NewTestHelper(t)
			output, err := helper.UseCase().Execute(context.Background(), tC.input)
			assert.Zero(t, output)
			var verr *usecase.ValidationError
			if assert.ErrorAs(t, err, &verr) {
				assert.ElementsMatch(t, tC.expected, verr.Errors())
			}
		})
	}
}

func TestRegister_ReturnsError_WhenUsernameAlreadyExists(t *testing.T) {
	helper := NewTestHelper(t)

	username := helper.ValidUsername()
	input := helper.ValidInput()
	input.Username = username.String()
	helper.FakeUserExistsChecker.Usernames.Add(username)

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	assert.ErrorIs(t, err, register.ErrUsernameAlreadyExists)
}

func TestRegister_ShouldUseClock(t *testing.T) {
	helper := NewTestHelper(t)
	input := helper.ValidInput()

	output, err := helper.UseCase().Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotZero(t, output)
	assert.Equal(t, helper.FakeClock.UtcNow(), output.User.CreatedAt)
}

func TestRegister_SavesNewUser(t *testing.T) {
	helper := NewTestHelper(t)
	input := helper.ValidInput()

	_, err := helper.UseCase().Execute(context.Background(), input)

	require.NoError(t, err)
	savedUsers := helper.FakeUserWriter.SavedUsers()
	require.Len(t, savedUsers, 1)
	user := savedUsers[0]
	assert.Equal(t, input.Username, user.Username().String())
}

func TestRegister_SavesNewCredential(t *testing.T) {
	helper := NewTestHelper(t)

	password := helper.ValidPlainPassword()
	input := register.Input{
		Username: helper.ValidUsername().String(),
		Password: password.Value(),
	}

	_, err := helper.UseCase().Execute(context.Background(), input)

	require.NoError(t, err)
	savedCreds := helper.FakeCredentialWriter.SavedCredentials()
	require.Len(t, savedCreds, 1)
	verify, _ := helper.FakePasswordHasher.Verify(password, savedCreds[0].Secret())
	assert.True(t, verify)
}

func TestRegister_ReturnsError_WhenUserWriterFails(t *testing.T) {
	helper := NewTestHelper(t)
	input := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeUserWriter.SetError(expectedErr)

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	assert.ErrorIs(t, err, expectedErr)
}

func TestRegister_ReturnsError_WhenCredentialWriterFails(t *testing.T) {
	helper := NewTestHelper(t)
	input := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeCredentialWriter.SetError(expectedErr)

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	assert.ErrorIs(t, err, expectedErr)
}

func TestRegister_ReturnsError_WhenPasswordHasherFails(t *testing.T) {
	helper := NewTestHelper(t)
	input := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakePasswordHasher.SetError(expectedErr)

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	assert.ErrorIs(t, err, expectedErr)
}
