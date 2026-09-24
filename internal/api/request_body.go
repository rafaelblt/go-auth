package api

import (
	"encoding/json"
	"net/http"
)

func decodeJSONBody(r *http.Request, body any) error {
	return json.NewDecoder(r.Body).Decode(body)
}
