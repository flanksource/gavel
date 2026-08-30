package parsers

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// GoModuleName reads the module path from rootWorkDir/go.mod.
//
// Parsers need it to normalize file paths: builds run with -trimpath report
// locations as import paths (github.com/org/repo/pkg/foo_test.go) rather than
// filesystem paths, so stripping the module prefix is what makes them relative
// to the work dir.
func GoModuleName(rootWorkDir string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(rootWorkDir, "go.mod"))
	if err != nil {
		return "", false
	}
	f, err := modfile.Parse("go.mod", data, nil)
	if err != nil || f.Module == nil {
		return "", false
	}
	name := strings.TrimSpace(f.Module.Mod.Path)
	return name, name != ""
}
