package api

import (
	"net/http"

	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
)

var errorFieldCatalog = map[string]string{
	register.FieldUsername: "username",
	register.FieldPassword: "password",
}

var kindStatusCatalog = map[usecase.ErrorKind]int{
	usecase.ErrorKindConflict:     http.StatusConflict,
	usecase.ErrorKindUnauthorized: http.StatusUnauthorized,
}

var kindMessageCatalog = map[usecase.ErrorKind]string{
	usecase.ErrorKindConflict:     "A conflict error occurred.",
	usecase.ErrorKindUnauthorized: "Not authorized.",
}
