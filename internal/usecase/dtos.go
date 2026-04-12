package usecase

import (
	"errors"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
)

type UserDTO struct {
	id        string
	username  string
	status    string
	createdAt time.Time
	updatedAt time.Time
}

func (dto UserDTO) ID() string           { return dto.id }
func (dto UserDTO) Username() string     { return dto.username }
func (dto UserDTO) Status() string       { return dto.status }
func (dto UserDTO) CreatedAt() time.Time { return dto.createdAt }
func (dto UserDTO) UpdatedAt() time.Time { return dto.updatedAt }
func (dto UserDTO) IsZero() bool         { return dto.id == "" }

func MapUserToDTO(user *domain.User) (UserDTO, error) {
	if user == nil {
		return UserDTO{}, errors.New("the user cannot be nil to map to dto")
	}
	if user.IsZero() {
		return UserDTO{}, errors.New("the user cannot be zero to map to dto")
	}
	dto := UserDTO{
		id:        user.ID().Value().String(),
		username:  user.Username().String(),
		status:    user.Status().String(),
		createdAt: user.CreatedAt(),
		updatedAt: user.UpdatedAt(),
	}
	return dto, nil
}
