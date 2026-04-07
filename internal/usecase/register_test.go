package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Ptr[T any](value T) *T { return &value }

type RegisterTestHelper struct {
	t                     *testing.T
	FakeUserExistsChecker *FakeUserExistsChecker
	FakeUnitOfWork        *FakeUnitOfWork
	FakeUserWriter        *FakeUserWriter
	FakeCredentialWriter  *FakeCredentialWriter
	FakePasswordHasher    *FakePasswordHasher
	FakeClock             *FakeClock
}

func NewRegisterTestHelper(t *testing.T) RegisterTestHelper {
	helper := RegisterTestHelper{
		t:                     t,
		FakeUserExistsChecker: Ptr(NewFakeUserExistsChecker()),
		FakeUserWriter:        Ptr(NewFakeUserWriter()),
		FakeCredentialWriter:  Ptr(NewFakeCredentialWriter()),
		FakePasswordHasher:    Ptr(NewFakePasswordHasher()),
		FakeClock:             Ptr(NewFakeClock(time.Now().UTC())),
	}
	deps := usecase.UowDeps{
		UserWriter:       helper.FakeUserWriter,
		CredentialWriter: helper.FakeCredentialWriter,
	}
	helper.FakeUnitOfWork = Ptr(NewFakeUnitOfWork(deps))
	return helper
}
func (helper RegisterTestHelper) UseCase() usecase.Register {
	helper.t.Helper()
	uc, err := usecase.NewRegister(usecase.RegisterConfig{
		UserExistsChecker: helper.FakeUserExistsChecker,
		UnitOfWork:        helper.FakeUnitOfWork,
		PasswordHasher:    helper.FakePasswordHasher,
		Clock:             helper.FakeClock,
	})
	require.NoError(helper.t, err)
	return uc
}
func (helper RegisterTestHelper) ValidInput() usecase.RegisterInput {
	helper.t.Helper()
	return usecase.RegisterInput{
		Username: helper.ValidUsername().String(),
		Password: helper.ValidPlainPassword().Value(),
	}
}
func (helper RegisterTestHelper) ValidUsername() domain.Username {
	helper.t.Helper()
	username, err := domain.NewUsername("username")
	require.NoError(helper.t, err)
	return username
}
func (helper RegisterTestHelper) ValidPlainPassword() domain.PlainPassword {
	helper.t.Helper()
	pwd, err := domain.NewPlainPassword("12345678")
	require.NoError(helper.t, err)
	return pwd
}

func TestNewRegister(t *testing.T) {
	testCases := []struct {
		desc      string
		config    usecase.RegisterConfig
		expectErr bool
	}{
		{
			desc: "valid case",
			config: usecase.RegisterConfig{
				UserExistsChecker: Ptr(NewFakeUserExistsChecker()),
				UnitOfWork:        Ptr(NewFakeUnitOfWork(usecase.UowDeps{})),
				PasswordHasher:    Ptr(NewFakePasswordHasher()),
				Clock:             Ptr(NewFakeClock(time.Now().UTC())),
			},
			expectErr: false,
		},
		{
			desc: "user exists checker nil",
			config: usecase.RegisterConfig{
				UserExistsChecker: nil,
				UnitOfWork:        Ptr(NewFakeUnitOfWork(usecase.UowDeps{})),
				PasswordHasher:    Ptr(NewFakePasswordHasher()),
				Clock:             Ptr(NewFakeClock(time.Now().UTC())),
			},
			expectErr: true,
		},
		{
			desc: "uow nil",
			config: usecase.RegisterConfig{
				UserExistsChecker: Ptr(NewFakeUserExistsChecker()),
				UnitOfWork:        nil,
				PasswordHasher:    Ptr(NewFakePasswordHasher()),
				Clock:             Ptr(NewFakeClock(time.Now().UTC())),
			},
			expectErr: true,
		},
		{
			desc: "password hasher nil",
			config: usecase.RegisterConfig{
				UserExistsChecker: Ptr(NewFakeUserExistsChecker()),
				UnitOfWork:        Ptr(NewFakeUnitOfWork(usecase.UowDeps{})),
				PasswordHasher:    nil,
				Clock:             Ptr(NewFakeClock(time.Now().UTC())),
			},
			expectErr: true,
		},
		{
			desc: "clock nil",
			config: usecase.RegisterConfig{
				UserExistsChecker: Ptr(NewFakeUserExistsChecker()),
				UnitOfWork:        Ptr(NewFakeUnitOfWork(usecase.UowDeps{})),
				PasswordHasher:    Ptr(NewFakePasswordHasher()),
				Clock:             nil,
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			uc, err := usecase.NewRegister(tC.config)
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
	helper := NewRegisterTestHelper(t)
	input := helper.ValidInput()

	output, err := helper.UseCase().Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotZero(t, output)
	assert.Equal(t, input.Username, output.User.Username())
}

func TestRegister_ReturnsValidationError_WhenInputIsInvalid(t *testing.T) {
	testCases := []struct {
		desc     string
		input    usecase.RegisterInput
		expected []error
	}{
		{
			desc: "username too short",
			input: usecase.RegisterInput{
				Username: "x",
				Password: "12345678",
			},
			expected: []error{usecase.ErrRegisterUsernameTooShort},
		},
		{
			desc: "username too long",
			input: usecase.RegisterInput{
				Username: strings.Repeat("a", domain.UsernameMaxLen+1),
				Password: "12345678",
			},
			expected: []error{usecase.ErrRegisterUsernameTooLong},
		},
		{
			desc: "password too short",
			input: usecase.RegisterInput{
				Username: "username",
				Password: strings.Repeat("a", domain.PlainPasswordMinLen-1),
			},
			expected: []error{usecase.ErrRegisterPasswordTooShort},
		},
		{
			desc: "password too long",
			input: usecase.RegisterInput{
				Username: "username",
				Password: strings.Repeat("a", domain.PlainPasswordMaxLen+1),
			},
			expected: []error{usecase.ErrRegisterPasswordTooLong},
		},
		{
			desc: "username and password too short",
			input: usecase.RegisterInput{
				Username: strings.Repeat("a", domain.UsernameMinLen-1),
				Password: strings.Repeat("a", domain.PlainPasswordMinLen-1),
			},
			expected: []error{
				usecase.ErrRegisterUsernameTooShort,
				usecase.ErrRegisterPasswordTooShort,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper := NewRegisterTestHelper(t)
			output, err := helper.UseCase().Execute(context.Background(), tC.input)
			assert.Zero(t, output)
			testutil.RequireErrors(t, err, tC.expected...)
		})
	}
}

func TestRegister_ReturnsError_WhenUsernameAlreadyExists(t *testing.T) {
	helper := NewRegisterTestHelper(t)

	username := helper.ValidUsername()
	input := helper.ValidInput()
	input.Username = username.String()
	helper.FakeUserExistsChecker.Usernames.Add(username)

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	assert.ErrorIs(t, err, usecase.ErrRegisterUsernameAlreadyExists)
}

func TestRegister_ShouldUseClock(t *testing.T) {
	helper := NewRegisterTestHelper(t)
	input := helper.ValidInput()

	output, err := helper.UseCase().Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotZero(t, output)
	assert.Equal(t, helper.FakeClock.UtcNow(), output.User.CreatedAt())
}

func TestRegister_SavesNewUser(t *testing.T) {
	helper := NewRegisterTestHelper(t)
	input := helper.ValidInput()

	_, err := helper.UseCase().Execute(context.Background(), input)

	require.NoError(t, err)
	savedUsers := helper.FakeUserWriter.SavedUsers()
	require.Len(t, savedUsers, 1)
	user := savedUsers[0]
	assert.Equal(t, input.Username, user.Username().String())
}

func TestRegister_SavesNewCredential(t *testing.T) {
	helper := NewRegisterTestHelper(t)

	password := helper.ValidPlainPassword()
	input := usecase.RegisterInput{
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
	helper := NewRegisterTestHelper(t)
	input := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeUserWriter.SetError(expectedErr)

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	assert.ErrorIs(t, err, expectedErr)
}

func TestRegister_ReturnsError_WhenCredentialWriterFails(t *testing.T) {
	helper := NewRegisterTestHelper(t)
	input := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeCredentialWriter.SetError(expectedErr)

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	assert.ErrorIs(t, err, expectedErr)
}

func TestRegister_ReturnsError_WhenPasswordHasherFails(t *testing.T) {
	helper := NewRegisterTestHelper(t)
	input := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakePasswordHasher.SetError(expectedErr)

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	assert.ErrorIs(t, err, expectedErr)
}
