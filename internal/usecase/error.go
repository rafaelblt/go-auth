package usecase

type ErrorKind string

const (
	ErrorKindConflict     ErrorKind = "conflict"
	ErrorKindUnauthorized ErrorKind = "unauthorized"
)

type UseCaseError struct {
	code   string
	kind   ErrorKind
	reason string
}

func NewError(code string, kind ErrorKind) UseCaseError {
	return UseCaseError{
		code: code,
		kind: kind,
	}
}

func NewErrorWithReason(code string, kind ErrorKind, reason string) UseCaseError {
	return UseCaseError{
		code:   code,
		kind:   kind,
		reason: reason,
	}
}

func (uce UseCaseError) Code() string {
	return uce.code
}
func (uce UseCaseError) Kind() ErrorKind {
	return uce.kind
}
func (uce UseCaseError) Reason() string {
	return uce.reason
}
func (uce UseCaseError) Error() string {
	return uce.code
}
