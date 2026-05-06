package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorData `json:"error"`
}

type FieldErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ValidationErrors = []FieldErrorData

type ValidationErrorResponse struct {
	Errors map[string]ValidationErrors `json:"errors"`
}

var InternalServerErrorResponse = ErrorResponse{
	Error: ErrorData{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "An internal error occurred.",
	},
}

var invalidJSONBodyErrorResponse = ErrorResponse{
	Error: ErrorData{
		Code:    "INVALID_BODY",
		Message: "...", // TODO
	},
}

func internalError(ctx context.Context, w http.ResponseWriter, msg string, err error) {
	logger := loggerFrom(ctx)
	logger.Error(msg, "error", err)
	w.WriteHeader(http.StatusInternalServerError)

	resp := InternalServerErrorResponse
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("failed to encode internal server error response", "error", err)
	}
}

func validationError(ctx context.Context, w http.ResponseWriter, resp ValidationErrorResponse) {
	logger := loggerFrom(ctx)

	attrs := make([]any, 0, len(resp.Errors))
	for field, verrs := range resp.Errors {
		errCodes := make([]string, 0, len(verrs))
		for _, err := range verrs {
			errCodes = append(errCodes, err.Code)
		}
		attrs = append(attrs, slog.Any(field, errCodes))
	}
	args := slog.Group("fields", attrs...)

	logger.Info("validation failed", args)
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(resp)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("failed to encode validation error response", "error", err)
	}
}

func badRequestError(ctx context.Context, w http.ResponseWriter, msg string) {
	logger := loggerFrom(ctx)
	logger.Info("bad")
}

func invalidJSONBodyError(ctx context.Context, w http.ResponseWriter, err error) {
	logger := loggerFrom(ctx)
	logger.Info("invalid json body", "error", err)

	w.WriteHeader(http.StatusBadRequest)

	resp := invalidJSONBodyErrorResponse
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		internalError(ctx, w, "failed to encode invalid json body error response", err)
	}
}
