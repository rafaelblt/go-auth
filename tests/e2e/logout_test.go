package e2e

import (
	"io"
	"net/http"
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type LogoutRequestBody struct {
	RefreshToken string `json:"refresh_token"`
}

const LogoutPath = "/v1/auth/logout"

func loginTokens(t *testing.T, env TestEnv, username, password string) LoginResponseBody {
	t.Helper()
	resp := env.Client.Post(t, LoginPath, LoginRequestBody{Username: username, Password: password})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	return DecodeBody[LoginResponseBody](t, resp)
}

func requireNoContent(t *testing.T, resp *http.Response) {
	t.Helper()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Empty(t, body)
	assert.Empty(t, resp.Header.Get("Content-Type"))
}

func requireRefreshStatus(t *testing.T, env TestEnv, refreshToken string, status int) {
	t.Helper()
	resp := env.Client.Post(t, RefreshPath, RefreshRequestBody{RefreshToken: refreshToken})
	require.Equal(t, status, resp.StatusCode)
}

func TestLogout_EndsTheSessionOfTheToken(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	tokens := loginTokens(t, env, usr.Username().String(), pwd.Value())

	resp := env.Client.Post(t, LogoutPath, LogoutRequestBody{RefreshToken: tokens.RefreshToken.Value})

	requireNoContent(t, resp)
	requireRefreshStatus(t, env, tokens.RefreshToken.Value, http.StatusUnauthorized)
}

func TestLogout_KeepsTheOtherSessionsOfTheUser(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	ended := loginTokens(t, env, usr.Username().String(), pwd.Value())
	kept := loginTokens(t, env, usr.Username().String(), pwd.Value())

	resp := env.Client.Post(t, LogoutPath, LogoutRequestBody{RefreshToken: ended.RefreshToken.Value})

	requireNoContent(t, resp)
	requireRefreshStatus(t, env, kept.RefreshToken.Value, http.StatusOK)
}

// An attacker who refreshes first leaves the client holding a spent token, and
// the client's logout with it has to end the attacker's chain too.
func TestLogout_EndsTheSession_WhenTokenIsAlreadyUsed(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	spent := loginTokens(t, env, usr.Username().String(), pwd.Value()).RefreshToken.Value
	resp := env.Client.Post(t, RefreshPath, RefreshRequestBody{RefreshToken: spent})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	successor := DecodeBody[RefreshResponseBody](t, resp).RefreshToken.Value

	resp = env.Client.Post(t, LogoutPath, LogoutRequestBody{RefreshToken: spent})

	requireNoContent(t, resp)
	requireRefreshStatus(t, env, successor, http.StatusUnauthorized)
}

func TestLogout_EndsTheSession_WhenTokenIsExpired(t *testing.T) {
	env := testApp.NewEnv(t)
	token, raw := env.Fixtures.CreateRefreshTokenExpired(t)

	resp := env.Client.Post(t, LogoutPath, LogoutRequestBody{RefreshToken: raw})

	requireNoContent(t, resp)
	env.Asserts.RequireSessionIsRevoked(t, token.SessionID())
}

func TestLogout_ReturnsNoContent_WhenSessionIsAlreadyRevoked(t *testing.T) {
	env := testApp.NewEnv(t)
	token, raw := env.Fixtures.CreateRefreshTokenWithSessionRevoked(t)

	resp := env.Client.Post(t, LogoutPath, LogoutRequestBody{RefreshToken: raw})

	requireNoContent(t, resp)
	env.Asserts.RequireSessionIsRevoked(t, token.SessionID())
}

// A client whose answer was lost sends the logout again, and gets the same one.
func TestLogout_ReturnsNoContent_WhenRepeated(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	tokens := loginTokens(t, env, usr.Username().String(), pwd.Value())
	reqBody := LogoutRequestBody{RefreshToken: tokens.RefreshToken.Value}

	requireNoContent(t, env.Client.Post(t, LogoutPath, reqBody))
	requireNoContent(t, env.Client.Post(t, LogoutPath, reqBody))
}

// Access tokens cannot be revoked: see docs/access-tokens.md#access-tokens-cannot-be-revoked.
func TestLogout_LeavesTheAccessTokenValid(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	tokens := loginTokens(t, env, usr.Username().String(), pwd.Value())

	resp := env.Client.Post(t, LogoutPath, LogoutRequestBody{RefreshToken: tokens.RefreshToken.Value})
	requireNoContent(t, resp)

	resp = env.Client.Post(t, VerifyPath, VerifyRequestBody{AccessToken: tokens.AccessToken.Value})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, usr.ID().String(), DecodeBody[VerifyResponseBody](t, resp).UserID)
}

func TestLogout_ReturnsInvalidTokenResponse_WhenTokenIsNotKnown(t *testing.T) {
	env := testApp.NewEnv(t)
	testCases := []struct {
		desc string
		body any
	}{
		{desc: "missing", body: LogoutRequestBody{}},
		{desc: "under another field name", body: map[string]string{"refreshToken": sessiontest.NewRefreshTokenSecret(t).Value()}},
		{desc: "malformed", body: LogoutRequestBody{RefreshToken: "invalid"}},
		{desc: "never issued", body: LogoutRequestBody{RefreshToken: sessiontest.NewRefreshTokenSecret(t).Value()}},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			resp := env.Client.Post(t, LogoutPath, tC.body)

			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			respBody := DecodeBody[ErrorResponseBody](t, resp)
			assert.Equal(t, InvalidTokenCode, respBody.Error.Code)
			assert.NotZero(t, respBody.Error.Message)
		})
	}
}

func TestLogout_ReturnsMethodNotAllowed(t *testing.T) {
	env := testApp.NewEnv(t)

	resp := env.Client.Get(t, LogoutPath)

	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	assert.Equal(t, "POST", resp.Header.Get("Allow"))
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, MethodNotAllowedCode, respBody.Error.Code)
}
