package api

import (
	"time"

	"github.com/rafaelblt/go-auth/internal/usecase"
)

type user struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func mapUserDTO(dto usecase.UserDTO) user {
	if dto.IsZero() {
		panic("the mapUserDTO() func received a zero UserDTO")
	}
	user := user{
		ID:        dto.ID(),
		Username:  dto.Username(),
		Status:    dto.Status(),
		CreatedAt: dto.CreatedAt(),
		UpdatedAt: dto.UpdatedAt(),
	}
	return user
}

type accessToken struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
	ExpiresIn int64     `json:"expires_in"`
}

func mapAccessTokenDTO(dto usecase.AccessTokenDTO) accessToken {
	if dto.IsZero() {
		panic("the mapAccessTokenDTO() func received a zero AccessTokenDTO")
	}
	token := accessToken{
		Value:     dto.Value(),
		ExpiresAt: dto.ExpiresAt(),
		ExpiresIn: int64(dto.ExpiresIn() / time.Second),
	}
	return token
}

type refreshToken struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
	ExpiresIn int64     `json:"expires_in"`
}

func mapRefreshTokenDTO(dto usecase.RefreshTokenDTO) refreshToken {
	token := refreshToken{}
	if dto.IsZero() {
		panic("the mapRefreshTokenDTO() func received a zero RefreshTokenDTO")
	}
	token.Value = dto.Value()
	token.ExpiresAt = dto.ExpiresAt()
	token.ExpiresIn = int64(dto.ExpiresIn() / time.Second)
	return token
}
