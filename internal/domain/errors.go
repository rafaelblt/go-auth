package domain

import "errors"

var ErrUserIDZero = errors.New("the user id is zero")
var ErrUsernameZero = errors.New("the username is zero")
var ErrHashedPasswordZero = errors.New("the hashed password is zero")
