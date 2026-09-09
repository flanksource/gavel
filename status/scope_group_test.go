package status

import (
	"testing"

	"github.com/flanksource/repomap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func scopeGroupPaths(g ScopeGroup) []string {
	paths := make([]string, len(g.Files))
	for i, f := range g.Files {
		paths[i] = f.Path
	}
	return paths
}

func TestGroupByScope(t *testing.T) {
	files := []FileStatus{
		{Path: "api/a.go", FileMap: &repomap.FileMap{Language: "go", Scopes: repomap.Scopes{repomap.ScopeType("api")}}},
		{Path: "api/a_test.go", FileMap: &repomap.FileMap{Language: "go", Scopes: repomap.Scopes{repomap.ScopeTypeTest}}},
		{Path: "api/b.go", FileMap: &repomap.FileMap{Language: "go", Scopes: repomap.Scopes{repomap.ScopeType("api")}}},
		{Path: "README.md"},
	}

	groups := GroupByScope(files)

	// Render order: feature scopes first, then unscoped/general, then test-only.
	require.Len(t, groups, 3)

	assert.Equal(t, "go · api", groups[0].Label)
	assert.Equal(t, []string{"api/a.go", "api/b.go"}, scopeGroupPaths(groups[0]))

	assert.Equal(t, string(repomap.ScopeTypeGeneral), groups[1].Label)
	assert.Equal(t, []string{"README.md"}, scopeGroupPaths(groups[1]))

	assert.Equal(t, "go · "+string(repomap.ScopeTypeTest), groups[2].Label)
	assert.Equal(t, []string{"api/a_test.go"}, scopeGroupPaths(groups[2]))
}

func TestResultScopeGroupsUsesFiles(t *testing.T) {
	r := &Result{Files: []FileStatus{
		{Path: "x.go", FileMap: &repomap.FileMap{Language: "go", Scopes: repomap.Scopes{repomap.ScopeType("api")}}},
	}}

	groups := r.ScopeGroups()
	require.Len(t, groups, 1)
	assert.Equal(t, "go · api", groups[0].Label)
	assert.Equal(t, []string{"x.go"}, scopeGroupPaths(groups[0]))
}
