package api

import (
	"encoding/json"
	"mime"
	"net/http"
)

// See docs/api/reference.md#request-bodies.
const requestBodyMaxBytes = 64 << 10

func hasJSONContentType(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && mediaType == "application/json"
}

func decodeJSONBody(r *http.Request, body any) error {
	return json.NewDecoder(r.Body).Decode(body)
}
