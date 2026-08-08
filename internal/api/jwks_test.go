package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
