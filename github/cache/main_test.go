// Runs the package's tests with an empty home and cache directory, so job logs never land in the
// developer's cache and their ~/.gavel.yaml storage can't redirect them.
package cache

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "github-cache-home-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create a test home directory: %v\n", err)
		os.Exit(1)
	}
	os.Setenv("HOME", home)
	os.Setenv("XDG_CACHE_HOME", home+"/.cache")
	code := m.Run()
	os.RemoveAll(home)
	os.Exit(code)
}
