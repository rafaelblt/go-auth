package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/testutil/domaintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type RegisterTestHelper struct {
	t *testing.T
}

func NewRegisterTestHelper(t *testing.T) RegisterTestHelper {
	t.Helper()
	return RegisterTestHelper{t}
}

func (helper RegisterTestHelper) NewBody(username, password string) string {
	return fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
}

func (helper RegisterTestHelper) URL() string {
	return testServer.URL + "/auth/register"
}

func (helper RegisterTestHelper) SendRequest(body string) *http.Response {
	helper.t.Helper()
	reader := strings.NewReader(body)
	response, err := http.Post(helper.URL(), "application/json", reader)
	require.NoError(helper.t, err)
	return response
}

func (helper RegisterTestHelper) SaveUser(t *testing.T, user *domain.User) {
	t.Helper()
	repo, err := infra.NewUserRepo(testDB.Pool())
	require.NoError(t, err)
	require.NoError(t, repo.Save(context.Background(), user))
}

func DecodeResponseBody[T any](t *testing.T, response *http.Response) T {
	defer response.Body.Close()

	var decoded T
	require.NoError(t, json.NewDecoder(response.Body).Decode(&decoded))

	return decoded
}

func TestRegister_ReturnsSuccessResponse(t *testing.T) {
	helper := NewRegisterTestHelper(t)
	body := `{"username":"maria","password":"12345678"}`

	response := helper.SendRequest(body)

	decoded := DecodeResponseBody[registerResponseBody](t, response)
	assert.NotZero(t, decoded.User.ID)
	assert.Equal(t, "maria", decoded.User.Username)
	assert.Equal(t, "active", decoded.User.Status)
	assert.NotZero(t, decoded.User.CreatedAt)
	assert.NotZero(t, decoded.User.UpdatedAt)
}

func TestRegister_ReturnsValidationError(t *testing.T) {
	helper := NewRegisterTestHelper(t)
	testCases := []struct {
		desc     string
		body     string
		expected map[string]ValidationErrors
	}{
		{
			desc: "username too short",
			body: helper.NewBody(
				strings.Repeat("a", domain.UsernameMinLen-1),
				"12345678",
			),
			expected: map[string]ValidationErrors{
				"username": {registerUsernameTooShortError},
			},
		},
		{
			desc: "username too long",
			body: helper.NewBody(
				strings.Repeat("a", domain.UsernameMaxLen+1),
				"12345678",
			),
			expected: map[string]ValidationErrors{
				"username": {registerUsernameTooLongError},
			},
		},
		{
			desc: "password too short",
			body: helper.NewBody(
				"rafael",
				strings.Repeat("a", domain.PlainPasswordMinLen-1),
			),
			expected: map[string]ValidationErrors{
				"password": {registerPasswordTooShortError},
			},
		},
		{
			desc: "password too long",
			body: helper.NewBody(
				"rafael",
				strings.Repeat("a", domain.PlainPasswordMaxLen+1),
			),
			expected: map[string]ValidationErrors{
				"password": {registerPasswordTooLongError},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			response := helper.SendRequest(tC.body)
			decoded := DecodeResponseBody[validationErrorBody](t, response)
			assert.Equal(t, http.StatusUnprocessableEntity, response.StatusCode)
			assert.Equal(t, tC.expected, decoded.Errors)
		})
	}
}

func TestRegister_ReturnsUsernameAlreadyExists(t *testing.T) {
	helper := NewRegisterTestHelper(t)
	user := domaintest.DefaultUser(t)
	helper.SaveUser(t, user)

	body := helper.NewBody(user.Username().String(), "12345678")
	response := helper.SendRequest(body)

	decoded := DecodeResponseBody[errorBody](t, response)
	assert.Equal(t, http.StatusConflict, response.StatusCode)
	assert.Equal(t, registerUsernameAlreadyExistsError, decoded)
}

func TestRegister_ReturnsInvalidJSONBody(t *testing.T) {
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
			helper := NewRegisterTestHelper(t)
			response := helper.SendRequest(tC.body)
			decoded := DecodeResponseBody[errorBody](t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
			assert.Equal(t, invalidJSONBodyErrorBody, decoded)
		})
	}
}
