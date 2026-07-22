package api

import (
	"context"
	"errors"

	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/validation"
)

func translateError(ctx context.Context, err error) response {
	var uerr usecase.UseCaseError
	if errors.As(err, &uerr) {
		return translateUseCaseError(ctx, uerr)
	}

	var verr validation.ValidationError
	if errors.As(err, &verr) {
		return translateValidationError(ctx, verr)
	}

	loggerFrom(ctx).Error("unexpected error for translation", "error", err)
	return internalServerError()
}

func translateUseCaseError(ctx context.Context, uerr usecase.UseCaseError) response {
	logger := loggerFrom(ctx)

	status, ok := kindStatusCatalog[uerr.Kind()]
	if !ok {
		logger.Error("error kind not found in status catalog", "kind", uerr.Kind())
		return internalServerError()
	}

	msg, ok := kindMessageCatalog[uerr.Kind()]
	if !ok {
		logger.Warn("error kind not found in message catalog, using fallback", "kind", uerr.Kind())
		msg = "An error occurred." // fallback
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

func translateValidationError(ctx context.Context, verr validation.ValidationError) response {
	fieldErrs := make([]fieldErrorData, len(verr.Errors()))
	pairsToLog := make([]string, len(verr.Errors()))

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
		pairsToLog[i] = field + " " + issue.Code()
	}

	loggerFrom(ctx).Info("validation error", "pairs", pairsToLog)
	body := validationErrorBody{Errors: fieldErrs}
	return validationError(body)
}
