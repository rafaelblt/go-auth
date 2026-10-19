package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MeRequestBody struct {
	AccessToken string `json:"access_token"`
}

type MeResponseBody struct {
	User User `json:"user"`
}

const MePath = "/v1/auth/me"

func loginAccessToken(t *testing.T, env TestEnv, username, password string) string {
	t.Helper()
	resp := env.Client.Post(t, LoginPath, LoginRequestBody{Username: username, Password: password})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	return DecodeBody[LoginResponseBody](t, resp).AccessToken.Value
}

func TestMe_ReturnsUserOfAccessToken(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	accessToken := loginAccessToken(t, env, usr.Username().String(), pwd.Value())

	resp := env.Client.Post(t, MePath, MeRequestBody{AccessToken: accessToken})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	respBody := DecodeBody[MeResponseBody](t, resp)
	assert.Equal(t, usr.ID().String(), respBody.User.ID)
	assert.Equal(t, usr.Username().String(), respBody.User.Username)
}

func TestMe_ReturnsMethodNotAllowed(t *testing.T) {
	env := testApp.NewEnv(t)

	resp := env.Client.Get(t, MePath)

	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	assert.Equal(t, "POST", resp.Header.Get("Allow"))
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, MethodNotAllowedCode, respBody.Error.Code)
}

func TestMe_ReturnsInvalidTokenResponse_WhenTokenIsNotValid(t *testing.T) {
	env := testApp.NewEnv(t)
	jwks := DecodeBody[JWKSResponseBody](t, env.Client.Get(t, JWKSPath))
	require.NotEmpty(t, jwks.Keys)
	forged := forgedAccessToken(t, jwks.Keys[0].Kid)
	testCases := []struct {
		desc  string
		token string
	}{
		{desc: "missing", token: ""},
		{desc: "malformed", token: "not-a-jwt"},
		{desc: "signed with a key go-auth does not hold", token: forged},
		{desc: "with an unknown alg", token: withHeader(forged, `{"alg":"ES999","typ":"JWT"}`)},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			resp := env.Client.Post(t, MePath, MeRequestBody{AccessToken: tC.token})

			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			respBody := DecodeBody[ErrorResponseBody](t, resp)
			assert.Equal(t, InvalidTokenCode, respBody.Error.Code)
			assert.NotZero(t, respBody.Error.Message)
		})
	}
}

func TestMe_ReturnsInvalidTokenResponse_WhenUserOfTokenIsGone(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	accessToken := loginAccessToken(t, env, usr.Username().String(), pwd.Value())
	env.Fixtures.DeleteUser(t, usr.ID())

	resp := env.Client.Post(t, MePath, MeRequestBody{AccessToken: accessToken})

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, InvalidTokenCode, respBody.Error.Code)
}
