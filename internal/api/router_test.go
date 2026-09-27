package api

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubUseCase[In, Out any] struct{}

func (stubUseCase[In, Out]) Execute(context.Context, In) (Out, error) {
	var out Out
	return out, nil
}

type panickingUseCase[In, Out any] struct{}

func (panickingUseCase[In, Out]) Execute(context.Context, In) (Out, error) {
	panic("boom")
}

func newRouterTestDependencies(t *testing.T) (Dependencies, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	return Dependencies{
		Logger:            slog.New(slog.NewJSONHandler(buf, nil)),
		Clock:             porttest.NewFakeClock(),
		Register:          stubUseCase[register.Input, register.Output]{},
		Login:             stubUseCase[login.Input, login.Output]{},
		Refresh:           stubUseCase[refresh.Input, refresh.Output]{},
		PublicKeyProvider: NewFakePublicKeyProvider(),
	}, buf
}

func TestNewRouter_ReturnsError_WhenADependencyIsNil(t *testing.T) {
	testCases := []struct {
		desc     string
		nilOne   func(*Dependencies)
		expected string
	}{
		{desc: "logger", nilOne: func(d *Dependencies) { d.Logger = nil }, expected: "logger nil"},
		{desc: "clock", nilOne: func(d *Dependencies) { d.Clock = nil }, expected: "clock nil"},
		{desc: "register", nilOne: func(d *Dependencies) { d.Register = nil }, expected: "register nil"},
		{desc: "login", nilOne: func(d *Dependencies) { d.Login = nil }, expected: "login nil"},
		{desc: "refresh", nilOne: func(d *Dependencies) { d.Refresh = nil }, expected: "refresh nil"},
		{desc: "public key provider", nilOne: func(d *Dependencies) { d.PublicKeyProvider = nil }, expected: "public key provider nil"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			deps, _ := newRouterTestDependencies(t)
			tC.nilOne(&deps)

			handler, err := NewRouter(Config{Dependencies: deps})

			assert.EqualError(t, err, tC.expected)
			assert.Nil(t, handler)
		})
	}
}

func TestNewRouter_AnswersUnknownRoute_AndLogsItsLifecycle(t *testing.T) {
	deps, buf := newRouterTestDependencies(t)
	handler, err := NewRouter(Config{Dependencies: deps})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/unknown", nil))

	require.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.Equal(t, routeNotFoundError().Body, decodeErrorBody(t, recorder))
	lines := loggedLines(t, buf)
	require.NotEmpty(t, lines)
	assert.Equal(t, "request received", lines[0]["msg"])
	last := lines[len(lines)-1]
	assert.Equal(t, "request finished", last["msg"])
	assert.Equal(t, float64(http.StatusNotFound), last["status"])
	assert.Equal(t, float64(0), last["duration"])
}

func TestNewRouter_AnswersInternalServerError_AndLogsItsLifecycle_WhenHandlerPanics(t *testing.T) {
	deps, buf := newRouterTestDependencies(t)
	deps.Login = panickingUseCase[login.Input, login.Output]{}
	handler, err := NewRouter(Config{Dependencies: deps})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(`{"username":"alice","password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, internalServerError().Body, decodeErrorBody(t, recorder))
	lines := loggedLines(t, buf)
	require.Len(t, lines, 3)
	assert.Equal(t, "request received", lines[0]["msg"])
	assert.Equal(t, "panic recovered", lines[1]["msg"])
	assert.Equal(t, lines[0]["request_id"], lines[1]["request_id"])
	assert.Equal(t, "request finished", lines[2]["msg"])
	assert.Equal(t, float64(http.StatusInternalServerError), lines[2]["status"])
}
