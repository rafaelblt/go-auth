package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/rafaelblt/go-auth/internal/validation"
)

type errorBody struct {
	Error errorData `json:"error"`
}

type errorData struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Fields  []fieldErrorData `json:"fields,omitempty"`
}

type fieldErrorData struct {
	Field   string         `json:"field"`
	Code    string         `json:"code"`
	Details map[string]any `json:"details"`
}

func errorResponse(status int, data errorData) response {
	return response{StatusCode: status, Body: errorBody{Error: data}}
}

func validationError(fields []fieldErrorData) response {
	return errorResponse(http.StatusUnprocessableEntity, errorData{
		Code:    "validation_failed",
		Message: "The input failed validation.",
		Fields:  fields,
	})
}

func internalServerError() response {
	return errorResponse(http.StatusInternalServerError, errorData{
		Code:    "internal_server_error",
		Message: "An internal error occurred.",
	})
}

func invalidJSONBodyError() response {
	return errorResponse(http.StatusBadRequest, errorData{
		Code:    "invalid_json_body",
		Message: "Request body is not valid JSON.",
	})
}

func unsupportedMediaTypeError() response {
	return errorResponse(http.StatusUnsupportedMediaType, errorData{
		Code:    "unsupported_media_type",
		Message: "Content-Type must be application/json.",
	})
}

func requestBodyTooLargeError() response {
	return errorResponse(http.StatusRequestEntityTooLarge, errorData{
		Code:    "request_body_too_large",
		Message: "Request body is too large.",
	})
}

func tooManyRequestsError() response {
	return errorResponse(http.StatusTooManyRequests, errorData{
		Code:    "too_many_requests",
		Message: "Too many requests.",
	})
}

func routeNotFoundError() response {
	return errorResponse(http.StatusNotFound, errorData{
		Code:    "route_not_found",
		Message: "Route not found.",
	})
}

func methodNotAllowedError() response {
	return errorResponse(http.StatusMethodNotAllowed, errorData{
		Code:    "method_not_allowed",
		Message: "Method not allowed for this route.",
	})
}

var errorFieldCatalog = map[string]string{
	register.FieldUsername: "username",
	register.FieldPassword: "password",
}

var kindStatusCatalog = map[usecase.ErrorKind]int{
	usecase.ErrorKindConflict:     http.StatusConflict,
	usecase.ErrorKindUnauthorized: http.StatusUnauthorized,
}

var kindMessageCatalog = map[usecase.ErrorKind]string{
	usecase.ErrorKindConflict:     "A conflict error occurred.",
	usecase.ErrorKindUnauthorized: "Not authorized.",
}

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
	logUseCaseError(logger, uerr)

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

	return errorResponse(status, errorData{Code: uerr.Code(), Message: msg})
}

// logUseCaseError records what the client is not told: the reason behind an
// error whose code is deliberately generic. Errors without one carry it in
// their code already, so the field is left out rather than logged empty.
func logUseCaseError(logger *slog.Logger, uerr usecase.UseCaseError) {
	attrs := []any{"code", uerr.Code(), "kind", uerr.Kind()}
	if reason := uerr.Reason(); reason != "" {
		attrs = append(attrs, "reason", reason)
	}
	logger.Info("use case error", attrs...)
}

func translateValidationError(ctx context.Context, verr validation.ValidationError) response {
	logger := loggerFrom(ctx)
	fieldErrs := make([]fieldErrorData, len(verr.Errors()))
	pairsToLog := make([]string, len(verr.Errors()))

	for i, ferr := range verr.Errors() {
		issue := ferr.Issue()
		field, ok := errorFieldCatalog[ferr.Field()]
		if !ok {
			logger.Warn("error field not found in field catalog, using raw name",
				"field", ferr.Field())
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

	logger.Info("validation error", "pairs", pairsToLog)
	return validationError(fieldErrs)
}
