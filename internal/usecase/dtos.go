// Package usecase holds what the use cases share: the DTOs their outputs carry,
// the mappers that are the only way to build one, and UseCaseError. Each use
// case itself lives in a subpackage, and that boundary is what makes a mapper
// the only way to build a DTO that is not zero.
//
// See docs/architecture/usecases/README.md, and
// docs/development/decisions/0034-protected-dtos.md.
package usecase

import (
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/port"
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

func MapUserToDTO(user *user.User) UserDTO {
	if user == nil {
		panic("cannot map a nil user to dto")
	}
	if user.IsZero() {
		panic("cannot map a zero user to dto")
	}
	dto := UserDTO{
		id:        user.ID().Value().String(),
		username:  user.Username().String(),
		status:    user.Status().String(),
		createdAt: user.CreatedAt(),
		updatedAt: user.UpdatedAt(),
	}
	return dto
}

type AccessTokenDTO struct {
	value     string
	expiresAt time.Time
}

func (dto AccessTokenDTO) Value() string        { return dto.value }
func (dto AccessTokenDTO) ExpiresAt() time.Time { return dto.expiresAt }
func (dto AccessTokenDTO) IsZero() bool         { return dto.value == "" }

func MapAccessTokenIssuedToDTO(issued port.AccessTokenIssued) AccessTokenDTO {
	if issued.Token.IsZero() {
		panic("cannot map a zero access token to dto")
	}
	dto := AccessTokenDTO{
		value:     issued.Token.Value(),
		expiresAt: issued.ExpiresAt,
	}
	return dto
}

type RefreshTokenDTO struct {
	value     string
	expiresAt time.Time
}

func (dto RefreshTokenDTO) Value() string        { return dto.value }
func (dto RefreshTokenDTO) ExpiresAt() time.Time { return dto.expiresAt }
func (dto RefreshTokenDTO) IsZero() bool         { return dto.value == "" }

// MapRefreshTokenToDTO takes the secret apart from the token because the token
// keeps only its hash.
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
		value:     secret.Value(),
		expiresAt: token.ExpiresAt(),
	}
	return dto
}
