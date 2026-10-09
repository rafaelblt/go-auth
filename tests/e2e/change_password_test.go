package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ChangePasswordRequestBody struct {
	Username        string `json:"username"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type ChangePasswordResponseBody struct {
	User User `json:"user"`
}

const ChangePasswordPath = "/v1/auth/change-password"

const newPassword = "0k-nEw_pàsswörd"

func requireLoginStatus(t *testing.T, env TestEnv, username, password string, status int) {
	t.Helper()
	resp := env.Client.Post(t, LoginPath, LoginRequestBody{Username: username, Password: password})
	require.Equal(t, status, resp.StatusCode)
}

func TestChangePassword_ReplacesThePassword(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)

	resp := env.Client.Post(t, ChangePasswordPath, ChangePasswordRequestBody{
		Username:        usr.Username().String(),
		CurrentPassword: pwd.Value(),
		NewPassword:     newPassword,
	})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	respBody := DecodeBody[ChangePasswordResponseBody](t, resp)
	assert.Equal(t, usr.ID().String(), respBody.User.ID)
	assert.Equal(t, usr.Username().String(), respBody.User.Username)
	requireLoginStatus(t, env, usr.Username().String(), pwd.Value(), http.StatusUnauthorized)
	requireLoginStatus(t, env, usr.Username().String(), newPassword, http.StatusOK)
}

func TestChangePassword_RevokesEverySessionOfTheUser(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	login := LoginRequestBody{Username: usr.Username().String(), Password: pwd.Value()}
	refreshTokens := []string{}
	for range 2 {
		resp := env.Client.Post(t, LoginPath, login)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		refreshTokens = append(refreshTokens, DecodeBody[LoginResponseBody](t, resp).RefreshToken.Value)
	}

	resp := env.Client.Post(t, ChangePasswordPath, ChangePasswordRequestBody{
		Username:        usr.Username().String(),
		CurrentPassword: pwd.Value(),
		NewPassword:     newPassword,
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	for _, token := range refreshTokens {
		resp := env.Client.Post(t, RefreshPath, RefreshRequestBody{RefreshToken: token})
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, InvalidTokenCode, DecodeBody[ErrorResponseBody](t, resp).Error.Code)
	}
}

func TestChangePassword_ReturnsInvalidCredentials_AndKeepsThePassword(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	testCases := []struct {
		desc string
		body ChangePasswordRequestBody
	}{
		{
			desc: "unknown username",
			body: ChangePasswordRequestBody{Username: "rafaelblt", CurrentPassword: pwd.Value(), NewPassword: newPassword},
		},
		{
			desc: "incorrect current password",
			body: ChangePasswordRequestBody{Username: usr.Username().String(), CurrentPassword: pwd.Value() + "X", NewPassword: newPassword},
		},
		{
			desc: "malformed username",
			body: ChangePasswordRequestBody{Username: "R", CurrentPassword: pwd.Value(), NewPassword: newPassword},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			resp := env.Client.Post(t, ChangePasswordPath, tC.body)

			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			respBody := DecodeBody[ErrorResponseBody](t, resp)
			assert.Equal(t, InvalidCredentialsCode, respBody.Error.Code)
			assert.NotZero(t, respBody.Error.Message)
		})
	}
	requireLoginStatus(t, env, usr.Username().String(), pwd.Value(), http.StatusOK)
}

func TestChangePassword_ReturnsValidationErrorResponse_WhenNewPasswordIsInvalid(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)

	resp := env.Client.Post(t, ChangePasswordPath, ChangePasswordRequestBody{
		Username:        usr.Username().String(),
		CurrentPassword: pwd.Value(),
		NewPassword:     "1",
	})

	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, ValidationFailedCode, respBody.Error.Code)
	require.Len(t, respBody.Error.Fields, 1)
	assert.Equal(t, "new_password", respBody.Error.Fields[0].Field)
	assert.Equal(t, TooShortCode, respBody.Error.Fields[0].Code)
	requireLoginStatus(t, env, usr.Username().String(), pwd.Value(), http.StatusOK)
}

// Whichever way the two requests interleave, the second to write finds the
// current password it was given already replaced: by the guard on the hash
// when both read the old one, or by the comparison when it reads the new one.
func TestChangePassword_AppliesOneOfTwoConcurrentChanges(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	newPasswords := []string{"first-new-password", "second-new-password"}
	url := "http://" + testApp.cfg.Address() + ChangePasswordPath
	statuses := make([]int, len(newPasswords))
	errs := make([]error, len(newPasswords))

	// The requests are sent with net/http directly: env.Client stops the test
	// on failure, which only the test goroutine may do.
	var wg sync.WaitGroup
	for i, next := range newPasswords {
		body, err := json.Marshal(ChangePasswordRequestBody{
			Username:        usr.Username().String(),
			CurrentPassword: pwd.Value(),
			NewPassword:     next,
		})
		require.NoError(t, err)
		wg.Go(func() {
			resp, err := http.Post(url, "application/json", bytes.NewReader(body))
			if err != nil {
				errs[i] = err
				return
			}
			resp.Body.Close()
			statuses[i] = resp.StatusCode
		})
	}
	wg.Wait()

	require.NoError(t, errors.Join(errs...))
	assert.ElementsMatch(t, []int{http.StatusOK, http.StatusUnauthorized}, statuses)
	winner := newPasswords[0]
	if statuses[1] == http.StatusOK {
		winner = newPasswords[1]
	}
	requireLoginStatus(t, env, usr.Username().String(), winner, http.StatusOK)
}

func TestChangePassword_ReturnsMethodNotAllowed(t *testing.T) {
	env := testApp.NewEnv(t)

	resp := env.Client.Get(t, ChangePasswordPath)

	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	assert.Equal(t, "POST", resp.Header.Get("Allow"))
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, MethodNotAllowedCode, respBody.Error.Code)
}
