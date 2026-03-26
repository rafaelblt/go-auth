package domain_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCredential(t *testing.T) {
	userID := domain.NewUserID()
	secret, err := domain.NewCredentialSecret("secret")
	require.NoError(t, err)
	testCases := []struct {
		desc      string
		params    domain.NewCredentialParams
		expectErr bool
	}{
		{
			desc:      "all params zero",
			params:    domain.NewCredentialParams{},
			expectErr: true,
		},
		{
			desc: "user id zero",
			params: domain.NewCredentialParams{
				UserID:    domain.UserID{},
				Kind:      domain.CredentialKindPassword,
				Provider:  domain.CredentialProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "kind zero",
			params: domain.NewCredentialParams{
				UserID:    userID,
				Kind:      "",
				Provider:  domain.CredentialProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "provider zero",
			params: domain.NewCredentialParams{
				UserID:    userID,
				Kind:      domain.CredentialKindPassword,
				Provider:  domain.CredentialProvider{},
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "secret zero",
			params: domain.NewCredentialParams{
				UserID:    userID,
				Kind:      domain.CredentialKindPassword,
				Provider:  domain.CredentialProviderLocal,
				Secret:    domain.CredentialSecret{},
				CreatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "valid case",
			params: domain.NewCredentialParams{
				UserID:    userID,
				Kind:      domain.CredentialKindPassword,
				Provider:  domain.CredentialProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
			},
			expectErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			credential, err := domain.NewCredential(tC.params)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, credential)
			} else {
				require.NoError(t, err)
				require.NotZero(t, credential)
				assert.NotZero(t, credential.ID())
				assert.Equal(t, credential.UserID(), tC.params.UserID)
				assert.Equal(t, credential.Kind(), tC.params.Kind)
				assert.Equal(t, credential.Provider(), tC.params.Provider)
				assert.Equal(t, credential.Secret(), tC.params.Secret)
				assert.Equal(t, credential.CreatedAt(), tC.params.CreatedAt)
				assert.Equal(t, credential.UpdatedAt(), tC.params.CreatedAt)
			}
		})
	}
}

func TestRestoreCredential(t *testing.T) {
	secret, err := domain.NewCredentialSecret("secret")
	require.NoError(t, err)
	testCases := []struct {
		desc      string
		params    domain.CredentialRestoreParams
		expectErr bool
	}{
		{
			desc:      "all params zero",
			params:    domain.CredentialRestoreParams{},
			expectErr: true,
		},
		{
			desc: "id zero",
			params: domain.CredentialRestoreParams{
				ID:        domain.CredentialID{},
				UserID:    domain.NewUserID(),
				Kind:      domain.CredentialKindPassword,
				Provider:  domain.CredentialProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "user id zero",
			params: domain.CredentialRestoreParams{
				ID:        domain.NewCredentialID(),
				UserID:    domain.UserID{},
				Kind:      domain.CredentialKindPassword,
				Provider:  domain.CredentialProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "kind invalid",
			params: domain.CredentialRestoreParams{
				ID:        domain.NewCredentialID(),
				UserID:    domain.NewUserID(),
				Kind:      "",
				Provider:  domain.CredentialProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "provider zero",
			params: domain.CredentialRestoreParams{
				ID:        domain.NewCredentialID(),
				UserID:    domain.NewUserID(),
				Kind:      domain.CredentialKindPassword,
				Provider:  domain.CredentialProvider{},
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "secret zero",
			params: domain.CredentialRestoreParams{
				ID:        domain.NewCredentialID(),
				UserID:    domain.NewUserID(),
				Kind:      domain.CredentialKindPassword,
				Provider:  domain.CredentialProviderLocal,
				Secret:    domain.CredentialSecret{},
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "created at zero",
			params: domain.CredentialRestoreParams{
				ID:        domain.NewCredentialID(),
				UserID:    domain.NewUserID(),
				Kind:      domain.CredentialKindPassword,
				Provider:  domain.CredentialProviderLocal,
				Secret:    secret,
				CreatedAt: time.Time{},
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "updated at zero",
			params: domain.CredentialRestoreParams{
				ID:        domain.NewCredentialID(),
				UserID:    domain.NewUserID(),
				Kind:      domain.CredentialKindPassword,
				Provider:  domain.CredentialProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Time{},
			},
			expectErr: true,
		},
		{
			desc: "valid case",
			params: domain.CredentialRestoreParams{
				ID:        domain.NewCredentialID(),
				UserID:    domain.NewUserID(),
				Kind:      domain.CredentialKindPassword,
				Provider:  domain.CredentialProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			credential, err := domain.RestoreCredential(tC.params)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, credential)
			} else {
				require.NoError(t, err)
				require.NotZero(t, credential)
				assert.Equal(t, credential.ID(), tC.params.ID)
				assert.Equal(t, credential.UserID(), tC.params.UserID)
				assert.Equal(t, credential.Kind(), tC.params.Kind)
				assert.Equal(t, credential.Provider(), tC.params.Provider)
				assert.Equal(t, credential.Secret(), tC.params.Secret)
				assert.Equal(t, credential.CreatedAt(), tC.params.CreatedAt)
				assert.Equal(t, credential.UpdatedAt(), tC.params.UpdatedAt)
			}
		})
	}
}

func TestNewCredentialProvider(t *testing.T) {
	testCases := []struct {
		desc      string
		value     string
		expectErr bool
	}{
		{desc: "empty value", value: "", expectErr: true},
		{desc: "reserved provider", value: "local", expectErr: true},
		{desc: "valid value", value: "github", expectErr: false},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			provider, err := domain.NewCredentialProvider(tC.value)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, provider)
			} else {
				require.NoError(t, err)
				require.NotZero(t, provider)
				assert.Equal(t, provider.String(), tC.value)
			}
		})
	}
}

func TestNewCredentialSecret(t *testing.T) {
	testCases := []struct {
		desc      string
		value     string
		expectErr bool
	}{
		{desc: "empty value", value: "", expectErr: true},
		{desc: "valid value", value: "secret", expectErr: false},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			secret, err := domain.NewCredentialSecret(tC.value)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, secret)
			} else {
				require.NoError(t, err)
				require.NotZero(t, secret)
				assert.Equal(t, secret.Value(), tC.value)
			}
		})
	}
}
