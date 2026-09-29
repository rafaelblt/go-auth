package docs

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const (
	openAPISpecFile = "docs/api/openapi.yaml"
	routerFile      = "internal/api/router.go"
)

var (
	routePattern = regexp.MustCompile(`mux\.Handle(?:Func)?\("([A-Z]+) (/[^"]*)"`)
	httpMethods  = map[string]bool{
		"get": true, "put": true, "post": true, "delete": true,
		"options": true, "head": true, "patch": true, "trace": true,
	}
)

// Every route the router registers is in the OpenAPI spec, and the reverse. A
// route added without the spec leaves generated clients without it, and nothing
// else notices.
func TestOpenAPISpecMatchesTheRoutes(t *testing.T) {
	router := readFile(t, filepath.Join(repoRoot, routerFile))
	routes := []string{}
	for _, match := range routePattern.FindAllStringSubmatch(router, -1) {
		routes = append(routes, match[1]+" "+match[2])
	}
	require.NotEmpty(t, routes, "no mux.Handle or mux.HandleFunc call found in %s;"+
		" the way the router registers routes may have changed", routerFile)

	var spec struct {
		Paths map[string]map[string]any `yaml:"paths"`
	}
	content := readFile(t, filepath.Join(repoRoot, openAPISpecFile))
	require.NoError(t, yaml.Unmarshal([]byte(content), &spec))

	operations := []string{}
	for path, item := range spec.Paths {
		for key := range item {
			if httpMethods[key] {
				operations = append(operations, strings.ToUpper(key)+" "+path)
			}
		}
	}

	require.ElementsMatch(t, routes, operations,
		"the routes in %s and the operations in %s differ", routerFile, openAPISpecFile)
}
