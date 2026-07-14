package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/validation"
)

func writeInvalidJSONBodyError(ctx context.Context, w http.ResponseWriter) {
	writeJSON(ctx, w, http.StatusBadRequest, invalidJSONBodyErrorBody)
}

func writeInternalServerError(ctx context.Context, w http.ResponseWriter) {
	writeJSON(ctx, w, http.StatusInternalServerError, internalServerErrorBody)
}

func writeValidationError(ctx context.Context, w http.ResponseWriter, verr validation.ValidationError) {
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

	writeJSON(ctx, w, http.StatusUnprocessableEntity, body)
}

func writeUseCaseError(ctx context.Context, w http.ResponseWriter, uerr usecase.UseCaseError) {
	logger := loggerFrom(ctx)

	status, ok := kindStatusCatalog[uerr.Kind()]
	if !ok {
		logger.Error("error kind not found in status catalog", "kind", uerr.Kind())
		writeInternalServerError(ctx, w)
		return
	}

	msg, ok := kindMessageCatalog[uerr.Kind()]
	if !ok {
		logger.Warn("error kind not found in message catalog, using fallback", "kind", uerr.Kind())
		msg = "An error occurred."
	}

	body := errorBody{Error: errorData{
		Code:    uerr.Code(),
		Message: msg,
	}}

	writeJSON(ctx, w, status, body)
}

func writeJSON(ctx context.Context, w http.ResponseWriter, status int, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		loggerFrom(ctx).Error("json marshal failed", "error", err, "value", v)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}
