package session

import "errors"

var (
	ErrSessionNotFound       = errors.New("session: session not found")
	ErrSessionAlreadyRevoked = errors.New("session: session already revoked")
	ErrTokenInvalid          = errors.New("session: token invalid")
	ErrTokenExpired          = errors.New("session: token expired")
	ErrTokenAlreadyUsed      = errors.New("session: token already used")
)
