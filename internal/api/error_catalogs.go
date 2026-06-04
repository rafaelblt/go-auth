package api

import (
	"github.com/rafaelblt/go-auth/internal/usecase/register"
)

var errorFieldCatalog = map[string]string{
	register.FieldUsername: "username",
	register.FieldPassword: "password",
}
