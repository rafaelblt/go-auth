package e2e

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type VerifyRequestBody struct {
	AccessToken string `json:"access_token"`
}

type VerifyResponseBody struct {
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	ExpiresIn int64     `json:"expires_in"`
}

const VerifyPath = "/v1/auth/verify"

// forgedAccessToken is well formed and unexpired, and names kid, but is signed
// with a key go-auth does not hold.
func forgedAccessToken(t *testing.T, kid string) string {
	t.Helper()
	_, private, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.RegisteredClaims{
		Subject:   user.NewID().String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
	token.Header["kid"] = kid
	raw, err := token.SignedString(private)
	require.NoError(t, err)
	return raw
}

func withHeader(token, header string) string {
	parts := strings.Split(token, ".")
	parts[0] = base64.RawURLEncoding.EncodeToString([]byte(header))
	return strings.Join(parts, ".")
}

func TestVerify_ReturnsUserOfAccessToken(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	loginResp := env.Client.Post(t, LoginPath, LoginRequestBody{
		Username: usr.Username().String(),
		Password: pwd.Value(),
	})
	require.Equal(t, http.StatusOK, loginResp.StatusCode)
	accessToken := DecodeBody[LoginResponseBody](t, loginResp).AccessToken

	resp := env.Client.Post(t, VerifyPath, VerifyRequestBody{AccessToken: accessToken.Value})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	respBody := DecodeBody[VerifyResponseBody](t, resp)
	assert.Equal(t, usr.ID().String(), respBody.UserID)
	assert.Equal(t, accessToken.ExpiresAt, respBody.ExpiresAt)
	assert.LessOrEqual(t, respBody.ExpiresIn, accessToken.ExpiresIn)
	assert.GreaterOrEqual(t, respBody.ExpiresIn, int64((AccessTokenTTL-5*time.Second)/time.Second))
}

func TestVerify_ReturnsMethodNotAllowed(t *testing.T) {
	env := testApp.NewEnv(t)

	resp := env.Client.Get(t, VerifyPath)

	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	assert.Equal(t, "POST", resp.Header.Get("Allow"))
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, MethodNotAllowedCode, respBody.Error.Code)
}

func TestVerify_ReturnsInvalidTokenResponse_WhenTokenIsNotValid(t *testing.T) {
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
			resp := env.Client.Post(t, VerifyPath, VerifyRequestBody{AccessToken: tC.token})

			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			respBody := DecodeBody[ErrorResponseBody](t, resp)
			assert.Equal(t, InvalidTokenCode, respBody.Error.Code)
			assert.NotZero(t, respBody.Error.Message)
		})
	}
}
