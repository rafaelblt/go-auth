package api

import (
	"encoding/base64"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/port"
)

type jwksBody struct {
	Keys []jwkData `json:"keys"`
}

type jwkData struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type jwksHandler struct {
	k port.PublicKeyProvider
}

func (h *jwksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	keys := h.k.PublicKeys()
	body := jwksBody{Keys: make([]jwkData, len(keys))}

	for i, key := range keys {
		jwk := jwkData{
			Kty: key.Type,
			Crv: key.Curve,
			X:   base64.RawURLEncoding.EncodeToString(key.Key),
			Use: "sig",
			Alg: key.Algorithm,
			Kid: key.ID,
		}
		body.Keys[i] = jwk
	}

	resp := response{
		StatusCode: http.StatusOK,
		Body:       body,
	}
	writeJSON(ctx, w, resp)
}
