package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/testutil/apitest"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapUserDTO(t *testing.T) {
	dto := apitest.NewUserDTO(t, nil)

	user := mapUserDTO(dto)

	assert.NotZero(t, user)
	assert.Equal(t, dto.ID(), user.ID)
	assert.Equal(t, dto.Username(), user.Username)
	assert.Equal(t, dto.Status(), user.Status)
	assert.Equal(t, dto.CreatedAt(), user.CreatedAt)
	assert.Equal(t, dto.UpdatedAt(), user.UpdatedAt)
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
	assert.Equal(t, dto.ExpiresAt(), retrieved.ExpiresAt)
	assert.Equal(t, int64(dto.ExpiresIn()/time.Second), retrieved.ExpiresIn)
	assert.Equal(t, int64(1800), retrieved.ExpiresIn)
}

func TestMapAccessTokenDTO_PanicsWithZeroDTO(t *testing.T) {
	dto := usecase.AccessTokenDTO{}
	assert.Panics(t, func() { mapAccessTokenDTO(dto) })
}

func TestMapRefreshTokenDTO(t *testing.T) {
	dto := apitest.NewRefreshTokenDTO(t)

	retrieved := mapRefreshTokenDTO(dto)

	require.NotZero(t, retrieved)
	assert.Equal(t, dto.Value(), retrieved.Value)
	assert.Equal(t, dto.ExpiresAt(), retrieved.ExpiresAt)
	assert.Equal(t, int64(dto.ExpiresIn()/time.Second), retrieved.ExpiresIn)
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
			ID:        output.User.ID(),
			Username:  output.User.Username(),
			Status:    output.User.Status(),
			CreatedAt: output.User.CreatedAt(),
			UpdatedAt: output.User.UpdatedAt(),
		},
	}
	assert.Equal(t, expectedBody, actualBody)
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
		RefreshToken: apitest.NewRefreshTokenDTO(t),
	}

	resp := loginEncoder(output)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.IsType(t, tokensResponseBody{}, resp.Body)
	actualBody := resp.Body.(tokensResponseBody)
	expectedBody := tokensResponseBody{
		AccessToken: token{
			Value:     output.AccessToken.Value(),
			ExpiresAt: output.AccessToken.ExpiresAt(),
			ExpiresIn: int64(output.AccessToken.ExpiresIn() / time.Second),
		},
		RefreshToken: token{
			Value:     output.RefreshToken.Value(),
			ExpiresAt: output.RefreshToken.ExpiresAt(),
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
		RefreshToken: apitest.NewRefreshTokenDTO(t),
	}

	resp := refreshEncoder(output)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.IsType(t, tokensResponseBody{}, resp.Body)
	actualBody := resp.Body.(tokensResponseBody)
	expectedBody := tokensResponseBody{
		AccessToken: token{
			Value:     output.AccessToken.Value(),
			ExpiresAt: output.AccessToken.ExpiresAt(),
			ExpiresIn: int64(output.AccessToken.ExpiresIn() / time.Second),
		},
		RefreshToken: token{
			Value:     output.RefreshToken.Value(),
			ExpiresAt: output.RefreshToken.ExpiresAt(),
			ExpiresIn: int64(output.RefreshToken.ExpiresIn() / time.Second),
		},
	}
	assert.Equal(t, expectedBody, actualBody)
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
