// Package docs checks the few documentation rules that are worth not having to
// remember. See docs/development/decisions/README.md.
package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	decisionsDir = "../../docs/development/decisions"
	codeRoot     = "../.."
)

// Every accepted record must be linked from a comment on the code it shaped.
// The rule is in the decisions README; this is only so that forgetting it fails
// instead of going unnoticed.
func TestEveryAcceptedDecisionIsLinkedFromCode(t *testing.T) {
	code := readAllGoFiles(t)

	for _, name := range decisionFiles(t) {
		if !isAccepted(t, filepath.Join(decisionsDir, name)) {
			continue
		}
		if !strings.Contains(code, name) {
			t.Errorf("no .go file mentions %s;"+
				" add a comment linking it from the code it shaped", name)
		}
	}
}

// decisionFiles returns the NNNN-slug.md file names, template.md and README.md
// excluded.
func decisionFiles(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(decisionsDir)
	if err != nil {
		t.Fatalf("read decisions dir: %v", err)
	}

	names := []string{}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".md") && name[0] >= '0' && name[0] <= '9' {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		t.Fatal("no decision records found")
	}
	return names
}

func isAccepted(t *testing.T, path string) bool {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return strings.Contains(string(content), "**Status:** Accepted")
}

// readAllGoFiles concatenates every .go file in the repository, so that a
// single Contains answers whether anything mentions a record.
func readAllGoFiles(t *testing.T) string {
	t.Helper()
	all := strings.Builder{}

	err := filepath.WalkDir(codeRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "ignore") {
			return filepath.SkipDir
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".go") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			all.Write(content)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repository: %v", err)
	}
	return all.String()
}
