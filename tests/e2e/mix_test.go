package e2e

import (
	"net/http"
	"testing"

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
