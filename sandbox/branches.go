// Branch lookups on serve's copy of a repository: the branches a member can pick to run a job on,
// and the commit a picked branch is at.
package sandbox

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrUnknownBranch is a branch name that doesn't name one of the repository's branches.
var ErrUnknownBranch = errors.New("unknown branch")

// Branches are a repository's branch names, sorted, and its default branch.
type Branches struct {
	Default string   `json:"default"`
	Names   []string `json:"names"`
}

// Branches lists repo's branches, fetching its GitHub mirror first.
func (s *Sources) Branches(ctx context.Context, repo Repo) (Branches, error) {
	dir, err := s.sourceDir(ctx, repo)
	if err != nil {
		return Branches{}, err
	}
	out, err := runGit(ctx, nil, "-C", dir, "for-each-ref", "--sort=refname", "--format=%(refname:strip=2)", "refs/heads")
	if err != nil {
		return Branches{}, fmt.Errorf("list the branches of %s: %w", repo.label(), err)
	}
	branches := Branches{Names: strings.Fields(out)}
	if head, err := runGit(ctx, nil, "-C", dir, "symbolic-ref", "--quiet", "--short", "HEAD"); err == nil {
		branches.Default = strings.TrimSpace(head)
	}
	return branches, nil
}

// BranchCommit is the commit repo's branch is at.
func (s *Sources) BranchCommit(ctx context.Context, repo Repo, branch string) (string, error) {
	if _, err := runGit(ctx, nil, "check-ref-format", "--branch", branch); err != nil || strings.HasPrefix(branch, "-") {
		return "", fmt.Errorf("%w %q: not a valid branch name", ErrUnknownBranch, branch)
	}
	dir, err := s.sourceDir(ctx, repo)
	if err != nil {
		return "", err
	}
	sha, err := runGit(ctx, nil, "-C", dir, "rev-parse", "--verify", "--quiet", "--end-of-options", "refs/heads/"+branch+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("%w %q: %s has no such branch", ErrUnknownBranch, branch, repo.label())
	}
	return strings.TrimSpace(sha), nil
}
