package porttest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/password"
)

type FakePasswordWriter struct {
	data []*password.Password
	err  error
}

func NewFakePasswordWriter() *FakePasswordWriter {
	return &FakePasswordWriter{data: []*password.Password{}}
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

func (w *FakePasswordWriter) SetError(err error) { w.err = err }
