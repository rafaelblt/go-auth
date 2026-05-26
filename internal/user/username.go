package user

import (
	"errors"
	"strings"

	"github.com/rafaelblt/go-auth/internal/validation"
)

type Username struct {
	value string
}

const (
	UsernameMinLen = 3
	UsernameMaxLen = 32
)

var ErrUsernameTooLong = errors.New("the username value is too long")
var ErrUsernameTooShort = errors.New("the username value is too short")

func NewUsername(value string) (Username, error) {
	normalized := normalizeUsername(value)

	errs := ValidateUsername(normalized)
	if len(errs) > 0 {
		return Username{}, validation.NewValidationError(errs)
	}

	return Username{normalized}, nil
}

func ValidateUsername(value string) []error {
	normalized := normalizeUsername(value)

	var errs []error

	if len(normalized) < UsernameMinLen {
		errs = append(errs, ErrUsernameTooShort)
	}
	if len(normalized) > UsernameMaxLen {
		errs = append(errs, ErrUsernameTooLong)
	}

	return errs
}

func normalizeUsername(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func (u Username) IsZero() bool   { return u.value == "" }
func (u Username) String() string { return u.value }
