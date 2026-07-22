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

func registerDecoder(r *http.Request) (register.Input, error) {
	var body registerRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
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
	resp := response{StatusCode: http.StatusOK, Body: body}
	return resp
}

func registerSuccessLog(ctx context.Context, out register.Output) {
	loggerFrom(ctx).Info("success register", "user_id", out.User.ID)
}
