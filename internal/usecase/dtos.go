package usecase

import (
	"errors"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
)

type UserDTO interface {
	ID() string
	Username() string
	Status() string
	CreatedAt() time.Time
	UpdatedAt() time.Time
}

type userDTO struct {
	id        string
	username  string
	status    string
	createdAt time.Time
	updatedAt time.Time
}
func (dto userDTO) ID() string           { return dto.id }
func (dto userDTO) Username() string     { return dto.username }
func (dto userDTO) Status() string       { return dto.status }
func (dto userDTO) CreatedAt() time.Time { return dto.createdAt }
func (dto userDTO) UpdatedAt() time.Time { return dto.updatedAt }

func MapUserToDTO(user *domain.User) (UserDTO, error) {
	if user.IsZero() {
		return nil, errors.New("the user cannot be zero to map to dto")
	}
	dto := userDTO{
		id:        user.ID().Value().String(),
		username:  user.Username().String(),
		status:    user.Status().String(),
		createdAt: user.CreatedAt(),
		updatedAt: user.UpdatedAt(),
	}
	return dto, nil
}
