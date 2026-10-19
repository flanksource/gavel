package provider

import (
	"context"
	"fmt"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/pr/model"
	"os/exec"
	"regexp"
	"strings"
)

type RunTarget struct {
	RepositoryURL string `json:"repositoryUrl"`
	Branch        string `json:"branch"`
	HeadSHA       string `json:"headSha"`
	// Rebase lets PushHead rewrite the source branch: the run rebases it onto
	// its base, so the new head no longer descends from HeadSHA. The push stays
	// leased on HeadSHA either way.
	Rebase bool `json:"rebase,omitempty"`
}

var commitSHA = regexp.MustCompile(`^[a-fA-F0-9]{40}([a-fA-F0-9]{24})?$`)

func TargetOf(pr *model.PRInfo) (*RunTarget, error) {
	if pr == nil || pr.State != "OPEN" {
		return nil, fmt.Errorf("PR execution requires an open pull request")
	}
	if pr.HeadRepositoryURL == "" || pr.HeadRefName == "" || !commitSHA.MatchString(pr.HeadRefOID) {
		return nil, fmt.Errorf("PR response has no complete source repository, branch and head SHA")
	}
	return &RunTarget{RepositoryURL: pr.HeadRepositoryURL, Branch: pr.HeadRefName, HeadSHA: pr.HeadRefOID}, nil
}

func Revision(ctx context.Context, dir, ref string) (string, error) {
	if commitSHA.MatchString(ref) {
		return strings.ToLower(ref), nil
	}
	return targetGit(ctx, dir, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
}

func targetGit(ctx context.Context, dir string, args ...string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func (t *RunTarget) FetchHead(ctx context.Context, dir string) error {
	if t == nil || t.RepositoryURL == "" || !commitSHA.MatchString(t.HeadSHA) {
		return fmt.Errorf("PR fetch requires a complete source target")
	}
	if _, err := targetGit(ctx, dir, "check-ref-format", "refs/heads/"+t.Branch); err != nil {
		return fmt.Errorf("invalid PR source branch: %w", err)
	}
	if _, err := targetGit(ctx, dir, "fetch", "--no-tags", "--no-recurse-submodules", "--", t.RepositoryURL, "refs/heads/"+t.Branch); err != nil {
		return err
	}
	head, err := Revision(ctx, dir, "FETCH_HEAD")
	if err != nil {
		return err
	}
	if head != t.HeadSHA {
		return fmt.Errorf("PR source branch advanced from %s to %s; refresh the PR before continuing", t.HeadSHA, head)
	}
	return nil
}

func (t *RunTarget) PushHead(ctx context.Context, dir string) (string, error) {
	if err := t.FetchHead(ctx, dir); err != nil {
		return "", err
	}
	head, err := Revision(ctx, dir, "HEAD")
	if err != nil {
		return "", err
	}
	if _, err := targetGit(ctx, dir, "merge-base", "--is-ancestor", t.HeadSHA, head); err != nil && !t.Rebase {
		return "", fmt.Errorf("PR push must fast-forward the source branch: %w", err)
	}
	if head != t.HeadSHA {
		// The lease closes the gap between FetchHead's check and the push: a
		// push that lands in between is refused instead of overwritten.
		lease := fmt.Sprintf("--force-with-lease=refs/heads/%s:%s", t.Branch, t.HeadSHA)
		if _, err := targetGit(ctx, dir, "push", "--porcelain", lease, "--", t.RepositoryURL, "HEAD:refs/heads/"+t.Branch); err != nil {
			return "", err
		}
	}
	t.HeadSHA = head
	return head, nil
}

func (t *RunTarget) Checkout() *shell.Checkout {
	return &shell.Checkout{Mode: shell.CheckoutLocal, Worktree: &shell.Worktree{Mode: shell.WorktreeNew, Prefix: "gavel-pr", Base: t.HeadSHA, Uncommitted: shell.CloneSkip, Ignored: shell.CloneClone}}
}
