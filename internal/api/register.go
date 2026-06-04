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
	User userResource `json:"user"`
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
	logger := loggerFrom(ctx)

	logger.Info("register request received")

	logger.Info("decoding register request")
	var reqBody registerRequestBody
	if err := json.NewDecoder(request.Body).Decode(&reqBody); err != nil {
		logger.Info("failed to decode register request", "error", err)
		return invalidJSONBodyError()
	}

	logger.Info("executing register use case")
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

	logger.Info("mapping register output to response")
	resource, err := mapUserDTOToResource(output.User)
	if err != nil {
		logger.Error("failed to map user dto to resource", "error", err)
		return internalServerError()
	}

	response := response{
		StatusCode: http.StatusOK,
		Body:       registerResponseBody{User: resource},
	}

	logger.Info("register response successfully returned")
	return response
}
