package validation

import (
	"errors"
	"fmt"
)

type Accumulator struct {
	result []FieldError
	fatal  error
}

func NewAccumulator() *Accumulator {
	result := make([]FieldError, 0)
	return &Accumulator{
		result: result,
		fatal:  nil,
	}
}

func (acc *Accumulator) Add(field string, err error) {
	if err == nil || acc.fatal != nil {
		return
	}

	var issues Issues
	if errors.As(err, &issues) {
		for _, iss := range issues {
			fieldErr := FieldError{field: field, issue: iss}
			acc.result = append(acc.result, fieldErr)
		}
		return
	}

	acc.fatal = fmt.Errorf("unexpected error: %w", err)
}

func (acc *Accumulator) Err() error {
	if acc.fatal != nil {
		return acc.fatal
	}
	if len(acc.result) > 0 {
		return ValidationError{acc.result}
	}
	return nil
}
