package porttest

import (
	"context"
	"errors"

	"github.com/rafaelblt/go-auth/internal/domain/password"
)

type FakePasswordWriter struct {
	data        []*password.Password
	hashUpdates []FakePasswordHashUpdate
	err         error
}

type FakePasswordHashUpdate struct {
	Password *password.Password
	Previous password.Hashed
}

func NewFakePasswordWriter() *FakePasswordWriter {
	return &FakePasswordWriter{
		data:        []*password.Password{},
		hashUpdates: []FakePasswordHashUpdate{},
	}
}

func (w *FakePasswordWriter) Add(
	ctx context.Context, pwd *password.Password,
) error {
	if w.err != nil {
		return w.err
	}
	w.data = append(w.data, pwd)
	return nil
}

func (w *FakePasswordWriter) UpdateHash(
	ctx context.Context, pwd *password.Password, previous password.Hashed,
) error {
	if w.err != nil {
		return w.err
	}
	if pwd == nil {
		return errors.New("password nil")
	}
	if pwd.IsZero() {
		return errors.New("password zero")
	}
	if previous.IsZero() {
		return errors.New("previous hash zero")
	}
	w.hashUpdates = append(w.hashUpdates, FakePasswordHashUpdate{
		Password: pwd,
		Previous: previous,
	})
	return nil
}

func (w *FakePasswordWriter) SavedPasswords() []*password.Password {
	return w.data
}

func (w *FakePasswordWriter) CheckHashIsSaved(h password.Hashed) bool {
	for _, pwd := range w.data {
		if pwd.Hash() == h {
			return true
		}
	}
	return false
}

func (w *FakePasswordWriter) HashUpdates() []FakePasswordHashUpdate {
	return w.hashUpdates
}

func (w *FakePasswordWriter) SetError(err error) { w.err = err }
