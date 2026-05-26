package infra_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/testutil/credentialtest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapUserToModel(t *testing.T) {
	testCases := []struct {
		desc      string
		user      *user.User
		expectErr bool
	}{
		{
			desc:      "default user",
			user:      usertest.NewUser(t, nil),
			expectErr: false,
		},
		{
			desc:      "user nil",
			user:      nil,
			expectErr: true,
		},
		{
			desc:      "user zero",
			user:      &user.User{},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			model, err := infra.MapUserToModel(tC.user)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, model)
			} else {
				require.NoError(t, err)
				idBytes, err := tC.user.ID().Value().MarshalBinary()
				require.NoError(t, err)
				assert.Equal(t, idBytes, model.ID)
				assert.Equal(t, tC.user.Username().String(), model.Username)
				assert.Equal(t, tC.user.Status().String(), model.Status)
				assert.Equal(t, tC.user.CreatedAt(), model.CreatedAt)
				assert.Equal(t, tC.user.UpdatedAt(), model.UpdatedAt)
			}
		})
	}
}

func TestMapCredentialToModel(t *testing.T) {
	testCases := []struct {
		desc      string
		cred      *credential.Credential
		expectErr bool
	}{
		{
			desc:      "password credential",
			cred:      credentialtest.PasswordCredential(t),
			expectErr: false,
		},
		{
			desc:      "credential nil",
			cred:      nil,
			expectErr: true,
		},
		{
			desc:      "credential zero",
			cred:      &credential.Credential{},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			model, err := infra.MapCredentialToModel(tC.cred)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, model)
			} else {
				require.NoError(t, err)
				credIDBytes, err := tC.cred.ID().Value().MarshalBinary()
				require.NoError(t, err)
				userIDBytes, err := tC.cred.UserID().Value().MarshalBinary()
				require.NoError(t, err)
				// asserts
				assert.Equal(t, credIDBytes, model.ID)
				assert.Equal(t, userIDBytes, model.UserID)
				assert.Equal(t, tC.cred.Kind().String(), model.Kind)
				assert.Equal(t, tC.cred.Provider().String(), model.Provider)
				assert.Equal(t, tC.cred.Secret().Value(), model.Secret)
				assert.Equal(t, tC.cred.CreatedAt(), model.CreatedAt)
				assert.Equal(t, tC.cred.UpdatedAt(), model.UpdatedAt)
			}
		})
	}
}
