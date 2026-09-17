package porttest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/user"
)

type FakePasswordReader struct {
	data map[password.ID]*password.Password
	err  error
}

func NewFakePasswordReader() *FakePasswordReader {
	reader := FakePasswordReader{
		data: make(map[password.ID]*password.Password),
		err:  nil,
	}
	return &reader
}

func (r *FakePasswordReader) FindByID(
	ctx context.Context, id password.ID,
) (*password.Password, error) {
	pwd, exists := r.data[id]
	if exists {
		return pwd, nil
	}
	return nil, nil
}

func (r *FakePasswordReader) FindByUserID(
	ctx context.Context, userID user.ID,
) (*password.Password, error) {
	for _, pwd := range r.data {
		if pwd.UserID() == userID {
			return pwd, nil
		}
	}
	return nil, nil
}

func (r FakePasswordReader) InsertPassword(pwd *password.Password) {
	r.data[pwd.ID()] = pwd
}
