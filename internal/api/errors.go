package api

import (
	"encoding/json"
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

func writeInternalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(InternalServerErrorResponse)
}

func writeValidationError(w http.ResponseWriter, resp ValidationErrorResponse) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(resp)
}
