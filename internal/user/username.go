package user

import (
	"errors"
	"strings"

	"github.com/rafaelblt/go-auth/internal/validation"
)

type Username struct {
	value string
}

var ErrUsernameEmpty = errors.New("the username value is empty")

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

	if len(normalized) == 0 {
		errs = append(errs, ErrUsernameEmpty)
	}

	return errs
}

func normalizeUsername(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func (u Username) IsZero() bool   { return u.value == "" }
func (u Username) String() string { return u.value }
