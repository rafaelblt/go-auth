// Package me returns the user an access token was issued to. It checks the
// token as verify does, then reads the user its sub names: nothing about the
// token's session, so the tokens of a revoked session still answer until they
// expire.
//
// See docs/architecture/usecases.md#me.
package me

import (
	"context"
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase"
)

type Me struct {
	validator port.AccessTokenValidator
	users     port.UserReader
}

type Input struct {
	AccessToken string
}

type Output struct {
	User usecase.UserDTO
}

var ErrTokenInvalid = usecase.NewErrorWithReason(
	"invalid_token",
	usecase.ErrorKindUnauthorized,
	"invalid token",
)

var ErrTokenExpired = usecase.NewErrorWithReason(
	"invalid_token",
	usecase.ErrorKindUnauthorized,
	"token expired",
)

var ErrUserNotFound = usecase.NewErrorWithReason(
	"invalid_token",
	usecase.ErrorKindUnauthorized,
	"user not found",
)

func (uc *Me) Execute(ctx context.Context, in Input) (Output, error) {
	claims, err := uc.validator.Validate(in.AccessToken)
	switch {
	case errors.Is(err, session.ErrTokenInvalid):
		return Output{}, ErrTokenInvalid
	case errors.Is(err, session.ErrTokenExpired):
		return Output{}, ErrTokenExpired
	case err != nil:
		return Output{}, fmt.Errorf("access token validator failed: %w", err)
	}

	usr, err := uc.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return Output{}, fmt.Errorf("find user by id failed: %w", err)
	}
	if usr == nil {
		return Output{}, ErrUserNotFound
	}

	dto := usecase.MapUserToDTO(usr)
	return Output{User: dto}, nil
}
