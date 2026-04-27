package runners

import (
	"go/build"
	"path/filepath"
)

// matchesBuildContext reports whether file (a single _test.go path) is
// compiled under the default Go build context (current GOOS/GOARCH and the
// caller's build tags). Files excluded by //go:build constraints return
// false; matching, malformed, or unreadable files return true so we keep
// the previous permissive behavior on edge cases.
//
// Used by package discovery so directories whose only test files are
// excluded by build constraints (e.g. //go:build never) are not surfaced
// as runnable packages.
func matchesBuildContext(file string) bool {
	dir, name := filepath.Split(file)
	if dir == "" {
		dir = "."
	}
	matched, err := build.Default.MatchFile(dir, name)
	if err != nil {
		return true
	}
	return matched
}
