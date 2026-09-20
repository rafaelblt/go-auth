package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil"
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
				validation.NewFieldError("field", validation.IssueTooLong(1, validation.UnitCodePoint)),
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

// contextWithLoggedLines returns a context whose logger writes one JSON
// object per line into the returned buffer, so a test can read what was
// logged.
func contextWithLoggedLines(t *testing.T) (context.Context, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, nil))
	return context.WithValue(t.Context(), loggerKey, logger), buf
}

func loggedLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	lines := []map[string]any{}
	decoder := json.NewDecoder(buf)
	for decoder.More() {
		line := map[string]any{}
		require.NoError(t, decoder.Decode(&line))
		lines = append(lines, line)
	}
	return lines
}

func TestTranslateUseCaseError_LogsReason(t *testing.T) {
	ctx, buf := contextWithLoggedLines(t)
	err := usecase.NewErrorWithReason(
		"INVALID_CREDENTIALS", usecase.ErrorKindUnauthorized, "password mismatch",
	)

	translateUseCaseError(ctx, err)

	line := testutil.Only(t, loggedLines(t, buf))
	assert.Equal(t, "use case error", line["msg"])
	assert.Equal(t, "INVALID_CREDENTIALS", line["code"])
	assert.Equal(t, string(usecase.ErrorKindUnauthorized), line["kind"])
	assert.Equal(t, "password mismatch", line["reason"])
}

func TestTranslateUseCaseError_OmitsEmptyReason(t *testing.T) {
	ctx, buf := contextWithLoggedLines(t)
	err := usecase.NewError("USERNAME_ALREADY_EXISTS", usecase.ErrorKindConflict)

	translateUseCaseError(ctx, err)

	line := testutil.Only(t, loggedLines(t, buf))
	assert.Equal(t, "use case error", line["msg"])
	assert.Equal(t, "USERNAME_ALREADY_EXISTS", line["code"])
	assert.NotContains(t, line, "reason")
}

// The reason is logged even when the kind is missing from a catalog, since
// that bug must not also cost the record of what the client was refused for.
func TestTranslateUseCaseError_LogsReason_WhenKindIsNotInCatalog(t *testing.T) {
	ctx, buf := contextWithLoggedLines(t)
	err := usecase.NewErrorWithReason("SOME_CODE", "unmapped kind", "some reason")

	resp := translateUseCaseError(ctx, err)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	lines := loggedLines(t, buf)
	require.NotEmpty(t, lines)
	assert.Equal(t, "some reason", lines[0]["reason"])
}

func TestTranslateValidationError(t *testing.T) {
	testCases := []struct {
		desc string
		err  validation.ValidationError
	}{
		{
			desc: "1 field error",
			err: validation.NewValidationError(
				validation.NewFieldError("single", validation.IssueTooLong(10, validation.UnitCodePoint)),
			),
		},
		{
			desc: "2 field errors",
			err: validation.NewValidationError(
				validation.NewFieldError("pair1", validation.IssueTooLong(2, validation.UnitCodePoint)),
				validation.NewFieldError("pair2", validation.IssueTooShort(1, validation.UnitCodePoint)),
			),
		},
		{
			desc: "3 field errors",
			err: validation.NewValidationError(
				validation.NewFieldError("family1", validation.IssueTooLong(2, validation.UnitCodePoint)),
				validation.NewFieldError("family2", validation.IssueTooShort(4, validation.UnitCodePoint)),
				validation.NewFieldError("family3", validation.IssueTooLong(6, validation.UnitCodePoint)),
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
