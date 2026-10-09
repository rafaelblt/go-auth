package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/changepassword"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/rafaelblt/go-auth/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decodeErrorBody(t *testing.T, recorder *httptest.ResponseRecorder) errorBody {
	t.Helper()
	var body errorBody
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	return body
}

func TestTranslateError(t *testing.T) {
	testCases := []struct {
		desc           string
		err            error
		expectedStatus int
	}{
		{
			desc:           "use case error",
			err:            usecase.NewError("coding_errors", usecase.ErrorKindConflict),
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
			err:            usecase.NewError("password_not_found", usecase.ErrorKindUnauthorized),
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    kindMessageCatalog[usecase.ErrorKindUnauthorized],
		},
		{
			desc:           "conflict error",
			err:            usecase.NewError("username_strange", usecase.ErrorKindConflict),
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

func TestTranslateUseCaseError_LogsReason(t *testing.T) {
	ctx, buf := contextWithLoggedLines(t)
	err := usecase.NewErrorWithReason(
		"invalid_credentials", usecase.ErrorKindUnauthorized, "password mismatch",
	)

	translateUseCaseError(ctx, err)

	line := testutil.Only(t, loggedLines(t, buf))
	assert.Equal(t, "use case error", line["msg"])
	assert.Equal(t, "invalid_credentials", line["code"])
	assert.Equal(t, string(usecase.ErrorKindUnauthorized), line["kind"])
	assert.Equal(t, "password mismatch", line["reason"])
}

func TestTranslateUseCaseError_OmitsEmptyReason(t *testing.T) {
	ctx, buf := contextWithLoggedLines(t)
	err := usecase.NewError("username_already_exists", usecase.ErrorKindConflict)

	translateUseCaseError(ctx, err)

	line := testutil.Only(t, loggedLines(t, buf))
	assert.Equal(t, "use case error", line["msg"])
	assert.Equal(t, "username_already_exists", line["code"])
	assert.NotContains(t, line, "reason")
}

// The reason is logged even when the kind is missing from a catalog, since
// that bug must not also cost the record of what the client was refused for.
func TestTranslateUseCaseError_LogsReason_WhenKindIsNotInCatalog(t *testing.T) {
	ctx, buf := contextWithLoggedLines(t)
	err := usecase.NewErrorWithReason("some_code", "unmapped kind", "some reason")

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
			// A buffered logger, not t.Context(): none of these field names is
			// in errorFieldCatalog, so each one warns, and the warnings would
			// otherwise reach the test output.
			ctx, _ := contextWithLoggedLines(t)

			resp := translateValidationError(ctx, tC.err)
			assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

			require.IsType(t, errorBody{}, resp.Body)
			body := resp.Body.(errorBody)
			assert.Equal(t, "validation_failed", body.Error.Code)
			assert.NotEmpty(t, body.Error.Message)

			expected := []fieldErrorData{}
			for _, fErr := range tC.err.Errors() {
				data := fieldErrorData{
					Field:   fErr.Field(),
					Code:    fErr.Issue().Code(),
					Details: fErr.Issue().Details(),
				}
				expected = append(expected, data)
			}

			assert.Equal(t, expected, body.Error.Fields)
		})
	}
}

// A field missing from errorFieldCatalog is answered with its internal name,
// which breaks the snake_case of every other field in the body. Nothing reaches
// that path today, so the warning is what would make it visible if something
// did.
func TestTranslateValidationError_WarnsOnFieldNotInCatalog(t *testing.T) {
	ctx, buf := contextWithLoggedLines(t)
	verr := validation.NewValidationError(
		validation.NewFieldError("Email", validation.IssueRequired()),
	)

	resp := translateValidationError(ctx, verr)

	require.IsType(t, errorBody{}, resp.Body)
	body := resp.Body.(errorBody)
	assert.Equal(t, "Email", testutil.Only(t, body.Error.Fields).Field)

	lines := loggedLines(t, buf)
	require.NotEmpty(t, lines)
	assert.Equal(t, "error field not found in field catalog, using raw name", lines[0]["msg"])
	assert.Equal(t, "WARN", lines[0]["level"])
	assert.Equal(t, "Email", lines[0]["field"])
}

func TestTranslateValidationError_DoesNotWarnForAKnownField(t *testing.T) {
	ctx, buf := contextWithLoggedLines(t)
	verr := validation.NewValidationError(
		validation.NewFieldError(register.FieldUsername, validation.IssueRequired()),
	)

	translateValidationError(ctx, verr)

	line := testutil.Only(t, loggedLines(t, buf))
	assert.Equal(t, "validation error", line["msg"])
}

// Every error shares one envelope, and fields is in it only for a validation
// error: a client reads error.code first, whatever the status.
func TestTranslateError_UsesOneEnvelope(t *testing.T) {
	testCases := []struct {
		desc       string
		err        error
		wantFields bool
	}{
		{
			desc:       "use case error",
			err:        usecase.NewError("username_already_exists", usecase.ErrorKindConflict),
			wantFields: false,
		},
		{
			desc: "validation error",
			err: validation.NewValidationError(
				validation.NewFieldError(register.FieldUsername, validation.IssueTooShort(3, validation.UnitCodePoint)),
			),
			wantFields: true,
		},
		{
			desc:       "unexpected error",
			err:        errors.New("unexpected"),
			wantFields: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, _ := contextWithLoggedLines(t)
			resp := translateError(ctx, tC.err)

			raw, err := json.Marshal(resp.Body)
			require.NoError(t, err)
			body := map[string]map[string]any{}
			require.NoError(t, json.Unmarshal(raw, &body))

			require.Contains(t, body, "error")
			assert.NotEmpty(t, body["error"]["code"])
			assert.NotEmpty(t, body["error"]["message"])
			if tC.wantFields {
				assert.Contains(t, body["error"], "fields")
			} else {
				assert.NotContains(t, body["error"], "fields")
			}
		})
	}
}

func TestErrorFieldCatalog(t *testing.T) {
	testCases := []struct {
		desc  string
		field string
	}{
		{
			desc:  "register username field",
			field: register.FieldUsername,
		},
		{
			desc:  "register password field",
			field: register.FieldPassword,
		},
		{
			desc:  "change password new password field",
			field: changepassword.FieldNewPassword,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			retrieved, ok := errorFieldCatalog[tC.field]
			require.True(t, ok)
			assert.NotEmpty(t, retrieved)
		})
	}
}

func TestKindStatusCatalog(t *testing.T) {
	testCases := []struct {
		desc string
		kind usecase.ErrorKind
	}{
		{
			desc: "kind conflict",
			kind: usecase.ErrorKindConflict,
		},
		{
			desc: "kind unauthorized",
			kind: usecase.ErrorKindUnauthorized,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			retrieved, ok := kindStatusCatalog[tC.kind]
			require.True(t, ok)
			status := http.StatusText(retrieved)
			assert.NotZero(t, status)
		})
	}
}

func TestKindMessageCatalog(t *testing.T) {
	testCases := []struct {
		desc string
		kind usecase.ErrorKind
	}{
		{
			desc: "kind conflict",
			kind: usecase.ErrorKindConflict,
		},
		{
			desc: "kind unauthorized",
			kind: usecase.ErrorKindUnauthorized,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			retrieved, ok := kindMessageCatalog[tC.kind]
			require.True(t, ok)
			assert.NotEmpty(t, retrieved)
		})
	}
}
