// Package land lands a todo run step's worktree commits: it cherry-picks the
// agent's work onto the checkout's current branch (merge) or opens a pull
// request carrying it (pr), then removes the run's worktree and branch once
// every commit is confirmed landed, and records the landing on the todo.
//
// The agent's work is Setup..Head of the run's recorded worktree: Base..Setup
// is Captain's setup snapshot of work-in-progress copied from the source
// checkout, which is never landed.
package land

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/aiflags"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/github/prcreate"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

const actor = "gavel"

var (
	ErrNotNative        = errors.New("landing a run requires native TODO storage")
	ErrInvalidOptions   = errors.New("invalid land options")
	ErrNoRunWorkspace   = errors.New("nothing to land")
	ErrNoCommits        = errors.New("run made no commits to land")
	ErrWorktreeDirty    = errors.New("run worktree was kept with uncommitted changes")
	ErrCheckoutNotReady = errors.New("checkout is not ready to land onto")
)

// Provider is the native TODO storage a landing reads runs from and records to.
type Provider interface {
	Captain() *captaindb.DB
	Repository() *native.Repository
}

type Options struct {
	Via native.LandingVia
	// Base is the ref a PR branches from; empty resolves pr.base from
	// .gavel.yaml, then prcreate.DefaultBase. A merge lands onto the checked-out
	// branch, so it takes no base.
	Base  string
	Draft bool
	// Flags and Deps drive the PR's AI content and GitHub calls.
	Flags aiflags.ModelFlags
	Deps  prcreate.Deps
}

// Validate refuses a via other than merge or pr, and PR-only options on a merge.
func (o Options) Validate() error {
	switch o.Via {
	case native.LandingPR:
		return nil
	case native.LandingMerge:
		if strings.TrimSpace(o.Base) != "" || o.Draft {
			return fmt.Errorf("%w: base and draft only apply to a pr landing", ErrInvalidOptions)
		}
		return nil
	default:
		return fmt.Errorf("%w: via %q must be %q or %q", ErrInvalidOptions, o.Via, native.LandingMerge, native.LandingPR)
	}
}

// ConflictError reports a merge whose cherry-pick conflicted. The cherry-pick
// was aborted, so the checkout is back where it started.
type ConflictError struct {
	Branch string
	Paths  []string
	Err    error
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("cherry-picking the run's commits onto %s conflicted in %s; the cherry-pick was aborted",
		e.Branch, strings.Join(e.Paths, ", "))
}

func (e *ConflictError) Unwrap() error { return e.Err }

// CleanupError reports a landing that was recorded but whose worktree or branch
// could not be removed. Land returns the recorded landing alongside it.
type CleanupError struct{ Err error }

func (e *CleanupError) Error() string {
	return "the run landed, but its worktree cleanup failed: " + e.Err.Error()
}

func (e *CleanupError) Unwrap() error { return e.Err }

// Land lands the todo's newest run step that recorded a worktree. A non-nil
// landing returned with a *CleanupError was recorded; only its cleanup failed.
func Land(ctx context.Context, provider Provider, todo *types.TODO, opts Options) (*native.RunLanding, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	if provider == nil || provider.Captain() == nil || provider.Repository() == nil {
		return nil, ErrNotNative
	}
	if todo == nil {
		return nil, errors.New("land: todo is nil")
	}
	issueID, err := uuid.Parse(strings.TrimSpace(todo.ID))
	if err != nil {
		return nil, fmt.Errorf("native TODO has invalid ID %q: %w", todo.ID, err)
	}
	run, err := provider.Repository().LatestRunWorktree(ctx, provider.Captain(), issueID)
	if errors.Is(err, native.ErrNoRunWorkspace) {
		return nil, fmt.Errorf("%w: %w", ErrNoRunWorkspace, err)
	}
	if err != nil {
		return nil, err
	}
	if landed := run.Landing; landed != nil {
		return nil, fmt.Errorf("%w: run %s landed via %s on %s at %s", native.ErrAlreadyLanded,
			run.PromptRunID, landed.Via, landed.TargetBranch, landed.LandedSHA)
	}
	wt := &run.Worktree
	commits, err := landableCommits(run.PromptRunID, wt)
	if err != nil {
		return nil, err
	}
	landing, err := landCommits(ctx, wt, commits, opts)
	if err != nil {
		return nil, err
	}
	landing.PromptRunID = run.PromptRunID
	landing.CommitCount = len(commits)
	cleanupErr := cleanup(wt, &landing)
	recorded, err := provider.Repository().RecordLanding(ctx, issueID, landing, actor)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("run %s landed on %s at %s but recording it failed: %w",
			run.PromptRunID, landing.TargetBranch, landing.LandedSHA, err), cleanupErr)
	}
	if cleanupErr != nil {
		return recorded, &CleanupError{Err: cleanupErr}
	}
	return recorded, nil
}

func landCommits(ctx context.Context, wt *api.WorktreeState, commits []string, opts Options) (native.RunLanding, error) {
	if opts.Via == native.LandingMerge {
		return merge(wt.Repo, wt.Branch, commits)
	}
	base, err := prcreate.ResolveBase(wt.Repo, opts.Base)
	if err != nil {
		return native.RunLanding{}, err
	}
	res, err := prcreate.Create(ctx, wt.Repo, prcreate.Input{
		SHAs: commits, Base: base, Draft: opts.Draft, Flags: opts.Flags, Deps: opts.Deps,
	})
	if err != nil {
		return native.RunLanding{}, err
	}
	number := res.PR.Number
	return native.RunLanding{
		Via: native.LandingPR, TargetBranch: res.PR.Base, LandedSHA: res.TopicHead,
		PRNumber: &number, PRURL: res.PR.URL,
	}, nil
}

// landableCommits is the agent's work, Setup..Head oldest first.
func landableCommits(runID uuid.UUID, wt *api.WorktreeState) ([]string, error) {
	if wt.Repo == "" || wt.Branch == "" || wt.Setup == "" || wt.Head == "" {
		return nil, fmt.Errorf("run %s recorded an incomplete worktree (repo %q, branch %q, setup %q, head %q)",
			runID, wt.Repo, wt.Branch, wt.Setup, wt.Head)
	}
	if wt.Kept && len(wt.Dirty) > 0 {
		return nil, fmt.Errorf("%w: %s on %s holds %s (kept: %s); land those edits by hand",
			ErrWorktreeDirty, wt.Path, wt.Branch, strings.Join(wt.Dirty, ", "), wt.KeptReason)
	}
	out, err := captureGit(wt.Repo, "rev-list", "--reverse", wt.Setup+".."+wt.Head)
	if err != nil {
		return nil, fmt.Errorf("list the commits of run %s: %w", runID, err)
	}
	commits := strings.Fields(out)
	if len(commits) == 0 {
		return nil, fmt.Errorf("%w: run %s has nothing past its setup (%s..%s)", ErrNoCommits, runID, short(wt.Setup), short(wt.Head))
	}
	return commits, nil
}

func short(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
