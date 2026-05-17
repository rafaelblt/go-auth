package domain

import (
	"errors"
	"strings"
)

type PlainPassword struct {
	value string
}

var ErrPlainPasswordTooShort = errors.New("the plain password is too short")
var ErrPlainPasswordTooLong = errors.New("the plain password is too long")

var (
	PlainPasswordMaxLen = 32
	PlainPasswordMinLen = 8
)

func NewPlainPassword(value string) (PlainPassword, error) {
	normalized := normalizePlainPassword(value)
	errs := ValidatePlainPassword(normalized)
	if len(errs) > 0 {
		return PlainPassword{}, &ValidationError{errs}
	}
	return PlainPassword{value: normalized}, nil
}

func ValidatePlainPassword(value string) []error {
	normalized := normalizePlainPassword(value)

	var errs []error

	if len(normalized) < PlainPasswordMinLen {
		errs = append(errs, ErrPlainPasswordTooShort)
	}
	if len(normalized) > PlainPasswordMaxLen {
		errs = append(errs, ErrPlainPasswordTooLong)
	}

	return errs
}

func normalizePlainPassword(value string) string {
	return strings.TrimSpace(value)
}

func (p PlainPassword) Value() string { return p.value }
func (p PlainPassword) IsZero() bool { return p.value == "" }
