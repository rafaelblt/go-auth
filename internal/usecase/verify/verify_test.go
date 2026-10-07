package verify_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var validInput = verify.Input{AccessToken: "access token"}

// SUCCESS

func TestVerify_ValidatesTheAccessToken(t *testing.T) {
	helper := NewTestHelper(t)

	_, err := helper.UseCase().Execute(context.Background(), validInput)

	require.NoError(t, err)
	raw := testutil.Only(t, helper.FakeAccessTokenValidator.Raws())
	assert.Equal(t, validInput.AccessToken, raw)
}

func TestVerify_ReturnsUserIDAndExpiryOfToken(t *testing.T) {
	helper := NewTestHelper(t)
	claims := helper.FakeAccessTokenValidator.Claims()

	out, err := helper.UseCase().Execute(context.Background(), validInput)

	require.NoError(t, err)
	assert.Equal(t, claims.UserID.String(), out.UserID)
	assert.Equal(t, claims.ExpiresAt, out.ExpiresAt)
}

func TestVerify_ReturnsExpiresIn_FromNowRoundedDown(t *testing.T) {
	helper := NewTestHelper(t)
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	helper.FakeClock.SetNow(now)
	helper.FakeAccessTokenValidator.SetClaims(port.AccessTokenClaims{
		UserID:    user.NewID(),
		ExpiresAt: now.Add(22*time.Minute + 22*time.Second + 999*time.Millisecond),
	})

	out, err := helper.UseCase().Execute(context.Background(), validInput)

	require.NoError(t, err)
	assert.Equal(t, 22*time.Minute+22*time.Second, out.ExpiresIn)
}

// REJECTED TOKENS

func TestVerify_ReturnsTokenInvalidError(t *testing.T) {
	helper := NewTestHelper(t)
	helper.FakeAccessTokenValidator.SetError(session.ErrTokenInvalid)

	out, err := helper.UseCase().Execute(context.Background(), validInput)

	assert.ErrorIs(t, err, verify.ErrTokenInvalid)
	assert.Zero(t, out)
}

func TestVerify_ReturnsTokenExpiredError(t *testing.T) {
	helper := NewTestHelper(t)
	helper.FakeAccessTokenValidator.SetError(session.ErrTokenExpired)

	out, err := helper.UseCase().Execute(context.Background(), validInput)

	assert.ErrorIs(t, err, verify.ErrTokenExpired)
	assert.Zero(t, out)
}

// Every rejection answers the client with one code: whatever failed, the
// token is no good. Only the reason, which stays on the server, tells them
// apart.
func TestVerify_RejectionsShareOneCodeAndDifferInReason(t *testing.T) {
	rejections := []usecase.UseCaseError{
		verify.ErrTokenInvalid,
		verify.ErrTokenExpired,
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

func TestVerify_ReturnsError_WhenValidatorFails(t *testing.T) {
	helper := NewTestHelper(t)
	expectedErr := errors.New("internal error")
	helper.FakeAccessTokenValidator.SetError(expectedErr)

	out, err := helper.UseCase().Execute(context.Background(), validInput)

	assert.Zero(t, out)
	assert.ErrorIs(t, err, expectedErr)
	var uerr usecase.UseCaseError
	assert.False(t, errors.As(err, &uerr), "expected an unexpected error, got use case error %v", err)
}
