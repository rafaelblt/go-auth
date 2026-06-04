package register

import (
	"context"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/rafaelblt/go-auth/internal/validation"
)

type Register struct {
	userExists port.UserExistsChecker
	uow        port.UnitOfWork
	hasher     port.PasswordHasher
	clock      port.Clock
}

var ErrUsernameAlreadyExists = usecase.NewError(
	"USERNAME_ALREADY_EXISTS",
	usecase.ErrorKindConflict,
)

func (uc Register) Execute(ctx context.Context, input Input) (Output, error) {
	validation := validation.NewAccumulator()

	username, err := user.NewUsername(input.Username)
	validation.Add(FieldUsername, err)
	password, err := credential.NewPlainPassword(input.Password)
	validation.Add(FieldPassword, err)

	err = validation.Err()
	if err != nil {
		return Output{}, err
	}

	err = uc.checkUsernameExists(ctx, username)
	if err != nil {
		return Output{}, err
	}

	hashed, err := uc.hasher.Hash(password)
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

	cred, err := uc.createCredential(credential.CreationParams{
		UserID:    usr.ID(),
		Kind:      credential.KindPassword,
		Provider:  credential.ProviderLocal,
		Secret:    hashed,
		CreatedAt: now,
	})
	if err != nil {
		return Output{}, err
	}

	err = uc.save(ctx, usr, cred)
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

func (uc Register) createCredential(params credential.CreationParams) (*credential.Credential, error) {
	cred, err := credential.NewCredential(params)
	if err != nil {
		return nil, fmt.Errorf("credential creation failed: %w", err)
	}
	return cred, nil
}

func (uc Register) save(ctx context.Context, user *user.User, cred *credential.Credential) error {
	return uc.uow.Do(ctx, func(deps port.UowDeps) error {
		err := deps.UserWriter.Save(ctx, user)
		if err != nil {
			return fmt.Errorf("user writer save failed: %w", err)
		}
		err = deps.CredentialWriter.Save(ctx, cred)
		if err != nil {
			return fmt.Errorf("credential writer save failed: %w", err)
		}
		return nil
	})
}
