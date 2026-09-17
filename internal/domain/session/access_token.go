package session

import (
	"errors"
)

type AccessToken struct {
	value string
}

func NewAccessToken(value string) (AccessToken, error) {
	if value == "" {
		return AccessToken{}, errors.New("value cannot be empty")
	}
	token := AccessToken{
		value: value,
	}
	return token, nil
}

func (t AccessToken) Value() string { return t.value }
func (t AccessToken) IsZero() bool  { return t.value == "" }
