package credential

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCredential(t *testing.T) {
	secret, err := NewSecret("secret")
	require.NoError(t, err)
	testCases := []struct {
		desc      string
		params    CreationParams
		expectErr bool
	}{
		{
			desc:      "all params zero",
			params:    CreationParams{},
			expectErr: true,
		},
		{
			desc: "user id zero",
			params: CreationParams{
				UserID:    user.ID{},
				Kind:      KindPassword,
				Provider:  ProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "kind zero",
			params: CreationParams{
				UserID:    user.NewID(),
				Kind:      Kind{},
				Provider:  ProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "provider zero",
			params: CreationParams{
				UserID:    user.NewID(),
				Kind:      KindPassword,
				Provider:  Provider{},
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "secret zero",
			params: CreationParams{
				UserID:    user.NewID(),
				Kind:      KindPassword,
				Provider:  ProviderLocal,
				Secret:    Secret{},
				CreatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "valid case",
			params: CreationParams{
				UserID:    user.NewID(),
				Kind:      KindPassword,
				Provider:  ProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
			},
			expectErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			credential, err := NewCredential(tC.params)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, credential)
				return
			}
			require.NoError(t, err)
			require.NotZero(t, credential)
			assert.NotZero(t, credential.ID())
			assert.Equal(t, credential.UserID(), tC.params.UserID)
			assert.Equal(t, credential.Kind(), tC.params.Kind)
			assert.Equal(t, credential.Provider(), tC.params.Provider)
			assert.Equal(t, credential.Secret(), tC.params.Secret)
			assert.Equal(t, credential.CreatedAt(), tC.params.CreatedAt)
			assert.Equal(t, credential.UpdatedAt(), tC.params.CreatedAt)
		})
	}
}

func TestRestoreCredential(t *testing.T) {
	secret, err := NewSecret("secret")
	require.NoError(t, err)
	testCases := []struct {
		desc      string
		params    RestoreParams
		expectErr bool
	}{
		{
			desc:      "all params zero",
			params:    RestoreParams{},
			expectErr: true,
		},
		{
			desc: "id zero",
			params: RestoreParams{
				ID:        ID{},
				UserID:    user.NewID(),
				Kind:      KindPassword,
				Provider:  ProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "user id zero",
			params: RestoreParams{
				ID:        NewID(),
				UserID:    user.ID{},
				Kind:      KindPassword,
				Provider:  ProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "kind zero",
			params: RestoreParams{
				ID:        NewID(),
				UserID:    user.NewID(),
				Kind:      Kind{},
				Provider:  ProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "provider zero",
			params: RestoreParams{
				ID:        NewID(),
				UserID:    user.NewID(),
				Kind:      KindPassword,
				Provider:  Provider{},
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "secret zero",
			params: RestoreParams{
				ID:        NewID(),
				UserID:    user.NewID(),
				Kind:      KindPassword,
				Provider:  ProviderLocal,
				Secret:    Secret{},
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "created at zero",
			params: RestoreParams{
				ID:        NewID(),
				UserID:    user.NewID(),
				Kind:      KindPassword,
				Provider:  ProviderLocal,
				Secret:    secret,
				CreatedAt: time.Time{},
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: true,
		},
		{
			desc: "updated at zero",
			params: RestoreParams{
				ID:        NewID(),
				UserID:    user.NewID(),
				Kind:      KindPassword,
				Provider:  ProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Time{},
			},
			expectErr: true,
		},
		{
			desc: "valid case",
			params: RestoreParams{
				ID:        NewID(),
				UserID:    user.NewID(),
				Kind:      KindPassword,
				Provider:  ProviderLocal,
				Secret:    secret,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			expectErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			credential, err := RestoreCredential(tC.params)
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
