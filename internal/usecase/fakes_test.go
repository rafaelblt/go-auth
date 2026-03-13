package usecase_test

import (
	"context"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/shared"
)

type FakePasswordHasher struct {
	ReceivedPasswords shared.Set[domain.PlainPassword]
}
func NewFakePasswordHasher() FakePasswordHasher {
	return FakePasswordHasher{ReceivedPasswords: shared.NewSet[domain.PlainPassword]()}
}
func (hasher FakePasswordHasher) Hash(plain domain.PlainPassword) (domain.HashedPassword, error) {
	hasher.ReceivedPasswords.Add(plain)
	return domain.NewHashedPassword(fmt.Sprintf("hash <%s>", plain.Value()))
}
func (hasher FakePasswordHasher) Verify(
	plain domain.PlainPassword, hash domain.HashedPassword,
) (bool, error) {
	expectedHash, _ := hasher.Hash(plain)
	return hash.Value() == expectedHash.Value(), nil
}

type FakeUserFinder struct {
	Data map[domain.UserID]domain.User
}
func (finder FakeUserFinder) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	usr, exists := finder.Data[id]
	if exists {
		return &usr, nil
	}
	return nil, nil
}
func (finder FakeUserFinder) FindByUsername(ctx context.Context, username domain.Username) (*domain.User, error) {
	for _, usr := range finder.Data {
		if usr.Username() == username {
			return &usr, nil
		}
	}
	return nil, nil
}

type FakeUserExistsChecker struct {
	Usernames shared.Set[domain.Username]
}
func NewFakeUserExistsChecker() FakeUserExistsChecker {
	return FakeUserExistsChecker{Usernames: shared.NewSet[domain.Username]()}
}
func (checker FakeUserExistsChecker) ExistsByUsername(ctx context.Context, username domain.Username) (bool, error) {
	return checker.Usernames.Contains(username), nil
}

type FakeUserSaver struct {
	SavedUsers []*domain.User
}
func NewFakeUserSaver() FakeUserSaver {
	return FakeUserSaver{SavedUsers: []*domain.User{}}
}
func (saver *FakeUserSaver) Save(ctx context.Context, user *domain.User) error {
	saver.SavedUsers = append(saver.SavedUsers, user)
	return nil
}
func (saver FakeUserSaver) UsernameIsSaved(username domain.Username) bool {
	for _, usr := range saver.SavedUsers {
		if usr.Username() == username {
			return true
		}
	}
	return false
}

type FakeUserCredentialsSaver struct {
	SavedCredentials []*domain.UserCredentials
}
func NewFakeUserCredentialsSaver() FakeUserCredentialsSaver {
	return FakeUserCredentialsSaver{SavedCredentials: []*domain.UserCredentials{}}
}
func (saver *FakeUserCredentialsSaver) Save(
	ctx context.Context, credentials *domain.UserCredentials,
) error {
	saver.SavedCredentials = append(saver.SavedCredentials, credentials)
	return nil
}

type FakeClock struct {
	tm time.Time
}
func NewFakeClock(tm time.Time) FakeClock {
	return FakeClock{tm: tm}
}
func (clk FakeClock) UtcNow() time.Time {
	return clk.tm
}

