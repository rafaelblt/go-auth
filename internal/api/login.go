package api

import (
	"context"
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
	accessToken := mapAccessTokenDTO(out.AccessToken)
	refreshToken := mapRefreshTokenDTO(out.RefreshToken)

	body := loginResponseBody{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	resp := response{
		StatusCode: http.StatusOK,
		Body:       body,
	}

	return resp
}

func loginSuccessLog(ctx context.Context, out login.Output) {
	loggerFrom(ctx).Info("success login",
		"user_id", out.UserID,
		"session_id", out.SessionID)
}
