package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil/apitest"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
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

func TestRegister_ReturnsUsernameAlreadyExists(t *testing.T) {
	helper := NewRegisterTestHelper(t)

	err := register.ErrUsernameAlreadyExists
	helper.FakeRegister.SetError(err)

	req := helper.NewRequestWithFields("maria", "12345678")
	resp := helper.Handler().Handle(req)

	require.Equal(t, http.StatusConflict, resp.StatusCode)
	require.IsType(t, errorBody{}, resp.Body)
	body := resp.Body.(errorBody)
	assert.Equal(t, err.Code(), body.Error.Code)
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
	helper := NewRegisterTestHelper(t)
	req := helper.NewRequestWithFields("username", "password")
	helper.FakeRegister.SetError(errors.New("internal error"))

	resp := helper.Handler().Handle(req)

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.IsType(t, errorBody{}, resp.Body)
	body := resp.Body.(errorBody)
	assert.Equal(t, internalServerErrorBody, body)
}

func TestRegister_ReturnsValidationError(t *testing.T) {
	helper := NewRegisterTestHelper(t)

	issue1 := validation.IssueTooShort(50)
	issue2 := validation.IssueTooLong(50)
	ferr1 := validation.NewFieldError("field1", issue1)
	ferr2 := validation.NewFieldError("field2", issue2)
	verr := validation.NewValidationError(ferr1, ferr2)
	helper.FakeRegister.SetError(verr)

	req := helper.NewRequestWithFields("username", "password")
	resp := helper.Handler().Handle(req)

	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	require.IsType(t, validationErrorBody{}, resp.Body)

	body := resp.Body.(validationErrorBody)
	expected := validationErrorBody{Errors: fieldErrors{
		fieldErrorData{Field: ferr1.Field(), Code: ferr1.Issue().Code(), Details: ferr1.Issue().Details()},
		fieldErrorData{Field: ferr2.Field(), Code: ferr2.Issue().Code(), Details: ferr2.Issue().Details()},
	}}
	assert.Equal(t, expected, body)
}
