package usecase

import (
	"context"
	"errors"

	"github.com/rafaelblt/go-auth/internal/domain"
)

type Register struct {
	userExists UserExistsChecker
	userSaver  UserSaver
	credSaver  UserCredentialsSaver
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
	UserExistsChecker    UserExistsChecker
	UserSaver            UserSaver
	UserCredentialsSaver UserCredentialsSaver
	PasswordHasher       PasswordHasher
	Clock                Clock
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
	if config.UserSaver == nil {
		return Register{}, errors.New("user saver cannot be nil")
	}
	if config.UserCredentialsSaver == nil {
		return Register{}, errors.New("user credentials saver cannot be nil")
	}
	if config.PasswordHasher == nil {
		return Register{}, errors.New("password hasher cannot be nil")
	}
	if config.Clock == nil {
		return Register{}, errors.New("clock cannot be nil")
	}
	return Register{
		userExists: config.UserExistsChecker,
		userSaver:  config.UserSaver,
		credSaver:  config.UserCredentialsSaver,
		hasher:     config.PasswordHasher,
		clock:      config.Clock,
	}, nil
}

func (uc Register) Execute(ctx context.Context, input RegisterInput) (RegisterOutput, error) {
	username, password, err := uc.validateFields(input.Username, input.Password)
	if err != nil {
		return RegisterOutput{}, err
	}
	exists, err := uc.userExists.ExistsByUsername(ctx, username)
	if err != nil {
		return RegisterOutput{}, err
	}
	if exists {
		return RegisterOutput{}, ErrRegisterUsernameTaken
	}
	hashedPwd, err := uc.hasher.Hash(password)
	if err != nil {
		return RegisterOutput{}, err
	}
	user, credentials, err := domain.NewUserWithPassword(
		username, hashedPwd, uc.clock.UtcNow(),
	)
	err = uc.userSaver.Save(ctx, user)
	if err != nil {
		return RegisterOutput{}, err
	}
	err = uc.credSaver.Save(ctx, credentials)
	if err != nil {
		return RegisterOutput{}, err
	}
	return RegisterOutput{User: user}, nil
}

func (uc Register) validateFields(
	username string, password string,
) (domain.Username, domain.PlainPassword, error) {
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
