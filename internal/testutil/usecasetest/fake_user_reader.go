package usecasetest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/user"
)

type FakeUserReader struct {
	data map[user.ID]*user.User
}

func NewFakeUserReader() *FakeUserReader {
	return &FakeUserReader{
		data: make(map[user.ID]*user.User),
	}
}

func (read FakeUserReader) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
	usr, exists := read.data[id]
	if exists {
		return usr, nil
	}
	return nil, nil
}

func (read FakeUserReader) FindByUsername(ctx context.Context, username user.Username) (*user.User, error) {
	for _, usr := range read.data {
		if usr.Username() == username {
			return usr, nil
		}
	}
	return nil, nil
}

func (read FakeUserReader) InsertUser(user *user.User) {
	read.data[user.ID()] = user
}
