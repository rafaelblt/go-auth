package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSON(t *testing.T) {
	testCases := []struct {
		desc    string
		resp    response
		wantErr bool
	}{
		{
			desc: "valid map body",
			resp: response{
				StatusCode: http.StatusIMUsed,
				Body: map[string]string{
					"map key": "map value",
				},
			},
		},
		{
			desc: "invalid response body",
			resp: response{
				StatusCode: http.StatusIMUsed,
				Body:       new(chan string),
			},
			wantErr: true,
		},
		{
			desc: "valid struct body",
			resp: response{
				StatusCode: http.StatusIMUsed,
				Body:       FakeResponseBody{RespValue: "x"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			writeJSON(t.Context(), recorder, tC.resp)

			if tC.wantErr {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assert.Empty(t, recorder.Body)
				return
			}
			assert.Equal(t, tC.resp.StatusCode, recorder.Code)
			expected, err := json.Marshal(tC.resp.Body)
			require.NoError(t, err)
			assert.JSONEq(t, string(expected), recorder.Body.String())
		})
	}
}

func TestWriteJSON_LogsBodyTypeWithoutBody_WhenMarshalFails(t *testing.T) {
	type unmarshalableBody struct {
		Token   string   `json:"token"`
		Channel chan int `json:"channel"`
	}
	token := "token value"
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	ctx := context.WithValue(t.Context(), loggerKey, logger)
	resp := response{
		StatusCode: http.StatusOK,
		Body:       unmarshalableBody{Token: token},
	}

	writeJSON(ctx, httptest.NewRecorder(), resp)

	var entry map[string]any
	require.NoError(t, json.Unmarshal(logs.Bytes(), &entry))
	assert.Equal(t, "api.unmarshalableBody", entry["body_type"])
	assert.NotContains(t, logs.String(), token)
}
