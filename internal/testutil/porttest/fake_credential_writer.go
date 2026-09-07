package porttest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/password"
)

type FakeCredentialWriter struct {
	data []*password.Credential
	err  error
}

func NewFakeCredentialWriter() *FakeCredentialWriter {
	return &FakeCredentialWriter{data: []*password.Credential{}}
}

func (w *FakeCredentialWriter) Add(
	ctx context.Context, cred *password.Credential,
) error {
	if w.err != nil {
		return w.err
	}
	w.data = append(w.data, cred)
	return nil
}

func (w *FakeCredentialWriter) SavedCredentials() []*password.Credential {
	return w.data
}

func (w *FakeCredentialWriter) CheckHashIsSaved(h password.Hashed) bool {
	for _, cred := range w.data {
		if cred.Hash() == h {
			return true
		}
	}
	return false
}

func (w *FakeCredentialWriter) SetError(err error) { w.err = err }
