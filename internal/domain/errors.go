package domain

import (
	"errors"
	"strings"
)

type ValidationError struct {
	errs []error
}

func (ve *ValidationError) Error() string {
	msgs := make([]string, len(ve.errs))
    for i, err := range ve.errs {
        msgs[i] = err.Error()
    }
    return strings.Join(msgs, "; ")
}

func (ve *ValidationError) Unwrap() []error {
    return ve.errs
}

var ErrUserIDZero = errors.New("the user id is zero")
var ErrUsernameZero = errors.New("the username is zero")
var ErrHashedPasswordZero = errors.New("the hashed password is zero")
var ErrPasswordCredentialZero = errors.New("the password credential is zero")

var ErrUsernameAlreadyExists = errors.New("the username already exists")
