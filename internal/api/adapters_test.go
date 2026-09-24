package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fake Use Case

type FakeInput struct {
	InValue string
}

type FakeOutput struct {
	OutValue string
}

type FakeUseCase struct {
	err      error
	contexts []context.Context
	inputs   []FakeInput
}

func NewFakeUseCase() *FakeUseCase {
	return &FakeUseCase{
		inputs:   make([]FakeInput, 0),
		contexts: make([]context.Context, 0),
	}
}

func (uc *FakeUseCase) Execute(ctx context.Context, in FakeInput) (FakeOutput, error) {
	uc.contexts = append(uc.contexts, ctx)
	uc.inputs = append(uc.inputs, in)

	if uc.err != nil {
		return FakeOutput{}, uc.err
	}
	return FakeOutput{OutValue: in.InValue}, nil
}

// Fake Success Logger

type FakeSuccessLogger struct {
	contexts []context.Context
	outputs  []FakeOutput
}

func NewFakeSuccessLogger() *FakeSuccessLogger {
	return &FakeSuccessLogger{
		contexts: make([]context.Context, 0),
		outputs:  make([]FakeOutput, 0),
	}
}

func (l *FakeSuccessLogger) Log(ctx context.Context, out FakeOutput) {
	l.contexts = append(l.contexts, ctx)
	l.outputs = append(l.outputs, out)
}

// Fake Request & Response Body

type FakeRequestBody struct {
	ReqValue string `json:"req_value"`
}

type FakeResponseBody struct {
	RespValue string `json:"resp_value"`
}

// Fake Decoder

type FakeDecoder struct {
	err      error
	requests []*http.Request
}

func NewFakeDecoder() *FakeDecoder {
	return &FakeDecoder{requests: make([]*http.Request, 0)}
}

func (d *FakeDecoder) Decode(r *http.Request) (FakeInput, error) {
	d.requests = append(d.requests, r)

	if d.err != nil {
		return FakeInput{}, d.err
	}

	var body FakeRequestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return FakeInput{}, err
	}

	in := FakeInput{InValue: body.ReqValue}
	return in, nil
}

// Fake Encoder

type FakeEncoder struct {
	statusCode int
	outputs    []FakeOutput
}

func NewFakeEncoder() *FakeEncoder {
	return &FakeEncoder{
		statusCode: http.StatusIMUsed,
		outputs:    make([]FakeOutput, 0),
	}
}

func (e *FakeEncoder) Encode(out FakeOutput) response {
	e.outputs = append(e.outputs, out)
	body := FakeResponseBody{RespValue: out.OutValue}
	return response{StatusCode: e.statusCode, Body: body}
}

// Test Helper

type TestAdapterUseCaseHelper struct {
	t                 *testing.T
	fakeUseCase       *FakeUseCase
	fakeDecoder       *FakeDecoder
	fakeEncoder       *FakeEncoder
	fakeSuccessLogger *FakeSuccessLogger
}

func NewTestAdapterUseCaseHelper(t *testing.T) *TestAdapterUseCaseHelper {
	return &TestAdapterUseCaseHelper{
		t:                 t,
		fakeUseCase:       NewFakeUseCase(),
		fakeDecoder:       NewFakeDecoder(),
		fakeEncoder:       NewFakeEncoder(),
		fakeSuccessLogger: NewFakeSuccessLogger(),
	}
}

func (h *TestAdapterUseCaseHelper) ExpectedStatusCode() int {
	return h.fakeEncoder.statusCode
}

func (h *TestAdapterUseCaseHelper) NewHandler() http.HandlerFunc {
	h.t.Helper()
	return adaptUseCase(useCaseAdapterParams[FakeInput, FakeOutput]{
		UseCase:    h.fakeUseCase,
		Decoder:    h.fakeDecoder.Decode,
		Encoder:    h.fakeEncoder.Encode,
		SuccessLog: h.fakeSuccessLogger.Log,
	})
}

func (h *TestAdapterUseCaseHelper) Handle(w http.ResponseWriter, r *http.Request) {
	h.t.Helper()
	handler := h.NewHandler()
	handler(w, r)
}

func (h *TestAdapterUseCaseHelper) NewRequest(value string) *http.Request {
	h.t.Helper()

	body := fmt.Sprintf(`{"req_value":"%s"}`, value)
	reader := strings.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "url", reader)
	require.NoError(h.t, err)
	req.Header.Set("Content-Type", "application/json")

	return req
}

func (h *TestAdapterUseCaseHelper) DecodeSuccessBody(r *httptest.ResponseRecorder) FakeResponseBody {
	h.t.Helper()
	require.NotNil(h.t, r, "recorder nil")

	var body FakeResponseBody
	require.NoError(h.t, json.Unmarshal(r.Body.Bytes(), &body))

	return body
}

func (h *TestAdapterUseCaseHelper) DecodeErrorBody(r *httptest.ResponseRecorder) errorBody {
	h.t.Helper()
	require.NotNil(h.t, r, "recorder nil")

	var body errorBody
	require.NoError(h.t, json.Unmarshal(r.Body.Bytes(), &body))

	return body
}

func (h *TestAdapterUseCaseHelper) ContextsProvidedToUseCase() []context.Context {
	return h.fakeUseCase.contexts
}

func (h *TestAdapterUseCaseHelper) InputsProvidedToUseCase() []FakeInput {
	return h.fakeUseCase.inputs
}

func (h *TestAdapterUseCaseHelper) RequestsProvidedToDecoder() []*http.Request {
	return h.fakeDecoder.requests
}

func (h *TestAdapterUseCaseHelper) OutputsProvidedToDecoder() []FakeOutput {
	return h.fakeEncoder.outputs
}

func (h *TestAdapterUseCaseHelper) SetErrorInUseCase(err error) {
	h.fakeUseCase.err = err
}

// Tests

func TestAdaptUseCase_ExecutesUseCaseWithInput(t *testing.T) {
	helper := NewTestAdapterUseCaseHelper(t)
	value := "valueststest"

	helper.Handle(httptest.NewRecorder(), helper.NewRequest(value))

	input := testutil.Only(t, helper.InputsProvidedToUseCase())
	assert.Equal(t, value, input.InValue)
}

func TestAdaptUseCase_ExecutesUseCaseWithContext(t *testing.T) {
	helper := NewTestAdapterUseCaseHelper(t)
	req := helper.NewRequest("value for tests in test file")

	helper.Handle(httptest.NewRecorder(), req)

	ctx := testutil.Only(t, helper.ContextsProvidedToUseCase())
	assert.Equal(t, req.Context(), ctx)
}

func TestAdaptUseCase_WritesSuccessResponse(t *testing.T) {
	helper := NewTestAdapterUseCaseHelper(t)
	recorder := httptest.NewRecorder()
	value := "1 + 1 = 3"

	helper.Handle(recorder, helper.NewRequest(value))

	require.Equal(t, helper.ExpectedStatusCode(), recorder.Code)
	resp := helper.DecodeSuccessBody(recorder)
	assert.Equal(t, value, resp.RespValue)
}

func TestAdaptUseCase_TranslatesError_WhenUseCaseFails(t *testing.T) {
	helper := NewTestAdapterUseCaseHelper(t)
	recorder := httptest.NewRecorder()

	err := usecase.NewError("ERROR_CODE", usecase.ErrorKindConflict)
	helper.SetErrorInUseCase(err)

	helper.Handle(recorder, helper.NewRequest("req value"))

	expectedResp := translateError(context.Background(), err)
	assert.Equal(t, expectedResp.StatusCode, recorder.Code)
	actualBody := helper.DecodeErrorBody(recorder)
	assert.Equal(t, expectedResp.Body, actualBody)
}

func TestAdaptUseCase_WritesInvalidJSONBodyError_WhenDecoderReturnsError(t *testing.T) {
	helper := NewTestAdapterUseCaseHelper(t)
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "url", strings.NewReader("invalid json"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	helper.Handle(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	body := helper.DecodeErrorBody(recorder)
	assert.Equal(t, invalidJSONBodyError().Body, body)
}

func TestAdaptUseCase_WritesUnsupportedMediaTypeError_WhenContentTypeIsNotJSON(t *testing.T) {
	testCases := []struct {
		desc        string
		contentType string
	}{
		{desc: "no header", contentType: ""},
		{desc: "text plain", contentType: "text/plain;charset=UTF-8"},
		{desc: "url encoded form", contentType: "application/x-www-form-urlencoded"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper := NewTestAdapterUseCaseHelper(t)
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest("POST", "url", strings.NewReader(`{"req_value":"v"}`))
			require.NoError(t, err)
			if tC.contentType != "" {
				req.Header.Set("Content-Type", tC.contentType)
			}

			helper.Handle(recorder, req)

			require.Equal(t, http.StatusUnsupportedMediaType, recorder.Code)
			body := helper.DecodeErrorBody(recorder)
			assert.Equal(t, unsupportedMediaTypeError().Body, body)
			assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			assert.Empty(t, helper.RequestsProvidedToDecoder())
			assert.Empty(t, helper.InputsProvidedToUseCase())
		})
	}
}

func TestAdaptUseCase_LogsContentType_WhenContentTypeIsNotJSON(t *testing.T) {
	helper := NewTestAdapterUseCaseHelper(t)
	ctx, buf := contextWithLoggedLines(t)
	req, err := http.NewRequest("POST", "url", strings.NewReader(`{"req_value":"v"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")

	helper.Handle(httptest.NewRecorder(), req.WithContext(ctx))

	line := testutil.Only(t, loggedLines(t, buf))
	assert.Equal(t, "unsupported media type error", line["msg"])
	assert.Equal(t, "INFO", line["level"])
	assert.Equal(t, "text/plain;charset=UTF-8", line["content_type"])
}

func TestAdaptUseCase_UseSuccessLog_WithContext(t *testing.T) {
	helper := NewTestAdapterUseCaseHelper(t)
	req := helper.NewRequest("zero value in golang")

	helper.Handle(httptest.NewRecorder(), req)

	ctx := testutil.Only(t, helper.fakeSuccessLogger.contexts)
	assert.Equal(t, req.Context(), ctx)
}

func TestAdaptUseCase_UseSuccessLog_WithOutput(t *testing.T) {
	helper := NewTestAdapterUseCaseHelper(t)
	value := "zero value in golang"

	helper.Handle(httptest.NewRecorder(), helper.NewRequest(value))

	output := testutil.Only(t, helper.fakeSuccessLogger.outputs)
	assert.Equal(t, value, output.OutValue)
}
