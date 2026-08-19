package migrations

import (
	"embed"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

//go:embed *.sql
var FS embed.FS

var ErrNoMigrations = errors.New("no migrations found")

func Latest() (uint, error) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		return 0, fmt.Errorf("read embedded migrations failed: %w", err)
	}

	var latest uint
	var found bool

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		prefix, _, ok := strings.Cut(name, "_")
		if !ok {
			continue
		}

		version, err := strconv.ParseUint(prefix, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid migration version in %q: %w", name, err)
		}

		if !found || uint(version) > latest {
			latest = uint(version)
			found = true
		}
	}

	if !found {
		return 0, ErrNoMigrations
	}

	return latest, nil
}
