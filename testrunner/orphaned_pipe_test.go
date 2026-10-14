package testrunner

import (
	"errors"
	osexec "os/exec"
	"testing"

	"github.com/flanksource/clicky/exec"
)

// A backgrounded sleep inherits the shell's stdout and outlives it, which is
// what a Go toolchain helper or a test's grandchild does to a package run.
func runWithOrphan(t *testing.T, exitCode string) *exec.ExecResult {
	t.Helper()
	return exec.NewExec("sh", "-c", "sleep 4 & exit "+exitCode).WithProcessGroup().Run().Result()
}

func TestClearOrphanedPipeErrorPassesACleanExit(t *testing.T) {
	result := runWithOrphan(t, "0")
	if !errors.Is(result.Error, osexec.ErrWaitDelay) {
		t.Fatalf("precondition: want ErrWaitDelay from the orphaned pipe, got %v", result.Error)
	}
	if got := clearOrphanedPipeError(result, "./pkg"); got.Error != nil || got.ExitCode != 0 {
		t.Fatalf("a package that exited 0 must pass, got exit %d err %v", got.ExitCode, got.Error)
	}
}

func TestClearOrphanedPipeErrorKeepsANonZeroExit(t *testing.T) {
	result := runWithOrphan(t, "3")
	if got := clearOrphanedPipeError(result, "./pkg"); got.Error == nil {
		t.Fatalf("a package that exited %d must keep its error", got.ExitCode)
	}
}
