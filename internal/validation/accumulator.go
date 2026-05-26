package validation

import (
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/shared"
)

type Accumulator struct {
	result *FieldValidationError
	fatal  error
}

func NewAccumulator() *Accumulator {
	fieldErrs := shared.NewSet[FieldError]()
	fieldValErr := &FieldValidationError{fieldErrs}
	return &Accumulator{
		result: fieldValErr,
		fatal:  nil,
	}
}

func (acc *Accumulator) Add(field string, err error) {
	if err == nil || acc.fatal != nil {
		return
	}
	var verr *ValidationError
	if errors.As(err, &verr) {
		for _, e := range verr.Errors() {
			acc.result.add(FieldError{field, e})
		}
		return
	}
	acc.fatal = fmt.Errorf("unexpected non-validation error: %w", err)
}

func (acc *Accumulator) Err() error {
	if acc.fatal != nil {
		return acc.fatal
	}
	if acc.result.Len() > 0 {
		return acc.result
	}
	return nil
}
