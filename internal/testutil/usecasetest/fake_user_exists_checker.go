package usecasetest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/user"
)

type FakeUserExistsChecker struct {
	data shared.Set[user.Username]
}

func NewFakeUserExistsChecker() FakeUserExistsChecker {
	return FakeUserExistsChecker{data: shared.NewSet[user.Username]()}
}
func (ch FakeUserExistsChecker) ExistsByUsername(ctx context.Context, username user.Username) (bool, error) {
	return ch.data.Contains(username), nil
}

func (ch *FakeUserExistsChecker) InsertUsername(username user.Username) {
	ch.data.Add(username)
}
