package porttest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/password"
	"github.com/rafaelblt/go-auth/internal/user"
)

type FakeCredentialReader struct {
	data map[password.ID]*password.Credential
	err  error
}

func NewFakeCredentialReader() *FakeCredentialReader {
	reader := FakeCredentialReader{
		data: make(map[password.ID]*password.Credential),
		err:  nil,
	}
	return &reader
}

func (r *FakeCredentialReader) FindByID(
	ctx context.Context, id password.ID,
) (*password.Credential, error) {
	cred, exists := r.data[id]
	if exists {
		return cred, nil
	}
	return nil, nil
}

func (r *FakeCredentialReader) FindByUserID(
	ctx context.Context, userID user.ID,
) (*password.Credential, error) {
	for _, cred := range r.data {
		if cred.UserID() == userID {
			return cred, nil
		}
	}
	return nil, nil
}

func (r FakeCredentialReader) InsertCredential(cred *password.Credential) {
	r.data[cred.ID()] = cred
}
