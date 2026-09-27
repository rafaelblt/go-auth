package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func decodeRouteErrorBody(t *testing.T, recorder *httptest.ResponseRecorder) errorBody {
	t.Helper()
	var body errorBody
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	return body
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
			assert.Equal(t, routeNotFoundError().Body, decodeRouteErrorBody(t, recorder))
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
			assert.Equal(t, methodNotAllowedError().Body, decodeRouteErrorBody(t, recorder))
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
	handler := adaptMiddleware(logging(logger))(newRouteErrorsTestHandler())
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)

	handler.ServeHTTP(httptest.NewRecorder(), req)

	lines := loggedLines(t, buf)
	require.NotEmpty(t, lines)
	last := lines[len(lines)-1]
	assert.Equal(t, "request finished", last["msg"])
	assert.Equal(t, float64(http.StatusNotFound), last["status"])
}
