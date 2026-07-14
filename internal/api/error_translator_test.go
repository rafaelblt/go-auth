package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestTranslateErrorFromUseCase(t *testing.T) {
	testCases := []struct {
		desc       string
		err        error
		wantStatus int
	}{
		{
			desc:       "use case error",
			err:        usecase.NewError("CODING_ERRORS", usecase.ErrorKindConflict),
			wantStatus: http.StatusConflict,
		},
		{
			desc: "validation error",
			err: validation.NewValidationError(
				validation.NewFieldError("field", validation.IssueTooLong(1)),
			),
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			desc:       "unexpected error",
			err:        errors.New("this is a unexpected error from the end"),
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			translateErrorFromUseCase(t.Context(), recorder, tC.err)
			assert.Equal(t, tC.wantStatus, recorder.Code)
		})
	}
}
