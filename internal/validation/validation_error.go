package validation

import (
	"strings"
)

type ValidationError struct {
	errs []FieldError
}

func NewValidationError(errs ...FieldError) ValidationError {
	return ValidationError{errs}
}

func (ve ValidationError) Error() string {
	msgs := make([]string, len(ve.errs))

	for i, err := range ve.errs {
		msgs[i] = err.Error()
	}

	return strings.Join(msgs, "; ")
}

func (ve ValidationError) Errors() []FieldError {
	return ve.errs
}
