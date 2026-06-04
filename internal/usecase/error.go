package usecase

import "fmt"

type ErrorKind string

const ErrorKindConflict ErrorKind = "conflict"

type UseCaseError struct {
	code string
	kind ErrorKind
}

func NewError(code string, kind ErrorKind) UseCaseError {
	return UseCaseError{
		code: code,
		kind: kind,
	}
}

func (uce UseCaseError) Code() string {
	return uce.code
}
func (uce UseCaseError) Kind() ErrorKind {
	return uce.kind
}
func (uce UseCaseError) Error() string {
	return fmt.Sprintf("[%s] %s", uce.code, uce.kind)
}
