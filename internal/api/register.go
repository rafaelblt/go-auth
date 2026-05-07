package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/usecase"
)

var Dependencies *infra.DependencyContainer

type RegisterBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	User UserResource `json:"user"`
}

type RegisterHandler struct {
	uc usecase.Register
}

var registerUsernameAlreadyExistsError = ErrorResponse{ErrorData{
	Code:    "USERNAME_ALREADY_EXISTS",
	Message: "The provided username already exists.",
}}

var (
	registerUsernameTooLongError = FieldErrorData{
		Code:    "USERNAME_TOO_LONG",
		Message: "The provided username is too long.",
	}
	registerUsernameTooShortError = FieldErrorData{
		Code:    "USERNAME_TOO_SHORT",
		Message: "The provided username is too short.",
	}
	registerPasswordTooLongError = FieldErrorData{
		Code:    "PASSWORD_TOO_LONG",
		Message: "The provided password is too long.",
	}
	registerPasswordTooShortError = FieldErrorData{
		Code:    "PASSWORD_TOO_SHORT",
		Message: "The provided password is too short.",
	}
)

func NewRegisterHandler(uc usecase.Register) RegisterHandler {
	return RegisterHandler{uc}
}

func (h RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Handle(w, r)
}

func (handler RegisterHandler) Handle(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	logger := loggerFrom(ctx)
	logger.Info("request received")

	input, err := handler.decodeRequestToInput(request)
	if err != nil {
		invalidJSONBodyError(ctx, writer, err)
		return
	}

	output, err := handler.uc.Execute(ctx, input)
	if err != nil {
		handler.handleUseCaseError(ctx, writer, err)
		return
	}

	handler.success(ctx, writer, output)
}

func (h RegisterHandler) decodeRequestToInput(r *http.Request) (usecase.RegisterInput, error) {
	var body RegisterBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return usecase.RegisterInput{},
			fmt.Errorf("register body decode failed: %w", err)
	}
	input := usecase.RegisterInput{}
	input.Username = body.Username
	input.Password = body.Password
	return input, nil
}

func (h RegisterHandler) handleUseCaseError(ctx context.Context, w http.ResponseWriter, err error) {
	if errors.Is(err, usecase.ErrRegisterUsernameAlreadyExists) {
		h.usernameAlreadyExists(ctx, w)
		return
	}

	var verr usecase.ValidationError
	if errors.As(err, &verr) {
		response, err := h.mapValidationError(verr)
		if err != nil {
			internalError(ctx, w, "validation error mapping failed", err)
		} else {
			validationError(ctx, w, response)
		}
		return
	}

	internalError(ctx, w, "unexpected error from use case", err)
}

func (h RegisterHandler) usernameAlreadyExists(ctx context.Context, w http.ResponseWriter)  {
	w.WriteHeader(http.StatusConflict)
	response := registerUsernameAlreadyExistsError
	if err := json.NewEncoder(w).Encode(response); err != nil {
		internalError(ctx, w, "failed to encode username already exists response", err)
	}
}

func (h RegisterHandler) mapValidationError(verr usecase.ValidationError) (ValidationErrorResponse, error) {
	response := ValidationErrorResponse{Errors: map[string]ValidationErrors{}}
	usernameErrs := []FieldErrorData{}
	passwordErrs := []FieldErrorData{}

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
			return ValidationErrorResponse{},
				fmt.Errorf("unexpected validation error from use case: %w", err)
		}
	}

	if len(usernameErrs) > 0 {
		response.Errors["username"] = usernameErrs
	}
	if len(passwordErrs) > 0 {
		response.Errors["password"] = passwordErrs
	}
	return response, nil
}

func (h RegisterHandler) success(ctx context.Context, w http.ResponseWriter, output usecase.RegisterOutput) {
	resource, err := MapUserDTOToResource(output.User)
	if err != nil {
		internalError(ctx, w, "failed to map user dto to resource", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := RegisterResponse{User: resource}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		internalError(ctx, w, "failed to encode register success response", err)
	}
}
