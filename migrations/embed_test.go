package migrations

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLatest(t *testing.T) {
	latest, err := Latest()
	assert.NoError(t, err)
	assert.Equal(t, uint(5), latest)
}

var createTableRe = regexp.MustCompile(`(?is)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([a-z_][a-z0-9_]*)`)
var dropTableRe = regexp.MustCompile(`(?is)DROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?([a-z_][a-z0-9_]*)`)

// TestDownDropsTablesCreatedByUp protects against copy-paste between migration
// files, where a down drops a table that belongs to another migration.
func TestDownDropsTablesCreatedByUp(t *testing.T) {
	entries, err := FS.ReadDir(".")
	require.NoError(t, err)

	var found bool
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		found = true

		t.Run(name, func(t *testing.T) {
			up, err := FS.ReadFile(name)
			require.NoError(t, err)

			downName := strings.TrimSuffix(name, ".up.sql") + ".down.sql"
			down, err := FS.ReadFile(downName)
			require.NoError(t, err, "missing down migration %q", downName)

			created := matchTables(createTableRe, string(up))
			dropped := matchTables(dropTableRe, string(down))

			assert.ElementsMatch(t, created, dropped,
				"%q must drop exactly the tables created by %q", downName, name)
		})
	}

	require.True(t, found, "no up migration found")
}

func matchTables(re *regexp.Regexp, sql string) []string {
	matches := re.FindAllStringSubmatch(sql, -1)
	tables := make([]string, len(matches))
	for i, match := range matches {
		tables[i] = strings.ToLower(match[1])
	}
	return tables
}
