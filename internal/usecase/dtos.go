package usecase

import (
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/port"
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

// MapRefreshTokenToDTO takes the secret apart from the token because the token
// keeps only its hash. Both come from the same session.NewRefreshToken call.
func MapRefreshTokenToDTO(token *session.RefreshToken, secret session.RefreshTokenSecret) RefreshTokenDTO {
	if token == nil {
		panic("cannot map a nil refresh token to dto")
	}
	if token.IsZero() {
		panic("cannot map a zero refresh token to dto")
	}
	if secret.IsZero() {
		panic("cannot map a zero refresh token secret to dto")
	}
	dto := RefreshTokenDTO{
		Value:     secret.Value(),
		ExpiresAt: token.ExpiresAt(),
	}
	return dto
}

func (dto RefreshTokenDTO) IsZero() bool { return dto.Value == "" }
