package credentialtest

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/require"
)

func NewCredential(t *testing.T, override func(*credential.RestoreParams)) *credential.Credential {
	t.Helper()

	secret := MustSecret(t, "default secret")

	params := credential.RestoreParams{
		ID:        credential.NewID(),
		UserID:    user.NewID(),
		Kind:      credential.KindPassword,
		Provider:  credential.ProviderLocal,
		Secret:    secret,
		CreatedAt: time.Date(2007, 8, 9, 20, 45, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 21, 16, 0, 7, 0, time.UTC),
	}

	if override != nil {
		override(&params)
	}

	entity, err := credential.RestoreCredential(params)
	require.NoError(t, err, "credential restore failed")
	return entity
}

type CredentialOption func(*credential.RestoreParams)

func WithUserID(id user.ID) CredentialOption {
	return func(p *credential.RestoreParams) {
		p.UserID = id
	}
}

func PasswordCredential(t *testing.T, opts ...CredentialOption) *credential.Credential {
	t.Helper()

	secret, err := credential.NewSecret("aspovkjpok3q-9r8i4-asikf0opç")
	require.NoError(t, err)

	params := credential.RestoreParams{
		ID:        credential.NewID(),
		UserID:    user.NewID(),
		Kind:      credential.KindPassword,
		Provider:  credential.ProviderLocal,
		Secret:    secret,
		CreatedAt: time.Date(2026, 3, 25, 17, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 3, 25, 17, 0, 0, 0, time.UTC),
	}

	for _, opt := range opts {
		opt(&params)
	}

	entity, err := credential.RestoreCredential(params)
	require.NoError(t, err)

	return entity
}
