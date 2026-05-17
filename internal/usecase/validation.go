package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/shared"
)

type ValidationError struct {
	errs shared.Set[FieldError]
}

func (ve *ValidationError) Error() string {
	msgs := make([]string, ve.errs.Len())
    for i, err := range ve.errs.Values() {
        msgs[i] = err.Error()
    }
    return strings.Join(msgs, "; ")
}

func (ve *ValidationError) Errors() []FieldError {
	return ve.errs.Values()
}

func (ve *ValidationError) Len() int {
	return ve.errs.Len()
}

type FieldError struct {
	field string
	err   error
}

func (fe FieldError) Error() string {
    return fmt.Sprintf("[%s]: %s", fe.field, fe.err)
}

type ValidationAccumulator struct {
	verr  *ValidationError
	fatal error
}

func NewValidationAccumulator() *ValidationAccumulator {
    return &ValidationAccumulator{
        verr: &ValidationError{errs: shared.NewSet[FieldError]()},
        fatal: nil,
    }
}

func (acc *ValidationAccumulator) Add(field string, err error) {
	if err == nil || acc.fatal != nil {
		return
	}
	var verr *domain.ValidationError
	if errors.As(err, &verr) {
        for _, e := range verr.Unwrap() {
            acc.verr.errs.Add(FieldError{field, e})
        }
		return
	}
	acc.fatal = fmt.Errorf("unexpected non-validation error: %w", err)
}

func (acc *ValidationAccumulator) Result() error {
	if acc.fatal != nil {
		return acc.fatal
	}
	if acc.verr.Len() > 0 {
		return acc.verr
	}
	return nil
}

