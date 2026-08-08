package e2e

import (
	"crypto/ed25519"
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_Login_Refresh(t *testing.T) {
	env := testApp.NewEnv(t)

	// Register

	username := "swimmer"
	password := "water404500"
	registerReq := RegisterRequestBody{
		Username: username,
		Password: password,
	}
	registerResp := env.Client.Post(t, RegisterPath, registerReq)
	require.Equal(t, http.StatusOK, registerResp.StatusCode)

	// Login

	loginReq := LoginRequestBody{
		Username: username,
		Password: password,
	}
	loginResp := env.Client.Post(t, LoginPath, loginReq)
	require.Equal(t, http.StatusOK, loginResp.StatusCode)
	tokens := DecodeBody[LoginResponseBody](t, loginResp)

	// Refresh

	refreshReq := RefreshRequestBody{
		RefreshToken: tokens.RefreshToken.Value,
	}
	refreshResp := env.Client.Post(t, RefreshPath, refreshReq)
	require.Equal(t, http.StatusOK, refreshResp.StatusCode)
}

func TestAccessTokenFromLogin_WithJWK(t *testing.T) {
	env := testApp.NewEnv(t)

	// Login

	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	reqBody := LoginRequestBody{
		Username: usr.Username().String(),
		Password: pwd.Value(),
	}

	resp := env.Client.Post(t, LoginPath, reqBody)

	loginBody := DecodeBody[LoginResponseBody](t, resp)
	accessToken := loginBody.AccessToken.Value

	// JWKS

	resp = env.Client.Get(t, JWKSPath)
	jwksBody := DecodeBody[JWKSResponseBody](t, resp)
	publicKey, err := base64.RawURLEncoding.DecodeString(jwksBody.Keys[0].X)
	require.NoError(t, err, "raw url decode key failed")

	// Parse & Assert

	token, err := jwt.ParseWithClaims(
		accessToken,
		&jwt.RegisteredClaims{},
		func(t *jwt.Token) (any, error) {
			return  ed25519.PublicKey(publicKey), nil
		},
	)
	require.NoError(t, err)
	claims := token.Claims.(*jwt.RegisteredClaims)
	assert.Equal(t, usr.ID().String(), claims.Subject)
}
