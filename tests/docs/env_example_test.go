package docs

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	envExampleFile = ".env.example"
	envVarsFile    = "internal/config/env_vars.go"
)

var (
	envKeyPattern     = regexp.MustCompile(`Key:\s+"([A-Z0-9_]+)"`)
	envExamplePattern = regexp.MustCompile(`(?m)^(?:# )?([A-Z0-9_]+)=`)
)

// Every variable internal/config reads is in .env.example, and the reverse.
// Platforms import the variables from it, so one left out is one a deployment
// never learns about, and nothing else notices.
func TestEnvExampleMatchesTheConfigVariables(t *testing.T) {
	envVars := readFile(t, filepath.Join(repoRoot, envVarsFile))
	keys := []string{}
	for _, match := range envKeyPattern.FindAllStringSubmatch(envVars, -1) {
		keys = append(keys, match[1])
	}
	require.NotEmpty(t, keys, "no Key field found in %s; the way the variables are"+
		" declared may have changed", envVarsFile)

	example := readFile(t, filepath.Join(repoRoot, envExampleFile))
	listed := []string{}
	for _, match := range envExamplePattern.FindAllStringSubmatch(example, -1) {
		listed = append(listed, match[1])
	}

	require.ElementsMatch(t, keys, listed,
		"the variables in %s and the ones in %s differ", envVarsFile, envExampleFile)
}
