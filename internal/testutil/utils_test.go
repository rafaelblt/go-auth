package testutil_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestOnly_ReturnsTheSingleElement(t *testing.T) {
	type person struct {
		name string
		age  int
	}

	t.Run("int slice", func(t *testing.T) {
		result := testutil.Only(t, []int{42})

		assert.Equal(t, 42, result)
	})

	t.Run("string slice", func(t *testing.T) {
		result := testutil.Only(t, []string{"hello"})

		assert.Equal(t, "hello", result)
	})

	t.Run("struct slice", func(t *testing.T) {
		expected := person{name: "Alice", age: 30}

		result := testutil.Only(t, []person{expected})

		assert.Equal(t, expected, result)
	})

	t.Run("pointer slice", func(t *testing.T) {
		expected := &person{name: "Bob", age: 25}

		result := testutil.Only(t, []*person{expected})

		assert.Same(t, expected, result)
	})
}
