// Runs the package's tests with an empty home directory, so a storage section in the developer's
// ~/.gavel.yaml can't send their snapshots to a bucket.
package snapshots

import (
	"fmt"
	"os"
	"testing"

	"github.com/flanksource/gavel/internal/safefs"
)

func TestMain(m *testing.M) {
	previousHome, hadHome := os.LookupEnv("HOME")
	home, err := os.MkdirTemp("", "snapshots-home-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create a test home directory: %v\n", err)
		os.Exit(1)
	}
	os.Setenv("HOME", home)
	code := m.Run()
	// SafeRemove refuses the current home, so it is restored first.
	if hadHome {
		os.Setenv("HOME", previousHome)
	} else {
		os.Unsetenv("HOME")
	}
	if err := safefs.SafeRemove(home, safefs.RemoveContext{AllowedRoot: os.TempDir()}); err != nil {
		fmt.Fprintf(os.Stderr, "remove the test home directory: %v\n", err)
	}
	os.Exit(code)
}
