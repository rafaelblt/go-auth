// Package credential checks a username and a password for login and change
// password alike. Every failure answers the same way, and a missing account
// costs a bcrypt comparison too, so neither the response nor its timing says
// whether an account exists.
//
// See docs/architecture/usecases/README.md#the-shared-credential-check.
package credential

import (
	"context"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase"
)

var ErrUsernameMalformed = usecase.NewErrorWithReason(
	"invalid_credentials",
	usecase.ErrorKindUnauthorized,
	"malformed username",
)

var ErrPasswordMalformed = usecase.NewErrorWithReason(
	"invalid_credentials",
	usecase.ErrorKindUnauthorized,
	"malformed password",
)

var ErrUserNotFound = usecase.NewErrorWithReason(
	"invalid_credentials",
	usecase.ErrorKindUnauthorized,
	"user not found",
)

var ErrPasswordNotFound = usecase.NewErrorWithReason(
	"invalid_credentials",
	usecase.ErrorKindUnauthorized,
	"password not found",
)

var ErrPasswordMismatch = usecase.NewErrorWithReason(
	"invalid_credentials",
	usecase.ErrorKindUnauthorized,
	"password mismatch",
)

// ErrPasswordChanged rejects credentials that were checked, but whose password
// was replaced before the use case wrote what depended on them.
var ErrPasswordChanged = usecase.NewErrorWithReason(
	"invalid_credentials",
	usecase.ErrorKindUnauthorized,
	"password changed",
)

type Checker struct {
	users      port.UserReader
	passwords  port.PasswordReader
	pwdChecker port.PasswordChecker
	dummyHash  password.Hashed
}

// Parse reads a username and a password sent as credentials. A value that
// breaks a format rule is rejected like a wrong password, never answered with
// a 422, which would say which of the two was wrong.
func Parse(username, plain string) (user.Username, password.Plain, error) {
	parsedUsername, issues := user.NewUsername(username)
	if !issues.IsEmpty() {
		return user.Username{}, password.Plain{}, ErrUsernameMalformed
	}
	parsedPlain, issues := password.NewPlain(plain)
	if !issues.IsEmpty() {
		return user.Username{}, password.Plain{}, ErrPasswordMalformed
	}
	return parsedUsername, parsedPlain, nil
}

// Check finds the account of username and verifies plain against its
// password, returning both.
func (c Checker) Check(
	ctx context.Context, username user.Username, plain password.Plain,
) (*user.User, *password.Password, error) {
	usr, err := c.users.FindByUsername(ctx, username)
	if err != nil {
		return nil, nil, fmt.Errorf("find user by username failed: %w", err)
	}
	if usr == nil {
		return nil, nil, c.rejectWithDummyVerify(plain, ErrUserNotFound)
	}

	pwd, err := c.passwords.FindByUserID(ctx, usr.ID())
	if err != nil {
		return nil, nil, fmt.Errorf("find password by user id failed: %w", err)
	}
	if pwd == nil {
		return nil, nil, c.rejectWithDummyVerify(plain, ErrPasswordNotFound)
	}

	ok, err := c.pwdChecker.Verify(plain, pwd.Hash())
	if err != nil {
		return nil, nil, fmt.Errorf("password verification failed: %w", err)
	}
	if !ok {
		return nil, nil, ErrPasswordMismatch
	}

	return usr, pwd, nil
}

// rejectWithDummyVerify spends the time of a real password check before
// returning rejection, so response time does not reveal whether the account
// exists. The result of the check is deliberately discarded: the call is here
// for the time it takes, and removing it reopens the timing difference.
//
// See docs/development/decisions/0045-login-dummy-hash.md.
func (c Checker) rejectWithDummyVerify(plain password.Plain, rejection error) error {
	_, err := c.pwdChecker.Verify(plain, c.dummyHash)
	if err != nil {
		return fmt.Errorf("dummy password verification failed: %w", err)
	}
	return rejection
}
