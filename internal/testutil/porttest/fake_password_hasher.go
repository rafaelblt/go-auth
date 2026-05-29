package porttest

import (
	"fmt"

	"github.com/rafaelblt/go-auth/internal/credential"
)

type FakePasswordHasher struct {
	data map[credential.PlainPassword]credential.Secret
	err  error
}

func NewFakePasswordHasher() FakePasswordHasher {
	return FakePasswordHasher{
		data: make(map[credential.PlainPassword]credential.Secret),
		err:  nil,
	}
}

func (hsh *FakePasswordHasher) Hash(plain credential.PlainPassword) (credential.Secret, error) {
	if hsh.err != nil {
		return credential.Secret{}, hsh.err
	}

	secret, err := credential.NewSecret(fmt.Sprintf("fake-hash<%s>", plain.Value()))
	if err != nil {
		e := fmt.Errorf("failed to create credential secret in fake hasher: %w", err)
		return credential.Secret{}, e
	}

	hsh.data[plain] = secret
	return secret, hsh.err
}

func (hsh *FakePasswordHasher) SetError(err error) {
	hsh.err = err
}

func (hsh *FakePasswordHasher) GetSecretByPassword(plain credential.PlainPassword) (credential.Secret, bool) {
	secret, ok := hsh.data[plain]
	return secret, ok
}
