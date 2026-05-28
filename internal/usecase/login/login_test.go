package login_test

import (
	"context"
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/credential"
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

	output, err := helper.UseCase().Execute(
		context.Background(), helper.ValidInput(),
	)

	require.Error(t, err)
	require.Zero(t, output)
	assert.ErrorIs(t, err, login.ErrInvalidCredentials)
}

func TestLogin_ReturnsInvalidCredentials_WhenPasswordIsIncorrect(t *testing.T) {
	helper := NewTestHelper(t)
	usr, pwd := helper.GetUserAndPassword()

	output, err := helper.UseCase().Execute(context.Background(), login.Input{
		Username: usr.Username().String(),
		Password: pwd.Value() + "abc",
	})

	require.Error(t, err)
	require.Zero(t, output)
	assert.ErrorIs(t, err, login.ErrInvalidCredentials)
}

func TestLogin_ShouldUseAccessTokenIssuer_AndReturnAccessToken(t *testing.T) {
	helper := NewTestHelper(t)
	usr, pwd := helper.GetUserAndPassword()

	output, err := helper.UseCase().Execute(context.Background(), login.Input{
		Username: usr.Username().String(),
		Password: pwd.Value(),
	})

	require.NoError(t, err)
	payload := helper.FakeAccessTokenIssuer.LastPayload()
	issued := helper.FakeAccessTokenIssuer.LastIssued()
	assert.Equal(t, usr.ID(), payload.UserID)
	assert.Equal(t, issued.Token.Value(), output.AccessToken.Value)
	assert.Equal(t, issued.ExpiresAt, output.AccessToken.ExpiresAt)
}

func TestLogin_ShouldUseRefreshTokenIssuer_AndReturnRefreshToken(t *testing.T) {
	helper := NewTestHelper(t)
	usr, pwd := helper.GetUserAndPassword()

	output, err := helper.UseCase().Execute(context.Background(), login.Input{
		Username: usr.Username().String(),
		Password: pwd.Value(),
	})

	require.NoError(t, err)
	payload := helper.FakeRefreshTokenIssuer.LastPayload()
	issued := helper.FakeRefreshTokenIssuer.LastIssued()
	assert.Equal(t, usr.ID(), payload.UserID)
	assert.Equal(t, issued.RawValue, output.RefreshToken.Value)
	assert.Equal(t, issued.Token.ExpiresAt(), output.RefreshToken.ExpiresAt)
}
