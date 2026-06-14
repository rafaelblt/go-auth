package session

import (
	"errors"
)

type RefreshTokenHash struct {
	value string
}

func NewRefreshTokenHash(value string) (RefreshTokenHash, error) {
	if len(value) == 0 {
		return RefreshTokenHash{}, errors.New("value empty")
	}
	obj := RefreshTokenHash{value: value}
	return obj, nil
}

func (h RefreshTokenHash) Value() string { return h.value }
func (h RefreshTokenHash) IsZero() bool { return h.value == "" }
