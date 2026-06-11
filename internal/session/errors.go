package session

import "errors"

var (
	ErrTokenInvalid   = errors.New("session: token invalid")
	ErrTokenExpired   = errors.New("session: token expired")
)
