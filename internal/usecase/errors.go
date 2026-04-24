package usecase

import (
	"github.com/rafaelblt/go-auth/internal/shared"
)

type ValidationError struct {
	errs shared.Set[error]
}

func newValidationError(errs ...error) ValidationError {
	set := shared.NewSetFrom(errs...)
	return ValidationError{set}
}

func (ve ValidationError) Error() string {
	return "Validation Errors" // TODO: improve the error message
}

func (ve ValidationError) Errors() shared.Set[error] {
	return ve.errs
}

func (ve ValidationError) Contains(target error) bool {
	if target == nil { return false }
	return ve.errs.Contains(target)
}
