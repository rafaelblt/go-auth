package bootstrap

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/infra/bcrypt"
	"github.com/rafaelblt/go-auth/internal/infra/jwt"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/infra/refreshtoken"
	"github.com/rafaelblt/go-auth/internal/port"
)

type infraDeps struct {
	Clock                 *infra.SystemClock
	UnitOfWork            *postgres.UnitOfWork
	Users                 *postgres.UserRepo
	Credentials           *postgres.CredentialRepo
	PasswordHasher        *bcrypt.Hasher
	AccessTokenService    *jwt.AccessTokenService
	RefreshTokenGenerator *refreshtoken.Generator
}

func newInfra(ctx context.Context, cfg Config) (infraDeps, error) {
	clock := infra.NewSystemClock()

	pool, err := buildPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return infraDeps{}, err
	}

	uow, err := buildUnitOfWork(pool)
	if err != nil {
		return infraDeps{}, err
	}

	users, err := buildUserRepo(pool)
	if err != nil {
		return infraDeps{}, err
	}

	creds, err := buildCredentialRepo(pool)
	if err != nil {
		return infraDeps{}, err
	}

	hasher, err := buildPasswordHasher(cfg.BcryptCost)
	if err != nil {
		return infraDeps{}, err
	}

	accessTokens, err := buildAccessTokenService(clock)
	if err != nil {
		return infraDeps{}, err
	}

	refreshGenerator, err := buildRefreshTokenGenerator()
	if err != nil {
		return infraDeps{}, err
	}

	deps := infraDeps{
		Clock:                 clock,
		UnitOfWork:            uow,
		Users:                 users,
		Credentials:           creds,
		PasswordHasher:        hasher,
		AccessTokenService:    accessTokens,
		RefreshTokenGenerator: refreshGenerator,
	}

	return deps, nil
}

func buildPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	pool, err := postgres.NewPool(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("pgx pool creation failed: %w", err)
	}
	return pool, nil
}

func buildUnitOfWork(beginner postgres.TxBeginner) (*postgres.UnitOfWork, error) {
	uow, err := postgres.NewUnitOfWork(beginner)
	if err != nil {
		return nil, fmt.Errorf("unit of work creation failed: %w", err)
	}
	return uow, nil
}

func buildUserRepo(db postgres.DB) (*postgres.UserRepo, error) {
	repo, err := postgres.NewUserRepo(db)
	if err != nil {
		return nil, fmt.Errorf("user repo creation failed: %w", err)
	}
	return repo, nil
}

func buildCredentialRepo(db postgres.DB) (*postgres.CredentialRepo, error) {
	repo, err := postgres.NewCredentialRepo(db)
	if err != nil {
		return nil, fmt.Errorf("credential repo creation failed: %w", err)
	}
	return repo, nil
}

func buildPasswordHasher(cost int) (*bcrypt.Hasher, error) {
	hasher, err := bcrypt.NewHasher(bcrypt.Config{Cost: cost})
	if err != nil {
		return nil, fmt.Errorf("bcrypt hasher creation failed: %w", err)
	}
	return hasher, nil
}

func buildAccessTokenService(clock port.Clock) (*jwt.AccessTokenService, error) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("ed25519 generate key failed: %w", err)
	}

	signer, err := jwt.NewEd25519(jwt.Ed25519Config{
		PrivateKey: privateKey,
		Clock:      clock,
	})
	if err != nil {
		return nil, fmt.Errorf("Ed25519Signer creation failed: %w", err)
	}

	service, err := jwt.NewAccessTokenService(jwt.AccessTokenServiceConfig{
		Signer: signer,
		Clock:  clock,
	})
	if err != nil {
		return nil, fmt.Errorf("AccessService creation failed: %w", err)
	}

	return service, nil
}

func buildRefreshTokenGenerator() (*refreshtoken.Generator, error) {
	generator := refreshtoken.NewGenerator()
	return generator, nil
}
