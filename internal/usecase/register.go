package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/domain"
)

type Register struct {
	userExists UserExistsChecker
	uow        UnitOfWork
	hasher     PasswordHasher
	clock      Clock
}

type RegisterInput struct {
	Username string
	Password string
}

type RegisterOutput struct {
	User UserDTO
}

type RegisterConfig struct {
	UserExistsChecker UserExistsChecker
	UnitOfWork        UnitOfWork
	PasswordHasher    PasswordHasher
	Clock             Clock
}

var ErrRegisterUsernameTooShort = errors.New("the provided username is too short")
var ErrRegisterUsernameTooLong = errors.New("the provided username is too long")
var ErrRegisterPasswordTooShort = errors.New("the provided password is too short")
var ErrRegisterPasswordTooLong = errors.New("the provided password is too long")
var ErrRegisterUsernameAlreadyExists = errors.New("the provided username is already registered")

var registerValidationMap = ErrorsMap{
	domain.ErrUsernameTooShort:      ErrRegisterUsernameTooShort,
	domain.ErrUsernameTooLong:       ErrRegisterUsernameTooLong,
	domain.ErrPlainPasswordTooShort: ErrRegisterPasswordTooShort,
	domain.ErrPlainPasswordTooLong:  ErrRegisterPasswordTooLong,
}

func NewRegister(config RegisterConfig) (Register, error) {
	if config.UserExistsChecker == nil {
		return Register{}, errors.New("user exists checker cannot be nil")
	}
	if config.UnitOfWork == nil {
		return Register{}, errors.New("unit of work cannot be nil")
	}
	if config.PasswordHasher == nil {
		return Register{}, errors.New("password hasher cannot be nil")
	}
	if config.Clock == nil {
		return Register{}, errors.New("clock cannot be nil")
	}
	uc := Register{
		userExists: config.UserExistsChecker,
		uow:        config.UnitOfWork,
		hasher:     config.PasswordHasher,
		clock:      config.Clock,
	}
	return uc, nil
}

func (uc Register) Execute(ctx context.Context, input RegisterInput) (RegisterOutput, error) {
	err := uc.validateInput(input)
	if err != nil {
		return RegisterOutput{}, err
	}

	username := uc.convertUsername(input.Username)
	password := uc.convertPassword(input.Password)

	err = uc.checkUsernameExists(ctx, username)
	if err != nil {
		return RegisterOutput{}, err
	}

	hashed, err := uc.hasher.Hash(password)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("password hasher failed: %w", err)
	}

	now := uc.clock.UtcNow()

	user, err := uc.createUser(domain.NewUserParams{
		Username:  username,
		CreatedAt: now,
	})
	if err != nil {
		return RegisterOutput{}, err
	}
	cred, err := uc.createCredential(domain.NewCredentialParams{
		UserID:    user.ID(),
		Kind:      domain.CredentialKindPassword,
		Provider:  domain.CredentialProviderLocal,
		Secret:    hashed,
		CreatedAt: now,
	})
	if err != nil {
		return RegisterOutput{}, err
	}

	err = uc.save(ctx, user, cred)
	if err != nil {
		return RegisterOutput{}, err
	}

	dto, err := MapUserToDTO(user)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("user dto mapping failed: %w", err)
	}

	return RegisterOutput{User: dto}, nil
}

func (uc Register) validateInput(input RegisterInput) error {
	usernameErrs := domain.ValidateUsername(input.Username)
	passwordErrs := domain.ValidatePlainPassword(input.Password)

	allErrs := append(usernameErrs, passwordErrs...)
	mappeds, unexpecteds := MapErrors(allErrs, registerValidationMap)

	if len(unexpecteds) > 0 {
		return fmt.Errorf("unexpected validation errors from domain: %w", errors.Join(unexpecteds...))
	}
	if len(mappeds) > 0 {
		return newValidationError(mappeds...)
	}
	return nil
}

func (uc Register) convertUsername(username string) domain.Username {
	converted, err := domain.NewUsername(username)
	if err != nil {
		panic(fmt.Sprintf("username here should be valid, but it contains an error: %v", err))
	}
	return converted
}

func (uc Register) convertPassword(password string) domain.PlainPassword {
	converted, err := domain.NewPlainPassword(password)
	if err != nil {
		panic(fmt.Sprintf("plain password here should be valid, but it contains an error: %v", err))
	}
	return converted
}

func (uc Register) checkUsernameExists(ctx context.Context, username domain.Username) error {
	exists, err := uc.userExists.ExistsByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("user exists checker failed: %w", err)
	}
	if exists {
		return ErrRegisterUsernameAlreadyExists
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
	return uc.uow.Do(ctx, func(deps UowDeps) error {
		err := deps.UserWriter.Save(ctx, user)
		if err != nil {
			if errors.Is(err, domain.ErrUsernameAlreadyExists) {
				return ErrRegisterUsernameAlreadyExists
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
