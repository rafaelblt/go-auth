package domain

import "strings"

type PlainPassword struct {
	value string
}

var ErrPlainPasswordEmpty = NewDomainError("PLAIN_PASSWORD_EMPTY", "the plain password is empty")

func NewPlainPassword(value string) (PlainPassword, error) {
	normalized := normalizePlainPassword(value)
	if normalized == "" {
		return PlainPassword{}, ErrPlainPasswordEmpty
	}
	return PlainPassword{value: normalized}, nil
}

func ValidatePlainPassword(value string) []DomainError {
	normalized := normalizePlainPassword(value)

	var errs []DomainError

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
