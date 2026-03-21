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
func (hasher FakePasswordHasher) Hash(plain domain.PlainPassword) (domain.CredentialSecret, error) {
	hasher.ReceivedPasswords.Add(plain)
	return domain.NewCredentialSecret(fmt.Sprintf("hash <%s>", plain.Value()))
}
func (hasher FakePasswordHasher) Verify(
	plain domain.PlainPassword, hash domain.CredentialSecret,
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

type FakeUserWriter struct {
	SavedUsers []*domain.User
}
func NewFakeUserWriter() FakeUserWriter {
	return FakeUserWriter{SavedUsers: []*domain.User{}}
}
func (writer *FakeUserWriter) Save(ctx context.Context, user *domain.User) error {
	writer.SavedUsers = append(writer.SavedUsers, user)
	return nil
}
func (writer FakeUserWriter) UsernameIsSaved(username domain.Username) bool {
	for _, usr := range writer.SavedUsers {
		if usr.Username() == username {
			return true
		}
	}
	return false
}

type FakeCredentialWriter struct {
	SavedCredentials []*domain.Credential
}
func NewFakeCredentialWriter() FakeCredentialWriter {
	return FakeCredentialWriter{SavedCredentials: []*domain.Credential{}}
}
func (writer *FakeCredentialWriter) Save(
	ctx context.Context, credential *domain.Credential,
) error {
	writer.SavedCredentials = append(writer.SavedCredentials, credential)
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

