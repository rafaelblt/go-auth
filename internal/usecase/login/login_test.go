package login_test

import (
	"context"
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin_ReturnsOutput_WhenInputIsValid(t *testing.T) {
	helper := NewTestHelper(t)
	usr, pwd := helper.GetUserAndPassword()

	output, err := helper.UseCase().Execute(context.Background(), login.Input{
		Username: usr.Username().String(),
		Password: pwd.Value(),
	})

	require.NoError(t, err)
	require.NotZero(t, output)
	assert.NotZero(t, output.AccessToken)
	assert.NotZero(t, output.RefreshToken)
}

func TestLogin_ReturnsInvalidCredentials_WhenInputIsInvalid(t *testing.T) {
	testCases := []struct {
		desc  string
		input login.Input
	}{
		{
			desc: "username too short",
			input: login.Input{
				Username: strings.Repeat("a", user.UsernameMinLen-1),
				Password: "12345678",
			},
		},
		{
			desc: "username too long",
			input: login.Input{
				Username: strings.Repeat("a", user.UsernameMaxLen+1),
				Password: "12345678",
			},
		},
		{
			desc: "password too short",
			input: login.Input{
				Username: "username",
				Password: strings.Repeat("a", credential.PlainPasswordMinLen-1),
			},
		},
		{
			desc: "password too long",
			input: login.Input{
				Username: "username",
				Password: strings.Repeat("a", credential.PlainPasswordMaxLen+1),
			},
		},
		{
			desc: "username and password too short",
			input: login.Input{
				Username: strings.Repeat("a", user.UsernameMinLen-1),
				Password: strings.Repeat("a", credential.PlainPasswordMinLen-1),
			},
		},
		{
			desc: "username and password too long",
			input: login.Input{
				Username: strings.Repeat("a", user.UsernameMaxLen+1),
				Password: strings.Repeat("a", credential.PlainPasswordMaxLen+1),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper := NewTestHelper(t)
			output, err := helper.UseCase().Execute(context.Background(), tC.input)
			assert.Zero(t, output)
			require.Error(t, err)
			assert.ErrorIs(t, err, login.ErrInvalidCredentials)
		})
	}
}

func TestLogin_ReturnsInvalidCredentials_WhenUsernameNotExists(t *testing.T) {
	helper := NewTestHelper(t)

	output, err := helper.UseCase().Execute(context.Background(), login.Input{
		Username: usertest.MustUsername(t, "username").String(),
		Password: "210-9i)S_D(Akfvfc12)",
	})

	require.Error(t, err)
	require.Zero(t, output)
	assert.ErrorIs(t, err, login.ErrInvalidCredentials)
}

func TestLogin_ReturnsInvalidCredentials_WhenPasswordIsIncorrect(t *testing.T) {
	helper := NewTestHelper(t)
	usr, pwd := helper.GetUserAndPassword()

	output, err := helper.UseCase().Execute(context.Background(), login.Input{
		Username: usr.Username().String(),
		Password: pwd.Value() + "INCORRECT",
	})

	require.Error(t, err)
	require.Zero(t, output)
	assert.ErrorIs(t, err, login.ErrInvalidCredentials)
}

func TestLogin_ShouldIssueAccessToken_AndReturnToken(t *testing.T) {
	helper := NewTestHelper(t)
	usr, pwd := helper.GetUserAndPassword()

	output, err := helper.UseCase().Execute(context.Background(), login.Input{
		Username: usr.Username().String(),
		Password: pwd.Value(),
	})
	require.NoError(t, err)

	payloads := helper.FakeAccessTokenIssuer.Payloads()
	require.Len(t, payloads, 1)
	payload := payloads[0]
	issueds := helper.FakeAccessTokenIssuer.Issueds()
	require.Len(t, issueds, 1)
	issued := issueds[0]

	assert.Equal(t, usr.ID(), payload.UserID)
	assert.Equal(t, issued.Token.Value(), output.AccessToken.Value)
	assert.Equal(t, issued.ExpiresAt, output.AccessToken.ExpiresAt)
}

func TestLogin_ShouldGenerateRefreshToken_AndReturnToken(t *testing.T) {
	helper := NewTestHelper(t)

	output, err := helper.UseCase().Execute(context.Background(), helper.ValidInput())
	require.NoError(t, err)

	generated := helper.FakeRefreshTokenGenerator.Generated()
	assert.Len(t, generated, 1)
	assert.Equal(t, generated[0].Raw, output.RefreshToken.Value)
}

func TestLogin_ShouldSaveSession(t *testing.T) {
	helper := NewTestHelper(t)

	_, err := helper.UseCase().Execute(context.Background(), helper.ValidInput())
	require.NoError(t, err)

	saved := helper.FakeUnitOfWork.FakeSessionWriter.Data()
	assert.Len(t, saved, 1)
	session := saved[0]

	assert.Equal(t, helper.FakeClock.Now(), session.IssuedAt())
}

func TestLogin_ShouldSaveRefreshToken(t *testing.T) {
	helper := NewTestHelper(t)

	_, err := helper.UseCase().Execute(context.Background(), helper.ValidInput())
	require.NoError(t, err)

	session := helper.FakeUnitOfWork.FakeSessionWriter.Data()[0]
	expectedHash := helper.FakeRefreshTokenGenerator.Generated()[0].Hash
	saved := helper.FakeUnitOfWork.FakeRefreshTokenWriter.Adds()
	assert.Len(t, saved, 1)
	token := saved[0]

	assert.Equal(t, session.ID(), token.SessionID())
	assert.Equal(t, expectedHash, token.Hash())
	assert.False(t, token.HasParent())
	assert.Equal(t, helper.FakeClock.Now(), token.IssuedAt())
	assert.Equal(t, helper.FakeClock.Now().Add(helper.RefreshTokenTTL), token.ExpiresAt())
}
