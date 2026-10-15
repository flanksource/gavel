// Runs the package's tests with an empty home directory, so a storage section in the developer's
// ~/.gavel.yaml can't send their snapshots to a bucket.
package snapshots

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "snapshots-home-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create a test home directory: %v\n", err)
		os.Exit(1)
	}
	os.Setenv("HOME", home)
	code := m.Run()
	os.RemoveAll(home)
	os.Exit(code)
}
