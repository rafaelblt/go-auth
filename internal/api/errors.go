package api

import (
	"net/http"
)

// Single Error

type errorBody struct {
	Error errorData `json:"error"`
}

type errorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Validation Error

type validationErrorBody struct {
	Errors fieldErrors `json:"errors"`
}

type validationErrors map[string]fieldErrors
type fieldErrors = []fieldErrorData

type fieldErrorData struct {
	Field   string         `json:"field"`
	Code    string         `json:"code"`
	Details map[string]any `json:"details"`
}

// Built Errors

var internalServerErrorBody = errorBody{
	Error: errorData{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "An internal error occurred.",
	},
}

var invalidJSONBodyErrorBody = errorBody{
	Error: errorData{
		Code:    "INVALID_BODY",
		Message: "...", // TODO
	},
}

func validationError(body validationErrorBody) response {
	return response{
		StatusCode: http.StatusUnprocessableEntity,
		Body:       body,
	}
}

func internalServerError() response {
	return response{
		StatusCode: http.StatusInternalServerError,
		Body:       internalServerErrorBody,
	}
}

func invalidJSONBodyError() response {
	return response{
		StatusCode: http.StatusBadRequest,
		Body:       invalidJSONBodyErrorBody,
	}
}
