package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/config"
	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/infra/bcrypt"
	"github.com/rafaelblt/go-auth/internal/infra/jwt"
	"github.com/rafaelblt/go-auth/internal/infra/jwt/ed25519"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/infra/ratelimit"
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
	SigningKeys        *postgres.SigningKeyRepo
	PasswordHasher     *bcrypt.Hasher
	Ed25519Keyring     *ed25519.Keyring
	Ed25519Signer      *ed25519.Signer
	AccessTokenService *jwt.AccessTokenService
	RateLimiter        *ratelimit.InMemory
}

func (deps infraDeps) close() {
	if deps.pool != nil {
		deps.pool.Close()
	}
}

func newInfra(ctx context.Context, cfg config.Config) (deps infraDeps, err error) {
	defer func() {
		if err != nil {
			deps.close()
			deps = infraDeps{}
		}
	}()

	clock := infra.NewSystemClock()
	deps.Clock = clock

	pool, err := buildPool(ctx, cfg.DatabaseURL())
	if err != nil {
		return deps, err
	}
	deps.pool = pool

	uow, err := buildUnitOfWork(pool)
	if err != nil {
		return deps, err
	}
	deps.UnitOfWork = uow

	users, err := buildUserRepo(pool)
	if err != nil {
		return deps, err
	}
	deps.Users = users

	passwords, err := buildPasswordRepo(pool)
	if err != nil {
		return deps, err
	}
	deps.Passwords = passwords

	sessions, err := buildSessionRepo(pool)
	if err != nil {
		return deps, err
	}
	deps.Sessions = sessions

	refreshTokens, err := buildRefreshTokenRepo(pool)
	if err != nil {
		return deps, err
	}
	deps.RefreshTokens = refreshTokens

	signingKeys, err := buildSigningKeyRepo(pool)
	if err != nil {
		return deps, err
	}
	deps.SigningKeys = signingKeys

	hasher, err := buildPasswordHasher(cfg.BcryptCost())
	if err != nil {
		return deps, err
	}
	deps.PasswordHasher = hasher

	rateLimiter, err := buildRateLimiter(clock)
	if err != nil {
		return deps, err
	}
	deps.RateLimiter = rateLimiter

	return deps, nil
}

// buildSigning builds what signs access tokens. It runs after the schema check,
// because the keyring reads the signing keys from the database.
func (deps *infraDeps) buildSigning(ctx context.Context, cfg config.Config) error {
	keyring, err := buildEd25519Keyring(ctx, deps.SigningKeys, deps.Clock, cfg.SigningKeyEncryptionKey())
	if err != nil {
		return err
	}
	deps.Ed25519Keyring = keyring

	signer, err := buildEd25519Signer(keyring, deps.Clock)
	if err != nil {
		return err
	}
	deps.Ed25519Signer = signer

	accessTokens, err := buildAccessTokenService(signer, deps.Clock, cfg.AccessTokenTTL())
	if err != nil {
		return err
	}
	deps.AccessTokenService = accessTokens

	return nil
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

func buildSigningKeyRepo(db postgres.DB) (*postgres.SigningKeyRepo, error) {
	repo, err := postgres.NewSigningKeyRepo(db)
	if err != nil {
		return nil, fmt.Errorf("signing key repo creation failed: %w", err)
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

const (
	signingKeyRotationInterval = 7 * 24 * time.Hour
	signingKeyPublishBefore    = 24 * time.Hour
	// An hour past the longest access token, to absorb clock skew between instances.
	signingKeyPublishAfter = jwt.MaxExpiration + time.Hour
)

func buildEd25519Keyring(ctx context.Context, store port.SigningKeyStore, clock port.Clock, encryptionKey []byte) (*ed25519.Keyring, error) {
	keyring, err := ed25519.NewKeyring(ctx, ed25519.KeyringConfig{
		KeyStore:         store,
		Clock:            clock,
		RotationInterval: signingKeyRotationInterval,
		PublishBefore:    signingKeyPublishBefore,
		PublishAfter:     signingKeyPublishAfter,
		EncryptionKey:    encryptionKey,
	})
	if err != nil {
		return nil, fmt.Errorf("new ed25519 keyring failed: %w", err)
	}
	return keyring, nil
}

func logSigningKeyEncryption(logger *slog.Logger, cfg config.Config) {
	if len(cfg.SigningKeyEncryptionKey()) == 0 {
		logger.Warn("signing keys are stored unencrypted; from v2 SIGNING_KEY_ENCRYPTION_KEY is required")
	}
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

func buildRateLimiter(clock port.Clock) (*ratelimit.InMemory, error) {
	limiter, err := ratelimit.NewInMemory(ratelimit.Config{Clock: clock})
	if err != nil {
		return nil, fmt.Errorf("rate limiter creation failed: %w", err)
	}
	return limiter, nil
}
