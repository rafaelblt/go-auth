package e2e

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/bootstrap"
	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/credentialtest"
	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type Fixtures struct {
	pool *pgxpool.Pool
	cfg  bootstrap.Config
}

func (f *Fixtures) SaveUser(t *testing.T, usr *user.User) {
	t.Helper()
	postgrestest.InsertUser(t, f.pool, usr)
}

func (f *Fixtures) SaveCredential(t *testing.T, cred *credential.Credential) {
	t.Helper()
	postgrestest.InsertCredential(t, f.pool, cred)
}

func (f *Fixtures) SaveSession(t *testing.T, sess *session.Session) {
	t.Helper()
	postgrestest.InsertSession(t, f.pool, sess)
}

func (f *Fixtures) SaveRefreshToken(t *testing.T, token *session.RefreshToken) {
	t.Helper()
	postgrestest.InsertRefreshToken(t, f.pool, token)
}

func (f *Fixtures) CreateUserAndPassword(t *testing.T) (*user.User, credential.PlainPassword) {
	usr := usertest.NewUser(t, nil)
	f.SaveUser(t, usr)

	plain := credentialtest.MustPlainPassword(t, "as-0dfo-01@$JU90512nsd")

	hash, err := bcrypt.GenerateFromPassword([]byte(plain.Value()), f.cfg.BcryptCost)
	require.NoError(t, err, "bcrypt generate from password failed")

	secret := credentialtest.MustSecret(t, string(hash))

	cred := credentialtest.NewCredential(t, func(p *credential.RestoreParams) {
		p.UserID = usr.ID()
		p.Kind = credential.KindPassword
		p.Provider = credential.ProviderLocal
		p.Secret = secret
	})
	f.SaveCredential(t, cred)

	return usr, plain
}

func (f *Fixtures) CreateSession(t *testing.T) *session.Session {
	t.Helper()

	usr := usertest.NewUser(t, nil)
	f.SaveUser(t, usr)

	sess := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		p.UserID = usr.ID()
	})
	f.SaveSession(t, sess)

	return sess
}

func (f *Fixtures) CreateSessionRevoked(t *testing.T) *session.Session {
	t.Helper()

	usr := usertest.NewUser(t, nil)
	f.SaveUser(t, usr)

	sess := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		p.UserID = usr.ID()
		p.RevokedAt = shared.Ptr(time.Now().UTC())
	})
	f.SaveSession(t, sess)

	return sess
}

func (f *Fixtures) CreateRefreshToken(t *testing.T) (*session.RefreshToken, string) {
	t.Helper()

	sess := f.CreateSession(t)

	raw, hash := f.generateRefreshTokenAndHash(t)
	token := sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.Hash = sessiontest.MustRefreshTokenHash(t, hash)
		p.SessionID = sess.ID()
		p.ExpiresAt = time.Now().UTC().Add(time.Hour)
	})
	f.SaveRefreshToken(t, token)

	return token, raw
}

func (f *Fixtures) CreateRefreshTokenExpired(t *testing.T) (*session.RefreshToken, string) {
	t.Helper()

	sess := f.CreateSession(t)

	raw, hash := f.generateRefreshTokenAndHash(t)
	token := sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.Hash = sessiontest.MustRefreshTokenHash(t, hash)
		p.SessionID = sess.ID()
		p.ExpiresAt = time.Now().UTC()
	})
	f.SaveRefreshToken(t, token)

	return token, raw
}

func (f *Fixtures) CreateRefreshTokenAlreadyUsed(t *testing.T) (*session.RefreshToken, string) {
	t.Helper()

	sess := f.CreateSession(t)

	raw, hash := f.generateRefreshTokenAndHash(t)
	token := sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.Hash = sessiontest.MustRefreshTokenHash(t, hash)
		p.SessionID = sess.ID()
		p.ExpiresAt = time.Now().UTC().Add(time.Hour)
		p.UsedAt = shared.Ptr(time.Now().UTC())
	})
	f.SaveRefreshToken(t, token)

	return token, raw
}

func (f *Fixtures) CreateRefreshTokenWithSessionRevoked(t *testing.T) (*session.RefreshToken, string) {
	t.Helper()

	sess := f.CreateSessionRevoked(t)

	raw, hash := f.generateRefreshTokenAndHash(t)
	token := sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.Hash = sessiontest.MustRefreshTokenHash(t, hash)
		p.SessionID = sess.ID()
		p.ExpiresAt = time.Now().UTC().Add(time.Hour)
	})
	f.SaveRefreshToken(t, token)

	return token, raw
}

func (f *Fixtures) generateRefreshTokenAndHash(t *testing.T) (string, []byte) {
	token := make([]byte, 16)

	_, err := rand.Read(token)
	require.NoError(t, err, "rand read failed")

	sum := sha256.Sum256(token)

	return base64.RawURLEncoding.EncodeToString(token), sum[:]
}
