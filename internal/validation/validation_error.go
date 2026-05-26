package validation

import (
	"strings"
)

type ValidationError struct {
	errs []error
}

func NewValidationError(errs []error) *ValidationError {
	if errs == nil {
		errs = []error{}
	}
	verr := ValidationError{errs}
	return &verr
}

func (ve *ValidationError) Error() string {
	msgs := make([]string, len(ve.errs))
	for i, err := range ve.errs {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "; ")
}

func (ve *ValidationError) Errors() []error {
	return ve.errs
}
