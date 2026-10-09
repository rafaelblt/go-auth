// Package changepassword replaces a user's password, given the username and the
// current one, and revokes every session of the user. It rejects credentials as
// login does: one answer for every failure, and a bcrypt comparison even when
// the account is missing.
//
// See docs/architecture/usecases/change-password.md.
package changepassword

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase"
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
	users      port.UserReader
	passwords  port.PasswordReader
	pwdChecker port.PasswordChecker
	hasher     port.PasswordHasher
	uow        port.UnitOfWork
	clock      port.Clock
	dummyHash  password.Hashed
}

var ErrUsernameMalformed = usecase.NewErrorWithReason(
	"invalid_credentials",
	usecase.ErrorKindUnauthorized,
	"malformed username",
)

var ErrPasswordMalformed = usecase.NewErrorWithReason(
	"invalid_credentials",
	usecase.ErrorKindUnauthorized,
	"malformed password",
)

var ErrUserNotFound = usecase.NewErrorWithReason(
	"invalid_credentials",
	usecase.ErrorKindUnauthorized,
	"user not found",
)

var ErrPasswordNotFound = usecase.NewErrorWithReason(
	"invalid_credentials",
	usecase.ErrorKindUnauthorized,
	"password not found",
)

var ErrPasswordMismatch = usecase.NewErrorWithReason(
	"invalid_credentials",
	usecase.ErrorKindUnauthorized,
	"password mismatch",
)

var ErrPasswordChanged = usecase.NewErrorWithReason(
	"invalid_credentials",
	usecase.ErrorKindUnauthorized,
	"password changed",
)

func (uc ChangePassword) Execute(ctx context.Context, input Input) (Output, error) {
	username, issues := user.NewUsername(input.Username)
	if !issues.IsEmpty() {
		return Output{}, ErrUsernameMalformed
	}
	current, issues := password.NewPlain(input.CurrentPassword)
	if !issues.IsEmpty() {
		return Output{}, ErrPasswordMalformed
	}

	acc := validation.NewAccumulator()
	next, nextIssues := password.NewPlain(input.NewPassword)
	acc.Add(FieldNewPassword, nextIssues)
	err := acc.Err()
	if err != nil {
		return Output{}, err
	}

	usr, pwd, err := uc.authenticate(ctx, username, current)
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

func (uc ChangePassword) authenticate(
	ctx context.Context, username user.Username, plain password.Plain,
) (*user.User, *password.Password, error) {
	usr, err := uc.users.FindByUsername(ctx, username)
	if err != nil {
		return nil, nil, fmt.Errorf("find user by username failed: %w", err)
	}
	if usr == nil {
		return nil, nil, uc.rejectWithDummyVerify(plain, ErrUserNotFound)
	}

	pwd, err := uc.passwords.FindByUserID(ctx, usr.ID())
	if err != nil {
		return nil, nil, fmt.Errorf("find password by user id failed: %w", err)
	}
	if pwd == nil {
		return nil, nil, uc.rejectWithDummyVerify(plain, ErrPasswordNotFound)
	}

	ok, err := uc.pwdChecker.Verify(plain, pwd.Hash())
	if err != nil {
		return nil, nil, fmt.Errorf("password verification failed: %w", err)
	}
	if !ok {
		return nil, nil, ErrPasswordMismatch
	}

	return usr, pwd, nil
}

// rejectWithDummyVerify spends the time of a real password check before
// returning rejection, as login does, so response time does not reveal whether
// the account exists. The result of the check is deliberately discarded: the
// call is here for the time it takes, and removing it reopens the timing
// difference.
//
// See docs/development/decisions/0045-login-dummy-hash.md.
func (uc ChangePassword) rejectWithDummyVerify(plain password.Plain, rejection error) error {
	_, err := uc.pwdChecker.Verify(plain, uc.dummyHash)
	if err != nil {
		return fmt.Errorf("dummy password verification failed: %w", err)
	}
	return rejection
}

// save replaces the hash and revokes every session of the user in one
// transaction. The hash is replaced only while it is still the one the current
// password was checked against: a change that loses that race is rejected like
// a wrong current password, which by then it is.
//
// See docs/architecture/usecases/change-password.md.
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
		return ErrPasswordChanged
	}
	if err != nil {
		return fmt.Errorf("password change saving failed: %w", err)
	}
	return nil
}
