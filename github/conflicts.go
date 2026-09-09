package github

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/flanksource/commons/logger"
)

// MergeConflict is one path that cannot be merged cleanly, tagged with the kind
// of conflict git reported for it (content, modify/delete, add/add, …).
type MergeConflict struct {
	Path string `json:"path"`
	Kind string `json:"kind,omitempty"`
}

// MergeConflictReport is the local reconstruction of the conflicts behind a PR
// GitHub reports as CONFLICTING. No GitHub API returns the conflicted paths —
// only the verdict — so the file list comes from replaying the same merge with
// `git merge-tree` against the two commits GitHub itself merges.
type MergeConflictReport struct {
	BaseRefName string          `json:"baseRefName"`
	BaseOID     string          `json:"baseOid,omitempty"`
	HeadRefName string          `json:"headRefName"`
	HeadOID     string          `json:"headOid,omitempty"`
	MergeBase   string          `json:"mergeBase,omitempty"`
	Files       []MergeConflict `json:"files,omitempty"`
	// Unavailable explains why Files is empty. A conflicting PR rendered with
	// no files and no explanation reads as "conflicts with nothing", so every
	// path that cannot produce a file list says why instead of going quiet.
	Unavailable string `json:"unavailable,omitempty"`
}

// ResolveCommands are the git commands that reproduce and fix the conflict in
// the user's checkout, in the `$ cmd` shape the gavel results section uses.
func (r MergeConflictReport) ResolveCommands() []string {
	if r.BaseRefName == "" {
		return nil
	}
	return []string{
		fmt.Sprintf("git fetch origin %s", r.BaseRefName),
		fmt.Sprintf("git switch %s", r.HeadRefName),
		fmt.Sprintf("git merge origin/%s", r.BaseRefName),
	}
}

// DetectMergeConflicts replays the PR's merge locally and reports which paths
// conflict. It returns nil for any PR GitHub does not call CONFLICTING — a
// clean PR has nothing to explain, and replaying its merge would spend a fetch
// on every poll of `pr status --follow`.
func DetectMergeConflicts(opts Options, pr *PRInfo) *MergeConflictReport {
	if pr == nil || !pr.IsConflicting() {
		return nil
	}

	report := &MergeConflictReport{
		BaseRefName: pr.BaseRefName,
		BaseOID:     pr.BaseRefOID,
		HeadRefName: pr.HeadRefName,
		HeadOID:     pr.HeadRefOID,
	}
	if report.BaseOID == "" || report.HeadOID == "" {
		report.Unavailable = "GitHub did not report both merge endpoints (the base branch may have been deleted)"
		return report
	}

	remote, err := conflictRemote(opts)
	if err != nil {
		report.Unavailable = err.Error()
		return report
	}
	if err := ensureCommits(opts.WorkDir, remote, pr.Number, pr.BaseRefName, report.BaseOID, report.HeadOID); err != nil {
		report.Unavailable = err.Error()
		return report
	}

	report.MergeBase = gitOutput(opts.WorkDir, "merge-base", report.BaseOID, report.HeadOID)
	files, err := mergeTreeConflicts(opts.WorkDir, report.BaseOID, report.HeadOID)
	if err != nil {
		report.Unavailable = err.Error()
		return report
	}
	if len(files) == 0 {
		// git and GitHub disagree. GitHub caches mergeability and recomputes it
		// asynchronously, so this is usually a stale verdict rather than a bug —
		// say so, instead of showing a conflict section with nothing in it.
		report.Unavailable = fmt.Sprintf("git merges %s into %s cleanly — GitHub's CONFLICTING verdict looks stale; push or reopen to make it recompute",
			report.BaseRefName, report.HeadRefName)
		return report
	}
	report.Files = files
	return report
}

// conflictRemote finds the remote that actually points at the PR's repository.
// Falling back to origin would be worse than reporting nothing: run from an
// unrelated checkout, `refs/pull/<n>/head` resolves to a different project's
// PR and the conflict list would be confidently wrong.
func conflictRemote(opts Options) (string, error) {
	repo, err := opts.resolveRepo()
	if err != nil {
		return "", fmt.Errorf("no local checkout to replay the merge in: %w", err)
	}

	out := gitOutput(opts.WorkDir, "remote")
	if out == "" {
		return "", fmt.Errorf("no git remotes found; run from a checkout of %s to list the conflicting files", repo)
	}
	for _, name := range strings.Fields(out) {
		url := gitOutput(opts.WorkDir, "remote", "get-url", name)
		if url == "" {
			continue
		}
		if parsed, err := parseGitHubRepo(url); err == nil && strings.EqualFold(parsed, repo) {
			return name, nil
		}
	}
	return "", fmt.Errorf("no git remote points at %s; run from a checkout of it to list the conflicting files", repo)
}

// ensureCommits makes both merge endpoints readable locally. It fetches the
// PR head ref and the base branch rather than the bare OIDs, because a fork's
// head commit is only reachable through refs/pull/<n>/head on the base repo.
// --refmap= and --no-write-fetch-head keep it an object-store-only fetch: a
// status command must not move the user's remote-tracking branches.
func ensureCommits(workDir, remote string, prNumber int, baseRef string, oids ...string) error {
	if missing := missingCommits(workDir, oids...); len(missing) == 0 {
		return nil
	}

	args := []string{"fetch", "--no-write-fetch-head", "--refmap=", "--quiet", remote,
		fmt.Sprintf("refs/pull/%d/head", prNumber)}
	if baseRef != "" {
		args = append(args, "refs/heads/"+baseRef)
	}
	logger.Debugf("fetching merge endpoints for PR #%d from %s", prNumber, remote)
	if out, err := gitRun(workDir, args...); err != nil {
		return fmt.Errorf("could not fetch the merge endpoints from %s: %w%s", remote, err, formatGitStderr(out))
	}

	if missing := missingCommits(workDir, oids...); len(missing) > 0 {
		return fmt.Errorf("commits %s are still missing after fetching from %s", strings.Join(missing, ", "), remote)
	}
	return nil
}

func missingCommits(workDir string, oids ...string) []string {
	var missing []string
	for _, oid := range oids {
		if _, err := gitRun(workDir, "cat-file", "-e", oid+"^{commit}"); err != nil {
			missing = append(missing, shortOID(oid))
		}
	}
	return missing
}

// conflictKindPattern captures the parenthesised kind out of git's own
// "CONFLICT (content): Merge conflict in go.mod" informational lines.
var conflictKindPattern = regexp.MustCompile(`^CONFLICT \(([^)]+)\)`)

// mergeTreeConflicts runs the merge without touching the working tree or index.
// `--write-tree` prints the merged tree OID, then the conflicted paths, then a
// blank line, then git's own conflict messages; it exits 1 when there are
// conflicts, 0 when the merge is clean, and >1 when it could not run at all.
func mergeTreeConflicts(workDir, baseOID, headOID string) ([]MergeConflict, error) {
	// core.quotePath=false keeps a non-ASCII path readable instead of rendering
	// it as the octal escapes git quotes by default.
	out, err := gitRun(workDir, "-c", "core.quotePath=false",
		"merge-tree", "--write-tree", "--name-only", baseOID, headOID)
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			return nil, fmt.Errorf("git merge-tree failed: %w%s", err, formatGitStderr(out))
		}
	}

	// A clean merge stops after the tree OID; a conflicted one continues with
	// the paths, a blank separator, and git's own conflict messages.
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 2 {
		return nil, nil
	}
	paths, messages := lines[1:], []string(nil)
	if idx := indexOfBlank(paths); idx >= 0 {
		paths, messages = paths[:idx], paths[idx+1:]
	}

	kinds := conflictKinds(paths, messages)
	conflicts := make([]MergeConflict, 0, len(paths))
	for _, path := range paths {
		conflicts = append(conflicts, MergeConflict{Path: path, Kind: kinds[path]})
	}
	return conflicts, nil
}

// conflictKinds attributes each "CONFLICT (kind)" message to a path. A message
// names its paths inline, so the longest path it contains wins — matching the
// first would tag go.mod for a message that is really about go.mod.orig.
func conflictKinds(paths []string, messages []string) map[string]string {
	kinds := make(map[string]string, len(paths))
	for _, msg := range messages {
		m := conflictKindPattern.FindStringSubmatch(msg)
		if m == nil {
			continue
		}
		best := ""
		for _, path := range paths {
			if strings.Contains(msg, path) && len(path) > len(best) {
				best = path
			}
		}
		if best != "" {
			kinds[best] = m[1]
		}
	}
	return kinds
}

func indexOfBlank(lines []string) int {
	for i, line := range lines {
		if line == "" {
			return i
		}
	}
	return -1
}

func shortOID(oid string) string {
	if len(oid) > 8 {
		return oid[:8]
	}
	return oid
}

// gitRun returns combined output so a failure carries git's stderr, which is
// the only place `merge-tree` and `fetch` explain themselves.
func gitRun(workDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func gitOutput(workDir string, args ...string) string {
	out, err := gitRun(workDir, args...)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func formatGitStderr(out string) string {
	out = strings.TrimSpace(out)
	if out == "" {
		return ""
	}
	return ": " + strings.ReplaceAll(out, "\n", "; ")
}
