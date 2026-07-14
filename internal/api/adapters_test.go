package api

import (
	"context"
	"encoding/json"
	"errors"
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
	if uc.err != nil {
		return FakeOutput{}, uc.err
	}
	return FakeOutput{OutValue: in.InValue}, nil
}

func (uc *FakeUseCase) Contexts() []context.Context {
	return uc.contexts
}

func (uc *FakeUseCase) Inputs() []FakeInput {
	return uc.inputs
}

// Fake Request & Response Body

type FakeRequestBody struct {
	ReqValue string `json:"req_value"`
}

type FakeResponseBody struct {
	RespValue string `json:"resp_value"`
}

// Test Helper

type TestAdapterUseCaseHelper struct {
	t           *testing.T
	code        int
	decoderErr  error
	encoderErr  error
	FakeUseCase *FakeUseCase
}

func NewTestAdapterUseCaseHelper(t *testing.T) *TestAdapterUseCaseHelper {
	return &TestAdapterUseCaseHelper{
		t:           t,
		FakeUseCase: NewFakeUseCase(),
		code:        http.StatusIMUsed,
	}
}

func (h *TestAdapterUseCaseHelper) ExpectedStatusCode() int {
	return h.code
}

func (h *TestAdapterUseCaseHelper) NewHandler() http.HandlerFunc {
	h.t.Helper()
	return adaptUseCase(h.FakeUseCase, h.FakeDecoder, h.FakeEncoder)
}

func (h *TestAdapterUseCaseHelper) NewRequest(value string) *http.Request {
	h.t.Helper()

	body := fmt.Sprintf(`{"req_value":"%s"}`, value)
	reader := strings.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "url", reader)
	require.NoError(h.t, err)

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

func (h *TestAdapterUseCaseHelper) SetErrorInUseCase(err error) {
	h.FakeUseCase.err = err
}

func (h *TestAdapterUseCaseHelper) SetErrorInFakeDecoder(err error) {
	h.decoderErr = err
}

func (h *TestAdapterUseCaseHelper) SetErrorInFakeEncoder(err error) {
	h.encoderErr = err
}

func (h *TestAdapterUseCaseHelper) FakeDecoder(r *http.Request) (FakeInput, error) {
	if h.decoderErr != nil {
		return FakeInput{}, h.decoderErr
	}

	var body FakeRequestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return FakeInput{}, err
	}

	in := FakeInput{InValue: body.ReqValue}
	return in, nil
}

func (h *TestAdapterUseCaseHelper) FakeEncoder(w http.ResponseWriter, out FakeOutput) error {
	if h.encoderErr != nil {
		return h.encoderErr
	}

	body := FakeResponseBody{RespValue: out.OutValue}
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}

	w.WriteHeader(h.code)
	w.Write(buf)
	return nil
}

// Tests

func TestAdaptUseCase_ExecutesUseCaseWithInput(t *testing.T) {
	helper := NewTestAdapterUseCaseHelper(t)
	handler := helper.NewHandler()
	recorder := httptest.NewRecorder()
	value := "valueststest"

	handler(recorder, helper.NewRequest(value))

	input := testutil.Only(t, helper.FakeUseCase.Inputs())
	assert.Equal(t, value, input.InValue)
}

func TestAdaptUseCase_ExecutesUseCaseWithContext(t *testing.T) {
	helper := NewTestAdapterUseCaseHelper(t)
	handler := helper.NewHandler()
	recorder := httptest.NewRecorder()

	req := helper.NewRequest("value for tests in test file")
	handler(recorder, req)

	context := testutil.Only(t, helper.FakeUseCase.Contexts())
	assert.Equal(t, req.Context(), context)
}

func TestAdaptUseCase_WritesSuccessResponse(t *testing.T) {
	helper := NewTestAdapterUseCaseHelper(t)
	handler := helper.NewHandler()
	recorder := httptest.NewRecorder()
	value := "1 + 1 = 3"

	handler(recorder, helper.NewRequest(value))

	require.Equal(t, helper.ExpectedStatusCode(), recorder.Code)
	resp := helper.DecodeSuccessBody(recorder)
	assert.Equal(t, value, resp.RespValue)
}

func TestAdaptUseCase_TranslatesError_WhenUseCaseFails(t *testing.T) {
	helper := NewTestAdapterUseCaseHelper(t)
	handler := helper.NewHandler()
	recorder := httptest.NewRecorder()

	err := usecase.NewError("ERROR_CODE", usecase.ErrorKindConflict)
	helper.SetErrorInUseCase(err)

	handler(recorder, helper.NewRequest("req value"))

	require.Equal(t, http.StatusConflict, recorder.Code)
	resp := helper.DecodeErrorBody(recorder)
	assert.Equal(t, err.Code(), resp.Error.Code)
}

func TestAdaptUseCase_WritesInvalidJSONBodyError_WhenDecoderFails(t *testing.T) {
	helper := NewTestAdapterUseCaseHelper(t)
	handler := helper.NewHandler()
	recorder := httptest.NewRecorder()

	err := errors.New("random unexpected error")
	helper.SetErrorInFakeDecoder(err)

	handler(recorder, helper.NewRequest("valuable"))

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	resp := helper.DecodeErrorBody(recorder)
	assert.Equal(t, invalidJSONBodyErrorBody.Error.Code, resp.Error.Code)
	assert.Equal(t, invalidJSONBodyErrorBody.Error.Message, resp.Error.Message)
}

func TestAdaptUseCase_WritesInternalServerError_WhenEncoderFails(t *testing.T) {
	helper := NewTestAdapterUseCaseHelper(t)
	handler := helper.NewHandler()
	recorder := httptest.NewRecorder()

	err := errors.New("random unexpected error")
	helper.SetErrorInFakeEncoder(err)

	handler(recorder, helper.NewRequest("valuable"))

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	resp := helper.DecodeErrorBody(recorder)
	assert.Equal(t, internalServerErrorBody.Error.Code, resp.Error.Code)
	assert.Equal(t, internalServerErrorBody.Error.Message, resp.Error.Message)
}
