package usecasetest

import (
	"github.com/rafaelblt/go-auth/internal/credential"
)

type FakePasswordChecker struct {
	pairs  map[credential.Secret]credential.PlainPassword
	err    error
	defaut *bool
}

func NewFakePasswordChecker() *FakePasswordChecker {
	checker := FakePasswordChecker{
		pairs:  make(map[credential.Secret]credential.PlainPassword),
		err:    nil,
		defaut: nil,
	}
	return &checker
}

func (checker *FakePasswordChecker) Verify(
	plain credential.PlainPassword, secret credential.Secret,
) (bool, error) {
	if checker.err != nil {
		return false, checker.err
	}
	if checker.defaut != nil {
		return *checker.defaut, nil
	}

	expectedPlain, ok := checker.pairs[secret]
	if ok {
		return plain == expectedPlain, nil
	}

	return false, nil
}

func (checker *FakePasswordChecker) SetError(err error) {
	checker.err = err
}

func (checker *FakePasswordChecker) SetDefault(def bool) {
	checker.defaut = &def
}

func (checker *FakePasswordChecker) SetPair(
	plain credential.PlainPassword, secret credential.Secret,
) {
	checker.pairs[secret] = plain
}
