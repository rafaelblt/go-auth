package docs

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const decisionsDir = "docs/development/decisions"

// A record must be linked from a comment on the code it shaped, which the
// decisions README requires. The code is where someone about to break a decision
// is looking, and forgetting the link is silent.
func TestEveryAcceptedDecisionIsLinkedFromCode(t *testing.T) {
	code := allGoCode(t)

	for _, record := range decisionRecords(t) {
		if record.status != accepted {
			continue
		}
		if !strings.Contains(code, record.name) {
			t.Errorf("no .go file mentions %s; add a comment linking it from the"+
				" code it shaped", record.name)
		}
	}
}

// Superseding a record means repointing what referred to it, so that the code
// names the decision in force rather than the history behind it.
func TestSupersededDecisionsAreNotLinkedFromCode(t *testing.T) {
	code := allGoCode(t)

	for _, record := range decisionRecords(t) {
		if record.status != superseded {
			continue
		}
		if strings.Contains(code, record.name) {
			t.Errorf("a .go file still mentions %s, which is superseded; point it"+
				" at the record that replaced it", record.name)
		}
	}
}

// The invariant of decision 0034: a mapper is the only way to build a DTO, and
// that holds because every use case sits in a subpackage of internal/usecase,
// where the DTOs' unexported fields are out of reach. A use case in
// internal/usecase itself could fill a field by hand, the fields would still
// look protective, and nothing else would notice. An Execute method in a
// top-level file is what that would look like.
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
				" (decision 0034)", rel(path))
		}
	}
}

type decisionStatus int

const (
	accepted decisionStatus = iota
	superseded
	other
)

type decisionRecord struct {
	name   string // the file name, as a code comment would spell it
	status decisionStatus
}

// decisionRecords returns the NNNN-slug.md records, with README.md and
// template.md left out.
func decisionRecords(t *testing.T) []decisionRecord {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(repoRoot, decisionsDir, "[0-9]*.md"))
	require.NoError(t, err)
	require.NotEmpty(t, paths, "no decision records found")

	records := make([]decisionRecord, 0, len(paths))
	for _, path := range paths {
		records = append(records, decisionRecord{
			name:   filepath.Base(path),
			status: statusOf(readFile(t, path)),
		})
	}
	return records
}

func statusOf(record string) decisionStatus {
	switch {
	case strings.Contains(record, "**Status:** Accepted"):
		return accepted
	case strings.Contains(record, "**Status:** Superseded"):
		return superseded
	default:
		return other
	}
}

// allGoCode concatenates every .go file, so that one Contains answers whether
// anything at all mentions a record.
func allGoCode(t *testing.T) string {
	t.Helper()
	all := strings.Builder{}

	for _, path := range goFiles(t) {
		all.WriteString(readFile(t, path))
	}
	return all.String()
}
