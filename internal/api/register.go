package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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

var (
	errRegisterUsernameAlreadyExists = ErrorData{
		Code:    "USERNAME_ALREADY_EXISTS",
		Message: "The provided username already exists.",
	}
	errRegisterUsernameTooLong = FieldErrorData{
		Code:    "USERNAME_TOO_LONG",
		Message: "The provided username is too long.",
	}
	errRegisterUsernameTooShort = FieldErrorData{
		Code:    "USERNAME_TOO_SHORT",
		Message: "The provided username is too short.",
	}
	errRegisterPasswordTooLong = FieldErrorData{
		Code:    "PASSWORD_TOO_LONG",
		Message: "The provided password is too long.",
	}
	errRegisterPasswordTooShort = FieldErrorData{
		Code:    "PASSWORD_TOO_SHORT",
		Message: "The provided password is too short.",
	}
)

func NewRegisterHandler(uc usecase.Register) RegisterHandler {
	return RegisterHandler{uc}
}

func (handler RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.Handle(w, r)
}

func (handler RegisterHandler) Handle(writer http.ResponseWriter, request *http.Request) {
	Log(request.Context(), slog.LevelInfo, "register request received")

	input, err := handler.decodeRequestToInput(request)
	if err != nil {
		writeInternalServerError(writer)
		return
	}

	output, err := handler.uc.Execute(request.Context(), input)
	if err != nil {
		handler.writeError(writer, err)
		return
	}

	handler.writeSuccessResponse(writer, output)
}

func (h RegisterHandler) decodeRequestToInput(r *http.Request) (usecase.RegisterInput, error) {
	var body RegisterBody
	input := usecase.RegisterInput{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return input, err
	}
	input.Username = body.Username
	input.Password = body.Password
	return input, nil
}

func (h RegisterHandler) writeError(writer http.ResponseWriter, err error) {
	if errors.Is(err, usecase.ErrRegisterUsernameAlreadyExists) {
		response := ErrorResponse{Error: errRegisterUsernameAlreadyExists}
		writer.WriteHeader(http.StatusConflict)
		json.NewEncoder(writer).Encode(response)
		return
	}

	var validationErr usecase.ValidationError
	if errors.As(err, &validationErr) {
		h.writeValidationError(writer, validationErr)
		return
	}

	// TODO: log: unexpected error from register use case
	writeInternalServerError(writer)
}

func (h RegisterHandler) writeValidationError(w http.ResponseWriter, verr usecase.ValidationError) {
	response, err := h.mapValidationError(verr)
	if err != nil {
		/// TODO: log: unexpected validation error from register use case
		writeInternalServerError(w)
		return
	}
	writeValidationError(w, response)
}

func (h RegisterHandler) mapValidationError(verr usecase.ValidationError) (ValidationErrorResponse, error) {
	response := ValidationErrorResponse{Errors: map[string]ValidationErrors{}}
	usernameErrs := []FieldErrorData{}
	passwordErrs := []FieldErrorData{}

	for _, err := range verr.Errors().Values() {
		if errors.Is(err, usecase.ErrRegisterUsernameTooLong) {
			usernameErrs = append(usernameErrs, errRegisterUsernameTooLong)
		} else if errors.Is(err, usecase.ErrRegisterUsernameTooShort) {
			usernameErrs = append(usernameErrs, errRegisterUsernameTooShort)
		} else if errors.Is(err, usecase.ErrRegisterPasswordTooLong) {
			passwordErrs = append(passwordErrs, errRegisterPasswordTooLong)
		} else if errors.Is(err, usecase.ErrRegisterPasswordTooShort) {
			passwordErrs = append(passwordErrs, errRegisterPasswordTooShort)
		} else {
			return ValidationErrorResponse{},
				fmt.Errorf("unexpected validation error from register: %w", err)
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

func (h RegisterHandler) writeSuccessResponse(writer http.ResponseWriter, output usecase.RegisterOutput) {
	response := RegisterResponse{}
	resource, err := MapUserDTOToResource(output.User)
	if err != nil {
		// TODO: log: user dto mapping failed
		writeInternalServerError(writer)
		return
	}
	response.User = resource
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(response)
}
