package api

import (
	"net/http"
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

func validationError(fields []fieldErrorData) response {
	return response{
		StatusCode: http.StatusUnprocessableEntity,
		Body: errorBody{Error: errorData{
			Code:    "VALIDATION_FAILED",
			Message: "The input failed validation.",
			Fields:  fields,
		}},
	}
}

func internalServerError() response {
	return response{
		StatusCode: http.StatusInternalServerError,
		Body: errorBody{Error: errorData{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "An internal error occurred.",
		}},
	}
}

func invalidJSONBodyError() response {
	return response{
		StatusCode: http.StatusBadRequest,
		Body: errorBody{Error: errorData{
			Code:    "INVALID_JSON_BODY",
			Message: "Request body is not valid JSON.",
		}},
	}
}

func unsupportedMediaTypeError() response {
	return response{
		StatusCode: http.StatusUnsupportedMediaType,
		Body: errorBody{Error: errorData{
			Code:    "UNSUPPORTED_MEDIA_TYPE",
			Message: "Content-Type must be application/json.",
		}},
	}
}

func requestBodyTooLargeError() response {
	return response{
		StatusCode: http.StatusRequestEntityTooLarge,
		Body: errorBody{Error: errorData{
			Code:    "REQUEST_BODY_TOO_LARGE",
			Message: "Request body is too large.",
		}},
	}
}
