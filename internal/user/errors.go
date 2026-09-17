package user

import "errors"

var ErrUsernameAlreadyExists = errors.New("user: username already exists")
