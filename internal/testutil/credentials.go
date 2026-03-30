package testutil

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/require"
)

type CredentialOption func(*domain.CredentialRestoreParams)

func WithUserID(id domain.UserID) CredentialOption {
    return func(p *domain.CredentialRestoreParams) {
        p.UserID = id
    }
}

func PasswordCredential(t *testing.T, opts ...CredentialOption) *domain.Credential {
	t.Helper()

	secret, err := domain.NewCredentialSecret("aspovkjpok3q-9r8i4-asikf0opç")
	require.NoError(t, err)

	params := domain.CredentialRestoreParams{
		ID:        domain.NewCredentialID(),
		UserID:    domain.NewUserID(),
		Kind:      domain.CredentialKindPassword,
		Provider:  domain.CredentialProviderLocal,
		Secret:    secret,
		CreatedAt: time.Date(2026, 3, 25, 17, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 3, 25, 17, 0, 0, 0, time.UTC),
	}

	for _, opt := range opts {
        opt(&params)
    }

	entity, err := domain.RestoreCredential(params)
	require.NoError(t, err)

	return entity
}
