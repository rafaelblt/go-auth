package testutil

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
)

func MustPlainPassword(t *testing.T, value string) domain.PlainPassword {
	t.Helper()
	p, err := domain.NewPlainPassword(value)
	if err != nil {
		t.Fatalf("%s is invalid to plain password: %s", value, err)
	}
	return p
}

func MustCredentialSecret(t *testing.T, value string) domain.CredentialSecret {
	t.Helper()
	secret, err := domain.NewCredentialSecret(value)
	if err != nil {
		t.Fatalf("%s is invalid to credential secret: %s", value, err)
	}
	return secret
}
