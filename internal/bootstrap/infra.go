package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/config"
	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/infra/bcrypt"
	"github.com/rafaelblt/go-auth/internal/infra/jwt"
	"github.com/rafaelblt/go-auth/internal/infra/jwt/ed25519"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/port"
)

type infraDeps struct {
	pool *pgxpool.Pool

	Clock              *infra.SystemClock
	UnitOfWork         *postgres.UnitOfWork
	Users              *postgres.UserRepo
	Passwords          *postgres.PasswordRepo
	Sessions           *postgres.SessionRepo
	RefreshTokens      *postgres.RefreshTokenRepo
	PasswordHasher     *bcrypt.Hasher
	Ed25519KeyStore    *ed25519.KeyStoreInMemory
	Ed25519Keyring     *ed25519.Keyring
	Ed25519Signer      *ed25519.Signer
	AccessTokenService *jwt.AccessTokenService
}

func newInfra(ctx context.Context, cfg config.Config) (infraDeps, error) {
	deps := infraDeps{}

	clock := infra.NewSystemClock()
	deps.Clock = clock

	pool, err := buildPool(ctx, cfg.DatabaseURL())
	if err != nil {
		return infraDeps{}, err
	}
	deps.pool = pool

	uow, err := buildUnitOfWork(pool)
	if err != nil {
		return infraDeps{}, err
	}
	deps.UnitOfWork = uow

	users, err := buildUserRepo(pool)
	if err != nil {
		return infraDeps{}, err
	}
	deps.Users = users

	passwords, err := buildPasswordRepo(pool)
	if err != nil {
		return infraDeps{}, err
	}
	deps.Passwords = passwords

	sessions, err := buildSessionRepo(pool)
	if err != nil {
		return infraDeps{}, err
	}
	deps.Sessions = sessions

	refreshTokens, err := buildRefreshTokenRepo(pool)
	if err != nil {
		return infraDeps{}, err
	}
	deps.RefreshTokens = refreshTokens

	hasher, err := buildPasswordHasher(cfg.BcryptCost())
	if err != nil {
		return infraDeps{}, err
	}
	deps.PasswordHasher = hasher

	keystore := ed25519.NewKeyStoreInMemory()
	deps.Ed25519KeyStore = keystore

	keyring, err := buildEd25519Keyring(ctx, keystore)
	if err != nil {
		return infraDeps{}, err
	}
	deps.Ed25519Keyring = keyring

	signer, err := buildEd25519Signer(keyring, clock)
	if err != nil {
		return infraDeps{}, err
	}
	deps.Ed25519Signer = signer

	accessTokens, err := buildAccessTokenService(signer, clock, cfg.AccessTokenTTL())
	if err != nil {
		return infraDeps{}, err
	}
	deps.AccessTokenService = accessTokens

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

func buildPasswordRepo(db postgres.DB) (*postgres.PasswordRepo, error) {
	repo, err := postgres.NewPasswordRepo(db)
	if err != nil {
		return nil, fmt.Errorf("password repo creation failed: %w", err)
	}
	return repo, nil
}

func buildSessionRepo(db postgres.DB) (*postgres.SessionRepo, error) {
	repo, err := postgres.NewSessionRepo(db)
	if err != nil {
		return nil, fmt.Errorf("session repo creation failed: %w", err)
	}
	return repo, nil
}

func buildRefreshTokenRepo(db postgres.DB) (*postgres.RefreshTokenRepo, error) {
	repo, err := postgres.NewRefreshTokenRepo(db)
	if err != nil {
		return nil, fmt.Errorf("refresh token repo creation failed: %w", err)
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

func buildEd25519Keyring(ctx context.Context, keyStore ed25519.KeyStore) (*ed25519.Keyring, error) {
	keyring, err := ed25519.NewKeyring(ctx, ed25519.KeyringConfig{
		KeyStore: keyStore,
	})
	if err != nil {
		return nil, fmt.Errorf("new ed25519 keyring failed: %w", err)
	}
	return keyring, nil
}

func buildEd25519Signer(keyring *ed25519.Keyring, clock port.Clock) (*ed25519.Signer, error) {
	signer, err := ed25519.NewSigner(ed25519.SignerConfig{
		Keyring: keyring,
		Clock:   clock,
	})
	if err != nil {
		return nil, fmt.Errorf("new ed25519 signer failed: %w", err)
	}
	return signer, nil
}

func buildAccessTokenService(signer jwt.Signer, clock port.Clock, exp time.Duration) (*jwt.AccessTokenService, error) {
	service, err := jwt.NewAccessTokenService(jwt.AccessTokenServiceConfig{
		Signer:     signer,
		Clock:      clock,
		Expiration: exp,
	})
	if err != nil {
		return nil, fmt.Errorf("new access token service failed: %w", err)
	}
	return service, nil
}
