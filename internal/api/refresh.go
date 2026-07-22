package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
)

type refreshRequestBody struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponseBody struct {
	AccessToken  accessToken  `json:"access_token"`
	RefreshToken refreshToken `json:"refresh_token"`
}

type refreshUseCase interface {
	Execute(context.Context, refresh.Input) (refresh.Output, error)
}

func refreshDecoder(r *http.Request) (refresh.Input, error) {
	var body refreshRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return refresh.Input{}, err
	}
	in := refresh.Input{
		RefreshToken: body.RefreshToken,
	}
	return in, nil
}

func refreshEncoder(out refresh.Output) response {
	accessToken := mapAccessTokenDTO(out.AccessToken)
	refreshToken := mapRefreshTokenDTO(out.RefreshToken)

	body := refreshResponseBody{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	resp := response{
		StatusCode: http.StatusOK,
		Body:       body,
	}

	return resp
}

func refreshSuccessLog(ctx context.Context, out refresh.Output) {
	loggerFrom(ctx).Info("success refresh") // TODO: add user id?
}
