package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRouteErrorsTestHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /post-only", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /get-only", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /handler-not-found", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "handler's own 404", http.StatusNotFound)
	})
	return jsonRouteErrors(mux)
}

func TestJSONRouteErrors_WritesRouteNotFoundError_WhenNoRouteMatchesPath(t *testing.T) {
	testCases := []struct {
		desc        string
		method      string
		path        string
		contentType string
	}{
		{desc: "unknown path", method: http.MethodGet, path: "/unknown"},
		{desc: "unknown path not json", method: http.MethodPost, path: "/unknown", contentType: "text/plain"},
		{desc: "trailing slash", method: http.MethodGet, path: "/post-only/"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(tC.method, tC.path, strings.NewReader("body"))
			if tC.contentType != "" {
				req.Header.Set("Content-Type", tC.contentType)
			}

			newRouteErrorsTestHandler().ServeHTTP(recorder, req)

			require.Equal(t, http.StatusNotFound, recorder.Code)
			assert.Equal(t, routeNotFoundError().Body, decodeErrorBody(t, recorder))
			assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			assert.Empty(t, recorder.Header().Get("Allow"))
		})
	}
}

func TestJSONRouteErrors_WritesMethodNotAllowedError_WhenRouteMatchesPathButNotMethod(t *testing.T) {
	testCases := []struct {
		desc   string
		method string
		path   string
		allow  string
	}{
		{desc: "get on post route", method: http.MethodGet, path: "/post-only", allow: "POST"},
		{desc: "options on post route", method: http.MethodOptions, path: "/post-only", allow: "POST"},
		{desc: "post on get route", method: http.MethodPost, path: "/get-only", allow: "GET, HEAD"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(tC.method, tC.path, nil)

			newRouteErrorsTestHandler().ServeHTTP(recorder, req)

			require.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
			assert.Equal(t, methodNotAllowedError().Body, decodeErrorBody(t, recorder))
			assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			assert.Equal(t, tC.allow, recorder.Header().Get("Allow"))
		})
	}
}

func TestJSONRouteErrors_PassesResponseThrough_WhenRouteMatches(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/handler-not-found", nil)

	newRouteErrorsTestHandler().ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, "handler's own 404\n", recorder.Body.String())
	assert.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
}

func TestJSONRouteErrors_PassesRedirectThrough_WhenPathIsNotClean(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "//post-only", nil)

	newRouteErrorsTestHandler().ServeHTTP(recorder, req)

	require.Equal(t, http.StatusTemporaryRedirect, recorder.Code)
	assert.Equal(t, "/post-only", recorder.Header().Get("Location"))
}

func TestJSONRouteErrors_ReportsStatusToLogging_WhenNoRouteMatchesPath(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, nil))
	handler := logging(logger, porttest.NewFakeClock())(newRouteErrorsTestHandler())
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)

	handler.ServeHTTP(httptest.NewRecorder(), req)

	lines := loggedLines(t, buf)
	require.NotEmpty(t, lines)
	last := lines[len(lines)-1]
	assert.Equal(t, "request finished", last["msg"])
	assert.Equal(t, float64(http.StatusNotFound), last["status"])
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

func TestLogging_LogsRequestReceived_WithRequestFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, nil))
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler := logging(logger, porttest.NewFakeClock())(next)
	req := httptest.NewRequest(http.MethodPost, "/some/path", nil)

	handler.ServeHTTP(httptest.NewRecorder(), req)

	lines := loggedLines(t, buf)
	require.NotEmpty(t, lines)
	first := lines[0]
	assert.Equal(t, "request received", first["msg"])
	assert.Equal(t, "INFO", first["level"])
	requestID, ok := first["request_id"].(string)
	require.True(t, ok)
	_, err := uuid.Parse(requestID)
	assert.NoError(t, err)
	assert.Equal(t, req.Method, first["method"])
	assert.Equal(t, req.URL.Path, first["path"])
	assert.Equal(t, req.RemoteAddr, first["ip"])
}

func TestLogging_PutsTaggedLoggerInContext(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, nil))
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggerFrom(r.Context()).Info("inside handler")
	})
	handler := logging(logger, porttest.NewFakeClock())(next)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	lines := loggedLines(t, buf)
	require.Len(t, lines, 3)
	assert.Equal(t, "request received", lines[0]["msg"])
	assert.Equal(t, "inside handler", lines[1]["msg"])
	assert.Equal(t, lines[0]["request_id"], lines[1]["request_id"])
}

func TestLogging_LogsRequestFinished_WithStatus(t *testing.T) {
	testCases := []struct {
		desc     string
		next     http.HandlerFunc
		expected int
	}{
		{
			desc:     "handler writes a status",
			next:     func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) },
			expected: http.StatusTeapot,
		},
		{
			desc:     "handler writes nothing",
			next:     func(w http.ResponseWriter, r *http.Request) {},
			expected: http.StatusOK,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := slog.New(slog.NewJSONHandler(buf, nil))
			handler := logging(logger, porttest.NewFakeClock())(tC.next)

			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

			lines := loggedLines(t, buf)
			require.NotEmpty(t, lines)
			last := lines[len(lines)-1]
			assert.Equal(t, "request finished", last["msg"])
			assert.Equal(t, float64(tC.expected), last["status"])
			assert.Contains(t, last, "duration")
		})
	}
}

func TestLogging_LogsDurationFromClock(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, nil))
	clock := porttest.NewFakeClock()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clock.SetNow(clock.Now().Add(1500 * time.Millisecond))
	})
	handler := logging(logger, clock)(next)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	lines := loggedLines(t, buf)
	require.NotEmpty(t, lines)
	last := lines[len(lines)-1]
	assert.Equal(t, "request finished", last["msg"])
	assert.Equal(t, float64((1500 * time.Millisecond).Nanoseconds()), last["duration"])
}

func TestRecovery_WritesInternalServerError_WhenHandlerPanics(t *testing.T) {
	ctx, buf := contextWithLoggedLines(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("boom") })
	handler := recovery(next)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

	handler.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, internalServerError().Body, decodeErrorBody(t, recorder))
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	line := testutil.Only(t, loggedLines(t, buf))
	assert.Equal(t, "panic recovered", line["msg"])
	assert.Equal(t, "ERROR", line["level"])
	assert.Equal(t, "boom", line["panic"])
}

func TestRecovery_RepanicsErrAbortHandler(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic(http.ErrAbortHandler) })
	handler := recovery(next)

	assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	})
}

func TestLogging_LogsRequestFinished_AfterRecoveredPanic(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, nil))
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("boom") })
	handler := logging(logger, porttest.NewFakeClock())(recovery(next))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	lines := loggedLines(t, buf)
	require.NotEmpty(t, lines)
	last := lines[len(lines)-1]
	assert.Equal(t, "request finished", last["msg"])
	assert.Equal(t, float64(http.StatusInternalServerError), last["status"])
}
