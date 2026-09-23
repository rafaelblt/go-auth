package docs

import (
	"path/filepath"
	"strings"
	"testing"
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
