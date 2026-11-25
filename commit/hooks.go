package commit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/verify"
)

var ErrHookFailed = errors.New("commit hook failed")

type HookResult struct {
	Name     string `json:"name"`
	Skipped  bool   `json:"skipped,omitempty"`
	ExitCode int    `json:"exit_code,omitempty"`
	Err      error  `json:"-"`
}

func RunHooks(ctx context.Context, workDir string, hooks []verify.CommitHook, stagedFiles []string) ([]HookResult, error) {
	var results []HookResult
	for _, hook := range hooks {
		matches, err := hookMatchesFiles(hook, stagedFiles)
		if err != nil {
			return results, fmt.Errorf("%w: %s: invalid files glob: %w", ErrHookFailed, hook.Name, err)
		}
		if !matches {
			logger.V(1).Infof("commit hook %q skipped: no matching staged files", hook.Name)
			results = append(results, HookResult{Name: hook.Name, Skipped: true})
			continue
		}

		err = runMaintenanceTask(ctx, "Commit hook: "+hook.Name, func(ctx context.Context) error {
			logger.Infof("Running commit hook: %s", hook.Name)
			cmd := exec.CommandContext(ctx, "sh", "-c", hook.Run)
			cmd.Dir, cmd.Env = workDir, gitIndexEnv(workDir)
			cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
			return cmd.Run()
		})

		result := HookResult{Name: hook.Name}
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				result.ExitCode = exitErr.ExitCode()
			} else {
				result.ExitCode = -1
			}
			result.Err = err
			results = append(results, result)
			return results, fmt.Errorf("%w: %s: %w", ErrHookFailed, hook.Name, err)
		}
		results = append(results, result)
	}
	return results, nil
}

func hookMatchesFiles(hook verify.CommitHook, stagedFiles []string) (bool, error) {
	if len(hook.Files) == 0 {
		return true, nil
	}
	for _, glob := range hook.Files {
		for _, f := range stagedFiles {
			ok, err := doublestar.Match(glob, f)
			if err != nil {
				return false, fmt.Errorf("glob %q: %w", glob, err)
			}
			if ok {
				return true, nil
			}
		}
	}
	return false, nil
}
