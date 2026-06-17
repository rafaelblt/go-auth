package bootstrap

import (
	"fmt"

	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
)

type usecases struct {
	Register *register.Register
	Login    *login.Login
}

func newUsecases(cfg Config, deps infraDeps) (usecases, error) {
	regst, err := buildRegister(deps)
	if err != nil {
		return usecases{}, err
	}

	logn, err := buildLogin(cfg, deps)
	if err != nil {
		return usecases{}, err
	}

	uc := usecases{
		Register: &regst,
		Login:    &logn,
	}
	return uc, nil
}

func buildRegister(deps infraDeps) (register.Register, error) {
	uc, err := register.New(register.Config{
		UserExistsChecker: deps.Users,
		UnitOfWork:        deps.UnitOfWork,
		PasswordHasher:    deps.PasswordHasher,
		Clock:             deps.Clock,
	})
	if err != nil {
		return register.Register{}, fmt.Errorf("register creation failed: %w", err)
	}
	return uc, nil
}

func buildLogin(cfg Config, deps infraDeps) (login.Login, error) {
	uc, err := login.New(login.Config{
		UserReader:            deps.Users,
		CredentialReader:      deps.Credentials,
		PasswordChecker:       deps.PasswordHasher,
		AccessTokenIssuer:     deps.AccessTokenService,
		RefreshTokenGenerator: deps.RefreshTokenGenerator,
		UnitOfWork:            deps.UnitOfWork,
		Clock:                 deps.Clock,
		RefreshTokenTTL:       cfg.RefreshTokenTTL,
	})
	if err != nil {
		return login.Login{}, fmt.Errorf("login creation failed: %w", err)
	}
	return uc, nil
}
