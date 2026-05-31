package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/testutil/apitest"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/rafaelblt/go-auth/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type RegisterTestHelper struct {
	t            *testing.T
	FakeRegister *apitest.FakeRegister
}

func NewRegisterTestHelper(t *testing.T) RegisterTestHelper {
	t.Helper()
	fakeRegister := apitest.NewFakeRegister()
	return RegisterTestHelper{t, fakeRegister}
}

func (h RegisterTestHelper) Handler() registerHandler {
	return newRegisterHandler(h.FakeRegister)
}

func (h RegisterTestHelper) NewRequest(body string) *http.Request {
	h.t.Helper()
	reader := strings.NewReader(body)
	req, err := http.NewRequest(http.MethodPost, "url", reader)
	require.NoError(h.t, err)
	return req
}

func (h RegisterTestHelper) NewRequestWithFields(username, password string) *http.Request {
	body := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
	return h.NewRequest(body)
}

func TestRegister_ReturnsSuccessResponse(t *testing.T) {
	helper := NewRegisterTestHelper(t)
	handler := helper.Handler()
	req := helper.NewRequestWithFields("maria", "12345678")

	resp := handler.Handle(req)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.IsType(t, registerResponseBody{}, resp.Body)
	body := resp.Body.(registerResponseBody)
	assert.NotZero(t, body.User.ID)
	assert.Equal(t, "maria", body.User.Username)
	assert.Equal(t, "active", body.User.Status)
	assert.NotZero(t, body.User.CreatedAt)
	assert.NotZero(t, body.User.UpdatedAt)
}

func TestRegister_ReturnsValidationError(t *testing.T) {
	helper := NewRegisterTestHelper(t)
	handler := helper.Handler()
	testCases := []struct {
		desc     string
		err      *validation.FieldValidationError
		expected validationErrorBody
	}{
		{
			desc: "username too short",
			err: validation.NewFieldValidationError([]validation.FieldError{
				validation.NewFieldError(register.UsernameField, user.ErrUsernameTooShort),
			}),
			expected: validationErrorBody{map[string]ValidationErrors{
				"username": {registerUsernameTooShortError},
			}},
		},
		{
			desc: "username too long",
			err: validation.NewFieldValidationError([]validation.FieldError{
				validation.NewFieldError(register.UsernameField, user.ErrUsernameTooLong),
			}),
			expected: validationErrorBody{map[string]ValidationErrors{
				"username": {registerUsernameTooLongError},
			}},
		},
		{
			desc: "password too short",
			err: validation.NewFieldValidationError([]validation.FieldError{
				validation.NewFieldError(register.PasswordField, credential.ErrPlainPasswordTooShort),
			}),
			expected: validationErrorBody{map[string]ValidationErrors{
				"password": {registerPasswordTooShortError},
			}},
		},
		{
			desc: "password too long",
			err: validation.NewFieldValidationError([]validation.FieldError{
				validation.NewFieldError(register.PasswordField, credential.ErrPlainPasswordTooLong),
			}),
			expected: validationErrorBody{map[string]ValidationErrors{
				"password": {registerPasswordTooLongError},
			}},
		},
		{
			desc: "username and password too long",
			err: validation.NewFieldValidationError([]validation.FieldError{
				validation.NewFieldError(register.PasswordField, credential.ErrPlainPasswordTooLong),
				validation.NewFieldError(register.UsernameField, user.ErrUsernameTooLong),
			}),
			expected: validationErrorBody{map[string]ValidationErrors{
				"username": {registerUsernameTooLongError},
				"password": {registerPasswordTooLongError},
			}},
		},
		{
			desc: "username and password too short",
			err: validation.NewFieldValidationError([]validation.FieldError{
				validation.NewFieldError(register.PasswordField, credential.ErrPlainPasswordTooShort),
				validation.NewFieldError(register.UsernameField, user.ErrUsernameTooShort),
			}),
			expected: validationErrorBody{map[string]ValidationErrors{
				"username": {registerUsernameTooShortError},
				"password": {registerPasswordTooShortError},
			}},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper.FakeRegister.SetFieldValidationError(tC.err)
			req := helper.NewRequestWithFields("username", "password")
			resp := handler.Handle(req)

			require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			require.IsType(t, validationErrorBody{}, resp.Body)
			body := resp.Body.(validationErrorBody)
			assert.Equal(t, tC.expected, body)
		})
	}
}

func TestRegister_ReturnsUsernameAlreadyExists(t *testing.T) {
	helper := NewRegisterTestHelper(t)
	handler := helper.Handler()
	helper.FakeRegister.SetUsernameAlreadyExistsError()
	req := helper.NewRequestWithFields("rafael", "12345678")

	resp := handler.Handle(req)

	require.Equal(t, http.StatusConflict, resp.StatusCode)
	require.IsType(t, errorBody{}, resp.Body)
	body := resp.Body.(errorBody)
	assert.Equal(t, registerUsernameAlreadyExistsError, body)
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
			req := helper.NewRequest(tC.body)
			resp := helper.Handler().Handle(req)

			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
			require.IsType(t, errorBody{}, resp.Body)
			body := resp.Body.(errorBody)
			assert.Equal(t, invalidJSONBodyErrorBody, body)
		})
	}
}

func TestRegister_ReturnsInternalServerError(t *testing.T) {
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
			req := helper.NewRequest(tC.body)
			resp := helper.Handler().Handle(req)

			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
			require.IsType(t, errorBody{}, resp.Body)
			body := resp.Body.(errorBody)
			assert.Equal(t, invalidJSONBodyErrorBody, body)
		})
	}
}
