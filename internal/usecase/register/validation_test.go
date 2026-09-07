package register_test

import (
	"context"
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/password"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/rafaelblt/go-auth/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_ReturnsValidationError_WhenUsernameIsInvalid(t *testing.T) {
	helper := NewTestHelper(t)
	input := register.Input{
		Username: strings.Repeat("a", user.UsernameMinLen-1),
		Password: helper.ValidPlainPassword().Value(),
	}

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	var verr validation.ValidationError
	require.ErrorAs(t, err, &verr)
	expected := []validation.FieldError{validation.NewFieldError(
		register.FieldUsername,
		validation.IssueTooShort(user.UsernameMinLen, validation.UnitCodePoint),
	)}
	assert.ElementsMatch(t, expected, verr.Errors())
}

func TestRegister_ReturnsValidationError_WithUsernameTooLong(t *testing.T) {
	helper := NewTestHelper(t)
	input := register.Input{
		Username: strings.Repeat("a", user.UsernameMaxLen+1),
		Password: helper.ValidPlainPassword().Value(),
	}

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	var verr validation.ValidationError
	require.ErrorAs(t, err, &verr)
	expected := []validation.FieldError{validation.NewFieldError(
		register.FieldUsername,
		validation.IssueTooLong(user.UsernameMaxLen, validation.UnitCodePoint),
	)}
	assert.ElementsMatch(t, expected, verr.Errors())
}

func TestRegister_ReturnsValidationError_WithPasswordTooLong(t *testing.T) {
	helper := NewTestHelper(t)
	input := register.Input{
		Username: helper.ValidUsername().String(),
		Password: strings.Repeat("a", password.PlainMaxBytes+1),
	}

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	var verr validation.ValidationError
	require.ErrorAs(t, err, &verr)
	expected := []validation.FieldError{validation.NewFieldError(
		register.FieldPassword,
		validation.IssueTooLong(password.PlainMaxBytes, validation.UnitByte),
	)}
	assert.ElementsMatch(t, expected, verr.Errors())
}

func TestRegister_ReturnsValidationError_WithPasswordTooShort(t *testing.T) {
	helper := NewTestHelper(t)
	input := register.Input{
		Username: helper.ValidUsername().String(),
		Password: strings.Repeat("a", password.PlainMinCodePoints-1),
	}

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	var verr validation.ValidationError
	require.ErrorAs(t, err, &verr)
	expected := []validation.FieldError{validation.NewFieldError(
		register.FieldPassword,
		validation.IssueTooShort(password.PlainMinCodePoints, validation.UnitCodePoint),
	)}
	assert.ElementsMatch(t, expected, verr.Errors())
}

func TestRegister_ReturnsValidationError_WithUsernameAndPasswordTooShort(t *testing.T) {
	helper := NewTestHelper(t)
	input := register.Input{
		Username: strings.Repeat("a", user.UsernameMinLen-1),
		Password: strings.Repeat("a", password.PlainMinCodePoints-1),
	}

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	var verr validation.ValidationError
	require.ErrorAs(t, err, &verr)
	expected := []validation.FieldError{
		validation.NewFieldError(
			register.FieldUsername,
			validation.IssueTooShort(user.UsernameMinLen, validation.UnitCodePoint),
		),
		validation.NewFieldError(
			register.FieldPassword,
			validation.IssueTooShort(password.PlainMinCodePoints, validation.UnitCodePoint),
		),
	}
	assert.ElementsMatch(t, expected, verr.Errors())
}

func TestRegister_ReturnsValidationError_WithUsernameAndPasswordTooLong(t *testing.T) {
	helper := NewTestHelper(t)
	input := register.Input{
		Username: strings.Repeat("a", user.UsernameMaxLen+1),
		Password: strings.Repeat("a", password.PlainMaxBytes+1),
	}

	output, err := helper.UseCase().Execute(context.Background(), input)

	assert.Zero(t, output)
	var verr validation.ValidationError
	require.ErrorAs(t, err, &verr)
	expected := []validation.FieldError{
		validation.NewFieldError(
			register.FieldUsername,
			validation.IssueTooLong(user.UsernameMaxLen, validation.UnitCodePoint),
		),
		validation.NewFieldError(
			register.FieldPassword,
			validation.IssueTooLong(password.PlainMaxBytes, validation.UnitByte),
		),
	}
	assert.ElementsMatch(t, expected, verr.Errors())
}
