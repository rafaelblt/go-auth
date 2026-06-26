package porttest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/credential"
)

type FakeCredentialWriter struct {
	data []*credential.Credential
	err  error
}

func NewFakeCredentialWriter() *FakeCredentialWriter {
	return &FakeCredentialWriter{data: []*credential.Credential{}}
}

func (w *FakeCredentialWriter) Add(
	ctx context.Context, cred *credential.Credential,
) error {
	if w.err != nil {
		return w.err
	}
	w.data = append(w.data, cred)
	return nil
}

func (w *FakeCredentialWriter) SavedCredentials() []*credential.Credential {
	return w.data
}

func (w *FakeCredentialWriter) CheckSecretIsSaved(s credential.Secret) bool {
	for _, cred := range w.data {
		if cred.Secret() == s {
			return true
		}
	}
	return false
}

func (w *FakeCredentialWriter) SetError(err error) { w.err = err }
