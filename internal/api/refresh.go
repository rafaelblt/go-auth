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

type refreshHandler struct {
	uc refreshUseCase
}

func newRefreshHandler(uc refreshUseCase) *refreshHandler {
	return &refreshHandler{uc}
}

func (handler *refreshHandler) Handle(request *http.Request) response {
	ctx := request.Context()

	var reqBody refreshRequestBody
	if err := json.NewDecoder(request.Body).Decode(&reqBody); err != nil {
		return invalidJSONBodyError()
	}

	output, err := handler.uc.Execute(ctx, refresh.Input{
		RefreshToken: reqBody.RefreshToken,
	})
	if err != nil {
		return translateError(ctx, err)
	}

	accessToken := mapAccessTokenDTO(output.AccessToken)
	refreshToken := mapRefreshTokenDTO(output.RefreshToken)

	response := response{
		StatusCode: http.StatusOK,
		Body: refreshResponseBody{
			AccessToken: accessToken,
			RefreshToken: refreshToken,
		},
	}

	return response
}
