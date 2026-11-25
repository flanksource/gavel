// Package checkoutswitch moves a repository's primary checkout onto a local
// branch and, optionally, brings that branch level with its upstream, refusing
// whenever the move would lose uncommitted edits or commits that no other ref
// holds.
package checkoutswitch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
)

// ErrRefused marks a switch the checkout's state does not allow; the message
// says what to change first.
var ErrRefused = errors.New("switch refused")

// ErrInvalidOptions marks a malformed request.
var ErrInvalidOptions = errors.New("invalid switch options")

// Commit is one commit by sha and subject line.
type Commit struct {
	SHA     string `json:"sha"`
	Subject string `json:"subject"`
}

// Plan is what switching the primary checkout to Branch would do, read from
// the last-fetched remote-tracking refs.
type Plan struct {
	Repo string `json:"repo"`
	// Current is the branch the checkout is on now; "" when detached.
	Current string `json:"current"`
	Head    string `json:"head"`
	Branch  string `json:"branch"`
	// BranchHead is the tip of the local branch being switched to.
	BranchHead string `json:"branchHead"`
	// Upstream is the branch's configured upstream (e.g. origin/main); "" when
	// it tracks none. UpstreamHead is "" when the upstream ref is gone.
	Upstream     string `json:"upstream,omitempty"`
	UpstreamHead string `json:"upstreamHead,omitempty"`
	upstreamRef  string
	// Ahead counts the branch's commits not in the upstream, Behind the
	// upstream's commits not in the branch.
	Ahead  int `json:"ahead"`
	Behind int `json:"behind"`
	// Unpreserved lists the branch's commits that no other branch, remote ref
	// or tag reaches and that the upstream does not carry as an equivalent
	// patch: a reset to the upstream would orphan exactly these.
	Unpreserved []Commit `json:"unpreserved"`
	// DirtyCount counts tracked files with uncommitted changes in the checkout.
	DirtyCount int `json:"dirtyCount"`
	// CheckedOutIn is the linked worktree that has Branch checked out, if any.
	CheckedOutIn string `json:"checkedOutIn,omitempty"`
}

// Inspect plans switching the primary checkout of the repository containing
// repo to branch. It does not fetch.
func Inspect(ctx context.Context, repo, branch string) (Plan, error) {
	worktrees, err := gavelgit.ListWorktrees(repo)
	if err != nil {
		return Plan{}, err
	}
	plan := Plan{Repo: worktrees[0].Path, Branch: branch, Unpreserved: []Commit{}}
	if err := validBranch(ctx, plan.Repo, branch); err != nil {
		return Plan{}, err
	}
	if plan.BranchHead, err = revParse(ctx, plan.Repo, "refs/heads/"+branch); err != nil {
		return Plan{}, fmt.Errorf("%w: no local branch %q in %s", ErrInvalidOptions, branch, plan.Repo)
	}
	for _, wt := range worktrees[1:] {
		if wt.Branch == branch {
			plan.CheckedOutIn = wt.Path
		}
	}
	if plan.Current, err = run(ctx, plan.Repo, "branch", "--show-current"); err != nil {
		return Plan{}, err
	}
	if plan.Head, err = revParse(ctx, plan.Repo, "HEAD"); err != nil {
		return Plan{}, err
	}
	dirty, err := run(ctx, plan.Repo, "status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return Plan{}, err
	}
	if dirty != "" {
		plan.DirtyCount = len(strings.Split(dirty, "\n"))
	}
	if err := plan.readUpstream(ctx); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func (plan *Plan) readUpstream(ctx context.Context) error {
	upstream, err := run(ctx, plan.Repo, "for-each-ref", "--format=%(upstream)%00%(upstream:short)", "refs/heads/"+plan.Branch)
	if err != nil || upstream == "" {
		if err == nil {
			return plan.readUnpreserved(ctx)
		}
		return err
	}
	plan.upstreamRef, plan.Upstream, _ = strings.Cut(upstream, "\x00")
	if head, err := revParse(ctx, plan.Repo, plan.upstreamRef); err == nil {
		plan.UpstreamHead = head
	}
	if plan.UpstreamHead == "" {
		return plan.readUnpreserved(ctx)
	}
	counts, err := run(ctx, plan.Repo, "rev-list", "--left-right", "--count", plan.BranchHead+"..."+plan.UpstreamHead)
	if err != nil {
		return err
	}
	fields := strings.Fields(counts)
	if len(fields) != 2 {
		return fmt.Errorf("git rev-list --left-right --count in %s: unexpected output %q", plan.Repo, counts)
	}
	if plan.Ahead, err = strconv.Atoi(fields[0]); err != nil {
		return err
	}
	if plan.Behind, err = strconv.Atoi(fields[1]); err != nil {
		return err
	}
	return plan.readUnpreserved(ctx)
}

// readUnpreserved lists the branch's commits only it reaches, minus those
// whose patch the upstream already carries (a squash or cherry-pick landing).
func (plan *Plan) readUnpreserved(ctx context.Context) error {
	plan.Unpreserved = []Commit{}
	only, err := run(ctx, plan.Repo, "log", "--format=%H %s", "refs/heads/"+plan.Branch,
		"--not", "--exclude="+plan.Branch, "--branches", "--remotes", "--tags")
	if err != nil || only == "" {
		return err
	}
	landed := map[string]bool{}
	if plan.UpstreamHead != "" {
		cherry, err := run(ctx, plan.Repo, "cherry", plan.UpstreamHead, plan.BranchHead)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(cherry, "\n") {
			if sha, ok := strings.CutPrefix(line, "- "); ok {
				landed[sha] = true
			}
		}
	}
	for _, line := range strings.Split(only, "\n") {
		sha, subject, _ := strings.Cut(line, " ")
		if !landed[sha] {
			plan.Unpreserved = append(plan.Unpreserved, Commit{SHA: sha, Subject: subject})
		}
	}
	return nil
}

func validBranch(ctx context.Context, repo, branch string) error {
	if branch == "" {
		return fmt.Errorf("%w: branch is required", ErrInvalidOptions)
	}
	if _, err := run(ctx, repo, "check-ref-format", "--branch", branch); err != nil {
		return fmt.Errorf("%w: %q is not a valid branch name", ErrInvalidOptions, branch)
	}
	return nil
}

func revParse(ctx context.Context, repo, rev string) (string, error) {
	return run(ctx, repo, "rev-parse", "--verify", "--quiet", "--end-of-options", rev+"^{commit}")
}

// idle refuses while a merge, rebase, cherry-pick or revert is in progress.
func idle(ctx context.Context, repo string) error {
	for _, name := range []string{"rebase-merge", "rebase-apply", "MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD"} {
		path, err := run(ctx, repo, "rev-parse", "--path-format=absolute", "--git-path", name)
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%w: %s has %s in progress; finish or abort it first", ErrRefused, repo, name)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func run(ctx context.Context, repo string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", repo}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s in %s: %w: %s", strings.Join(args, " "), repo, err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}
