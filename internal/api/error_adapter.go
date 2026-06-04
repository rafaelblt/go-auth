package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/validation"
)

func adaptError(err error) (response, error) {
	var uerr usecase.UseCaseError
	if errors.As(err, &uerr) {
		return adaptUseCaseError(uerr)
	}

	var verr validation.ValidationError
	if errors.As(err, &verr) {
		return adaptValidationError(verr)
	}

	return response{}, fmt.Errorf("unexpected error: %w", err)
}

func adaptUseCaseError(uerr usecase.UseCaseError) (response, error) {
	var status int
	var msg string

	switch uerr.Kind() {
	case usecase.ErrorKindConflict:
		status = http.StatusConflict
		msg = "A conflict error occurred."
	}

	body := errorBody{Error: errorData{
		Code:    uerr.Code(),
		Message: msg,
	}}
	resp := response{
		StatusCode: status,
		Body:       body,
	}
	return resp, nil
}

func adaptValidationError(verr validation.ValidationError) (response, error) {
	fieldErrs := make([]fieldErrorData, len(verr.Errors()))

	for i, ferr := range verr.Errors() {
		issue := ferr.Issue()
		field, ok := errorFieldCatalog[ferr.Field()]
		if !ok {
			field = ferr.Field()
		}
		fieldErr := fieldErrorData{
			Field:   field,
			Code:    issue.Code(),
			Details: issue.Details(),
		}
		fieldErrs[i] = fieldErr
	}

	body := validationErrorBody{Errors: fieldErrs}
	resp := response{
		StatusCode: http.StatusUnprocessableEntity,
		Body:       body,
	}
	return resp, nil
}
