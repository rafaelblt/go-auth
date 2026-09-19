package apitest

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/usecase/register"
)

type FakeRegister struct {
	inputs []register.Input
	output register.Output
	err    error
}

func NewFakeRegister(t *testing.T) *FakeRegister {
	t.Helper()

	fr := FakeRegister{
		inputs: make([]register.Input, 0),
		output: register.Output{User: NewUserDTO(t, nil)},
		err:    nil,
	}
	return &fr
}

func (fr *FakeRegister) Execute(ctx context.Context, in register.Input) (register.Output, error) {
	fr.inputs = append(fr.inputs, in)

	if fr.err != nil {
		return register.Output{}, fr.err
	}

	return fr.output, nil
}

func (fr *FakeRegister) Output() register.Output {
	return fr.output
}

func (fr *FakeRegister) SetError(err error) {
	fr.err = err
}
