package register_test

import (
	"context"
	"errors"
	"testing"

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
				return
			}
			assert.NoError(t, err)
			assert.NotZero(t, uc)
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

func TestRegister_ReturnsError_WhenUsernameAlreadyExists(t *testing.T) {
	helper := NewTestHelper(t)

	username := helper.ValidUsername()
	input := helper.ValidInput()
	input.Username = username.String()
	helper.FakeUserExistsChecker.InsertUsername(username)

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
	assert.Equal(t, helper.FakeClock.Now(), output.User.CreatedAt)
}

func TestRegister_SavesNewUser(t *testing.T) {
	helper := NewTestHelper(t)
	input := helper.ValidInput()

	_, err := helper.UseCase().Execute(context.Background(), input)

	require.NoError(t, err)
	savedUsers := helper.FakeUnitOfWork.FakeUserWriter.SavedUsers()
	require.Len(t, savedUsers, 1)
	user := savedUsers[0]
	assert.Equal(t, input.Username, user.Username().String())
}

func TestRegister_SavesNewPassword_AndHashesIt(t *testing.T) {
	helper := NewTestHelper(t)

	plain := helper.ValidPlainPassword()
	input := register.Input{
		Username: helper.ValidUsername().String(),
		Password: plain.Value(),
	}

	_, err := helper.UseCase().Execute(context.Background(), input)

	require.NoError(t, err)

	hash, ok := helper.FakePasswordHasher.GetHashByPassword(plain)
	require.True(t, ok, "use case didnt hash the password")
	assert.True(t,
		helper.FakeUnitOfWork.FakePasswordWriter.CheckHashIsSaved(hash),
	)
}

func TestRegister_ReturnsError_WhenUserWriterFails(t *testing.T) {
	helper := NewTestHelper(t)
	input := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeUnitOfWork.FakeUserWriter.SetError(expectedErr)

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	assert.ErrorIs(t, err, expectedErr)
}

func TestRegister_ReturnsError_WhenPasswordWriterFails(t *testing.T) {
	helper := NewTestHelper(t)
	input := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeUnitOfWork.FakePasswordWriter.SetError(expectedErr)

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
