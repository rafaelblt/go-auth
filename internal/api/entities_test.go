package api

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil/apitest"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapUserDTO(t *testing.T) {
	dto := apitest.NewUserDTO(t, nil)

	user := mapUserDTO(dto)

	assert.NotZero(t, user)
	assert.Equal(t, dto.ID, user.ID)
	assert.Equal(t, dto.Username, user.Username)
	assert.Equal(t, dto.Status, user.Status)
	assert.Equal(t, dto.CreatedAt, user.CreatedAt)
	assert.Equal(t, dto.UpdatedAt, user.UpdatedAt)
}

func TestMapUserDTO_PanicsWithZeroDTO(t *testing.T) {
	dto := usecase.UserDTO{}
	assert.Panics(t, func() { mapUserDTO(dto) })
}

func TestMapAccessTokenDTO(t *testing.T) {
	dto := apitest.NewAccessTokenDTO(t)

	retrieved := mapAccessTokenDTO(dto)

	require.NotZero(t, retrieved)
	assert.Equal(t, dto.Value, retrieved.Value)
	assert.Equal(t, dto.ExpiresAt, retrieved.ExpiresAt)
}

func TestMapAccessTokenDTO_PanicsWithZeroDTO(t *testing.T) {
	dto := usecase.AccessTokenDTO{}
	assert.Panics(t, func() { mapAccessTokenDTO(dto) })
}

func TestMapRefreshTokenDTO(t *testing.T) {
	dto := apitest.NewRefreshTokenDTO(t)

	retrieved := mapRefreshTokenDTO(dto)

	require.NotZero(t, retrieved)
	assert.Equal(t, dto.Value, retrieved.Value)
	assert.Equal(t, dto.ExpiresAt, retrieved.ExpiresAt)
}

func TestMapRefreshTokenDTO_PanicsWithZeroDTO(t *testing.T) {
	dto := usecase.RefreshTokenDTO{}
	assert.Panics(t, func() { mapRefreshTokenDTO(dto) })
}
