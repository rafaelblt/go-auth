package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/testutil/apitest"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type RefreshTestHelper struct {
	t           *testing.T
	FakeRefresh *apitest.FakeRefresh
}

func NewRefreshTestHelper(t *testing.T) *RefreshTestHelper {
	t.Helper()
	fakeRefresh := apitest.NewFakeRefresh()
	return &RefreshTestHelper{t, fakeRefresh}
}

func (h *RefreshTestHelper) Handler() *refreshHandler {
	return newRefreshHandler(h.FakeRefresh)
}

func (h *RefreshTestHelper) NewRequest(body string) *http.Request {
	h.t.Helper()
	reader := strings.NewReader(body)
	req, err := http.NewRequest(http.MethodPost, "url", reader)
	require.NoError(h.t, err)
	return req
}

func (h *RefreshTestHelper) NewRequestWithToken(token string) *http.Request {
	body := fmt.Sprintf(`{"refresh_token":"%s"}`, token)
	return h.NewRequest(body)
}

func TestRefresh_ReturnsRefreshResponse(t *testing.T) {
	helper := NewRefreshTestHelper(t)
	req := helper.NewRequestWithToken("123")

	resp := helper.Handler().Handle(req)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.IsType(t, refreshResponseBody{}, resp.Body)

	body := resp.Body.(refreshResponseBody)
	output := testutil.Only(t, helper.FakeRefresh.Outputs())
	assert.Equal(t, output.AccessToken.Value, body.AccessToken.Value)
	assert.Equal(t, output.RefreshToken.Value, body.RefreshToken.Value)
}

func TestRefresh_ShouldGiveInputToUseCase(t *testing.T) {
	helper := NewRefreshTestHelper(t)
	token := "123"
	req := helper.NewRequestWithToken(token)

	resp := helper.Handler().Handle(req)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	input := testutil.Only(t, helper.FakeRefresh.Inputs())
	assert.Equal(t, token, input.RefreshToken)
}

func TestRefresh_TranslateUseCaseError(t *testing.T) {
	helper := NewRefreshTestHelper(t)

	err := refresh.ErrTokenInvalid
	helper.FakeRefresh.SetError(err)

	req := helper.NewRequestWithToken("secret")
	resp := helper.Handler().Handle(req)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.IsType(t, errorBody{}, resp.Body)
	body := resp.Body.(errorBody)
	assert.Equal(t, err.Code(), body.Error.Code)
}

func TestRefresh_ReturnsInvalidJSONBody(t *testing.T) {
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
			helper := NewRefreshTestHelper(t)
			req := helper.NewRequest(tC.body)
			resp := helper.Handler().Handle(req)

			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
			require.IsType(t, errorBody{}, resp.Body)
			body := resp.Body.(errorBody)
			assert.Equal(t, invalidJSONBodyErrorBody, body)
		})
	}
}
