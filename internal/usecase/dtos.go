package usecase

import (
	"time"

	"github.com/rafaelblt/go-auth/internal/user"
)

type UserDTO struct {
	ID        string
	Username  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (dto UserDTO) IsZero() bool { return dto.ID == "" }

func MapUserToDTO(user *user.User) UserDTO {
	if user == nil {
		panic("cannot map a nil user to dto")
	}
	if user.IsZero() {
		panic("cannot map a zero user to dto")
	}
	dto := UserDTO{
		ID:        user.ID().Value().String(),
		Username:  user.Username().String(),
		Status:    user.Status().String(),
		CreatedAt: user.CreatedAt(),
		UpdatedAt: user.UpdatedAt(),
	}
	return dto
}

type AccessTokenDTO struct {
	Value     string
}

func (dto AccessTokenDTO) IsZero() bool { return dto.Value == "" }

type RefreshTokenDTO struct {
	Value     string
	ExpiresAt time.Time
}

func (dto RefreshTokenDTO) IsZero() bool { return dto.Value == "" }
