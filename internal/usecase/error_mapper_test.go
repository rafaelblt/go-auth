package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapErrors_ReturnsExpectedAndUnexpectedErrors(t *testing.T) {
	err1 := errors.New("example error 1")
	err2 := errors.New("example error 2")
	err3 := errors.New("example error 3")
	errA := errors.New("example error A")
	errB := errors.New("example error B")
	errC := errors.New("example error C")
	testCases := []struct {
		desc       string
		given      []error
		errmap     ErrorsMap
		expected   []error
		unexpected []error
	}{
		{
			desc: "3 expected of 3 given errors",
			given: []error{err1, err2, err3},
			errmap: ErrorsMap{
				err1: errA,
				err2: errB,
				err3: errC},
			expected: []error{errA, errB, errC},
			unexpected: []error{},
		},
		{
			desc: "2 expected of 2 given errors",
			given: []error{err1, err2},
			errmap: ErrorsMap{
				err1: errA,
				err2: errB,
				err3: errC},
			expected: []error{errA, errB},
			unexpected: []error{},
		},
		{
			desc: "all given errors are unexpected",
			given: []error{err2, err3},
			errmap: ErrorsMap{
				err1: errA},
			expected: []error{},
			unexpected: []error{err2, err3},
		},
		{
			desc: "1 expected and 1 unexpected",
			given: []error{err1, err3},
			errmap: ErrorsMap{
				err1: errA,
				err2: errB},
			expected: []error{errA},
			unexpected: []error{err3},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mapped, unexpected := MapErrors(tC.given, tC.errmap)
			assert.Equal(t, tC.expected, mapped)
			assert.Equal(t, tC.unexpected, unexpected)
		})
	}
}
