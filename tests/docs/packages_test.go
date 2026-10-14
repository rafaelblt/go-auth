package docs

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Every package states what it is and names the document covering it. That
// comment is how a reader who opened a file finds docs/ at all, since this
// project keeps its reasoning there rather than in comments.
func TestEveryPackageHasAPackageComment(t *testing.T) {
	documented := map[string]bool{}
	packages := map[string]bool{}

	for _, path := range goFiles(t) {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		dir := filepath.Dir(path)
		packages[dir] = true

		body := readFile(t, path)
		if strings.HasPrefix(body, "// Package ") || strings.HasPrefix(body, "// Command ") {
			documented[dir] = true
		}
	}

	for dir := range packages {
		if !documented[dir] {
			t.Errorf("%s has no package comment; add one naming the document that"+
				" covers it, as every other package does", rel(dir))
		}
	}
}

// A mapper is the only way to build a DTO because every use case sits in a
// subpackage of internal/usecase, where the DTOs' unexported fields are out of
// reach. A use case in internal/usecase itself could fill a field by hand, the
// fields would still look protective, and nothing else would notice. An
// Execute method in a top-level file is what that would look like.
func TestNoUseCaseLivesInTheUsecasePackageItself(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(repoRoot, "internal/usecase/*.go"))
	require.NoError(t, err)
	require.NotEmpty(t, files)

	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		if strings.Contains(readFile(t, path), ") Execute(") {
			t.Errorf("%s declares Execute; a use case belongs in a subpackage of"+
				" internal/usecase, or its DTOs stop going through a mapper"+
				" (docs/architecture/usecases.md#dtos)", rel(path))
		}
	}
}
