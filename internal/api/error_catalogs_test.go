package api

import (
	"net/http"
	"testing"

	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorFieldCatalog(t *testing.T) {
	testCases := []struct {
		desc     string
		field    string
	}{
		{
			desc:     "register username field",
			field:    register.FieldUsername,
		},
		{
			desc:     "register password field",
			field:    register.FieldPassword,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			retrieved, ok := errorFieldCatalog[tC.field]
			require.True(t, ok)
			assert.NotEmpty(t, retrieved)
		})
	}
}

func TestKindStatusCatalog(t *testing.T) {
	testCases := []struct {
		desc string
		kind usecase.ErrorKind
	}{
		{
			desc: "kind conflict",
			kind: usecase.ErrorKindConflict,
		},
		{
			desc: "kind unauthorized",
			kind: usecase.ErrorKindUnauthorized,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			retrieved, ok := kindStatusCatalog[tC.kind]
			require.True(t, ok)
			status := http.StatusText(retrieved)
			assert.NotZero(t, status)
		})
	}
}

func TestKindMessageCatalog(t *testing.T) {
	testCases := []struct {
		desc string
		kind usecase.ErrorKind
	}{
		{
			desc: "kind conflict",
			kind: usecase.ErrorKindConflict,
		},
		{
			desc: "kind unauthorized",
			kind: usecase.ErrorKindUnauthorized,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			retrieved, ok := kindMessageCatalog[tC.kind]
			require.True(t, ok)
			assert.NotEmpty(t, retrieved)
		})
	}
}
