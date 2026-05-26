package usecasetest

import (
	"context"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/user"
)

// Unit Of Work

type FakeUnitOfWork struct {
	deps port.UowDeps
}

func NewFakeUnitOfWork(deps port.UowDeps) FakeUnitOfWork {
	return FakeUnitOfWork{deps}
}
func (uow FakeUnitOfWork) Do(ctx context.Context, fn func(deps port.UowDeps) error) error {
	return fn(uow.deps)
}

// Password Hasher

type FakePasswordHasher struct {
	received shared.Set[credential.PlainPassword]
	err      error
}

func NewFakePasswordHasher() FakePasswordHasher {
	return FakePasswordHasher{
		received: shared.NewSet[credential.PlainPassword](),
		err:      nil,
	}
}
func (hasher *FakePasswordHasher) Hash(plain credential.PlainPassword) (credential.Secret, error) {
	if hasher.err == nil {
		hasher.received.Add(plain)
		return credential.NewSecret(fmt.Sprintf("hash <%s>", plain.Value()))
	}
	return credential.Secret{}, hasher.err
}
func (hasher FakePasswordHasher) Verify(
	plain credential.PlainPassword, hash credential.Secret,
) (bool, error) {
	expectedHash, _ := hasher.Hash(plain)
	return hash.Value() == expectedHash.Value(), nil
}
func (hasher *FakePasswordHasher) SetError(err error) { hasher.err = err }

// User Finder

type FakeUserFinder struct {
	Data map[user.ID]user.User
}

func (finder FakeUserFinder) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
	usr, exists := finder.Data[id]
	if exists {
		return &usr, nil
	}
	return nil, nil
}
func (finder FakeUserFinder) FindByUsername(ctx context.Context, username user.Username) (*user.User, error) {
	for _, usr := range finder.Data {
		if usr.Username() == username {
			return &usr, nil
		}
	}
	return nil, nil
}

// User Exists Checker

type FakeUserExistsChecker struct {
	Usernames shared.Set[user.Username]
}

func NewFakeUserExistsChecker() FakeUserExistsChecker {
	return FakeUserExistsChecker{Usernames: shared.NewSet[user.Username]()}
}
func (checker FakeUserExistsChecker) ExistsByUsername(ctx context.Context, username user.Username) (bool, error) {
	return checker.Usernames.Contains(username), nil
}

// User Writer

type FakeUserWriter struct {
	saved []*user.User
	err   error
}

func NewFakeUserWriter() FakeUserWriter {
	return FakeUserWriter{
		saved: []*user.User{},
		err:   nil,
	}
}
func (writer *FakeUserWriter) Save(ctx context.Context, user *user.User) error {
	if writer.err == nil {
		writer.saved = append(writer.saved, user)
		return nil
	}
	return writer.err
}
func (writer FakeUserWriter) UsernameIsSaved(username user.Username) bool {
	for _, usr := range writer.saved {
		if usr.Username() == username {
			return true
		}
	}
	return false
}
func (w *FakeUserWriter) SavedUsers() []*user.User { return w.saved }
func (w *FakeUserWriter) SetError(err error) { w.err = err }

// Credential Writer

type FakeCredentialWriter struct {
	saved []*credential.Credential
	err   error
}

func NewFakeCredentialWriter() FakeCredentialWriter {
	return FakeCredentialWriter{saved: []*credential.Credential{}}
}
func (writer *FakeCredentialWriter) Save(
	ctx context.Context, credential *credential.Credential,
) error {
	if writer.err == nil {
		writer.saved = append(writer.saved, credential)
		return nil
	}
	return writer.err
}
func (w *FakeCredentialWriter) SavedCredentials() []*credential.Credential {
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
func (clk FakeClock) Now() time.Time {
	return clk.tm
}
