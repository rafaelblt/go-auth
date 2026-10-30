package config

import (
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"
)

// The rules config owns: the ones about a variable itself rather than about
// the code its value is passed to, whose rules that code's package exports.

func required(value string) error {
	if value == "" {
		return errors.New("must be set")
	}
	return nil
}

// hostPort is what net.Listen reads of an address before it resolves
// anything, so that an address the server could never listen on is reported
// with the other settings, not once the app is built.
func hostPort(address string) error {
	_, port, err := net.SplitHostPort(address)
	if err == nil {
		_, err = net.LookupPort("tcp", port)
	}
	if err != nil {
		return fmt.Errorf("must be host:port, such as localhost:8080 (%w)", err)
	}
	return nil
}

// oneOf is the rule of a setting with a closed set of values.
func oneOf[T ~string](allowed ...T) func(T) error {
	names := make([]string, len(allowed))
	for i, value := range allowed {
		names[i] = string(value)
	}
	list := strings.Join(names[:len(names)-1], ", ") + " or " + names[len(names)-1]

	return func(value T) error {
		if slices.Contains(allowed, value) {
			return nil
		}
		return fmt.Errorf("must be %s, not %q", list, value)
	}
}
