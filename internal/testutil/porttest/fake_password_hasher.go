package porttest

import (
	"fmt"

	"github.com/rafaelblt/go-auth/internal/password"
)

type FakePasswordHasher struct {
	data map[password.Plain]password.Hashed
	err  error
}

func NewFakePasswordHasher() FakePasswordHasher {
	return FakePasswordHasher{
		data: make(map[password.Plain]password.Hashed),
		err:  nil,
	}
}

func (hsh *FakePasswordHasher) Hash(plain password.Plain) (password.Hashed, error) {
	if hsh.err != nil {
		return password.Hashed{}, hsh.err
	}

	hashed, err := password.NewHashed(fmt.Sprintf("fake-hash<%s>", plain.Value()))
	if err != nil {
		e := fmt.Errorf("failed to create hashed password in fake hasher: %w", err)
		return password.Hashed{}, e
	}

	hsh.data[plain] = hashed
	return hashed, hsh.err
}

func (hsh *FakePasswordHasher) SetError(err error) {
	hsh.err = err
}

func (hsh *FakePasswordHasher) GetHashByPassword(plain password.Plain) (password.Hashed, bool) {
	hashed, ok := hsh.data[plain]
	return hashed, ok
}
