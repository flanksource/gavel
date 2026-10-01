package prcreate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/logger"
)

// preflight validates the source repo and base ref and resolves every SHA to
// its full form, preserving order.
func preflight(repoRoot string, in Input) ([]string, error) {
	if len(in.SHAs) == 0 {
		return nil, errors.New("prcreate: at least one SHA is required")
	}
	if strings.TrimSpace(in.Base) == "" {
		return nil, errors.New("prcreate: base ref is required")
	}
	if err := runGitQuiet(repoRoot, "rev-parse", "--git-dir"); err != nil {
		return nil, fmt.Errorf("not a git repository: %s", repoRoot)
	}
	if err := runGitQuiet(repoRoot, "rev-parse", "--verify", in.Base); err != nil {
		return nil, fmt.Errorf("base ref %q not resolvable", in.Base)
	}
	if err := RefuseInProgressOps(repoRoot); err != nil {
		return nil, err
	}
	if dirty, _ := gitWorkingTreeDirty(repoRoot); dirty {
		logger.Warnf("source repo has uncommitted changes; the new worktree is isolated, but check you're not mid-something")
	}
	full := make([]string, 0, len(in.SHAs))
	for _, sha := range in.SHAs {
		resolved, err := resolveCommit(repoRoot, sha, in.Mainline)
		if err != nil {
			return nil, err
		}
		full = append(full, resolved)
	}
	return full, nil
}

func resolveCommit(repoRoot, sha string, mainline int) (string, error) {
	if err := runGitQuiet(repoRoot, "cat-file", "-e", sha+"^{commit}"); err != nil {
		return "", fmt.Errorf("commit %q not found in %s", sha, repoRoot)
	}
	full, err := captureGit(repoRoot, "rev-parse", sha)
	if err != nil {
		return "", fmt.Errorf("resolve SHA %q: %w", sha, err)
	}
	if len(full) < 8 {
		return "", fmt.Errorf("git rev-parse returned unexpectedly short SHA %q", full)
	}
	parents, err := captureGit(repoRoot, "rev-list", "--parents", "-n", "1", full)
	if err != nil {
		return "", fmt.Errorf("rev-list parents: %w", err)
	}
	if len(strings.Fields(parents)) > 2 && mainline == 0 {
		return "", fmt.Errorf("commit %s is a merge commit; pass --mainline 1 (or 2) to choose the parent", full[:8])
	}
	return full, nil
}

// RefuseInProgressOps fails when the repo at repoRoot has a rebase,
// cherry-pick or merge in progress.
func RefuseInProgressOps(repoRoot string) error {
	gitDir, err := captureGit(repoRoot, "rev-parse", "--git-dir")
	if err != nil {
		return fmt.Errorf("locate .git: %w", err)
	}
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(repoRoot, gitDir)
	}
	for _, m := range []string{"REBASE_HEAD", "CHERRY_PICK_HEAD", "MERGE_HEAD"} {
		if _, err := os.Stat(filepath.Join(gitDir, m)); err == nil {
			return fmt.Errorf("%s has %s in progress; finish or abort it first", repoRoot, m)
		}
	}
	return nil
}

// splitBaseRef returns (branchForGitHub, refForGit). origin/main -> ("main", "origin/main").
func splitBaseRef(base string) (string, string) {
	branch := strings.TrimPrefix(base, "refs/heads/")
	// Only strip "origin/" since gavel pushes there exclusively; "feature/foo"
	// stays as is.
	branch = strings.TrimPrefix(branch, "origin/")
	return branch, base
}
