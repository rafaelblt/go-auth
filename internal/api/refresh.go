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

func refreshEncoder(w http.ResponseWriter, out refresh.Output) error {
	accessToken := mapAccessTokenDTO(out.AccessToken)
	refreshToken := mapRefreshTokenDTO(out.RefreshToken)

	body := refreshResponseBody{
		AccessToken: accessToken,
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
