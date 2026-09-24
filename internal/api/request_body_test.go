package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasJSONContentType(t *testing.T) {
	testCases := []struct {
		desc     string
		header   string
		expected bool
	}{
		{desc: "json", header: "application/json", expected: true},
		{desc: "json with charset", header: "application/json; charset=utf-8", expected: true},
		{desc: "json with upper case charset", header: "application/json;charset=UTF-8", expected: true},
		{desc: "json with another charset", header: "application/json;charset=latin1", expected: true},
		{desc: "json in another case", header: "Application/JSON", expected: true},
		{desc: "json with empty parameters", header: "application/json;", expected: true},
		{desc: "no header", header: "", expected: false},
		{desc: "text plain", header: "text/plain", expected: false},
		{desc: "text plain with charset", header: "text/plain;charset=UTF-8", expected: false},
		{desc: "url encoded form", header: "application/x-www-form-urlencoded", expected: false},
		{desc: "multipart form", header: "multipart/form-data; boundary=x", expected: false},
		{desc: "json suffix", header: "application/problem+json", expected: false},
		{desc: "json with malformed parameter", header: "application/json; foo", expected: false},
		{desc: "list of types", header: "application/json, text/plain", expected: false},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", nil)
			if tC.header != "" {
				req.Header.Set("Content-Type", tC.header)
			}

			assert.Equal(t, tC.expected, hasJSONContentType(req))
		})
	}
}

type testRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// These pin what docs/api/reference.md#request-bodies describes. Only ignoring
// unknown fields is a rule; the rest may still be tightened before v1, and a
// change here updates that section.
func TestDecodeJSONBody_Tolerances(t *testing.T) {
	testCases := []struct {
		desc     string
		body     string
		expected testRequestBody
	}{
		{desc: "null", body: `null`, expected: testRequestBody{}},
		{desc: "null with spaces", body: ` null `, expected: testRequestBody{}},
		{desc: "empty object", body: `{}`, expected: testRequestBody{}},
		{desc: "unknown field", body: `{"username":"a","user_name":"b"}`, expected: testRequestBody{Username: "a"}},
		{desc: "garbage after the value", body: `{"username":"a"}garbage`, expected: testRequestBody{Username: "a"}},
		{desc: "second value", body: `{"username":"a"}{"username":"b"}`, expected: testRequestBody{Username: "a"}},
		{desc: "trailing equals sign", body: `{"username":"a","password":"p"}=`, expected: testRequestBody{Username: "a", Password: "p"}},
		{desc: "field names in another case", body: `{"USERNAME":"a","PassWord":"p"}`, expected: testRequestBody{Username: "a", Password: "p"}},
		{desc: "repeated field", body: `{"username":"a","username":"b"}`, expected: testRequestBody{Username: "b"}},
		{desc: "repeated field in another case", body: `{"username":"a","USERNAME":"b"}`, expected: testRequestBody{Username: "b"}},
		{desc: "null field", body: `{"username":null}`, expected: testRequestBody{}},
		{desc: "invalid utf-8", body: "{\"password\":\"ab\xffcd\"}", expected: testRequestBody{Password: "ab�cd"}},
		{desc: "escaped lone surrogate", body: `{"password":"ab\ud800cd"}`, expected: testRequestBody{Password: "ab�cd"}},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", strings.NewReader(tC.body))
			var body testRequestBody

			err := decodeJSONBody(req, &body)

			require.NoError(t, err)
			assert.Equal(t, tC.expected, body)
		})
	}
}

func TestDecodeJSONBody_ReturnsError(t *testing.T) {
	testCases := []struct {
		desc string
		body string
	}{
		{desc: "empty", body: ``},
		{desc: "only spaces", body: `   `},
		{desc: "truncated", body: `{"username":"a"`},
		{desc: "array", body: `[]`},
		{desc: "string", body: `"text"`},
		{desc: "field of the wrong type", body: `{"username":1}`},
		{desc: "not json", body: `not json`},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", strings.NewReader(tC.body))
			var body testRequestBody

			err := decodeJSONBody(req, &body)

			assert.Error(t, err)
		})
	}
}
