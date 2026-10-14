// Package changepassword replaces a user's password, given the username and the
// current one, and revokes every session of the user. It rejects credentials as
// login does: one answer for every failure, and a bcrypt comparison even when
// the account is missing.
//
// See docs/architecture/usecases.md#change-password.
package changepassword

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/credential"
	"github.com/rafaelblt/go-auth/internal/validation"
)

const FieldNewPassword = "NewPassword"

type Input struct {
	Username        string
	CurrentPassword string
	NewPassword     string
}

type Output struct {
	User usecase.UserDTO
}

type ChangePassword struct {
	credentials credential.Checker
	hasher      port.PasswordHasher
	uow         port.UnitOfWork
	clock       port.Clock
}

func (uc ChangePassword) Execute(ctx context.Context, input Input) (Output, error) {
	username, current, err := credential.Parse(input.Username, input.CurrentPassword)
	if err != nil {
		return Output{}, err
	}

	acc := validation.NewAccumulator()
	next, nextIssues := password.NewPlainReplacing(input.NewPassword, current)
	acc.Add(FieldNewPassword, nextIssues)
	err = acc.Err()
	if err != nil {
		return Output{}, err
	}

	usr, pwd, err := uc.credentials.Check(ctx, username, current)
	if err != nil {
		return Output{}, err
	}

	hashed, err := uc.hasher.Hash(next)
	if err != nil {
		return Output{}, fmt.Errorf("password hashing failed: %w", err)
	}

	now := uc.clock.Now()
	previous := pwd.Hash()
	err = pwd.ChangeHash(hashed, now)
	if err != nil {
		return Output{}, fmt.Errorf("password hash change failed: %w", err)
	}

	err = uc.save(ctx, pwd, previous, now)
	if err != nil {
		return Output{}, err
	}

	dto := usecase.MapUserToDTO(usr)
	return Output{User: dto}, nil
}

// save replaces the hash and revokes every session of the user in one
// transaction. The hash is replaced only while it is still the one the current
// password was checked against: a change that loses that race is rejected like
// a wrong current password, which by then it is.
//
// See docs/architecture/usecases.md#change-password.
func (uc ChangePassword) save(
	ctx context.Context, pwd *password.Password, previous password.Hashed, now time.Time,
) error {
	err := uc.uow.Do(ctx, func(deps port.UowDeps) error {
		err := deps.PasswordWriter.UpdateHash(ctx, pwd, previous)
		if err != nil {
			return fmt.Errorf("password writer update hash failed: %w", err)
		}

		err = deps.SessionWriter.RevokeAllByUserID(ctx, pwd.UserID(), now)
		if err != nil {
			return fmt.Errorf("session writer revoke all failed: %w", err)
		}

		return nil
	})
	if errors.Is(err, password.ErrHashChanged) {
		return credential.ErrPasswordChanged
	}
	if err != nil {
		return fmt.Errorf("password change saving failed: %w", err)
	}
	return nil
}
