package domain

import "errors"

type HashedPassword struct {
	value string
}

var ErrHashedPasswordEmpty = errors.New("the hashed password is empty")

func NewHashedPassword(hash string) (HashedPassword, error) {
	if hash == "" {
		return HashedPassword{}, ErrHashedPasswordEmpty
	}
	return HashedPassword{value: hash}, nil
}

func (h HashedPassword) Value() string { return h.value }
func (h HashedPassword) IsZero() bool { return h.value == "" }
