package api

import (
	"errors"
	"time"

	"github.com/rafaelblt/go-auth/internal/usecase"
)

type UserResource struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Status    string `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func MapUserDTOToResource(dto usecase.UserDTO) (UserResource, error) {
	resource := UserResource{}
	if dto.IsZero() {
		return resource, errors.New("the user dto cannot be zero to map to user resource")
	}
	resource.ID = dto.ID()
	resource.Username = dto.Username()
	resource.Status = dto.Status()
	resource.CreatedAt = dto.CreatedAt()
	resource.UpdatedAt = dto.UpdatedAt()
	return resource, nil
}
