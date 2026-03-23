package usecase_test

import (
	"context"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/usecase"
)

// Unit Of Work

type FakeUnitOfWork struct {
	deps usecase.UowDeps
}

func NewFakeUnitOfWork(deps usecase.UowDeps) FakeUnitOfWork {
	return FakeUnitOfWork{deps}
}
func (uow FakeUnitOfWork) Do(ctx context.Context, fn func(deps usecase.UowDeps) error) error {
	return fn(uow.deps)
}

// Password Hasher

type FakePasswordHasher struct {
	received shared.Set[domain.PlainPassword]
	err      error
}

func NewFakePasswordHasher() FakePasswordHasher {
	return FakePasswordHasher{
		received: shared.NewSet[domain.PlainPassword](),
		err:      nil,
	}
}
func (hasher *FakePasswordHasher) Hash(plain domain.PlainPassword) (domain.CredentialSecret, error) {
	if hasher.err == nil {
		hasher.received.Add(plain)
		return domain.NewCredentialSecret(fmt.Sprintf("hash <%s>", plain.Value()))
	}
	return domain.CredentialSecret{}, hasher.err
}
func (hasher FakePasswordHasher) Verify(
	plain domain.PlainPassword, hash domain.CredentialSecret,
) (bool, error) {
	expectedHash, _ := hasher.Hash(plain)
	return hash.Value() == expectedHash.Value(), nil
}
func (hasher *FakePasswordHasher) SetError(err error) { hasher.err = err }

// User Finder

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

// User Exists Checker

type FakeUserExistsChecker struct {
	Usernames shared.Set[domain.Username]
}

func NewFakeUserExistsChecker() FakeUserExistsChecker {
	return FakeUserExistsChecker{Usernames: shared.NewSet[domain.Username]()}
}
func (checker FakeUserExistsChecker) ExistsByUsername(ctx context.Context, username domain.Username) (bool, error) {
	return checker.Usernames.Contains(username), nil
}

// User Writer

type FakeUserWriter struct {
	saved []*domain.User
	err   error
}

func NewFakeUserWriter() FakeUserWriter {
	return FakeUserWriter{
		saved: []*domain.User{},
		err:   nil,
	}
}
func (writer *FakeUserWriter) Save(ctx context.Context, user *domain.User) error {
	if writer.err == nil {
		writer.saved = append(writer.saved, user)
		return nil
	}
	return writer.err
}
func (writer FakeUserWriter) UsernameIsSaved(username domain.Username) bool {
	for _, usr := range writer.saved {
		if usr.Username() == username {
			return true
		}
	}
	return false
}
func (writer *FakeUserWriter) SetError(err error) {
	writer.err = err
}

// Credential Writer

type FakeCredentialWriter struct {
	saved []*domain.Credential
	err   error
}

func NewFakeCredentialWriter() FakeCredentialWriter {
	return FakeCredentialWriter{saved: []*domain.Credential{}}
}
func (writer *FakeCredentialWriter) Save(
	ctx context.Context, credential *domain.Credential,
) error {
	if writer.err == nil {
		writer.saved = append(writer.saved, credential)
		return nil
	}
	return writer.err
}
func (w *FakeCredentialWriter) SavedCredentials() []*domain.Credential {
	return w.saved
}
func (w *FakeCredentialWriter) SetError(err error) { w.err = err }

// Clock

type FakeClock struct {
	tm time.Time
}

func NewFakeClock(tm time.Time) FakeClock {
	return FakeClock{tm: tm}
}
func (clk FakeClock) UtcNow() time.Time {
	return clk.tm
}
