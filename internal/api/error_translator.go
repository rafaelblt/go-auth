package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/validation"
)

func translateError(ctx context.Context, err error) response {
	var uerr usecase.UseCaseError
	if errors.As(err, &uerr) {
		return translateUseCaseError(uerr)
	}

	var verr validation.ValidationError
	if errors.As(err, &verr) {
		return translateValidationError(verr)
	}

	logger := loggerFrom(ctx)
	logger.Error("unexpected error for translation", "error", err)
	return internalServerError()
}

func translateUseCaseError(uerr usecase.UseCaseError) response {
	var status int
	var msg string

	switch uerr.Kind() {
	case usecase.ErrorKindConflict:
		status = http.StatusConflict
		msg = "A conflict error occurred."
	case usecase.ErrorKindUnauthorized:
		status = http.StatusUnauthorized
		msg = "Not authorized."
	}

	body := errorBody{Error: errorData{
		Code:    uerr.Code(),
		Message: msg,
	}}
	resp := response{
		StatusCode: status,
		Body:       body,
	}
	return resp
}

func translateValidationError(verr validation.ValidationError) response {
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
	return resp
}
