package api

import (
	"encoding/json"
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
