package porttest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/user"
)

type FakeUserWriter struct {
	users []*user.User
	err   error
}

func NewFakeUserWriter() *FakeUserWriter {
	return &FakeUserWriter{
		users: make([]*user.User, 0),
		err:   nil,
	}
}
func (w *FakeUserWriter) Save(ctx context.Context, user *user.User) error {
	if w.err == nil {
		w.users = append(w.users, user)
		return nil
	}
	return w.err
}

func (w *FakeUserWriter) UsernameIsSaved(username user.Username) bool {
	for _, usr := range w.users {
		if usr.Username() == username {
			return true
		}
	}
	return false
}

func (w *FakeUserWriter) SavedUsers() []*user.User {
	return w.users
}

func (w *FakeUserWriter) SetError(err error) {
	w.err = err
}
