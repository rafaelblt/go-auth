package credentialtest

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/stretchr/testify/require"
)

func MustPlainPassword(t *testing.T, value string) credential.PlainPassword {
	t.Helper()
	pwd, issues := credential.NewPlainPassword(value)
	require.Truef(t, issues.IsEmpty(), "%q is invalid to plain password: %s", value, issues)
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

