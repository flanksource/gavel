// Run commits: the commits a sandbox job's run made come back to serve as a git bundle, and serve
// puts them on a branch of its own naming, in the pushed repository or on GitHub.
package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// RunBranchPrefix starts the name of every branch a run's commits go on.
const RunBranchPrefix = "gavel/"

// ErrBranchExists is a run branch that is already there: a run's branch is created once.
var ErrBranchExists = errors.New("the branch already exists")

// ErrNoRunBranch is a run whose commits are on no run branch.
var ErrNoRunBranch = errors.New("no run branch holds the run's commits")

// RunBranch is the branch serve puts the commits of run, a sandbox run of todo, on.
func RunBranch(todo, run uuid.UUID) string {
	return RunBranchPrefix + todo.String()[:8] + "-" + run.String()[:8]
}

// RunBranchAt is the branch, of those todo's runs put their commits on in repo, whose tip is
// head: a run's recorded head names its branch.
func (s *Sources) RunBranchAt(ctx context.Context, repo Repo, todo uuid.UUID, head string) (string, error) {
	dir, err := s.sourceDir(ctx, repo)
	if err != nil {
		return "", err
	}
	refs, err := runGit(ctx, nil, "-C", dir, "for-each-ref", "--format=%(objectname) %(refname:strip=2)",
		"refs/heads/"+RunBranchPrefix+todo.String()[:8]+"-*")
	if err != nil {
		return "", fmt.Errorf("list the run branches of todo %s in %s: %w", todo, repo.label(), err)
	}
	for _, line := range strings.Split(strings.TrimSpace(refs), "\n") {
		if sha, branch, ok := strings.Cut(line, " "); ok && sha == head {
			return branch, nil
		}
	}
	return "", fmt.Errorf("%w: none of todo %s's branches in %s is at %s", ErrNoRunBranch, todo, repo.label(), head)
}

// CommitSubjects lists the subjects of the commits base..head in repo, oldest first.
func (s *Sources) CommitSubjects(ctx context.Context, repo Repo, base, head string) ([]string, error) {
	for _, commit := range []string{base, head} {
		if strings.HasPrefix(commit, "-") {
			return nil, fmt.Errorf("commit %q can't start with \"-\"", commit)
		}
	}
	dir, err := s.sourceDir(ctx, repo)
	if err != nil {
		return nil, err
	}
	log, err := runGit(ctx, nil, "-C", dir, "log", "--reverse", "--format=%s", base+".."+head, "--")
	if err != nil {
		return nil, fmt.Errorf("list the commits %s..%s of %s: %w", base, head, repo.label(), err)
	}
	if strings.TrimSpace(log) == "" {
		return nil, nil
	}
	return strings.Split(strings.TrimSpace(log), "\n"), nil
}

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
	sha, done, err := receiveCommits(ctx, source, base, bundle)
	if err != nil {
		return "", err
	}
	defer done()

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

// commitSHA is a full SHA-1 or SHA-256 commit id, which is all a PR's base can be.
var commitSHA = regexp.MustCompile(`^([0-9a-f]{40}|[0-9a-f]{64})$`)

// PushPRHead fast-forwards the PR's source branch, branch of the repository at headURL, from base,
// the head the run started its turn at, to the one branch bundle holds. The push fails if the
// branch has moved since. It returns the branch's new commit.
func (s *Sources) PushPRHead(ctx context.Context, repo Repo, headURL, branch, base string, bundle io.Reader) (string, error) {
	if _, err := runGit(ctx, nil, "check-ref-format", "refs/heads/"+branch); err != nil || strings.HasPrefix(branch, "-") {
		return "", fmt.Errorf("%q is not a branch name: %v", branch, err)
	}
	if !commitSHA.MatchString(base) {
		return "", fmt.Errorf("the PR's base %q is not a commit SHA", base)
	}
	if repo.Source != "github" {
		return "", fmt.Errorf("repository %s has source %q; only GitHub repositories have PRs", repo.label(), repo.Source)
	}
	source, err := s.sourceDir(ctx, repo)
	if err != nil {
		return "", err
	}
	sha, done, err := receiveCommits(ctx, source, base, bundle)
	if err != nil {
		return "", err
	}
	defer done()
	header, err := s.authHeader(ctx, repo)
	if err != nil {
		return "", err
	}
	if _, err := runGit(ctx, nil, "-c", header, "-C", source, "push", "--quiet",
		"--force-with-lease=refs/heads/"+branch+":"+base, "--", headURL, sha+":refs/heads/"+branch); err != nil {
		return "", fmt.Errorf("push the PR's source branch %s: %w", branch, err)
	}
	return sha, nil
}

// receiveCommits reads the one branch bundle holds into source under a temporary ref, after
// checking it continues from base, and returns its commit and a func that drops the ref.
func receiveCommits(ctx context.Context, source, base string, bundle io.Reader) (string, func(), error) {
	file, err := os.CreateTemp("", "gavel-run-*.bundle")
	if err != nil {
		return "", nil, fmt.Errorf("save the run's commits: %w", err)
	}
	defer os.Remove(file.Name())
	if _, err := io.Copy(file, bundle); err != nil {
		_ = file.Close()
		return "", nil, fmt.Errorf("save the run's commits: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", nil, fmt.Errorf("save the run's commits: %w", err)
	}
	heads, err := runGit(ctx, nil, "bundle", "list-heads", file.Name())
	if err != nil {
		return "", nil, fmt.Errorf("read the run's bundle: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(heads), "\n")
	if len(lines) != 1 || lines[0] == "" {
		return "", nil, fmt.Errorf("the run's bundle has %d branches; a run's commits are on one", len(lines))
	}
	sha, ref, _ := strings.Cut(lines[0], " ")
	name := strings.TrimPrefix(ref, "refs/heads/")

	incoming := "refs/gavel/incoming/" + uuid.NewString()
	if _, err := runGit(ctx, nil, "-C", source, "fetch", "--quiet", file.Name(), ref+":"+incoming); err != nil {
		return "", nil, fmt.Errorf("read the run's commits on %s: %w", name, err)
	}
	done := func() { _, _ = runGit(context.WithoutCancel(ctx), nil, "-C", source, "update-ref", "-d", incoming) }
	if _, err := runGit(ctx, nil, "-C", source, "merge-base", "--is-ancestor", base, sha); err != nil {
		done()
		return "", nil, fmt.Errorf("the run's commits on %s don't continue from %s, the commit it started at", name, base)
	}
	return sha, done, nil
}
