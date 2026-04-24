package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type RegisterTestHelper struct {
	t    *testing.T
	pool *pgxpool.Pool
}

func NewRegisterTestHelper(t *testing.T) RegisterTestHelper {
	t.Helper()
	pool := dbProvider.NewPool(t)
	return RegisterTestHelper{t, pool}
}

func (helper RegisterTestHelper) Handler() RegisterHandler {
	deps, err := infra.NewDependencyContainer(
		context.Background(),
		infra.DependenciesConfig{DatabasePool: helper.pool},
	)
	require.NoError(helper.t, err)
	uc, err := deps.BuildRegister()
	require.NoError(helper.t, err)
	return RegisterHandler{uc}
}

func (helper RegisterTestHelper) NewRequest(
	username string, password string,
) *http.Request {
	body := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
	return httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
}

func (helper RegisterTestHelper) SaveUser(t *testing.T, user *domain.User) {
	t.Helper()
	repo, err := infra.NewUserRepo(helper.pool)
	require.NoError(t, err)
	require.NoError(t, repo.Save(context.Background(), user))
}

func DecodeResponse[T any](t *testing.T, body io.ReadCloser) T {
	defer body.Close()

	var response T
	require.NoError(t, json.NewDecoder(body).Decode(&response))

	return response
}

func TestRegister_ReturnsSuccessResponse(t *testing.T) {
	helper := NewRegisterTestHelper(t)
	request := helper.NewRequest("maria", "12345678")
	recorder := httptest.NewRecorder()

	helper.Handler().Handle(recorder, request)
	response := recorder.Result()
	decoded := DecodeResponse[RegisterResponse](t, response.Body)

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
		request  *http.Request
		expected map[string]ValidationErrors
	}{
		{
			desc: "username too short",
			request: helper.NewRequest(
				strings.Repeat("a", domain.UsernameMinLen-1),
				"12345678",
			),
			expected: map[string]ValidationErrors{
				"username": {errRegisterUsernameTooShort},
			},
		},
		{
			desc: "username too long",
			request: helper.NewRequest(
				strings.Repeat("a", domain.UsernameMaxLen+1),
				"12345678",
			),
			expected: map[string]ValidationErrors{
				"username": {errRegisterUsernameTooLong},
			},
		},
		{
			desc: "password too short",
			request: helper.NewRequest(
				"rafael",
				strings.Repeat("a", domain.PlainPasswordMinLen-1),
			),
			expected: map[string]ValidationErrors{
				"password": {errRegisterPasswordTooShort},
			},
		},
		{
			desc: "password too long",
			request: helper.NewRequest(
				"rafael",
				strings.Repeat("a", domain.PlainPasswordMaxLen+1),
			),
			expected: map[string]ValidationErrors{
				"password": {errRegisterPasswordTooLong},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			helper.Handler().Handle(recorder, tC.request)
			response := recorder.Result()

			decoded := DecodeResponse[ValidationErrorResponse](t, response.Body)
			assert.Equal(t, http.StatusUnprocessableEntity, response.StatusCode)
			assert.Equal(t, tC.expected, decoded.Errors)
		})
	}
}

func TestRegister_ReturnsUsernameAlreadyExists(t *testing.T) {
	helper := NewRegisterTestHelper(t)
	user := testutil.DefaultUser(t)
	helper.SaveUser(t, user)

	request := helper.NewRequest(user.Username().String(), "12345678")
	recorder := httptest.NewRecorder()
	helper.Handler().Handle(recorder, request)

	response := recorder.Result()
	decoded := DecodeResponse[ErrorResponse](t, response.Body)
	assert.Equal(t, http.StatusConflict, response.StatusCode)
	assert.Equal(t, errRegisterUsernameAlreadyExists, decoded.Error)
}
