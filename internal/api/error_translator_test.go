package api

import (
	"errors"
	"net/http"
	"testing"

	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslateError(t *testing.T) {
	testCases := []struct {
		desc           string
		err            error
		expectedStatus int
	}{
		{
			desc:           "use case error",
			err:            usecase.NewError("CODING_ERRORS", usecase.ErrorKindConflict),
			expectedStatus: http.StatusConflict,
		},
		{
			desc: "validation error",
			err: validation.NewValidationError(
				validation.NewFieldError("field", validation.IssueTooLong(1)),
			),
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			desc:           "unexpected error",
			err:            errors.New("this is a unexpected error from the end"),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			resp := translateError(t.Context(), tC.err)
			assert.Equal(t, tC.expectedStatus, resp.StatusCode)
		})
	}
}

func TestTranslateUseCaseError(t *testing.T) {
	testCases := []struct {
		desc           string
		err            usecase.UseCaseError
		expectedStatus int
		expectedMsg    string
	}{
		{
			desc:           "unauthorized error",
			err:            usecase.NewError("PASSWORD_NOT_FOUND", usecase.ErrorKindUnauthorized),
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    kindMessageCatalog[usecase.ErrorKindUnauthorized],
		},
		{
			desc:           "conflict error",
			err:            usecase.NewError("USERNAME_STRANGE", usecase.ErrorKindConflict),
			expectedStatus: http.StatusConflict,
			expectedMsg:    kindMessageCatalog[usecase.ErrorKindConflict],
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			resp := translateUseCaseError(t.Context(), tC.err)

			assert.Equal(t, tC.expectedStatus, resp.StatusCode)

			require.IsType(t, errorBody{}, resp.Body)
			body := resp.Body.(errorBody)
			assert.Equal(t, tC.err.Code(), body.Error.Code)
			assert.Equal(t, tC.expectedMsg, body.Error.Message)
		})
	}
}

func TestTranslateValidationError(t *testing.T) {
	testCases := []struct {
		desc string
		err  validation.ValidationError
	}{
		{
			desc: "1 field error",
			err: validation.NewValidationError(
				validation.NewFieldError("single", validation.IssueTooLong(10)),
			),
		},
		{
			desc: "2 field errors",
			err: validation.NewValidationError(
				validation.NewFieldError("pair1", validation.IssueTooLong(2)),
				validation.NewFieldError("pair2", validation.IssueTooShort(1)),
			),
		},
		{
			desc: "3 field errors",
			err: validation.NewValidationError(
				validation.NewFieldError("family1", validation.IssueTooLong(2)),
				validation.NewFieldError("family2", validation.IssueTooShort(4)),
				validation.NewFieldError("family3", validation.IssueTooLong(6)),
			),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			resp := translateValidationError(t.Context(), tC.err)
			assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

			require.IsType(t, validationErrorBody{}, resp.Body)
			body := resp.Body.(validationErrorBody)

			expected := fieldErrors{}
			for _, fErr := range tC.err.Errors() {
				data := fieldErrorData{
					Field:   fErr.Field(),
					Code:    fErr.Issue().Code(),
					Details: fErr.Issue().Details(),
				}
				expected = append(expected, data)
			}

			assert.Equal(t, expected, body.Errors)
		})
	}
}
