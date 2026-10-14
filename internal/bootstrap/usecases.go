package bootstrap

import (
	"fmt"

	"github.com/rafaelblt/go-auth/internal/config"
	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/port"

	"github.com/rafaelblt/go-auth/internal/usecase/changepassword"
	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/rafaelblt/go-auth/internal/usecase/sessioncleanup"
	"github.com/rafaelblt/go-auth/internal/usecase/verify"
)

type usecases struct {
	Register       *register.Register
	Login          *login.Login
	Refresh        *refresh.Refresh
	Verify         *verify.Verify
	ChangePassword *changepassword.ChangePassword
	// SessionCleanup is nil while SESSION_RETENTION is off.
	SessionCleanup *sessioncleanup.SessionCleanup
}

func newUsecases(cfg config.Config, deps infraDeps) (usecases, error) {
	regst, err := buildRegister(deps)
	if err != nil {
		return usecases{}, err
	}

	dummyHash, err := newDummyPasswordHash(deps.PasswordHasher)
	if err != nil {
		return usecases{}, err
	}

	logn, err := buildLogin(cfg, deps, dummyHash)
	if err != nil {
		return usecases{}, err
	}

	refr, err := buildRefresh(cfg, deps)
	if err != nil {
		return usecases{}, err
	}

	verf, err := buildVerify(deps)
	if err != nil {
		return usecases{}, err
	}

	chpw, err := buildChangePassword(deps, dummyHash)
	if err != nil {
		return usecases{}, err
	}

	clean, err := buildSessionCleanup(cfg, deps)
	if err != nil {
		return usecases{}, err
	}

	uc := usecases{
		Register:       &regst,
		Login:          &logn,
		Refresh:        refr,
		Verify:         verf,
		ChangePassword: &chpw,
		SessionCleanup: clean,
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

func buildLogin(cfg config.Config, deps infraDeps, dummyHash password.Hashed) (login.Login, error) {
	uc, err := login.New(login.Config{
		UserReader:        deps.Users,
		PasswordReader:    deps.Passwords,
		PasswordChecker:   deps.PasswordHasher,
		AccessTokenIssuer: deps.AccessTokenService,
		UnitOfWork:        deps.UnitOfWork,
		Clock:             deps.Clock,
		RefreshTokenTTL:   cfg.RefreshTokenTTL(),
		DummyPasswordHash: dummyHash,
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
		SessionReader:      deps.Sessions,
		AccessTokenIssuer:  deps.AccessTokenService,
		RefreshTokenReader: deps.RefreshTokens,
		UnitOfWork:         deps.UnitOfWork,
		Clock:              deps.Clock,
		RefreshTokenTTL:    cfg.RefreshTokenTTL(),
	})
	if err != nil {
		return nil, fmt.Errorf("refresh creation failed: %w", err)
	}
	return uc, nil
}

func buildVerify(deps infraDeps) (*verify.Verify, error) {
	uc, err := verify.New(verify.Config{
		AccessTokenValidator: deps.AccessTokenService,
		Clock:                deps.Clock,
	})
	if err != nil {
		return nil, fmt.Errorf("verify creation failed: %w", err)
	}
	return uc, nil
}

func buildChangePassword(deps infraDeps, dummyHash password.Hashed) (changepassword.ChangePassword, error) {
	uc, err := changepassword.New(changepassword.Config{
		UserReader:        deps.Users,
		PasswordReader:    deps.Passwords,
		PasswordChecker:   deps.PasswordHasher,
		PasswordHasher:    deps.PasswordHasher,
		UnitOfWork:        deps.UnitOfWork,
		Clock:             deps.Clock,
		DummyPasswordHash: dummyHash,
	})
	if err != nil {
		return changepassword.ChangePassword{}, fmt.Errorf("change password creation failed: %w", err)
	}
	return uc, nil
}

// One run of the session cleanup deletes at most sessionCleanupMaxBatches
// batches of sessionCleanupBatchSize sessions, each in a statement of its own.
const (
	sessionCleanupBatchSize  = 1000
	sessionCleanupMaxBatches = 10
)

// buildSessionCleanup returns nil when SESSION_RETENTION is off.
func buildSessionCleanup(cfg config.Config, deps infraDeps) (*sessioncleanup.SessionCleanup, error) {
	retention, on := cfg.SessionRetention().Duration()
	if !on {
		return nil, nil
	}

	uc, err := sessioncleanup.New(sessioncleanup.Config{
		SessionCleaner: deps.Sessions,
		Clock:          deps.Clock,
		Retention:      retention,
		BatchSize:      sessionCleanupBatchSize,
		MaxBatches:     sessionCleanupMaxBatches,
	})
	if err != nil {
		return nil, fmt.Errorf("session cleanup creation failed: %w", err)
	}
	return uc, nil
}
