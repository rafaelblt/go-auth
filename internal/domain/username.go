package domain

import "strings"

type Username struct {
	value string
}

var ErrUsernameEmpty = NewDomainError("USERNAME_EMPTY", "The username is empty.")

func NewUsername(value string) (Username, error) {
	normalized := normalizeUsername(value)
	if normalized == "" {
		return Username{}, ErrUsernameEmpty
	}
	return Username{value: normalized}, nil
}

func ValidateUsername(value string) []DomainError {
	normalized := normalizeUsername(value)

	var errs []DomainError

	if normalized == "" {
		errs = append(errs, ErrUsernameEmpty)
	}

	return errs
}

func normalizeUsername(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func (u Username) String() string {
	return u.value
}
