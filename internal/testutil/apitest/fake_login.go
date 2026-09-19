package apitest

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/usecase/login"
)

type FakeLogin struct {
	inputs []login.Input
	output login.Output
	err    error
}

func NewFakeLogin(t *testing.T) *FakeLogin {
	t.Helper()

	fl := FakeLogin{
		inputs: make([]login.Input, 0),
		output: login.Output{
			AccessToken:  NewAccessTokenDTO(t),
			RefreshToken: NewRefreshTokenDTO(t),
		},
		err: nil,
	}
	return &fl
}

func (fl *FakeLogin) Execute(ctx context.Context, in login.Input) (login.Output, error) {
	fl.inputs = append(fl.inputs, in)

	if fl.err != nil {
		return login.Output{}, fl.err
	}

	return fl.output, nil
}

func (fl *FakeLogin) Output() login.Output {
	return fl.output
}

func (fl *FakeLogin) SetError(err error) {
	fl.err = err
}
