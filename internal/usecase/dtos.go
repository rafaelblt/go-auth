package usecase

import (
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
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
	ExpiresAt time.Time
}

func MapAccessTokenIssuedToDTO(issued port.AccessTokenIssued) AccessTokenDTO {
	if issued.Token.IsZero() {
		panic("cannot map a zero access token to dto")
	}
	dto := AccessTokenDTO{
		Value:     issued.Token.Value(),
		ExpiresAt: issued.ExpiresAt,
	}
	return dto
}

func (dto AccessTokenDTO) IsZero() bool { return dto.Value == "" }

type RefreshTokenDTO struct {
	Value     string
	ExpiresAt time.Time
}

func MapRefreshTokenIssuedToDTO(issued port.RefreshTokenIssued) RefreshTokenDTO {
	if issued.Token == nil {
		panic("cannot map a nil refresh token to dto")
	}
	dto := RefreshTokenDTO{
		Value:     issued.RawValue,
		ExpiresAt: issued.Token.ExpiresAt(),
	}
	return dto
}

func (dto RefreshTokenDTO) IsZero() bool { return dto.Value == "" }
