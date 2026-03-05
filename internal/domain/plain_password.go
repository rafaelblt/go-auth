package domain

import (
	"errors"
	"strings"
)

type PlainPassword struct {
	value string
}

var ErrPlainPasswordEmpty = errors.New("the plain password is empty")

func NewPlainPassword(value string) (PlainPassword, error) {
	normalized := normalizePlainPassword(value)
	if normalized == "" {
		return PlainPassword{}, ErrPlainPasswordEmpty
	}
	return PlainPassword{value: normalized}, nil
}

func ValidatePlainPassword(value string) []error {
	normalized := normalizePlainPassword(value)

	var errs []error

	if normalized == "" {
		errs = append(errs, ErrPlainPasswordEmpty)
	}

	return errs
}

func normalizePlainPassword(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func (p PlainPassword) String() string {
	return p.value
}
