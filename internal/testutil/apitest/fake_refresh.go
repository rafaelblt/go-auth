package apitest

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
)

type FakeRefresh struct {
	inputs []refresh.Input
	output refresh.Output
	err    error
}

func NewFakeRefresh(t *testing.T) *FakeRefresh {
	t.Helper()

	fr := FakeRefresh{
		inputs: make([]refresh.Input, 0),
		output: refresh.Output{
			AccessToken:  NewAccessTokenDTO(t),
			RefreshToken: NewRefreshTokenDTO(t),
		},
		err: nil,
	}
	return &fr
}

func (r *FakeRefresh) Execute(ctx context.Context, in refresh.Input) (refresh.Output, error) {
	r.inputs = append(r.inputs, in)

	if r.err != nil {
		return refresh.Output{}, r.err
	}

	return r.output, nil
}

func (r *FakeRefresh) Inputs() []refresh.Input {
	return r.inputs
}

func (r *FakeRefresh) Output() refresh.Output {
	return r.output
}

func (r *FakeRefresh) SetError(err error) {
	r.err = err
}
