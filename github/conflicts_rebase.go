package github

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/gavel/pr/model"
)

// EnsurePRCommits makes the PR's head and base commits readable in the local
// object store, using the same object-only fetch conflict detection uses. A
// worktree based on the PR head, or a rebase onto its base, needs both.
func EnsurePRCommits(opts Options, pr *model.PRInfo) error {
	if pr == nil || pr.HeadRefOID == "" || pr.BaseRefOID == "" {
		return fmt.Errorf("PR does not report both its head and base commits")
	}
	remote, err := RepoRemote(opts)
	if err != nil {
		return err
	}
	return ensureCommits(opts.WorkDir, remote, pr.Number, pr.BaseRefName, pr.BaseRefOID, pr.HeadRefOID)
}

// BeginConflictRebase starts rebasing dir's checkout onto the PR's base and
// stops at the first commit that conflicts: the commits before it are already
// replayed, its conflicted paths carry markers, and the rebase waits for whoever
// resolves them to `git rebase --continue`. It returns the paths that stop left
// unmerged, tagged with the kinds the report replayed. Later commits may stop
// on conflicts of their own.
//
// The checkout must be clean and contain the PR head, or the rebase would
// rewrite unrelated work.
func BeginConflictRebase(dir string, report *MergeConflictReport) ([]MergeConflict, error) {
	if report == nil || len(report.Files) == 0 {
		reason := "no conflicted files were reported"
		if report != nil && report.Unavailable != "" {
			reason = report.Unavailable
		}
		return nil, fmt.Errorf("cannot start the conflict rebase: %s", reason)
	}
	if _, inProgress, err := RebaseInProgress(dir); err != nil {
		return nil, err
	} else if inProgress {
		return nil, fmt.Errorf("a rebase is already in progress in %s; continue or abort it first", dir)
	}
	if out, err := gitRun(dir, "status", "--porcelain", "--untracked-files=no"); err != nil {
		return nil, fmt.Errorf("git status in %s: %w%s", dir, err, formatGitStderr(out))
	} else if strings.TrimSpace(out) != "" {
		return nil, fmt.Errorf("%s has uncommitted changes; commit them or re-run with --worktree", dir)
	}
	if _, err := gitRun(dir, "merge-base", "--is-ancestor", report.HeadOID, "HEAD"); err != nil {
		return nil, fmt.Errorf("HEAD in %s does not contain the PR head %s (%s); switch to %s or re-run with --worktree",
			dir, shortOID(report.HeadOID), report.HeadRefName, report.HeadRefName)
	}

	before := gitOutput(dir, "rev-parse", "HEAD")
	out, err := gitRun(dir, "-c", "core.editor=true", "rebase", report.BaseOID)
	if err == nil {
		// merge-tree and a commit-by-commit replay disagree, so GitHub's verdict
		// is the odd one out. Put the branch back rather than hand the run a
		// rewritten head nobody asked to push.
		if out, err := gitRun(dir, "reset", "--keep", before); err != nil {
			return nil, fmt.Errorf("restore %s after a clean rebase: %w%s", shortOID(before), err, formatGitStderr(out))
		}
		return nil, fmt.Errorf("git rebased %s onto %s without conflicts; GitHub's CONFLICTING verdict looks stale", report.HeadRefName, report.BaseRefName)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		return nil, fmt.Errorf("git rebase onto %s: %w%s", report.BaseRefName, err, formatGitStderr(out))
	}

	unresolved, inProgress, err := RebaseInProgress(dir)
	if err != nil {
		return nil, err
	}
	if !inProgress || len(unresolved) == 0 {
		return nil, fmt.Errorf("git rebase onto %s failed without stopping on a conflict%s", report.BaseRefName, formatGitStderr(out))
	}
	kinds := make(map[string]string, len(report.Files))
	for _, f := range report.Files {
		kinds[f.Path] = f.Kind
	}
	files := make([]MergeConflict, 0, len(unresolved))
	for _, path := range unresolved {
		files = append(files, MergeConflict{Path: path, Kind: kinds[path]})
	}
	return files, nil
}

// RebaseInProgress reports whether dir is stopped mid-rebase and which paths
// block continuing it: index entries git still holds as unmerged, plus files
// that were `git add`-ed with conflict markers left in them.
func RebaseInProgress(dir string) (unresolved []string, inProgress bool, err error) {
	for _, state := range []string{"rebase-merge", "rebase-apply"} {
		out, err := gitRun(dir, "rev-parse", "--git-path", state)
		if err != nil {
			return nil, false, fmt.Errorf("locate %s in %s: %w%s", state, dir, err, formatGitStderr(out))
		}
		path := strings.TrimSpace(out)
		if !filepath.IsAbs(path) {
			path = filepath.Join(dir, path)
		}
		if _, err := os.Stat(path); err == nil {
			inProgress = true
			break
		} else if !os.IsNotExist(err) {
			return nil, false, err
		}
	}
	if !inProgress {
		return nil, false, nil
	}
	unresolved, err = unmergedPaths(dir)
	if err != nil {
		return nil, true, err
	}
	for _, path := range leftoverMarkerPaths(dir) {
		if !slices.Contains(unresolved, path) {
			unresolved = append(unresolved, path)
		}
	}
	return unresolved, true, nil
}

func unmergedPaths(dir string) ([]string, error) {
	out, err := gitRun(dir, "-c", "core.quotePath=false", "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil, fmt.Errorf("list unmerged paths in %s: %w%s", dir, err, formatGitStderr(out))
	}
	return nonEmptyLines(out), nil
}

// leftoverMarkerPaths reads `git diff --check`'s "leftover conflict marker"
// findings against HEAD. Other --check findings (trailing whitespace) are not
// conflicts and must not block the rebase, so only marker lines are kept. The
// command exits non-zero whenever it finds anything, so its error is expected.
func leftoverMarkerPaths(dir string) []string {
	out, _ := gitRun(dir, "-c", "core.quotePath=false", "diff", "--check", "HEAD")
	var paths []string
	for _, line := range strings.Split(out, "\n") {
		if !strings.HasSuffix(line, "leftover conflict marker") {
			continue
		}
		path, _, ok := strings.Cut(line, ":")
		if ok && !slices.Contains(paths, path) {
			paths = append(paths, path)
		}
	}
	return paths
}

func nonEmptyLines(out string) []string {
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimRight(line, "\r"); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
