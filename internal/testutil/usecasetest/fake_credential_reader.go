package usecasetest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/user"
)

type FakeCredentialReader struct {
	data map[credential.ID]*credential.Credential
	err  error
}

func NewFakeCredentialReader() *FakeCredentialReader {
	reader := FakeCredentialReader{
		data: make(map[credential.ID]*credential.Credential),
		err:  nil,
	}
	return &reader
}

func (r *FakeCredentialReader) FindByID(
	ctx context.Context, id credential.ID,
) (*credential.Credential, error) {
	cred, exists := r.data[id]
	if exists {
		return cred, nil
	}
	return nil, nil
}

func (r *FakeCredentialReader) FindByUserAndKind(
	ctx context.Context, userID user.ID, kind credential.Kind,
) (*credential.Credential, error) {
	for _, cred := range r.data {
		if cred.UserID() == userID && cred.Kind() == kind {
			return cred, nil
		}
	}
	return nil, nil
}

func (r FakeCredentialReader) InsertCredential(cred *credential.Credential) {
	r.data[cred.ID()] = cred
}
