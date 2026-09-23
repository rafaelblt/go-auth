package docs

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Every relative link between documents must resolve, file and anchor. The
// documentation cross-links densely, and a renamed heading breaks an anchor
// without breaking anything a reader would notice until they follow it.
func TestMarkdownLinksResolve(t *testing.T) {
	for _, path := range markdownFiles(t) {
		body := withoutCode(readFile(t, path))

		for _, match := range markdownURL.FindAllStringSubmatch(body, -1) {
			target := match[1]
			if isExternal(target) || isPlaceholder(target) {
				continue
			}

			file, anchor := splitAnchor(target)
			resolved := path
			if file != "" {
				resolved = filepath.Join(filepath.Dir(path), file)
				if !fileExists(resolved) {
					t.Errorf("%s links to %s, which does not exist", rel(path), target)
					continue
				}
			}
			if anchor != "" && !headingSlugs(t, resolved)[anchor] {
				t.Errorf("%s links to %s, but that document has no such heading",
					rel(path), target)
			}
		}
	}
}

// Package comments and decision links point at documents by path. They are the
// only route from a file to the document explaining it, so a stale one is worse
// than none.
func TestDocLinksInCodeCommentsResolve(t *testing.T) {
	docPath := regexp.MustCompile(`docs/[a-zA-Z0-9/._-]+\.md(#[a-z0-9_-]+)?`)

	for _, path := range goFiles(t) {
		for _, target := range docPath.FindAllString(readFile(t, path), -1) {
			file, anchor := splitAnchor(target)
			resolved := filepath.Join(repoRoot, file)

			if !fileExists(resolved) {
				t.Errorf("%s mentions %s, which does not exist", rel(path), target)
				continue
			}
			if anchor != "" && !headingSlugs(t, resolved)[anchor] {
				t.Errorf("%s mentions %s, but that document has no such heading",
					rel(path), target)
			}
		}
	}
}

func isExternal(target string) bool {
	return strings.HasPrefix(target, "http") || strings.HasPrefix(target, "mailto")
}

// isPlaceholder skips the NNNN-slug.md the template and the lifecycle rules use
// to show the shape of a record's file name.
func isPlaceholder(target string) bool {
	return strings.Contains(target, "NNNN")
}

func splitAnchor(target string) (file, anchor string) {
	file, anchor, _ = strings.Cut(target, "#")
	return file, anchor
}

func rel(path string) string {
	return strings.TrimPrefix(path, repoRoot+"/")
}
