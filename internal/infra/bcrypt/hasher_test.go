package bcrypt

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/testutil/credentialtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewHasher(t *testing.T) {
	testCases := []struct {
		desc      string
		config    Config
		expectErr bool
	}{
		{
			desc: "config with default cost",
			config: Config{
				Cost: bcrypt.DefaultCost,
			},
			expectErr: false,
		},
		{
			desc: "config with exactly min cost",
			config: Config{
				Cost: bcrypt.MinCost,
			},
			expectErr: false,
		},
		{
			desc: "config with exactly max cost",
			config: Config{
				Cost: bcrypt.MaxCost,
			},
			expectErr: false,
		},
		{
			desc: "cost less than bcrypt min cost",
			config: Config{
				Cost: bcrypt.MinCost - 1,
			},
			expectErr: true,
		},
		{
			desc: "cost greater than bcrypt max cost",
			config: Config{
				Cost: bcrypt.MaxCost + 1,
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			hasher, err := NewHasher(tC.config)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, hasher)
			} else {
				assert.NotZero(t, hasher)
				assert.NoError(t, err)
			}
		})
	}
}

func TestHash(t *testing.T) {
	cfg := Config{Cost: bcrypt.MinCost}
	hasher, err := NewHasher(cfg)
	require.NoError(t, err)

	testCases := []struct {
		desc      string
		password  credential.PlainPassword
		expectErr bool
	}{
		{
			desc:      "plain password zero",
			password:  credential.PlainPassword{},
			expectErr: true,
		},
		{
			desc:      "valid case",
			password:  credentialtest.MustPlainPassword(t, "12345678"),
			expectErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			secret, err := hasher.Hash(tC.password)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, secret)
				return
			} 
			require.NoError(t, err)
			require.NotZero(t, secret)
			assert.NotEqual(t, tC.password.Value(), secret.Value())
			assert.NoError(t, bcrypt.CompareHashAndPassword(
				[]byte(secret.Value()),
				[]byte(tC.password.Value()),
			))
		})
	}
}

func TestVerify_WithPlainPasswordZero(t *testing.T) {
	cfg := Config{Cost: bcrypt.MinCost}
	hasher, err := NewHasher(cfg)
	require.NoError(t, err)

	password := credential.PlainPassword{}
	secret := credentialtest.MustSecret(t, "secret")

	check, err := hasher.Verify(password, secret)

	assert.Error(t, err)
	assert.False(t, check)
}

func TestVerify_WithCredentialSecretZero(t *testing.T) {
	cfg := Config{Cost: bcrypt.MinCost}
	hasher, err := NewHasher(cfg)
	require.NoError(t, err)

	password := credentialtest.MustPlainPassword(t, "1240970sadpiogkj1")
	secret := credential.Secret{}

	check, err := hasher.Verify(password, secret)

	assert.Error(t, err)
	assert.False(t, check)
}

func TestVerify_WithValidPassword(t *testing.T) {
	cfg := Config{Cost: bcrypt.MinCost}
	hasher, err := NewHasher(cfg)
	require.NoError(t, err)

	password := credentialtest.MustPlainPassword(t, "1240970sadpiogkj1")
	secret, err := hasher.Hash(password)
	require.NoError(t, err)

	check, err := hasher.Verify(password, secret)

	assert.NoError(t, err)
	assert.True(t, check)
}
