package bootstrap

import (
	"fmt"

	"github.com/rafaelblt/go-auth/internal/config"
	"github.com/rafaelblt/go-auth/internal/password"
	"github.com/rafaelblt/go-auth/internal/port"

	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
)

type usecases struct {
	Register *register.Register
	Login    *login.Login
	Refresh  *refresh.Refresh
}

func newUsecases(cfg config.Config, deps infraDeps) (usecases, error) {
	regst, err := buildRegister(deps)
	if err != nil {
		return usecases{}, err
	}

	logn, err := buildLogin(cfg, deps)
	if err != nil {
		return usecases{}, err
	}

	refr, err := buildRefresh(cfg, deps)
	if err != nil {
		return usecases{}, err
	}

	uc := usecases{
		Register: &regst,
		Login:    &logn,
		Refresh:  refr,
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

func buildLogin(cfg config.Config, deps infraDeps) (login.Login, error) {
	dummyHash, err := newDummyPasswordHash(deps.PasswordHasher)
	if err != nil {
		return login.Login{}, err
	}

	uc, err := login.New(login.Config{
		UserReader:            deps.Users,
		PasswordReader:        deps.Passwords,
		PasswordChecker:       deps.PasswordHasher,
		AccessTokenIssuer:     deps.AccessTokenService,
		RefreshTokenGenerator: deps.RefreshTokenGenerator,
		UnitOfWork:            deps.UnitOfWork,
		Clock:                 deps.Clock,
		RefreshTokenTTL:       cfg.RefreshTokenTTL(),
		DummyPasswordHash:     dummyHash,
	})
	if err != nil {
		return login.Login{}, fmt.Errorf("login creation failed: %w", err)
	}
	return uc, nil
}

// newDummyPasswordHash hashes a fixed password with the configured hasher, so
// verifying it costs the same as verifying a stored hash.
func newDummyPasswordHash(hasher port.PasswordHasher) (password.Hashed, error) {
	plain, issues := password.NewPlain("go-auth-dummy-password")
	if !issues.IsEmpty() {
		return password.Hashed{}, fmt.Errorf("dummy plain password is invalid: %s", issues)
	}

	hash, err := hasher.Hash(plain)
	if err != nil {
		return password.Hashed{}, fmt.Errorf("dummy password hashing failed: %w", err)
	}
	return hash, nil
}

func buildRefresh(cfg config.Config, deps infraDeps) (*refresh.Refresh, error) {
	uc, err := refresh.New(refresh.Config{
		SessionReader:         deps.Sessions,
		AccessTokenIssuer:     deps.AccessTokenService,
		RefreshTokenResolver:  deps.RefreshTokenResolver,
		RefreshTokenGenerator: deps.RefreshTokenGenerator,
		UnitOfWork:            deps.UnitOfWork,
		Clock:                 deps.Clock,
		RefreshTokenTTL:       cfg.RefreshTokenTTL(),
	})
	if err != nil {
		return nil, fmt.Errorf("refresh creation failed: %w", err)
	}
	return uc, nil
}
