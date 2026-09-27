package api

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
)

type user struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func mapUserDTO(dto usecase.UserDTO) user {
	if dto.IsZero() {
		panic("the mapUserDTO() func received a zero UserDTO")
	}
	user := user{
		ID:        dto.ID(),
		Username:  dto.Username(),
		Status:    dto.Status(),
		CreatedAt: dto.CreatedAt(),
		UpdatedAt: dto.UpdatedAt(),
	}
	return user
}

type token struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
	ExpiresIn int64     `json:"expires_in"`
}

func mapAccessTokenDTO(dto usecase.AccessTokenDTO) token {
	if dto.IsZero() {
		panic("the mapAccessTokenDTO() func received a zero AccessTokenDTO")
	}
	return token{
		Value:     dto.Value(),
		ExpiresAt: dto.ExpiresAt(),
		ExpiresIn: int64(dto.ExpiresIn() / time.Second),
	}
}

func mapRefreshTokenDTO(dto usecase.RefreshTokenDTO) token {
	if dto.IsZero() {
		panic("the mapRefreshTokenDTO() func received a zero RefreshTokenDTO")
	}
	return token{
		Value:     dto.Value(),
		ExpiresAt: dto.ExpiresAt(),
		ExpiresIn: int64(dto.ExpiresIn() / time.Second),
	}
}

type credentialsRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokensResponseBody struct {
	AccessToken  token `json:"access_token"`
	RefreshToken token `json:"refresh_token"`
}

func tokensResponse(accessDTO usecase.AccessTokenDTO, refreshDTO usecase.RefreshTokenDTO) response {
	body := tokensResponseBody{
		AccessToken:  mapAccessTokenDTO(accessDTO),
		RefreshToken: mapRefreshTokenDTO(refreshDTO),
	}
	return response{StatusCode: http.StatusOK, Body: body}
}

type registerResponseBody struct {
	User user `json:"user"`
}

func registerDecoder(r *http.Request) (register.Input, error) {
	var body credentialsRequestBody
	if err := decodeJSONBody(r, &body); err != nil {
		return register.Input{}, err
	}
	in := register.Input{
		Username: body.Username,
		Password: body.Password,
	}
	return in, nil
}

func registerEncoder(out register.Output) response {
	usr := mapUserDTO(out.User)
	body := registerResponseBody{User: usr}
	resp := response{StatusCode: http.StatusCreated, Body: body}
	return resp
}

func registerSuccessLog(ctx context.Context, out register.Output) {
	loggerFrom(ctx).Info("success register", "user_id", out.User.ID())
}

func loginDecoder(r *http.Request) (login.Input, error) {
	var body credentialsRequestBody
	if err := decodeJSONBody(r, &body); err != nil {
		return login.Input{}, err
	}
	in := login.Input{
		Username: body.Username,
		Password: body.Password,
	}
	return in, nil
}

func loginEncoder(out login.Output) response {
	return tokensResponse(out.AccessToken, out.RefreshToken)
}

func loginSuccessLog(ctx context.Context, out login.Output) {
	loggerFrom(ctx).Info("success login",
		"user_id", out.UserID,
		"session_id", out.SessionID)
}

type refreshRequestBody struct {
	RefreshToken string `json:"refresh_token"`
}

func refreshDecoder(r *http.Request) (refresh.Input, error) {
	var body refreshRequestBody
	if err := decodeJSONBody(r, &body); err != nil {
		return refresh.Input{}, err
	}
	in := refresh.Input{
		RefreshToken: body.RefreshToken,
	}
	return in, nil
}

func refreshEncoder(out refresh.Output) response {
	return tokensResponse(out.AccessToken, out.RefreshToken)
}

func refreshSuccessLog(ctx context.Context, out refresh.Output) {
	loggerFrom(ctx).Info("success refresh",
		"user_id", out.UserID,
		"session_id", out.SessionID)
}

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
