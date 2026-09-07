package login_test

import (
	"context"
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/password"
	"github.com/rafaelblt/go-auth/internal/testutil"
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
				Password: strings.Repeat("a", password.PlainMinCodePoints-1),
			},
		},
		{
			desc: "password too long",
			input: login.Input{
				Username: "username",
				Password: strings.Repeat("a", password.PlainMaxBytes+1),
			},
		},
		{
			desc: "username and password too short",
			input: login.Input{
				Username: strings.Repeat("a", user.UsernameMinLen-1),
				Password: strings.Repeat("a", password.PlainMinCodePoints-1),
			},
		},
		{
			desc: "username and password too long",
			input: login.Input{
				Username: strings.Repeat("a", user.UsernameMaxLen+1),
				Password: strings.Repeat("a", password.PlainMaxBytes+1),
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

func TestLogin_IssuesAndReturnsAccessToken(t *testing.T) {
	helper := NewTestHelper(t)
	usr, pwd := helper.GetUserAndPassword()

	output, err := helper.UseCase().Execute(context.Background(), login.Input{
		Username: usr.Username().String(),
		Password: pwd.Value(),
	})
	require.NoError(t, err)

	payload := testutil.Only(t, helper.FakeAccessTokenIssuer.Payloads())
	assert.Equal(t, usr.ID(), payload.UserID)

	issued := testutil.Only(t, helper.FakeAccessTokenIssuer.Issueds())
	assert.Equal(t, issued.Token.Value(), output.AccessToken.Value)
	assert.Equal(t, issued.ExpiresAt, output.AccessToken.ExpiresAt)
}

func TestLogin_GeneratesAndReturnsRefreshToken(t *testing.T) {
	helper := NewTestHelper(t)

	output, err := helper.UseCase().Execute(context.Background(), helper.ValidInput())
	require.NoError(t, err)

	generated := testutil.Only(t, helper.FakeRefreshTokenGenerator.Generated())
	assert.Equal(t, generated.Raw, output.RefreshToken.Value)
}

func TestLogin_ShouldSaveSession(t *testing.T) {
	helper := NewTestHelper(t)

	_, err := helper.UseCase().Execute(context.Background(), helper.ValidInput())
	require.NoError(t, err)

	sess := testutil.Only(t, helper.FakeUnitOfWork.FakeSessionWriter.Adds())
	assert.Equal(t, helper.FakeClock.Now(), sess.CreatedAt())
}

func TestLogin_ShouldSaveRefreshToken(t *testing.T) {
	helper := NewTestHelper(t)

	_, err := helper.UseCase().Execute(context.Background(), helper.ValidInput())
	require.NoError(t, err)

	sess := testutil.Only(t, helper.FakeUnitOfWork.FakeSessionWriter.Adds())
	generated := testutil.Only(t, helper.FakeRefreshTokenGenerator.Generated())
	token := testutil.Only(t, helper.FakeUnitOfWork.FakeRefreshTokenWriter.Adds())

	assert.Equal(t, sess.ID(), token.SessionID())
	assert.Equal(t, generated.Hash, token.Hash())
	assert.False(t, token.HasParent())
	assert.Equal(t, helper.FakeClock.Now(), token.CreatedAt())
	assert.Equal(t, helper.FakeClock.Now().Add(helper.RefreshTokenTTL), token.ExpiresAt())
}
