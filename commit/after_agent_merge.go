package commit

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/flanksource/gavel/github"
)

// ErrMergeUnresolved is returned while an agent run's tree is mid-merge with
// paths still conflicted. It is not a failure of the turn: the merge stays in
// place for the next turn to finish, and the run's verify commands say which
// paths remain.
var ErrMergeUnresolved = errors.New("merge unresolved")

// concludeAgentMerge finishes a merge an agent resolved. git refuses the partial
// `commit -- <paths>` the ordinary pipeline cuts while MERGE_HEAD exists, and the
// pipeline's explicit-file staging resets the index — which would throw away the
// merge's already-staged side — so a merge is concluded here instead: stage the
// paths the turn touched, refuse while anything is still conflicted, then commit
// the whole index as the merge commit.
func concludeAgentMerge(ctx context.Context, opts Options) (*Result, error) {
	if len(opts.Files) > 0 {
		if out, err := gitCombined(opts.WorkDir, append([]string{"add", "--"}, opts.Files...)...); err != nil {
			return nil, fmt.Errorf("stage resolved merge paths: %w: %s", err, out)
		}
	}
	unresolved, _, err := github.MergeInProgress(opts.WorkDir)
	if err != nil {
		return nil, err
	}
	if len(unresolved) > 0 {
		return nil, fmt.Errorf("%w (%s)", ErrMergeUnresolved, strings.Join(unresolved, ", "))
	}
	if opts.DryRun {
		fmt.Fprintln(dryRunOutput, "would conclude the in-progress merge")
		return &Result{DryRun: true}, nil
	}
	if out, err := gitCombined(opts.WorkDir, "commit", "--no-edit"); err != nil {
		return nil, fmt.Errorf("conclude merge: %w: %s", err, out)
	}
	hash, err := gitCombined(opts.WorkDir, "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("read merge commit: %w: %s", err, hash)
	}
	message, err := gitCombined(opts.WorkDir, "log", "-1", "--format=%B")
	if err != nil {
		return nil, fmt.Errorf("read merge commit message: %w: %s", err, message)
	}
	result := &Result{Commits: []CommitResult{{Hash: strings.TrimSpace(hash), Message: strings.TrimSpace(message), Files: opts.Files}}}
	if opts.Push {
		if err := pushAfterCommit(ctx, opts, result); err != nil {
			return result, err
		}
	}
	return result, nil
}

func gitCombined(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}
