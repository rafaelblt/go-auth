package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/domain"
)

type Register struct {
	userExists UserExistsChecker
	userWriter UserWriter
	credWriter CredentialWriter
	hasher     PasswordHasher
	clock      Clock
}

type RegisterInput struct {
	Username string
	Password string
}

type RegisterOutput struct {
	User *domain.User
}

type RegisterConfig struct {
	UserExistsChecker UserExistsChecker
	UserWriter        UserWriter
	CredentialWriter  CredentialWriter
	PasswordHasher    PasswordHasher
	Clock             Clock
}

var ErrRegisterUsernameTooShort = errors.New("the provided username is too short")
var ErrRegisterUsernameTooLong = errors.New("the provided username is too long")
var ErrRegisterPasswordTooShort = errors.New("the provided password is too short")
var ErrRegisterPasswordTooLong = errors.New("the provided password is too long")
var ErrRegisterUsernameTaken = errors.New("the provided username is already registered")

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
	if config.UserWriter == nil {
		return Register{}, errors.New("user writer cannot be nil")
	}
	if config.CredentialWriter == nil {
		return Register{}, errors.New("credential writer cannot be nil")
	}
	if config.PasswordHasher == nil {
		return Register{}, errors.New("password hasher cannot be nil")
	}
	if config.Clock == nil {
		return Register{}, errors.New("clock cannot be nil")
	}
	uc := Register{
		userExists: config.UserExistsChecker,
		userWriter: config.UserWriter,
		credWriter: config.CredentialWriter,
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

	user, err := uc.createUser(ctx, domain.NewUserParams{
		Username:  username,
		CreatedAt: now,
	})
	if err != nil {
		return RegisterOutput{}, err
	}

	_, err = uc.createCredential(ctx, domain.NewCredentialParams{
		UserID:    user.ID(),
		Kind:      domain.CredentialKindPassword,
		Provider:  domain.CredentialProviderLocal,
		Secret:    hashed,
		CreatedAt: now,
	})
	if err != nil {
		return RegisterOutput{}, err
	}

	return RegisterOutput{User: user}, nil
}

func (uc Register) validateFields(username string, password string) (domain.Username, domain.PlainPassword, error) {
	usernameErrs := domain.ValidateUsername(username)
	passwordErrs := domain.ValidatePlainPassword(password)

	if len(usernameErrs) == 0 && len(passwordErrs) == 0 {
		usr, err := domain.NewUsername(username)
		if err != nil {
			panic("new username with errors after success validation")
		}
		pwd, err := domain.NewPlainPassword(password)
		if err != nil {
			panic("new plain password with errors after success validation")
		}
		return usr, pwd, nil
	}

	allErrs := append(usernameErrs, passwordErrs...)
	mappeds, unexpecteds := MapErrors(allErrs, registerValidationMap)

	var err error
	if len(unexpecteds) > 0 {
		err = errors.Join(unexpecteds...)
	} else {
		err = errors.Join(mappeds...)
	}

	return domain.Username{}, domain.PlainPassword{}, err
}

func (uc Register) validateInput(input RegisterInput) error {
	usernameErrs := domain.ValidateUsername(input.Username)
	passwordErrs := domain.ValidatePlainPassword(input.Password)

	allErrs := append(usernameErrs, passwordErrs...)
	mappeds, unexpecteds := MapErrors(allErrs, registerValidationMap)

	var err error = nil
	if len(unexpecteds) > 0 {
		err = errors.Join(unexpecteds...)
	} else {
		err = errors.Join(mappeds...)
	}

	return err
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
		return ErrRegisterUsernameTaken
	}
	return nil
}

func (uc Register) createUser(ctx context.Context, params domain.NewUserParams) (*domain.User, error) {
	user, err := domain.NewUser(domain.NewUserParams{
		Username:  params.Username,
		CreatedAt: params.CreatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("user creation failed: %w", err)
	}
	err = uc.userWriter.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("user writer save failed: %w", err)
	}
	return user, nil
}

func (uc Register) createCredential(ctx context.Context, params domain.NewCredentialParams) (*domain.Credential, error) {
	cred, err := domain.NewAuthCredential(params)
	if err != nil {
		return nil, fmt.Errorf("credential creation failed: %w", err)
	}
	err = uc.credWriter.Save(ctx, cred)
	if err != nil {
		return nil, fmt.Errorf("credential writer save failed: %w", err)
	}
	return cred, nil
}
