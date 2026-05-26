package validation

import (
	"fmt"
	"strings"

	"github.com/rafaelblt/go-auth/internal/shared"
)

type FieldValidationError struct {
	errs shared.Set[FieldError]
}

func (fve *FieldValidationError) Error() string {
	msgs := make([]string, fve.errs.Len())
	for i, err := range fve.errs.Values() {
		msgs[i] = fmt.Sprintf("%s: %s", err.Field(), err.Err())
	}
	return strings.Join(msgs, "; ")
}

func (fve *FieldValidationError) Errors() []FieldError {
	return fve.errs.Values()
}

func (fve *FieldValidationError) Len() int {
	return fve.errs.Len()
}

func (fve *FieldValidationError) add(err FieldError) {
	fve.errs.Add(err)
}
