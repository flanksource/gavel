package commit

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/flanksource/clicky/prompt"
	"github.com/flanksource/commons/logger"
)

// ErrGitLockHeld means a git command kept failing because another git process
// (or a crashed one) holds a .lock file in the repository.
var ErrGitLockHeld = errors.New("git lock file held")

const (
	gitLockRetries   = 3
	gitLockRetryBase = 500 * time.Millisecond
	// gitStaleLockAge is how old a lock nobody has open must be before it is
	// offered for removal.
	gitStaleLockAge = 5 * time.Second
)

var gitLockPattern = regexp.MustCompile(`Unable to create '([^']+\.lock)': File exists`)

// gitLockRetryDelay is the wait before the 1-based retry: exponential backoff
// with up to 100% jitter, so concurrent gavel commits don't retry in lockstep.
var gitLockRetryDelay = func(retry int) time.Duration {
	step := gitLockRetryBase << (retry - 1)
	return step + rand.N(step)
}

// gitLockPolicy decides what happens to a stale lock: Mode is the resolved
// --precommit mode (fail reports it, anything else asks), and AssumeYes removes
// it without asking.
type gitLockPolicy struct {
	Mode      string
	AssumeYes bool
}

type gitLockPolicyKey struct{}

// withGitLockPolicy attaches the stale-lock policy Run resolved from its
// options, so the git helpers it calls inherit it alongside any prompt.Scope.
func withGitLockPolicy(ctx context.Context, p gitLockPolicy) context.Context {
	return context.WithValue(ctx, gitLockPolicyKey{}, p)
}

// gitLockPolicyFrom returns the policy on ctx, defaulting to prompt like
// resolvePrecommitMode does.
func gitLockPolicyFrom(ctx context.Context) gitLockPolicy {
	if p, ok := ctx.Value(gitLockPolicyKey{}).(gitLockPolicy); ok {
		return p
	}
	return gitLockPolicy{Mode: CheckModePrompt}
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
// gitLockRetries times with jittered backoff. If the lock is still there and is
// stale — older than gitStaleLockAge with no process holding it open — it is
// removed (after asking, per the gitLockPolicy on ctx) and the command runs once
// more. At most one lock is removed per call.
func runGitRetryingLocks(ctx context.Context, workDir string, args ...string) ([]byte, error) {
	for retry := 1; ; retry++ {
		out, err := execGit(workDir, args)
		if err == nil {
			return out, nil
		}
		lock, held := heldGitLock(string(out))
		if !held {
			return out, err
		}
		if retry > gitLockRetries {
			return retryAfterStaleLock(ctx, workDir, args, lock, err)
		}
		delay := gitLockRetryDelay(retry)
		logger.Warnf("git %s: %s is held by another git process, retrying in %s (%d/%d)",
			args[0], lock, delay.Round(time.Millisecond), retry, gitLockRetries)
		time.Sleep(delay)
	}
}

func execGit(workDir string, args []string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	return cmd.CombinedOutput()
}

func retryAfterStaleLock(ctx context.Context, workDir string, args []string, lock string, gitErr error) ([]byte, error) {
	if !filepath.IsAbs(lock) {
		lock = filepath.Join(workDir, lock)
	}
	if reason := removeStaleGitLock(ctx, lock); reason != nil {
		return nil, fmt.Errorf("%w: %s still held after %d retries (%w): %w",
			ErrGitLockHeld, lock, gitLockRetries, reason, gitErr)
	}
	out, err := execGit(workDir, args)
	if err == nil {
		return out, nil
	}
	if next, held := heldGitLock(string(out)); held {
		return out, fmt.Errorf("%w: %s held after removing stale %s: %w", ErrGitLockHeld, next, lock, err)
	}
	return out, err
}

// gitLockDiagnosis is what is known about a held lock file.
type gitLockDiagnosis struct {
	info    os.FileInfo
	age     time.Duration
	holders []string
}

func (d gitLockDiagnosis) stale() bool {
	return d.age > gitStaleLockAge && len(d.holders) == 0
}

func (d gitLockDiagnosis) String() string {
	holder := "no process has it open"
	if len(d.holders) > 0 {
		holder = "open by pid " + strings.Join(d.holders, ", ")
	}
	return fmt.Sprintf("age %s, %d bytes, %s", d.age.Round(time.Second), d.info.Size(), holder)
}

func diagnoseGitLock(lock string) (gitLockDiagnosis, error) {
	info, err := os.Lstat(lock)
	if err != nil {
		return gitLockDiagnosis{}, fmt.Errorf("stat %s: %w", lock, err)
	}
	holders, err := lockHolders(lock)
	if err != nil {
		return gitLockDiagnosis{}, err
	}
	return gitLockDiagnosis{info: info, age: time.Since(info.ModTime()), holders: holders}, nil
}

// lockHolders lists the PIDs that have path open. lsof exits 1 with no output
// when nothing has it open; any other failure means holders cannot be verified.
func lockHolders(path string) ([]string, error) {
	cmd := exec.Command("lsof", "-t", "--", path)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 && len(out) == 0 && strings.TrimSpace(stderr.String()) == "" {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot verify who holds %s: lsof: %w: %s", path, err, strings.TrimSpace(stderr.String()))
	}
	return splitLines(string(out)), nil
}

// removeStaleGitLock removes lock when it is stale and the policy allows it,
// returning nil once removed or the reason it was kept.
func removeStaleGitLock(ctx context.Context, lock string) error {
	diag, err := diagnoseGitLock(lock)
	if err != nil {
		return err
	}
	if !diag.stale() {
		return fmt.Errorf("not stale: %s", diag)
	}
	if err := confirmStaleLockRemoval(ctx, lock, diag); err != nil {
		return err
	}
	current, err := os.Lstat(lock)
	if err != nil {
		return fmt.Errorf("re-stat %s: %w", lock, err)
	}
	if !os.SameFile(diag.info, current) || !current.ModTime().Equal(diag.info.ModTime()) || current.Size() != diag.info.Size() {
		return fmt.Errorf("changed while deciding, so it is in use")
	}
	if err := os.Remove(lock); err != nil {
		return fmt.Errorf("remove stale lock: %w", err)
	}
	logger.Warnf("removed stale git lock %s (%s)", lock, diag)
	return nil
}

func confirmStaleLockRemoval(ctx context.Context, lock string, diag gitLockDiagnosis) error {
	policy := gitLockPolicyFrom(ctx)
	switch {
	case policy.Mode == CheckModeFail:
		return fmt.Errorf("stale: %s; remove it if no git process is running, or rerun with --precommit=prompt", diag)
	case policy.AssumeYes:
		return nil
	case !stdinIsTerminal() && !prompt.HasInteractiveSink():
		return fmt.Errorf("stale: %s; cannot ask to remove it without a terminal", diag)
	}
	title := fmt.Sprintf("Stale git lock detected: %s (%s). Remove and retry?", lock, diag)
	if choice, ok := promptSelectIndex(ctx, title, []string{"Remove and retry", "Cancel"}); !ok || choice != 0 {
		return fmt.Errorf("stale: %s; removal declined", diag)
	}
	return nil
}
