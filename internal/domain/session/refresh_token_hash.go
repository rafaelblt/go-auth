package session

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"slices"
)

type RefreshTokenHash struct {
	value []byte
}

func NewRefreshTokenHash(value []byte) (RefreshTokenHash, error) {
	if len(value) == 0 {
		return RefreshTokenHash{}, errors.New("value empty")
	}
	if len(value) != sha256.Size {
		e := fmt.Errorf("value length %d, want %d", len(value), sha256.Size)
		return RefreshTokenHash{}, e
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
