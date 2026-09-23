package docs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const repoRoot = "../.."

// skippedDirs are not part of the documented codebase: .git is plumbing and
// ignore/ holds notes and drafts that are not published.
var skippedDirs = map[string]bool{".git": true, "ignore": true}

// rootMarkdown are the documents outside docs/ that are still part of it.
var rootMarkdown = []string{"README.md", "CONTRIBUTING.md", "AGENTS.md", "CLAUDE.md"}

// filesWithSuffix walks the repository and returns every file under dir whose
// name ends in suffix, as paths usable from this package.
func filesWithSuffix(t *testing.T, dir, suffix string) []string {
	t.Helper()
	found := []string{}

	err := filepath.WalkDir(filepath.Join(repoRoot, dir), func(
		path string, entry os.DirEntry, err error,
	) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && skippedDirs[entry.Name()] {
			return filepath.SkipDir
		}
		if !entry.IsDir() && strings.HasSuffix(path, suffix) {
			found = append(found, path)
		}
		return nil
	})
	require.NoError(t, err)

	return found
}

// markdownFiles returns every document: docs/ plus the ones in the repository
// root. A root file that does not exist is skipped, so renaming CLAUDE.md to
// AGENTS.md or back needs no change here.
func markdownFiles(t *testing.T) []string {
	t.Helper()
	files := filesWithSuffix(t, "docs", ".md")

	for _, name := range rootMarkdown {
		path := filepath.Join(repoRoot, name)
		if _, err := os.Stat(path); err == nil {
			files = append(files, path)
		}
	}
	return files
}

func goFiles(t *testing.T) []string {
	t.Helper()
	files := []string{}

	for _, dir := range []string{"internal", "cmd", "migrations", "tests"} {
		files = append(files, filesWithSuffix(t, dir, ".go")...)
	}
	return files
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}

var (
	fencedBlock = regexp.MustCompile("(?s)```.*?```")
	inlineCode  = regexp.MustCompile("`[^`\n]*`")
	headingLine = regexp.MustCompile(`(?m)^#{1,6} +(.+)$`)
	markdownURL = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]+)\)`)
	unslugged   = regexp.MustCompile(`[^a-z0-9 _-]`)
	inlineLink  = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
)

// withoutCode strips fenced blocks and inline code, so that a link shown as an
// example is not read as a link to follow.
func withoutCode(markdown string) string {
	return inlineCode.ReplaceAllString(fencedBlock.ReplaceAllString(markdown, ""), "")
}

// headingSlugs returns the anchors a document offers, by the rules GitHub uses:
// lower case, formatting and punctuation dropped, spaces turned into hyphens.
func headingSlugs(t *testing.T, path string) map[string]bool {
	t.Helper()
	slugs := map[string]bool{}

	for _, match := range headingLine.FindAllStringSubmatch(readFile(t, path), -1) {
		slugs[slugify(match[1])] = true
	}
	return slugs
}

func slugify(heading string) string {
	text := inlineLink.ReplaceAllString(heading, "$1")
	text = strings.NewReplacer("`", "", "*", "", "_", "_").Replace(text)
	text = unslugged.ReplaceAllString(strings.ToLower(text), "")
	return strings.ReplaceAll(strings.TrimSpace(text), " ", "-")
}
