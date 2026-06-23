package e2e

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/bootstrap"
	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/testutil/credentialtest"
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
	require.NotNil(t, usr, "user nil")
	require.NotZero(t, usr, "user zero")

	sql := `INSERT INTO users
			(id, username, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)`

	_, err := f.pool.Exec(context.Background(), sql,
		usr.ID().Value(),
		usr.Username().String(),
		usr.Status().String(),
		usr.CreatedAt(),
		usr.UpdatedAt(),
	)

	require.NoError(t, err)
}

func (f *Fixtures) SaveCredential(t *testing.T, cred *credential.Credential) {
	require.NotNil(t, cred, "credential nil")
	require.False(t, cred.IsZero(), "credential zero")

	sql := `INSERT INTO credentials
			(id, user_id, kind, provider, secret, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := f.pool.Exec(context.Background(), sql,
		cred.ID().Value(),
		cred.UserID().Value(),
		cred.Kind().String(),
		cred.Provider().String(),
		cred.Secret().Value(),
		cred.CreatedAt(),
		cred.UpdatedAt(),
	)

	require.NoError(t, err, "insert credential failed")
}

func (f *Fixtures) GetUserAndPassword(t *testing.T) (*user.User, credential.PlainPassword) {
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
