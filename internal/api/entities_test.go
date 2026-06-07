package api

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func TestMapUserDTO(t *testing.T) {
	dto := usecase.MapUserToDTO(usertest.NewUser(t, nil))

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
	assert.Panics(t, func() {mapUserDTO(dto)} )
}
