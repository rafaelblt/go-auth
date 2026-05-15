package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/usecase"
)

type registerRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerResponseBody struct {
	User UserResource `json:"user"`
}

type registerHandler struct {
	uc     usecase.Register
	logger *slog.Logger
}

var registerUsernameAlreadyExistsError = errorBody{errorData{
	Code:    "USERNAME_ALREADY_EXISTS",
	Message: "The provided username already exists.",
}}

// validation errors
var (
	registerUsernameTooLongError = fieldErrorData{
		Code:    "USERNAME_TOO_LONG",
		Message: "The provided username is too long.",
	}
	registerUsernameTooShortError = fieldErrorData{
		Code:    "USERNAME_TOO_SHORT",
		Message: "The provided username is too short.",
	}
	registerPasswordTooLongError = fieldErrorData{
		Code:    "PASSWORD_TOO_LONG",
		Message: "The provided password is too long.",
	}
	registerPasswordTooShortError = fieldErrorData{
		Code:    "PASSWORD_TOO_SHORT",
		Message: "The provided password is too short.",
	}
)

func newRegisterHandler(uc usecase.Register) registerHandler {
	return registerHandler{uc, slog.Default()}
}

func (handler registerHandler) Handle(request *http.Request) response {
	ctx := request.Context()
	logger := loggerFrom(ctx)
	handler.logger = logger

	logger.Info("register request received")

	logger.Info("decoding register request")
	var reqBody registerRequestBody
	if err := json.NewDecoder(request.Body).Decode(&reqBody); err != nil {
		logger.Info("failed to decode register request", "error", err)
		return invalidJSONBodyError()
	}

	logger.Info("executing register use case")
	output, err := handler.uc.Execute(ctx, usecase.RegisterInput{
		Username: reqBody.Username,
		Password: reqBody.Password,
	})
	if err != nil {
		return handler.handleUseCaseError(err)
	}

	logger.Info("registration completed successfully")

	logger.Info("mapping register output to response")
	resource, err := MapUserDTOToResource(output.User)
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

func (h registerHandler) handleUseCaseError(err error) response {
	if errors.Is(err, usecase.ErrRegisterUsernameAlreadyExists) {
		h.logger.Info("registration failed: username already exists")
		return h.usernameAlreadyExists()
	}

	var verr usecase.ValidationError
	if errors.As(err, &verr) {
		response, err := h.mapValidationError(verr)
		if err != nil {
			h.logger.Error("register validation error could not be mapped", "error", err)
			return internalServerError()
		}
		h.logger.Info("registration failed: validation error", makeValidationErrorLogFields(response.Errors))
		return validationError(response)
	}

	h.logger.Error("unexpected error from register use case", "error", err)
	return internalServerError()
}

func (h registerHandler) usernameAlreadyExists() response {
	body := registerUsernameAlreadyExistsError
	return response{
		StatusCode: http.StatusConflict,
		Body:       body,
	}
}

func (h registerHandler) mapValidationError(verr usecase.ValidationError) (validationErrorBody, error) {
	body := validationErrorBody{Errors: map[string]ValidationErrors{}}
	usernameErrs := []fieldErrorData{}
	passwordErrs := []fieldErrorData{}

	for _, err := range verr.Errors() {
		if errors.Is(err, usecase.ErrRegisterUsernameTooLong) {
			usernameErrs = append(usernameErrs, registerUsernameTooLongError)
		} else if errors.Is(err, usecase.ErrRegisterUsernameTooShort) {
			usernameErrs = append(usernameErrs, registerUsernameTooShortError)
		} else if errors.Is(err, usecase.ErrRegisterPasswordTooLong) {
			passwordErrs = append(passwordErrs, registerPasswordTooLongError)
		} else if errors.Is(err, usecase.ErrRegisterPasswordTooShort) {
			passwordErrs = append(passwordErrs, registerPasswordTooShortError)
		} else {
			return validationErrorBody{},
				fmt.Errorf("unexpected validation error from use case: %w", err)
		}
	}

	if len(usernameErrs) > 0 {
		body.Errors["username"] = usernameErrs
	}
	if len(passwordErrs) > 0 {
		body.Errors["password"] = passwordErrs
	}
	return body, nil
}
