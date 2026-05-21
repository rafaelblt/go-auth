package register

import (
	"context"
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/usecase"
)

type Register struct {
	userExists usecase.UserExistsChecker
	uow        usecase.UnitOfWork
	hasher     usecase.PasswordHasher
	clock      usecase.Clock
}

var ErrUsernameAlreadyExists = errors.New("the provided username is already registered")

func New(cfg Config) (Register, error) {
	if cfg.UserExistsChecker == nil {
		return Register{}, errors.New("user exists checker cannot be nil")
	}
	if cfg.UnitOfWork == nil {
		return Register{}, errors.New("unit of work cannot be nil")
	}
	if cfg.PasswordHasher == nil {
		return Register{}, errors.New("password hasher cannot be nil")
	}
	if cfg.Clock == nil {
		return Register{}, errors.New("clock cannot be nil")
	}
	uc := Register{
		userExists: cfg.UserExistsChecker,
		uow:        cfg.UnitOfWork,
		hasher:     cfg.PasswordHasher,
		clock:      cfg.Clock,
	}
	return uc, nil
}

func (uc Register) Execute(ctx context.Context, input Input) (Output, error) {
	validation := usecase.NewValidationAccumulator()

	username, err := domain.NewUsername(input.Username)
	validation.Add(UsernameField, err)
	password, err := domain.NewPlainPassword(input.Password)
	validation.Add(PasswordField, err)

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

	now := uc.clock.UtcNow()

	user, err := uc.createUser(domain.NewUserParams{
		Username:  username,
		CreatedAt: now,
	})
	if err != nil {
		return Output{}, err
	}

	cred, err := uc.createCredential(domain.NewCredentialParams{
		UserID:    user.ID(),
		Kind:      domain.CredentialKindPassword,
		Provider:  domain.CredentialProviderLocal,
		Secret:    hashed,
		CreatedAt: now,
	})
	if err != nil {
		return Output{}, err
	}

	err = uc.save(ctx, user, cred)
	if err != nil {
		return Output{}, err
	}

	dto := usecase.MapUserToDTO(user)
	return Output{User: dto}, nil
}

func (uc Register) checkUsernameExists(ctx context.Context, username domain.Username) error {
	exists, err := uc.userExists.ExistsByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("user exists checker failed: %w", err)
	}
	if exists {
		return ErrUsernameAlreadyExists
	}
	return nil
}

func (uc Register) createUser(params domain.NewUserParams) (*domain.User, error) {
	user, err := domain.NewUser(domain.NewUserParams{
		Username:  params.Username,
		CreatedAt: params.CreatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("user creation failed: %w", err)
	}
	return user, nil
}

func (uc Register) createCredential(params domain.NewCredentialParams) (*domain.Credential, error) {
	cred, err := domain.NewCredential(params)
	if err != nil {
		return nil, fmt.Errorf("credential creation failed: %w", err)
	}
	return cred, nil
}

func (uc Register) save(ctx context.Context, user *domain.User, cred *domain.Credential) error {
	return uc.uow.Do(ctx, func(deps usecase.UowDeps) error {
		err := deps.UserWriter.Save(ctx, user)
		if err != nil {
			if errors.Is(err, domain.ErrUsernameAlreadyExists) {
				return ErrUsernameAlreadyExists
			} else {
				return fmt.Errorf("user writer save failed: %w", err)
			}
		}
		err = deps.CredentialWriter.Save(ctx, cred)
		if err != nil {
			return fmt.Errorf("credential writer save failed: %w", err)
		}
		return nil
	})
}
