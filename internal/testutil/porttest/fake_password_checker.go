package porttest

import (
	"github.com/rafaelblt/go-auth/internal/password"
)

type FakePasswordChecker struct {
	pairs  map[password.Hashed]password.Plain
	err    error
	defaut *bool
}

func NewFakePasswordChecker() *FakePasswordChecker {
	checker := FakePasswordChecker{
		pairs:  make(map[password.Hashed]password.Plain),
		err:    nil,
		defaut: nil,
	}
	return &checker
}

func (checker *FakePasswordChecker) Verify(
	plain password.Plain, hash password.Hashed,
) (bool, error) {
	if checker.err != nil {
		return false, checker.err
	}
	if checker.defaut != nil {
		return *checker.defaut, nil
	}

	expectedPlain, ok := checker.pairs[hash]
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
	plain password.Plain, hash password.Hashed,
) {
	checker.pairs[hash] = plain
}
