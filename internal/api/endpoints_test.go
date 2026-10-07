package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/testutil/apitest"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/rafaelblt/go-auth/internal/usecase/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapUserDTO(t *testing.T) {
	dto := apitest.NewUserDTO(t, nil)

	user := mapUserDTO(dto)

	assert.NotZero(t, user)
	assert.Equal(t, dto.ID(), user.ID)
	assert.Equal(t, dto.Username(), user.Username)
}

func TestMapUserDTO_PanicsWithZeroDTO(t *testing.T) {
	dto := usecase.UserDTO{}
	assert.Panics(t, func() { mapUserDTO(dto) })
}

func TestMapAccessTokenDTO(t *testing.T) {
	dto := apitest.NewAccessTokenDTO(t)

	retrieved := mapAccessTokenDTO(dto)

	require.NotZero(t, retrieved)
	assert.Equal(t, dto.Value(), retrieved.Value)
	assert.Equal(t, "2031-03-14T15:09:26Z", retrieved.ExpiresAt)
	assert.Equal(t, int64(dto.ExpiresIn()/time.Second), retrieved.ExpiresIn)
	assert.Equal(t, int64(1800), retrieved.ExpiresIn)
}

func TestMapAccessTokenDTO_PanicsWithZeroDTO(t *testing.T) {
	dto := usecase.AccessTokenDTO{}
	assert.Panics(t, func() { mapAccessTokenDTO(dto) })
}

func TestMapRefreshTokenDTO(t *testing.T) {
	dto := apitest.NewRefreshTokenDTO(t, nil)

	retrieved := mapRefreshTokenDTO(dto)

	require.NotZero(t, retrieved)
	assert.Equal(t, dto.Value(), retrieved.Value)
	assert.Equal(t, "2026-06-17T23:40:00Z", retrieved.ExpiresAt)
	assert.Equal(t, int64(dto.ExpiresIn()/time.Second), retrieved.ExpiresIn)
}

func TestMapRefreshTokenDTO_FormatsExpiresAt(t *testing.T) {
	testCases := []struct {
		desc      string
		expiresAt time.Time
		expected  string
	}{
		{
			desc:      "whole seconds in UTC",
			expiresAt: time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC),
			expected:  "2026-09-15T12:00:00Z",
		},
		{
			desc:      "fraction dropped, not rounded",
			expiresAt: time.Date(2026, 9, 15, 12, 0, 0, 999999999, time.UTC),
			expected:  "2026-09-15T12:00:00Z",
		},
		{
			desc:      "other offset converted to UTC",
			expiresAt: time.Date(2026, 9, 15, 9, 0, 0, 0, time.FixedZone("UTC-3", -3*60*60)),
			expected:  "2026-09-15T12:00:00Z",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			dto := apitest.NewRefreshTokenDTO(t, func(p *session.RefreshTokenRestoreParams) { p.ExpiresAt = tC.expiresAt })

			retrieved := mapRefreshTokenDTO(dto)

			assert.Equal(t, tC.expected, retrieved.ExpiresAt)
		})
	}
}

func TestMapRefreshTokenDTO_PanicsWithZeroDTO(t *testing.T) {
	dto := usecase.RefreshTokenDTO{}
	assert.Panics(t, func() { mapRefreshTokenDTO(dto) })
}

func TestRegisterDecoder_ReturnsInput(t *testing.T) {
	body := credentialsRequestBody{
		Username: "username",
		Password: "password",
	}
	buf, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "localhost:3264", bytes.NewReader(buf))

	in, err := registerDecoder(req)

	require.NoError(t, err)
	assert.Equal(t, body.Username, in.Username)
	assert.Equal(t, body.Password, in.Password)
}

func TestRegisterDecoder_ReturnsError_WhenRequestBodyIsNil(t *testing.T) {
	req := httptest.NewRequest("POST", "localhost:3264", nil)

	in, err := registerDecoder(req)

	assert.Error(t, err)
	assert.Zero(t, in)
}

func TestRegisterEncoder_ReturnsResponse(t *testing.T) {
	output := register.Output{User: apitest.NewUserDTO(t, nil)}

	resp := registerEncoder(output)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	require.IsType(t, registerResponseBody{}, resp.Body)
	actualBody := resp.Body.(registerResponseBody)
	expectedBody := registerResponseBody{
		User: user{
			ID:       output.User.ID(),
			Username: output.User.Username(),
		},
	}
	assert.Equal(t, expectedBody, actualBody)
}

func TestRegisterEncoder_BodyIsOnlyUserIDAndUsername(t *testing.T) {
	output := register.Output{User: apitest.NewUserDTO(t, nil)}

	resp := registerEncoder(output)

	actualJSON, err := json.Marshal(resp.Body)
	require.NoError(t, err)
	expectedJSON, err := json.Marshal(map[string]map[string]string{
		"user": {
			"id":       output.User.ID(),
			"username": output.User.Username(),
		},
	})
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedJSON), string(actualJSON))
}

func TestLoginDecoder_ReturnsInput(t *testing.T) {
	body := credentialsRequestBody{
		Username: "username",
		Password: "password",
	}
	buf, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "localhost:2040", bytes.NewReader(buf))

	in, err := loginDecoder(req)

	require.NoError(t, err)
	assert.Equal(t, body.Username, in.Username)
	assert.Equal(t, body.Password, in.Password)
}

func TestLoginDecoder_ReturnsError_WhenRequestBodyIsNil(t *testing.T) {
	req := httptest.NewRequest("POST", "localhost:2040", nil)

	in, err := loginDecoder(req)

	assert.Error(t, err)
	assert.Zero(t, in)
}

func TestLoginEncoder_ReturnsResponse(t *testing.T) {
	output := login.Output{
		AccessToken:  apitest.NewAccessTokenDTO(t),
		RefreshToken: apitest.NewRefreshTokenDTO(t, nil),
	}

	resp := loginEncoder(output)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.IsType(t, tokensResponseBody{}, resp.Body)
	actualBody := resp.Body.(tokensResponseBody)
	expectedBody := tokensResponseBody{
		AccessToken: token{
			Value:     output.AccessToken.Value(),
			ExpiresAt: "2031-03-14T15:09:26Z",
			ExpiresIn: int64(output.AccessToken.ExpiresIn() / time.Second),
		},
		RefreshToken: token{
			Value:     output.RefreshToken.Value(),
			ExpiresAt: "2026-06-17T23:40:00Z",
			ExpiresIn: int64(output.RefreshToken.ExpiresIn() / time.Second),
		},
	}
	assert.Equal(t, expectedBody, actualBody)
}

func TestRefreshDecoder_ReturnsInput(t *testing.T) {
	body := refreshRequestBody{
		RefreshToken: "refresh token",
	}
	buf, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "localhost:0510", bytes.NewReader(buf))

	in, err := refreshDecoder(req)

	require.NoError(t, err)
	assert.Equal(t, body.RefreshToken, in.RefreshToken)
}

func TestRefreshDecoder_ReturnsError_WhenRequestBodyIsNil(t *testing.T) {
	req := httptest.NewRequest("POST", "localhost:0510", nil)

	in, err := refreshDecoder(req)

	assert.Error(t, err)
	assert.Zero(t, in)
}

func TestRefreshEncoder_ReturnsResponse(t *testing.T) {
	output := refresh.Output{
		AccessToken:  apitest.NewAccessTokenDTO(t),
		RefreshToken: apitest.NewRefreshTokenDTO(t, nil),
	}

	resp := refreshEncoder(output)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.IsType(t, tokensResponseBody{}, resp.Body)
	actualBody := resp.Body.(tokensResponseBody)
	expectedBody := tokensResponseBody{
		AccessToken: token{
			Value:     output.AccessToken.Value(),
			ExpiresAt: "2031-03-14T15:09:26Z",
			ExpiresIn: int64(output.AccessToken.ExpiresIn() / time.Second),
		},
		RefreshToken: token{
			Value:     output.RefreshToken.Value(),
			ExpiresAt: "2026-06-17T23:40:00Z",
			ExpiresIn: int64(output.RefreshToken.ExpiresIn() / time.Second),
		},
	}
	assert.Equal(t, expectedBody, actualBody)
}

func TestVerifyDecoder_ReturnsInput(t *testing.T) {
	body := verifyRequestBody{
		AccessToken: "access token",
	}
	buf, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "localhost:8080", bytes.NewReader(buf))

	in, err := verifyDecoder(req)

	require.NoError(t, err)
	assert.Equal(t, body.AccessToken, in.AccessToken)
}

func TestVerifyDecoder_ReturnsError_WhenRequestBodyIsNil(t *testing.T) {
	req := httptest.NewRequest("POST", "localhost:8080", nil)

	in, err := verifyDecoder(req)

	assert.Error(t, err)
	assert.Zero(t, in)
}

func TestVerifyEncoder_ReturnsResponse(t *testing.T) {
	output := verify.Output{
		UserID:    "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37",
		ExpiresAt: time.Date(2026, 9, 8, 9, 30, 0, 999999999, time.FixedZone("UTC-3", -3*60*60)),
		ExpiresIn: 1342*time.Second + 999*time.Millisecond,
	}

	resp := verifyEncoder(output)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	actualJSON, err := json.Marshal(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"user_id": "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37",
		"expires_at": "2026-09-08T12:30:00Z",
		"expires_in": 1342
	}`, string(actualJSON))
}

type FakePublicKeyProvider struct {
	keys []port.PublicKey
}

func NewFakePublicKeyProvider() *FakePublicKeyProvider {
	return &FakePublicKeyProvider{
		keys: make([]port.PublicKey, 0),
	}
}

func (fake *FakePublicKeyProvider) PublicKeys() []port.PublicKey {
	return fake.keys
}

func handlerAndFakeForTest(t *testing.T) (*jwksHandler, *FakePublicKeyProvider) {
	t.Helper()
	fake := NewFakePublicKeyProvider()
	handler := jwksHandler{fake}
	return &handler, fake
}

func decodeJWKSBody(t *testing.T, body []byte) jwksBody {
	t.Helper()
	require.NotNil(t, body)

	var b jwksBody
	require.NoError(t, json.Unmarshal(body, &b))

	return b
}

func TestJWKS(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "url:8080", nil)
	handler, fake := handlerAndFakeForTest(t)

	handler.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	body := decodeJWKSBody(t, recorder.Body.Bytes())
	require.Equal(t, len(fake.PublicKeys()), len(body.Keys))
	for i, expected := range fake.PublicKeys() {
		actual := body.Keys[i]
		assert.Equal(t, base64.RawURLEncoding.EncodeToString(expected.Key), actual.X)
		assert.Equal(t, expected.Algorithm, actual.Alg)
		assert.Equal(t, expected.ID, actual.Kid)
		assert.Equal(t, expected.Type, actual.Kty)
		assert.Equal(t, "sig", actual.Use)
	}
}
