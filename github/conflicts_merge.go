package github

import (
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/flanksource/gavel/pr/model"
)

// EnsurePRCommits makes the PR's head and base commits readable in the local
// object store, using the same object-only fetch conflict detection uses. A
// worktree based on the PR head, or a merge of its base, needs both.
func EnsurePRCommits(opts Options, pr *model.PRInfo) error {
	if pr == nil || pr.HeadRefOID == "" || pr.BaseRefOID == "" {
		return fmt.Errorf("PR does not report both its head and base commits")
	}
	remote, err := conflictRemote(opts)
	if err != nil {
		return err
	}
	return ensureCommits(opts.WorkDir, remote, pr.Number, pr.BaseRefName, pr.BaseRefOID, pr.HeadRefOID)
}

// BeginConflictMerge starts merging the PR's base into dir's checkout and stops
// with the conflicts in the tree: the cleanly merged side is staged, the
// conflicted paths carry markers, and MERGE_HEAD records the merge for whoever
// concludes it. It returns the paths git left unmerged, tagged with the kinds
// the report replayed.
//
// The checkout must be clean and contain the PR head, or the merge would be cut
// on top of unrelated work.
func BeginConflictMerge(dir string, report *MergeConflictReport) ([]MergeConflict, error) {
	if report == nil || len(report.Files) == 0 {
		reason := "no conflicted files were reported"
		if report != nil && report.Unavailable != "" {
			reason = report.Unavailable
		}
		return nil, fmt.Errorf("cannot start the conflict merge: %s", reason)
	}
	if _, inProgress, err := MergeInProgress(dir); err != nil {
		return nil, err
	} else if inProgress {
		return nil, fmt.Errorf("a merge is already in progress in %s; conclude or abort it first", dir)
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

	message := fmt.Sprintf("Merge %s into %s", report.BaseRefName, report.HeadRefName)
	out, err := gitRun(dir, "merge", "--no-ff", "--no-commit", "-m", message, report.BaseOID)
	var exitErr *exec.ExitError
	if err != nil && (!errors.As(err, &exitErr) || exitErr.ExitCode() != 1) {
		return nil, fmt.Errorf("git merge %s: %w%s", report.BaseRefName, err, formatGitStderr(out))
	}

	unresolved, err := unmergedPaths(dir)
	if err != nil {
		return nil, err
	}
	if len(unresolved) == 0 {
		return nil, fmt.Errorf("git merged %s into %s without conflicts; GitHub's CONFLICTING verdict looks stale", report.BaseRefName, report.HeadRefName)
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

// MergeInProgress reports whether dir is mid-merge and which paths still block
// concluding it: index entries git still holds as unmerged, plus files that were
// `git add`-ed with conflict markers left in them.
func MergeInProgress(dir string) (unresolved []string, inProgress bool, err error) {
	if _, err := gitRun(dir, "rev-parse", "-q", "--verify", "MERGE_HEAD"); err != nil {
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
// conflicts and must not block the merge, so only marker lines are kept. The
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
