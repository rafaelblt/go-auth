// Package register creates an account. It does not log the user in: it creates
// no session and no token.
//
// See docs/architecture/usecases/register.md.
package register

import (
	"context"
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/validation"
)

type Register struct {
	userExists port.UserExistsChecker
	uow        port.UnitOfWork
	hasher     port.PasswordHasher
	clock      port.Clock
}

var ErrUsernameAlreadyExists = usecase.NewError(
	"username_already_exists",
	usecase.ErrorKindConflict,
)

func (uc Register) Execute(ctx context.Context, input Input) (Output, error) {
	acc := validation.NewAccumulator()

	username, usernameIssues := user.NewUsername(input.Username)
	acc.Add(FieldUsername, usernameIssues)
	plain, plainIssues := password.NewPlain(input.Password)
	acc.Add(FieldPassword, plainIssues)

	err := acc.Err()
	if err != nil {
		return Output{}, err
	}

	// Checked here so a taken username costs one index lookup instead of a
	// full bcrypt hash. The insert in save is what actually decides it.
	err = uc.checkUsernameExists(ctx, username)
	if err != nil {
		return Output{}, err
	}

	hashed, err := uc.hasher.Hash(plain)
	if err != nil {
		return Output{}, fmt.Errorf("password hashing failed: %w", err)
	}

	now := uc.clock.Now()

	usr, err := uc.createUser(user.CreationParams{
		Username:  username,
		CreatedAt: now,
	})
	if err != nil {
		return Output{}, err
	}

	pwd, err := uc.createPassword(password.CreationParams{
		UserID:    usr.ID(),
		Hash:      hashed,
		CreatedAt: now,
	})
	if err != nil {
		return Output{}, err
	}

	err = uc.save(ctx, usr, pwd)
	if err != nil {
		return Output{}, err
	}

	dto := usecase.MapUserToDTO(usr)
	return Output{User: dto}, nil
}

func (uc Register) checkUsernameExists(ctx context.Context, username user.Username) error {
	exists, err := uc.userExists.ExistsByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("user exists checker failed: %w", err)
	}
	if exists {
		return ErrUsernameAlreadyExists
	}
	return nil
}

func (uc Register) createUser(params user.CreationParams) (*user.User, error) {
	usr, err := user.NewUser(params)
	if err != nil {
		return nil, fmt.Errorf("user creation failed: %w", err)
	}
	return usr, nil
}

func (uc Register) createPassword(params password.CreationParams) (*password.Password, error) {
	pwd, err := password.NewPassword(params)
	if err != nil {
		return nil, fmt.Errorf("password creation failed: %w", err)
	}
	return pwd, nil
}

// save inserts the user and its password in one transaction. The username is
// checked for free before the password is hashed, but only the insert decides
// it: a registration that loses the race for the username is reported as a
// conflict, not as an unexpected failure.
//
// See docs/development/decisions/0048-duplicate-username-reported-by-the-writer.md.
func (uc Register) save(ctx context.Context, usr *user.User, pwd *password.Password) error {
	err := uc.uow.Do(ctx, func(deps port.UowDeps) error {
		err := deps.UserWriter.Add(ctx, usr)
		if err != nil {
			return fmt.Errorf("user writer save failed: %w", err)
		}
		err = deps.PasswordWriter.Add(ctx, pwd)
		if err != nil {
			return fmt.Errorf("password writer save failed: %w", err)
		}
		return nil
	})
	if errors.Is(err, user.ErrUsernameAlreadyExists) {
		return ErrUsernameAlreadyExists
	}
	return err
}
