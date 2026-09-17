package session

import (
	"bytes"
	"errors"
	"slices"
)

type RefreshTokenHash struct {
	value []byte
}

func NewRefreshTokenHash(value []byte) (RefreshTokenHash, error) {
	if len(value) == 0 {
		return RefreshTokenHash{}, errors.New("value empty")
	}
	copy := slices.Clone(value)
	obj := RefreshTokenHash{value: copy}
	return obj, nil
}

func (h RefreshTokenHash) Value() []byte {
	return slices.Clone(h.value)
}

func (h RefreshTokenHash) IsZero() bool {
	return h.value == nil
}

func (h RefreshTokenHash) Equal(other RefreshTokenHash) bool {
	return bytes.Equal(h.value, other.value)
}
