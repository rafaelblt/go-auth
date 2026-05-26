package credential

import (
	"errors"
	"strings"

	"github.com/rafaelblt/go-auth/internal/validation"
)

type PlainPassword struct {
	value string
}

var ErrPlainPasswordTooShort = errors.New("the plain password is too short")
var ErrPlainPasswordTooLong = errors.New("the plain password is too long")

var (
	PlainPasswordMaxLen = 32
	PlainPasswordMinLen = 4
)

func NewPlainPassword(value string) (PlainPassword, error) {
	normalized := normalizePlainPassword(value)
	errs := validatePlainPassword(normalized)

	if len(errs) > 0 {
		return PlainPassword{}, validation.NewValidationError(errs)
	}

	return PlainPassword{value: normalized}, nil
}

func validatePlainPassword(value string) []error {
	var errs []error

	if len(value) < PlainPasswordMinLen {
		errs = append(errs, ErrPlainPasswordTooShort)
	}
	if len(value) > PlainPasswordMaxLen {
		errs = append(errs, ErrPlainPasswordTooLong)
	}

	return errs
}

func normalizePlainPassword(value string) string {
	return strings.TrimSpace(value)
}

func (p PlainPassword) Value() string { return p.value }
func (p PlainPassword) IsZero() bool  { return p.value == "" }
