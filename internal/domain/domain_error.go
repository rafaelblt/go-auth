package domain

import "fmt"

type DomainError struct {
	code    string
	message string
}

func (e DomainError) Code() string { return e.code }
func (e DomainError) Message() string { return e.message }

func (e DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code(), e.Message())
}

func NewDomainError(code string, message string) DomainError {
	return DomainError{code: code, message: message}
}
