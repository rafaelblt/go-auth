package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil/apitest"
	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type LoginTestHelper struct {
	t         *testing.T
	FakeLogin *apitest.FakeLogin
}

func NewLoginTestHelper(t *testing.T) LoginTestHelper {
	t.Helper()
	fakeLogin := apitest.NewFakeLogin()
	return LoginTestHelper{t, fakeLogin}
}

func (h LoginTestHelper) Handler() loginHandler {
	return newLoginHandler(h.FakeLogin)
}

func (h LoginTestHelper) NewRequest(body string) *http.Request {
	h.t.Helper()
	reader := strings.NewReader(body)
	req, err := http.NewRequest(http.MethodPost, "url", reader)
	require.NoError(h.t, err)
	return req
}

func (h LoginTestHelper) NewRequestWithFields(username, password string) *http.Request {
	body := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
	return h.NewRequest(body)
}

func TestLogin_ReturnsLoginResponse(t *testing.T) {
	helper := NewLoginTestHelper(t)
	req := helper.NewRequestWithFields("rblt", "123")

	resp := helper.Handler().Handle(req)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.IsType(t, loginResponseBody{}, resp.Body)

	body := resp.Body.(loginResponseBody)
	expectedOutput := helper.FakeLogin.DefaultOutput()
	assert.Equal(t, expectedOutput.AccessToken.Value, body.AccessToken.Value)
	assert.Equal(t, expectedOutput.RefreshToken.Value, body.RefreshToken.Value)
}

func TestLogin_TranslateUseCaseError(t *testing.T) {
	helper := NewLoginTestHelper(t)

	err := login.ErrInvalidCredentials
	helper.FakeLogin.SetError(err)

	req := helper.NewRequestWithFields("rblt", "123")
	resp := helper.Handler().Handle(req)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.IsType(t, errorBody{}, resp.Body)
	body := resp.Body.(errorBody)
	assert.Equal(t, err.Code(), body.Error.Code)
}

func TestLogin_ReturnsInvalidJSONBody(t *testing.T) {
	testCases := []struct {
		desc string
		body string
	}{
		{
			desc: "with only start bracket",
			body: "{",
		},
		{
			desc: "with only end bracket",
			body: "}",
		},
		{
			desc: "with random chars",
			body: "21j89kf dsag-ĺ1#fdsh",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper := NewLoginTestHelper(t)
			req := helper.NewRequest(tC.body)
			resp := helper.Handler().Handle(req)

			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
			require.IsType(t, errorBody{}, resp.Body)
			body := resp.Body.(errorBody)
			assert.Equal(t, invalidJSONBodyErrorBody, body)
		})
	}
}
