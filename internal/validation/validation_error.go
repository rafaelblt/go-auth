package validation

import (
	"strings"
)

type ValidationError struct {
	errs []FieldError
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
