package refresh_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SUCCESS

func TestRefresh_IssuesAccessToken(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.Seed()

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	require.NoError(t, err)

	payload := testutil.Only(t, helper.FakeAccessTokenIssuer.Payloads())
	assert.Equal(t, fixture.Session.UserID(), payload.UserID)

	issued := testutil.Only(t, helper.FakeAccessTokenIssuer.Issueds())
	assert.Equal(t, issued.Token.Value(), out.AccessToken.Value)
	assert.Equal(t, issued.ExpiresAt, out.AccessToken.ExpiresAt)
}

func TestRefresh_GeneratesAndReturnsRefreshToken(t *testing.T) {
	helper := NewTestHelper(t)
	in := helper.ValidInput()

	out, err := helper.UseCase().Execute(context.Background(), in)

	require.NoError(t, err)
	generated := testutil.Only(t, helper.FakeRefreshTokenGenerator.Generated())
	assert.Equal(t, generated.Raw, out.RefreshToken.Value)
	assert.Equal(t, helper.FakeClock.Now().Add(helper.RefreshTokenTTL), out.RefreshToken.ExpiresAt)
}

func TestRefresh_MarksUsedRefreshToken(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.Seed()

	_, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	require.NoError(t, err)
	marked := testutil.Only(t, helper.FakeUnitOfWork.FakeRefreshTokenWriter.MarkedUsed())
	usedAt, isUsed := marked.UsedAt()
	assert.True(t, isUsed)
	assert.Equal(t, helper.FakeClock.Now(), usedAt)
	assert.Equal(t, fixture.Token.ID(), marked.ID())
}

func TestRefresh_AddGeneratedRefreshToken(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.Seed()

	_, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	require.NoError(t, err)
	expectedHash := testutil.Only(t, helper.FakeRefreshTokenGenerator.Generated()).Hash
	added := testutil.Only(t, helper.FakeUnitOfWork.FakeRefreshTokenWriter.Adds())

	addedParentID, hasParent := added.ParentID()
	require.True(t, hasParent)
	assert.Equal(t, fixture.Token.ID(), addedParentID)
	assert.Equal(t, expectedHash, added.Hash())
	assert.Equal(t, fixture.Token.SessionID(), added.SessionID())
	assert.Equal(t, helper.FakeClock.Now(), added.CreatedAt())
	assert.Equal(t, helper.FakeClock.Now().Add(helper.RefreshTokenTTL), added.ExpiresAt())
}

// REJECTED TOKENS

func TestRefresh_ReturnsTokenInvalidError(t *testing.T) {
	helper := NewTestHelper(t)

	out, err := helper.UseCase().Execute(context.Background(), refresh.Input{RefreshToken: "INVALID TOKEN"})

	assert.ErrorIs(t, err, refresh.ErrTokenInvalid)
	assert.Zero(t, out)
	helper.AssertNoAccessTokenIssued()
	helper.AssertNoWrites()
}

func TestRefresh_ReturnsTokenExpiredError(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.SeedExpiredToken()

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	assert.ErrorIs(t, err, refresh.ErrTokenExpired)
	assert.Zero(t, out)
	helper.AssertNoAccessTokenIssued()
	helper.AssertNoWrites()
}

func TestRefresh_ReturnsSessionRevokedError(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.SeedRevokedSession()

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	assert.ErrorIs(t, err, refresh.ErrSessionRevoked)
	assert.Zero(t, out)
	helper.AssertNoAccessTokenIssued()
	helper.AssertNoWrites()
}

// REUSE DETECTION

func TestRefresh_ReturnsTokenAlreadyUsedError(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.SeedUsedToken()

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	assert.ErrorIs(t, err, refresh.ErrTokenAlreadyUsed)
	assert.Zero(t, out)
}

func TestRefresh_RevokesSessionOfTokenAlreadyUsed(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.SeedUsedToken()

	_, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	require.ErrorIs(t, err, refresh.ErrTokenAlreadyUsed)
	updated := testutil.Only(t, helper.FakeUnitOfWork.FakeSessionWriter.Updates())
	assert.Equal(t, fixture.Session.ID(), updated.ID())
	revokedAt, isRevoked := updated.RevokedAt()
	require.True(t, isRevoked)
	assert.Equal(t, helper.FakeClock.Now(), revokedAt)
}

func TestRefresh_DoesNotRotate_WhenTokenIsAlreadyUsed(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.SeedUsedToken()

	_, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	require.ErrorIs(t, err, refresh.ErrTokenAlreadyUsed)
	helper.AssertNoAccessTokenIssued()
	helper.AssertNoRefreshTokenWrites()
}

func TestRefresh_RevokesSession_WhenTokenIsSpentConcurrently(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.Seed()

	// The writer reports the token as spent, as it does when a concurrent
	// refresh marked it used between the read and the write.
	helper.FakeUnitOfWork.FakeRefreshTokenWriter.SetError(session.ErrTokenAlreadyUsed)

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	assert.Zero(t, out)
	require.ErrorIs(t, err, refresh.ErrTokenAlreadyUsed)
	updated := testutil.Only(t, helper.FakeUnitOfWork.FakeSessionWriter.Updates())
	assert.Equal(t, fixture.Session.ID(), updated.ID())
	revokedAt, isRevoked := updated.RevokedAt()
	require.True(t, isRevoked)
	assert.Equal(t, helper.FakeClock.Now(), revokedAt)
}

func TestRefresh_ReturnsError_WhenRevokingSessionOfTokenAlreadyUsedFails(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.SeedUsedToken()

	expectedErr := errors.New("internal error")
	helper.FakeUnitOfWork.SetError(expectedErr)

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	assert.Zero(t, out)
	assert.ErrorIs(t, err, expectedErr)
	assert.NotErrorIs(t, err, refresh.ErrTokenAlreadyUsed)
}

// UNEXPECTED ERRORS

func TestRefresh_ReturnsError_WhenRefreshTokenResolverFails(t *testing.T) {
	helper := NewTestHelper(t)
	in := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeRefreshTokenResolver.SetError(expectedErr)

	out, err := helper.UseCase().Execute(context.Background(), in)

	assert.Zero(t, out)
	assert.ErrorIs(t, err, expectedErr)
	AssertUnexpectedError(t, err)
	helper.AssertNoAccessTokenIssued()
	helper.AssertNoWrites()
}

func TestRefresh_ReturnsError_WhenSessionReaderFails(t *testing.T) {
	helper := NewTestHelper(t)
	in := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeSessionReader.SetError(expectedErr)

	out, err := helper.UseCase().Execute(context.Background(), in)

	assert.Zero(t, out)
	assert.ErrorIs(t, err, expectedErr)
	AssertUnexpectedError(t, err)
	helper.AssertNoAccessTokenIssued()
	helper.AssertNoWrites()
}

func TestRefresh_ReturnsError_WhenSessionOfTokenIsNotFound(t *testing.T) {
	helper := NewTestHelper(t)
	fixture := helper.SeedTokenWithoutSession()

	out, err := helper.UseCase().Execute(context.Background(), fixture.Input())

	assert.Zero(t, out)
	AssertUnexpectedError(t, err)
	helper.AssertNoAccessTokenIssued()
	helper.AssertNoWrites()
}

func TestRefresh_ReturnsError_WhenAccessTokenIssuerFails(t *testing.T) {
	helper := NewTestHelper(t)
	in := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeAccessTokenIssuer.SetError(expectedErr)

	out, err := helper.UseCase().Execute(context.Background(), in)

	assert.Zero(t, out)
	assert.ErrorIs(t, err, expectedErr)
	AssertUnexpectedError(t, err)
	helper.AssertNoWrites()
}

func TestRefresh_ReturnsError_WhenRefreshTokenGeneratorFails(t *testing.T) {
	helper := NewTestHelper(t)
	in := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeRefreshTokenGenerator.SetError(expectedErr)

	out, err := helper.UseCase().Execute(context.Background(), in)

	assert.Zero(t, out)
	assert.ErrorIs(t, err, expectedErr)
	AssertUnexpectedError(t, err)
	helper.AssertNoWrites()
}

func TestRefresh_ReturnsError_WhenUnitOfWorkFails(t *testing.T) {
	helper := NewTestHelper(t)
	in := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeUnitOfWork.SetError(expectedErr)

	out, err := helper.UseCase().Execute(context.Background(), in)

	assert.Zero(t, out)
	assert.ErrorIs(t, err, expectedErr)
	AssertUnexpectedError(t, err)
}

func TestRefresh_ReturnsError_WhenRefreshTokenWriterFails(t *testing.T) {
	helper := NewTestHelper(t)
	in := helper.ValidInput()

	expectedErr := errors.New("internal error")
	helper.FakeUnitOfWork.FakeRefreshTokenWriter.SetError(expectedErr)

	out, err := helper.UseCase().Execute(context.Background(), in)

	assert.Zero(t, out)
	assert.ErrorIs(t, err, expectedErr)
	AssertUnexpectedError(t, err)
	assert.Empty(t, helper.FakeUnitOfWork.FakeSessionWriter.Updates(), "session updated")
}
