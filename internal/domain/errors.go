package domain

import "errors"

var ErrUserIDZero = errors.New("the user id is zero")
var ErrUsernameZero = errors.New("the username is zero")
var ErrHashedPasswordZero = errors.New("the hashed password is zero")
var ErrPasswordCredentialZero = errors.New("the password credential is zero")

var ErrUsernameAlreadyExists = errors.New("the username already exists")
