package commit

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"os/exec"
	"regexp"
	"time"

	"github.com/flanksource/commons/logger"
)

// ErrGitLockHeld means a git command kept failing because another git process
// (or a crashed one) holds a .lock file in the repository.
var ErrGitLockHeld = errors.New("git lock file held")

const (
	gitLockRetries   = 3
	gitLockRetryBase = 500 * time.Millisecond
)

var gitLockPattern = regexp.MustCompile(`Unable to create '([^']+\.lock)': File exists`)

// gitLockRetryDelay is the wait before the 1-based retry: exponential backoff
// with up to 100% jitter, so concurrent gavel commits don't retry in lockstep.
var gitLockRetryDelay = func(retry int) time.Duration {
	step := gitLockRetryBase << (retry - 1)
	return step + rand.N(step)
}

// heldGitLock returns the lock file git reported it could not create because
// it already exists (index.lock, HEAD.lock, refs/heads/<branch>.lock, ...).
func heldGitLock(output string) (string, bool) {
	match := gitLockPattern.FindStringSubmatch(output)
	if match == nil {
		return "", false
	}
	return match[1], true
}

// runGitRetryingLocks runs `git args...` in workDir and returns its combined
// output. When git fails because a lock file is held, it retries up to
// gitLockRetries times with jittered backoff. The lock file is never removed:
// it may belong to a live git process.
func runGitRetryingLocks(workDir string, args ...string) ([]byte, error) {
	for retry := 1; ; retry++ {
		cmd := exec.Command("git", args...)
		cmd.Dir = workDir
		out, err := cmd.CombinedOutput()
		if err == nil {
			return out, nil
		}
		lock, held := heldGitLock(string(out))
		if !held {
			return out, err
		}
		if retry > gitLockRetries {
			return out, fmt.Errorf("%w: %s still held after %d retries (another git process is running, or a crashed one left it behind — remove it only if no git process is running): %w",
				ErrGitLockHeld, lock, gitLockRetries, err)
		}
		delay := gitLockRetryDelay(retry)
		logger.Warnf("git %s: %s is held by another git process, retrying in %s (%d/%d)",
			args[0], lock, delay.Round(time.Millisecond), retry, gitLockRetries)
		time.Sleep(delay)
	}
}
