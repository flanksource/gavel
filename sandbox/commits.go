// Run commits: the commits a sandbox job's run made come back to serve as a git bundle, and serve
// puts them on a branch of its own naming, in the pushed repository or on GitHub.
package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
)

// RunBranchPrefix starts the name of every branch a run's commits go on.
const RunBranchPrefix = "gavel/"

// ErrBranchExists is a run branch that is already there: a run's branch is created once.
var ErrBranchExists = errors.New("the branch already exists")

// PushRunCommits creates branch, which must start with RunBranchPrefix and not exist yet, at the
// one branch bundle holds, after checking it continues from base, the commit the run started at.
// It returns the branch's commit.
func (s *Sources) PushRunCommits(ctx context.Context, repo Repo, base, branch string, bundle io.Reader) (string, error) {
	if !strings.HasPrefix(branch, RunBranchPrefix) || strings.HasPrefix(branch, "-") {
		return "", fmt.Errorf("a run's commits go on a %s branch, not %q", RunBranchPrefix, branch)
	}
	if _, err := runGit(ctx, nil, "check-ref-format", "refs/heads/"+branch); err != nil {
		return "", fmt.Errorf("%q is not a branch name: %w", branch, err)
	}
	source, err := s.sourceDir(ctx, repo)
	if err != nil {
		return "", err
	}
	file, err := os.CreateTemp("", "gavel-run-*.bundle")
	if err != nil {
		return "", fmt.Errorf("save the run's commits: %w", err)
	}
	defer os.Remove(file.Name())
	if _, err := io.Copy(file, bundle); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("save the run's commits: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("save the run's commits: %w", err)
	}
	heads, err := runGit(ctx, nil, "bundle", "list-heads", file.Name())
	if err != nil {
		return "", fmt.Errorf("read the run's bundle: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(heads), "\n")
	if len(lines) != 1 || lines[0] == "" {
		return "", fmt.Errorf("the run's bundle has %d branches; a run's commits are on one", len(lines))
	}
	sha, ref, _ := strings.Cut(lines[0], " ")

	incoming := "refs/gavel/incoming/" + uuid.NewString()
	if _, err := runGit(ctx, nil, "-C", source, "fetch", "--quiet", file.Name(), ref+":"+incoming); err != nil {
		return "", fmt.Errorf("read the run's commits on %s: %w", strings.TrimPrefix(ref, "refs/heads/"), err)
	}
	defer func() { _, _ = runGit(context.WithoutCancel(ctx), nil, "-C", source, "update-ref", "-d", incoming) }()
	if _, err := runGit(ctx, nil, "-C", source, "merge-base", "--is-ancestor", base, sha); err != nil {
		return "", fmt.Errorf("the run's commits on %s don't continue from %s, the commit it started at", strings.TrimPrefix(ref, "refs/heads/"), base)
	}

	switch repo.Source {
	case "push":
		if _, err := runGit(ctx, nil, "-C", source, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch); err == nil {
			return "", fmt.Errorf("create branch %s of %s: %w", branch, repo.label(), ErrBranchExists)
		}
		if _, err := runGit(ctx, nil, "-C", source, "update-ref", "refs/heads/"+branch, sha, ""); err != nil {
			return "", fmt.Errorf("create branch %s of %s: %w", branch, repo.label(), err)
		}
	case "github":
		header, err := s.authHeader(ctx, repo)
		if err != nil {
			return "", err
		}
		existing, err := runGit(ctx, nil, "-c", header, "ls-remote", s.githubURL(repo), "refs/heads/"+branch)
		if err != nil {
			return "", fmt.Errorf("look for branch %s of %s on GitHub: %w", branch, repo.label(), err)
		}
		if strings.TrimSpace(existing) != "" {
			return "", fmt.Errorf("create branch %s of %s on GitHub: %w", branch, repo.label(), ErrBranchExists)
		}
		// The mirror's origin pushes everything; this pushes the one branch, to its URL. The lease
		// refuses a branch created since the check above.
		if _, err := runGit(ctx, nil, "-c", header, "-C", source, "push", "--quiet",
			"--force-with-lease=refs/heads/"+branch+":", s.githubURL(repo), sha+":refs/heads/"+branch); err != nil {
			return "", fmt.Errorf("create branch %s of %s on GitHub: %w", branch, repo.label(), err)
		}
	}
	return sha, nil
}
