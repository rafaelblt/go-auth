package refresh_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefresh_IssuesAccessToken(t *testing.T) {
	helper := NewTestHelper(t)
	sess, raw := helper.GetSessionAndTokenRaw()

	uc := helper.UseCase()
	out, err := uc.Execute(context.Background(), refresh.Input{RefreshToken: raw})

	require.NoError(t, err)

	payload := testutil.Only(t, helper.FakeAccessTokenIssuer.Payloads())
	assert.Equal(t, sess.UserID(), payload.UserID)

	issued := testutil.Only(t, helper.FakeAccessTokenIssuer.Issueds())
	assert.Equal(t, issued.Token.Value(), out.AccessToken.Value)
	assert.Equal(t, issued.ExpiresAt, out.AccessToken.ExpiresAt)
}

func TestRefresh_GeneratesAndReturnsRefreshToken(t *testing.T) {
	helper := NewTestHelper(t)
	uc := helper.UseCase()
	in := helper.ValidInput()

	out, err := uc.Execute(context.Background(), in)

	require.NoError(t, err)
	generated := testutil.Only(t, helper.FakeRefreshTokenGenerator.Generated())
	assert.Equal(t, generated.Raw, out.RefreshToken.Value)
}

func TestRefresh_UpdatesUsedRefreshToken(t *testing.T) {
	helper := NewTestHelper(t)
	token, raw := helper.GetRefreshTokenAndRaw()

	uc := helper.UseCase()
	_, err := uc.Execute(context.Background(), refresh.Input{RefreshToken: raw})

	require.NoError(t, err)
	updated := testutil.Only(t, helper.FakeUnitOfWork.FakeRefreshTokenWriter.Updates())
	usedAt, isUsed := updated.UsedAt()
	assert.True(t, isUsed)
	assert.Equal(t, helper.FakeClock.Now(), usedAt)
	assert.Equal(t, token.ID(), updated.ID())
}

func TestRefresh_AddGeneratedRefreshToken(t *testing.T) {
	helper := NewTestHelper(t)
	token, raw := helper.GetRefreshTokenAndRaw()

	uc := helper.UseCase()
	_, err := uc.Execute(context.Background(), refresh.Input{
		RefreshToken: raw,
	})
	require.NoError(t, err)

	expectedHash := testutil.Only(t, helper.FakeRefreshTokenGenerator.Generated()).Hash
	added := testutil.Only(t, helper.FakeUnitOfWork.FakeRefreshTokenWriter.Adds())

	addedParentID, hasParent := added.ParentID()
	require.True(t, hasParent)
	assert.Equal(t, token.ID(), addedParentID)
	assert.Equal(t, expectedHash, added.Hash())
	assert.Equal(t, token.SessionID(), added.SessionID())
	assert.Equal(t, helper.FakeClock.Now().Add(helper.RefreshTokenTTL), added.ExpiresAt())
}

func TestRefresh_ReturnsTokenAlreadyUsedError(t *testing.T) {
	helper := NewTestHelper(t)
	raw := helper.GetTokenAlreadyUsed()

	uc := helper.UseCase()
	out, err := uc.Execute(context.Background(), refresh.Input{RefreshToken: raw})

	assert.ErrorIs(t, err, refresh.ErrTokenAlreadyUsed)
	assert.Zero(t, out)
}

func TestRefresh_ReturnsTokenExpiredError(t *testing.T) {
	helper := NewTestHelper(t)
	raw := helper.GetTokenExpired()

	uc := helper.UseCase()
	out, err := uc.Execute(context.Background(), refresh.Input{RefreshToken: raw})

	assert.ErrorIs(t, err, refresh.ErrTokenExpired)
	assert.Zero(t, out)
}

func TestRefresh_ReturnsTokenInvalidError(t *testing.T) {
	helper := NewTestHelper(t)
	raw := "INVALID TOKEN"

	uc := helper.UseCase()
	out, err := uc.Execute(context.Background(), refresh.Input{RefreshToken: raw})

	assert.ErrorIs(t, err, refresh.ErrTokenInvalid)
	assert.Zero(t, out)
}

func TestRefresh_ReturnsSessionRevokedError(t *testing.T) {
	helper := NewTestHelper(t)
	raw := helper.GetTokenWithSessionRevoked()

	uc := helper.UseCase()
	out, err := uc.Execute(context.Background(), refresh.Input{RefreshToken: raw})

	assert.ErrorIs(t, err, refresh.ErrSessionRevoked)
	assert.Zero(t, out)
}

func TestRefresh_RevokesSessionOfTokenAlreadyUsed(t *testing.T) {
	helper := NewTestHelper(t)
	raw := helper.GetTokenAlreadyUsed()

	uc := helper.UseCase()
	_, err := uc.Execute(context.Background(), refresh.Input{RefreshToken: raw})

	require.ErrorIs(t, err, refresh.ErrTokenAlreadyUsed)
	updated := testutil.Only(t, helper.FakeUnitOfWork.FakeSessionWriter.Updates())
	revokedAt, isRevoked := updated.RevokedAt()
	require.True(t, isRevoked)
	assert.Equal(t, helper.FakeClock.Now(), revokedAt)
}
