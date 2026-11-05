package uirexport

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/utils"
)

// module is the Go module uir indexed: the one holding the run's working directory. Export paths are
// relative to its root, with forward slashes, and its packages are named by import path.
type module struct {
	root string
	path string
	// roots caches the module root of each directory looked up.
	roots map[string]string
}

func findModule(workDir string) (*module, error) {
	root := utils.FindNearestGoModRoot(workDir)
	if root == "" {
		return nil, fmt.Errorf("no go.mod at or above %s", workDir)
	}
	path, ok := parsers.GoModuleName(root)
	if !ok {
		return nil, fmt.Errorf("read the module path from %s", filepath.Join(root, "go.mod"))
	}
	return &module{root: root, path: path, roots: map[string]string{}}, nil
}

// rootOf is the root of the module holding dir, or "" outside any.
func (m *module) rootOf(dir string) string {
	if root, ok := m.roots[dir]; ok {
		return root
	}
	root := utils.FindNearestGoModRoot(dir)
	m.roots[dir] = root
	return root
}

// relative is file's module-relative slash path, or "" when the file lies outside the module or in a
// nested module, whose files uir does not serve for this root. A relative file is taken from workDir.
func (m *module) relative(workDir, file string) string {
	if file == "" {
		return ""
	}
	if !filepath.IsAbs(file) {
		file = filepath.Join(workDir, file)
	}
	rel, err := filepath.Rel(m.root, file)
	if err != nil || !filepath.IsLocal(rel) || m.rootOf(filepath.Dir(file)) != m.root {
		return ""
	}
	return filepath.ToSlash(rel)
}

// holds reports whether test's package lies in the module itself, not in a nested module gavel ran as
// a root of its own, whose files uir does not serve for this root.
func (m *module) holds(test parsers.Test) bool {
	return m.rootOf(filepath.Join(test.WorkDir, test.PackagePath)) == m.root
}

// importPath is the Go import path of the package directory dir in the module.
func (m *module) importPath(dir string) (string, bool) {
	rel, err := filepath.Rel(m.root, dir)
	if err != nil || !filepath.IsLocal(rel) {
		return "", false
	}
	if rel == "." {
		return m.path, true
	}
	return m.path + "/" + filepath.ToSlash(rel), true
}

// profileFile maps a coverage profile's file name, an import path plus file name, to a module-relative
// path, or "" when it names a file outside the module.
func (m *module) profileFile(name string) string {
	rel, ok := strings.CutPrefix(name, m.path+"/")
	if !ok {
		return ""
	}
	return m.relative(m.root, filepath.FromSlash(rel))
}
