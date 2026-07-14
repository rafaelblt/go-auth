package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/usecase/login"
)

type loginRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponseBody struct {
	AccessToken  accessToken  `json:"access_token"`
	RefreshToken refreshToken `json:"refresh_token"`
}

type loginUseCase interface {
	Execute(context.Context, login.Input) (login.Output, error)
}

func loginDecoder(r *http.Request) (login.Input, error) {
	var body loginRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return login.Input{}, err
	}
	in := login.Input{
		Username: body.Username,
		Password: body.Password,
	}
	return in, nil
}

func loginEncoder(w http.ResponseWriter, out login.Output) error {
	accessToken := mapAccessTokenDTO(out.AccessToken)
	refreshToken := mapRefreshTokenDTO(out.RefreshToken)

	body := loginResponseBody{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(buf)
	return err
}
