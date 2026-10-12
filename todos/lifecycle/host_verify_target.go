package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
)

// verifyWorktreePrefix names the branch of the throwaway worktree a verify step
// checks a run's commit out into. Captain's setup teardown removes the clean
// worktree and deletes the branch, which carries no commits past its base.
const verifyWorktreePrefix = "gavel-verify"

// RunWorktreeProvider is the provider capability a verify step reads the commit
// it checks from: the worktree the todo's newest run step recorded, and how its
// commits were landed. A todo no run recorded a worktree for is
// native.ErrNoRunWorkspace.
type RunWorktreeProvider interface {
	LatestRunWorktree(ctx context.Context, todo *types.TODO) (*native.RunWorktree, error)
}

// verifyTarget is the commit a verify step checks. The run's own work lives on
// its worktree branch until it is landed, so the main checkout — which does not
// have it — is the target only when no run recorded a worktree, or the run was
// merged into it. A PR landing is verified at the pull request's topic head.
func (h *Host) verifyTarget(ctx context.Context, todo *types.TODO) (*types.VerifyTarget, error) {
	provider, ok := h.Provider.(RunWorktreeProvider)
	if !ok {
		return nil, fmt.Errorf("verify target: provider %T does not expose run worktrees", h.Provider)
	}
	run, err := provider.LatestRunWorktree(ctx, todo)
	if errors.Is(err, native.ErrNoRunWorkspace) {
		return &types.VerifyTarget{Source: types.VerifyTargetCheckout}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("verify target of %s: %w", todos.TODOReference(todo), err)
	}
	if landing := run.Landing; landing != nil {
		return landedTarget(run, landing)
	}
	wt := run.Worktree
	if strings.TrimSpace(wt.Branch) == "" || strings.TrimSpace(wt.Head) == "" {
		return nil, fmt.Errorf("verify target: run %s recorded an incomplete worktree (branch %q, head %q)",
			run.PromptRunID, wt.Branch, wt.Head)
	}
	if wt.Kept && len(wt.Dirty) > 0 {
		return nil, fmt.Errorf("verify target: run %s left uncommitted edits in %s on %s (%s; kept: %s); "+
			"its head %s does not hold them, so commit or land them before verifying",
			run.PromptRunID, wt.Path, wt.Branch, strings.Join(wt.Dirty, ", "), wt.KeptReason, shortSHA(wt.Head))
	}
	return &types.VerifyTarget{Branch: wt.Branch, SHA: wt.Head, Source: types.VerifyTargetRun}, nil
}

func landedTarget(run *native.RunWorktree, landing *native.RunLanding) (*types.VerifyTarget, error) {
	switch landing.Via {
	case native.LandingMerge:
		return &types.VerifyTarget{Source: types.VerifyTargetCheckout}, nil
	case native.LandingPR:
		if strings.TrimSpace(landing.LandedSHA) == "" {
			return nil, fmt.Errorf("verify target: run %s landed as %s without a topic head", run.PromptRunID, landing.PRURL)
		}
		return &types.VerifyTarget{SHA: landing.LandedSHA, Source: types.VerifyTargetPR, PR: landing.PRURL}, nil
	}
	return nil, fmt.Errorf("verify target: run %s landed via %q, neither %q nor %q",
		run.PromptRunID, landing.Via, native.LandingMerge, native.LandingPR)
}

// verifyCheckout is the setup that checks a worktree target out: a fresh
// worktree of the step's checkout at the target sha. The source checkout's
// uncommitted changes stay out — the target is a commit — while gitignored
// content (installed dependencies, build caches) is copied so the definition of
// done can run.
func verifyCheckout(target *types.VerifyTarget) *shell.Checkout {
	return &shell.Checkout{Mode: shell.CheckoutLocal, Worktree: &shell.Worktree{
		Mode: shell.WorktreeNew, Prefix: verifyWorktreePrefix, Base: target.SHA,
		Uncommitted: shell.CloneSkip, Ignored: shell.CloneClone,
	}}
}

// applyVerifyTarget resolves the commit a verify step checks and, for one that
// is not the main checkout, points the step's setup at a worktree of it.
func (h *Host) applyVerifyTarget(ctx context.Context, todo *types.TODO, prepared *preparedStep) error {
	target, err := h.verifyTarget(ctx, todo)
	if err != nil {
		return err
	}
	prepared.verifyTarget = target
	if !target.Worktree() {
		return nil
	}
	setup := *prepared.request.Setup
	setup.Checkout = verifyCheckout(target)
	prepared.request.Setup = &setup
	prepared.provenance = recordRuntimeFields(prepared.provenance, runtimeFields{
		Name: "lifecycle verify target", Paths: []string{"/setup/checkout"},
	})
	return nil
}

// verifiedTarget completes the target with the commit the check actually ran
// against: the worktree Captain's setup recorded for a run or PR target, whose
// base must be the target sha, or the main checkout's git state.
func verifiedTarget(target *types.VerifyTarget, workspace *api.WorkspaceRecord, workDir string) (types.VerifyTarget, types.BuildTestResultInfoOptions, error) {
	if target == nil {
		return types.VerifyTarget{}, types.BuildTestResultInfoOptions{}, fmt.Errorf("verify step resolved no target")
	}
	if !target.Worktree() {
		branch, commit, dirty, _ := todos.GetGitInfo(workDir)
		checkout := types.VerifyTarget{Branch: branch, SHA: commit, Source: types.VerifyTargetCheckout}
		return checkout, types.BuildTestResultInfoOptions{CWD: workDir, GitBranch: branch, GitCommit: commit, GitDirty: dirty}, nil
	}
	if workspace == nil || workspace.Worktree == nil {
		return types.VerifyTarget{}, types.BuildTestResultInfoOptions{},
			fmt.Errorf("verify step for %s %s recorded no worktree", target.Source, shortSHA(target.SHA))
	}
	wt := workspace.Worktree
	if wt.Base != target.SHA {
		return types.VerifyTarget{}, types.BuildTestResultInfoOptions{},
			fmt.Errorf("verify step checked out %s, not its %s target %s", shortSHA(wt.Base), target.Source, shortSHA(target.SHA))
	}
	return *target, types.BuildTestResultInfoOptions{
		CWD: wt.Path, GitBranch: target.Branch, GitCommit: shortSHA(wt.Base), GitDirty: len(wt.Dirty) > 0,
	}, nil
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
