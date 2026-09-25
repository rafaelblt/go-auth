package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil/apitest"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterDecoder_ReturnsInput(t *testing.T) {
	body := registerRequestBody{
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
