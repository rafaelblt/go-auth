package api

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"unicode/utf8"
)

// See docs/api/reference.md#request-bodies.
const requestBodyMaxBytes = 64 << 10

var errInvalidUTF8 = errors.New("request body is not valid UTF-8")

func hasJSONContentType(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && mediaType == "application/json"
}

// decodeJSONBody reads only the first JSON value, as before, and checks its
// bytes before encoding/json can turn invalid UTF-8 into U+FFFD. See
// docs/api/reference.md#request-bodies.
func decodeJSONBody(r *http.Request, body any) error {
	var raw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		return err
	}
	if !utf8.Valid(raw) {
		return errInvalidUTF8
	}
	return json.Unmarshal(raw, body)
}
