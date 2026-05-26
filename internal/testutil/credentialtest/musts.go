package credentialtest

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/stretchr/testify/require"
)

func MustPlainPassword(t *testing.T, value string) credential.PlainPassword {
	t.Helper()
	pwd, err := credential.NewPlainPassword(value)
	require.NoError(t, err)
	return pwd
}

func MustSecret(t *testing.T, value string) credential.Secret {
	t.Helper()
	secret, err := credential.NewSecret(value)
	if err != nil {
		t.Fatalf("%s is invalid to secret: %s", value, err)
	}
	return secret
}

