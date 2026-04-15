package verify

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/flanksource/commons/logger"
)

// RunPushHooks executes each HookStep sequentially as a shell command
// rooted at workDir. Used by `gavel test` to run the `pre:` and `post:`
// lists from .gavel.yaml around the test run.
//
// Semantics:
//   - Each step runs via `sh -c` so users can use pipes, &&, redirects, etc.
//   - cmd.Dir = workDir (the repo root or the pushed worktree).
//   - stdout/stderr stream to os.Stderr so output reaches the pusher's push
//     output via SSH (or the local terminal in `gavel test`).
//   - The first failing step aborts with a wrapped error; remaining steps
//     do not run.
//
// Callers handle post-hook "failure doesn't mask main exit" by running
// post hooks separately and ignoring/logging the returned error.
func RunPushHooks(workDir string, hooks []HookStep, phase string) error {
	for _, step := range hooks {
		if step.Run == "" {
			continue
		}
		name := step.Name
		if name == "" {
			name = phase
		}
		logger.Infof("===== %s: %s =====", phase, name)
		cmd := exec.Command("sh", "-c", step.Run)
		cmd.Dir = workDir
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s hook %q failed: %w", phase, name, err)
		}
	}
	return nil
}
