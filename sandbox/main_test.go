// Runs the package's tests with a git configuration of their own, which turns off git's automatic
// maintenance: a fetch or push starts it detached, and its repack races a test's TempDir cleanup.
// The test binary also stands in for the substrate environment's program and its job.
package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/internal/safefs"
)

func TestMain(m *testing.M) {
	// The substrate executor's tests run the test binary as the environment's program and as the
	// job.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case substrateRunTaskArg:
			os.Exit(runSubstrateTask())
		case substrateJobArg:
			os.Exit(runSubstrateJob(os.Args[2:]))
		}
	}
	dir, err := os.MkdirTemp("", "sandbox-gitconfig-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create a test git configuration: %v\n", err)
		os.Exit(1)
	}
	config := filepath.Join(dir, "gitconfig")
	if err := os.WriteFile(config, []byte("[maintenance]\n\tauto = false\n[gc]\n\tauto = 0\n"), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write the test git configuration: %v\n", err)
		os.Exit(1)
	}
	os.Setenv("GIT_CONFIG_GLOBAL", config)
	code := m.Run()
	if err := safefs.SafeRemove(dir, safefs.RemoveContext{AllowedRoot: os.TempDir()}); err != nil {
		fmt.Fprintf(os.Stderr, "remove the test git configuration: %v\n", err)
	}
	os.Exit(code)
}
