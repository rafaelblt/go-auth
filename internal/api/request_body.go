package api

import (
	"encoding/json"
	"mime"
	"net/http"
)

func hasJSONContentType(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && mediaType == "application/json"
}

func decodeJSONBody(r *http.Request, body any) error {
	return json.NewDecoder(r.Body).Decode(body)
}
