package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/usecase/register"
)

type registerRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerResponseBody struct {
	User user `json:"user"`
}

type registerUseCase interface {
	Execute(context.Context, register.Input) (register.Output, error)
}

type registerHandler struct {
	uc registerUseCase
}

func newRegisterHandler(uc registerUseCase) registerHandler {
	return registerHandler{uc}
}

func (handler registerHandler) Handle(request *http.Request) response {
	ctx := request.Context()

	var reqBody registerRequestBody
	if err := json.NewDecoder(request.Body).Decode(&reqBody); err != nil {
		return invalidJSONBodyError()
	}

	output, err := handler.uc.Execute(ctx, register.Input{
		Username: reqBody.Username,
		Password: reqBody.Password,
	})
	if err != nil {
		resp, err := adaptError(err)
		if err != nil {
			logger.Error("an error occurred while adapting the register error", "error", err)
			return internalServerError()
		}
		return resp
	}

	user := mapUserDTO(output.User)

	response := response{
		StatusCode: http.StatusOK,
		Body:       registerResponseBody{User: user},
	}

	return response
}
