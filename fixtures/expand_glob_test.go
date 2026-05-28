package fixtures

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsGlobRef(t *testing.T) {
	cases := map[string]bool{
		"@foo/*.md":     true,
		"@foo/bar.md":   false,
		"@a/[abc].txt":  true,
		"@dir/?.md":     true,
		"plain":         false,
		`\@literal`:     false,
		"@":             false,
		"@foo/**/*.md":  true,
	}
	for in, want := range cases {
		assert.Equal(t, want, isGlobRef(in), "isGlobRef(%q)", in)
	}
}

func makeTestNode(name, stdout, stderr string) *FixtureNode {
	return &FixtureNode{
		Name: name,
		Type: TestNode,
		Test: &FixtureTest{
			Name: name,
			Expected: Expectations{
				Stdout: stdout,
				Stderr: stderr,
			},
		},
	}
}

func TestExpandGlobInRows_StdoutGlob(t *testing.T) {
	dir := t.TempDir()
	goldenDir := filepath.Join(dir, "golden")
	require.NoError(t, os.MkdirAll(goldenDir, 0o755))
	for _, name := range []string{"alpha.md", "beta.md", "gamma.md"} {
		writeFile(t, filepath.Join(goldenDir, name), "body of "+name+"\n")
	}

	root := &FixtureNode{
		Name: "root",
		Type: FileNode,
		Children: []*FixtureNode{
			makeTestNode("render", "@golden/*.md", ""),
		},
	}

	require.NoError(t, expandGlobInRows(root, dir))
	require.Len(t, root.Children, 3)

	names := make([]string, 0, len(root.Children))
	filenames := make([]string, 0, len(root.Children))
	for _, c := range root.Children {
		names = append(names, c.Test.Name)
		filenames = append(filenames, c.Test.TemplateVars["filename"].(string))
		// Stdout should now be a concrete @<abs> reference, no glob chars.
		assert.False(t, isGlobRef(c.Test.Expected.Stdout))
		assert.Contains(t, c.Test.Expected.Stdout, "@"+goldenDir)
	}
	sort.Strings(filenames)
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, filenames)
	for _, n := range names {
		assert.Contains(t, n, "render [")
	}
}

func TestExpandGlobInRows_StderrGlob(t *testing.T) {
	dir := t.TempDir()
	errDir := filepath.Join(dir, "errors")
	require.NoError(t, os.MkdirAll(errDir, 0o755))
	writeFile(t, filepath.Join(errDir, "missing.txt"), "missing field\n")
	writeFile(t, filepath.Join(errDir, "bad.txt"), "bad input\n")

	root := &FixtureNode{
		Name: "root",
		Type: FileNode,
		Children: []*FixtureNode{
			makeTestNode("broken", "", "@errors/*.txt"),
		},
	}

	require.NoError(t, expandGlobInRows(root, dir))
	require.Len(t, root.Children, 2)
	for _, c := range root.Children {
		assert.Contains(t, c.Test.Expected.Stderr, "@"+errDir)
		assert.False(t, isGlobRef(c.Test.Expected.Stderr))
	}
}

func TestExpandGlobInRows_NoMatchIsError(t *testing.T) {
	dir := t.TempDir()
	root := &FixtureNode{
		Name: "root",
		Type: FileNode,
		Children: []*FixtureNode{
			makeTestNode("empty", "@nothing/*.md", ""),
		},
	}
	err := expandGlobInRows(root, dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "matched no files")
}

func TestExpandGlobInRows_NonGlobLeftAlone(t *testing.T) {
	dir := t.TempDir()
	root := &FixtureNode{
		Name: "root",
		Type: FileNode,
		Children: []*FixtureNode{
			makeTestNode("literal", "@golden/single.md", ""),
		},
	}
	require.NoError(t, expandGlobInRows(root, dir))
	require.Len(t, root.Children, 1)
	assert.Equal(t, "@golden/single.md", root.Children[0].Test.Expected.Stdout)
}

func TestGlobTemplateVars(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "golden")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	abs := filepath.Join(sub, "alpha.md")
	writeFile(t, abs, "x")

	vars := globTemplateVars(abs, dir)
	assert.Equal(t, "alpha", vars["filename"])
	assert.Equal(t, "alpha.md", vars["basename"])
	assert.Equal(t, ".md", vars["ext"])
	assert.Equal(t, filepath.Join("golden", "alpha.md"), vars["file"])
	assert.Equal(t, "golden", vars["dir"])
	assert.Equal(t, abs, vars["absfile"])
	assert.Equal(t, sub, vars["absdir"])
}
