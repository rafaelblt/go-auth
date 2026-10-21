package logout_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/usecase/logout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SUCCESS

func TestLogout_RevokesSessionOfToken(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.Seed()

	_, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	require.NoError(t, err)
	helper.AssertSessionRevokedNow(fixture.Session)
	helper.AssertNoRefreshTokenWrites()
}

func TestLogout_ReturnsIDsOfSessionAndItsUser(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.Seed()

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	require.NoError(t, err)
	assert.Equal(t, fixture.Session.UserID().String(), out.UserID)
	assert.Equal(t, fixture.Session.ID().String(), out.SessionID)
	assert.False(t, out.ReuseDetected)
}

func TestLogout_RevokesSession_WhenTokenIsExpired(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.SeedExpiredToken()

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	require.NoError(t, err)
	assert.False(t, out.ReuseDetected)
	helper.AssertSessionRevokedNow(fixture.Session)
	helper.AssertNoRefreshTokenWrites()
}

// REUSE

func TestLogout_RevokesSessionAndReportsReuse_WhenTokenIsAlreadyUsed(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.SeedUsedToken()

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	require.NoError(t, err)
	assert.True(t, out.ReuseDetected)
	assert.Equal(t, fixture.Session.ID().String(), out.SessionID)
	helper.AssertSessionRevokedNow(fixture.Session)
	helper.AssertNoRefreshTokenWrites()
}

// SESSION ALREADY REVOKED

func TestLogout_WritesNothing_WhenSessionIsAlreadyRevoked(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.SeedRevokedSession()

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	require.NoError(t, err)
	assert.Equal(t, fixture.Session.UserID().String(), out.UserID)
	assert.Equal(t, fixture.Session.ID().String(), out.SessionID)
	helper.AssertNoWrites()
}

// A revoked session takes precedence over a spent token, as at refresh.
func TestLogout_DoesNotReportReuse_WhenSessionIsAlreadyRevoked(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.SeedUsedTokenOfRevokedSession()

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	require.NoError(t, err)
	assert.False(t, out.ReuseDetected)
	helper.AssertNoWrites()
}

// Another revocation can commit between the read of the session and its
// revocation by the logout, which then leaves the session as it is.
func TestLogout_DoesNotReportReuse_WhenSessionIsRevokedFirst(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.SeedUsedToken()

	helper.FakeUnitOfWork.FakeSessionWriter.SetError(session.ErrSessionAlreadyRevoked)

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	require.NoError(t, err)
	assert.Equal(t, fixture.Session.UserID().String(), out.UserID)
	assert.Equal(t, fixture.Session.ID().String(), out.SessionID)
	assert.False(t, out.ReuseDetected)
}

// REJECTED TOKENS

func TestLogout_ReturnsTokenInvalidError(t *testing.T) {
	testCases := []struct {
		desc  string
		token string
	}{
		{desc: "malformed", token: "INVALID TOKEN"},
		{desc: "missing", token: ""},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper := NewTestHelper(t)

			out, err := helper.UseCase().Execute(context.Background(), logout.Input{RefreshToken: tC.token})

			assert.ErrorIs(t, err, logout.ErrTokenInvalid)
			assert.Zero(t, out)
			helper.AssertNoWrites()
		})
	}
}

func TestLogout_ReturnsTokenInvalidError_WhenTokenIsNotStored(t *testing.T) {
	helper := NewTestHelper(t)
	in := logout.Input{RefreshToken: sessiontest.NewRefreshTokenSecret(t).Value()}

	out, err := helper.UseCase().Execute(context.Background(), in)

	assert.ErrorIs(t, err, logout.ErrTokenInvalid)
	assert.Zero(t, out)
	helper.AssertNoWrites()
}

// The session cleanup can delete the session between the reads of the token
// and of its session.
func TestLogout_ReturnsTokenInvalidError_WhenSessionOfTokenIsGone(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.SeedTokenWithoutSession()

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	assert.ErrorIs(t, err, logout.ErrTokenInvalid)
	assert.Zero(t, out)
	helper.AssertNoWrites()
}

// The session cleanup can delete the session between its read and its
// revocation.
func TestLogout_ReturnsTokenInvalidError_WhenSessionIsGoneBeforeItIsRevoked(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.Seed()

	helper.FakeUnitOfWork.FakeSessionWriter.SetError(session.ErrSessionNotFound)

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	assert.ErrorIs(t, err, logout.ErrTokenInvalid)
	assert.Zero(t, out)
}

// UNEXPECTED ERRORS

func TestLogout_ReturnsError_WhenRefreshTokenReaderFails(t *testing.T) {
	helper := NewTestHelper(t)
	in := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeRefreshTokenReader.SetError(expectedErr)

	out, err := helper.UseCase().Execute(context.Background(), in)

	assert.Zero(t, out)
	assert.ErrorIs(t, err, expectedErr)
	AssertUnexpectedError(t, err)
	helper.AssertNoWrites()
}

func TestLogout_ReturnsError_WhenSessionReaderFails(t *testing.T) {
	helper := NewTestHelper(t)
	in := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeSessionReader.SetError(expectedErr)

	out, err := helper.UseCase().Execute(context.Background(), in)

	assert.Zero(t, out)
	assert.ErrorIs(t, err, expectedErr)
	AssertUnexpectedError(t, err)
	helper.AssertNoWrites()
}

func TestLogout_ReturnsError_WhenUnitOfWorkFails(t *testing.T) {
	helper := NewTestHelper(t)
	in := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeUnitOfWork.SetError(expectedErr)

	out, err := helper.UseCase().Execute(context.Background(), in)

	assert.Zero(t, out)
	assert.ErrorIs(t, err, expectedErr)
	AssertUnexpectedError(t, err)
}

func TestLogout_ReturnsError_WhenSessionWriterFails(t *testing.T) {
	helper := NewTestHelper(t)
	in := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeUnitOfWork.FakeSessionWriter.SetError(expectedErr)

	out, err := helper.UseCase().Execute(context.Background(), in)

	assert.Zero(t, out)
	assert.ErrorIs(t, err, expectedErr)
	AssertUnexpectedError(t, err)
}
