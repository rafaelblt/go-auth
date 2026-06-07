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

type loginHandler struct {
	uc loginUseCase
}

func newLoginHandler(uc loginUseCase) loginHandler {
	return loginHandler{uc}
}

func (handler loginHandler) Handle(request *http.Request) response {
	ctx := request.Context()

	var reqBody loginRequestBody
	if err := json.NewDecoder(request.Body).Decode(&reqBody); err != nil {
		return invalidJSONBodyError()
	}

	output, err := handler.uc.Execute(ctx, login.Input{
		Username: reqBody.Username,
		Password: reqBody.Password,
	})
	if err != nil {
		resp := translateError(ctx, err)
		return resp
	}

	accessToken := mapAccessTokenDTO(output.AccessToken)
	refreshToken := mapRefreshTokenDTO(output.RefreshToken)

	response := response{
		StatusCode: http.StatusOK,
		Body: loginResponseBody{
			AccessToken: accessToken,
			RefreshToken: refreshToken,
		},
	}

	return response
}
