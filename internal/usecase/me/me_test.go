package me_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/me"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var validInput = me.Input{AccessToken: "access token"}

// SUCCESS

func TestMe_ValidatesTheAccessToken(t *testing.T) {
	helper := NewTestHelper(t)
	helper.AddUserOfToken()

	_, err := helper.UseCase().Execute(context.Background(), validInput)

	require.NoError(t, err)
	raw := testutil.Only(t, helper.FakeAccessTokenValidator.Raws())
	assert.Equal(t, validInput.AccessToken, raw)
}

func TestMe_ReturnsUserOfToken(t *testing.T) {
	helper := NewTestHelper(t)
	usr := helper.AddUserOfToken()

	out, err := helper.UseCase().Execute(context.Background(), validInput)

	require.NoError(t, err)
	assert.Equal(t, usr.ID().String(), out.User.ID())
	assert.Equal(t, usr.Username().String(), out.User.Username())
}

// REJECTED TOKENS

func TestMe_ReturnsTokenInvalidError(t *testing.T) {
	helper := NewTestHelper(t)
	helper.AddUserOfToken()
	helper.FakeAccessTokenValidator.SetError(session.ErrTokenInvalid)

	out, err := helper.UseCase().Execute(context.Background(), validInput)

	assert.ErrorIs(t, err, me.ErrTokenInvalid)
	assert.Zero(t, out)
}

func TestMe_ReturnsTokenExpiredError(t *testing.T) {
	helper := NewTestHelper(t)
	helper.AddUserOfToken()
	helper.FakeAccessTokenValidator.SetError(session.ErrTokenExpired)

	out, err := helper.UseCase().Execute(context.Background(), validInput)

	assert.ErrorIs(t, err, me.ErrTokenExpired)
	assert.Zero(t, out)
}

func TestMe_ReturnsUserNotFoundError_WhenUserOfTokenIsNotStored(t *testing.T) {
	helper := NewTestHelper(t)

	out, err := helper.UseCase().Execute(context.Background(), validInput)

	assert.ErrorIs(t, err, me.ErrUserNotFound)
	assert.Zero(t, out)
}

// Every rejection answers the client with one code: whatever failed, the
// token is no good. Only the reason, which stays on the server, tells them
// apart.
func TestMe_RejectionsShareOneCodeAndDifferInReason(t *testing.T) {
	rejections := []usecase.UseCaseError{
		me.ErrTokenInvalid,
		me.ErrTokenExpired,
		me.ErrUserNotFound,
	}

	reasons := map[string]bool{}
	for _, rejection := range rejections {
		assert.Equal(t, "invalid_token", rejection.Code())
		assert.Equal(t, usecase.ErrorKindUnauthorized, rejection.Kind())
		assert.NotEmpty(t, rejection.Reason())
		reasons[rejection.Reason()] = true
	}
	assert.Len(t, reasons, len(rejections))
}

// UNEXPECTED ERRORS

func TestMe_ReturnsError_WhenValidatorFails(t *testing.T) {
	helper := NewTestHelper(t)
	expectedErr := errors.New("internal error")
	helper.FakeAccessTokenValidator.SetError(expectedErr)

	out, err := helper.UseCase().Execute(context.Background(), validInput)

	assert.Zero(t, out)
	assert.ErrorIs(t, err, expectedErr)
	var uerr usecase.UseCaseError
	assert.False(t, errors.As(err, &uerr), "expected an unexpected error, got use case error %v", err)
}

func TestMe_ReturnsError_WhenUserReaderFails(t *testing.T) {
	helper := NewTestHelper(t)
	helper.AddUserOfToken()
	expectedErr := errors.New("database down")
	helper.FakeUserReader.SetError(expectedErr)

	out, err := helper.UseCase().Execute(context.Background(), validInput)

	assert.Zero(t, out)
	assert.ErrorIs(t, err, expectedErr)
	var uerr usecase.UseCaseError
	assert.False(t, errors.As(err, &uerr), "expected an unexpected error, got use case error %v", err)
}
