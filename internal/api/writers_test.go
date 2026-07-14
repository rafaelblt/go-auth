package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteInvalidJSONBodyError_StatusCode(t *testing.T) {
	r := httptest.NewRecorder()

	writeInvalidJSONBodyError(t.Context(), r)

	assert.Equal(t, http.StatusBadRequest, r.Code)
}

func TestWriteInvalidJSONBodyError_ResponseBody(t *testing.T) {
	r := httptest.NewRecorder()

	writeInvalidJSONBodyError(t.Context(), r)

	var body errorBody
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &body))
	assert.Equal(t, invalidJSONBodyErrorBody, body)
}

func TestWriteInternalServerError_StatusCode(t *testing.T) {
	r := httptest.NewRecorder()

	writeInternalServerError(t.Context(), r)

	assert.Equal(t, http.StatusInternalServerError, r.Code)
}

func TestWriteInternalServerError_ResponseBody(t *testing.T) {
	r := httptest.NewRecorder()

	writeInternalServerError(t.Context(), r)

	var body errorBody
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &body))
	assert.Equal(t, internalServerErrorBody, body)
}

func TestWriteValidationError_StatusCode(t *testing.T) {
	recorder := httptest.NewRecorder()
	verr := validation.NewValidationError(
		validation.NewFieldError("field", validation.IssueTooLong(100)),
	)

	writeValidationError(t.Context(), recorder, verr)

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
}

func TestWriteValidationError_ResponseBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	fErr1 := validation.NewFieldError("field1", validation.IssueTooLong(100))
	fErr2 := validation.NewFieldError("field2", validation.IssueTooShort(5))
	fErr3 := validation.NewFieldError("field3", validation.IssueTooLong(4))
	verr := validation.NewValidationError(fErr1, fErr2, fErr3)

	writeValidationError(t.Context(), recorder, verr)

	body := validationErrorBody{Errors: fieldErrors{
		fieldErrorData{
			Field:   fErr1.Field(),
			Code:    fErr1.Issue().Code(),
			Details: fErr1.Issue().Details(),
		},
		fieldErrorData{
			Field:   fErr2.Field(),
			Code:    fErr2.Issue().Code(),
			Details: fErr2.Issue().Details(),
		},
		fieldErrorData{
			Field:   fErr3.Field(),
			Code:    fErr3.Issue().Code(),
			Details: fErr3.Issue().Details(),
		},
	}}
	// normalizing expected body
	b, err := json.Marshal(body)
	require.NoError(t, err)
	var expected validationErrorBody
	require.NoError(t, json.Unmarshal(b, &expected))

	var actual validationErrorBody
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &actual))
	assert.Equal(t, expected, actual)
}

func TestWriteUseCaseError(t *testing.T) {
	testCases := []struct {
		desc string
		err  usecase.UseCaseError
	}{
		{
			desc: "conflict error",
			err:  usecase.NewError("CODE_CODE_CODE", usecase.ErrorKindConflict),
		},
		{
			desc: "unauthorized error",
			err:  usecase.NewError("ERROR_ERROR_ERROR", usecase.ErrorKindUnauthorized),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			writeUseCaseError(t.Context(), recorder, tC.err)

			expectedStatus := kindStatusCatalog[tC.err.Kind()]
			assert.Equal(t, expectedStatus, recorder.Code)

			expectedCode := tC.err.Code()
			expectedMsg := kindMessageCatalog[tC.err.Kind()]
			var actualBody errorBody
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &actualBody))
			expectedBody := errorBody{Error: errorData{
				Code:    expectedCode,
				Message: expectedMsg,
			}}
			assert.Equal(t, expectedBody, actualBody)
		})
	}
}
