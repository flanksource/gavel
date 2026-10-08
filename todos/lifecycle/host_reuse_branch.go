package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
)

// reusableWorktree is the worktree the todo's newest run step recorded, when its
// branch still exists for another run to continue on.
func (h *Host) reusableWorktree(ctx context.Context, todo *types.TODO) (*api.WorktreeState, error) {
	provider, ok := h.Provider.(RunWorktreeProvider)
	if !ok {
		return nil, fmt.Errorf("reuse branch: provider %T does not expose run worktrees", h.Provider)
	}
	ref := todos.TODOReference(todo)
	run, err := provider.LatestRunWorktree(ctx, todo)
	if errors.Is(err, native.ErrNoRunWorkspace) {
		return nil, fmt.Errorf("todo %s has no reusable branch: %w", ref, err)
	}
	if err != nil {
		return nil, fmt.Errorf("reuse branch of %s: %w", ref, err)
	}
	wt := run.Worktree
	if wt.BranchDeleted || strings.TrimSpace(wt.Branch) == "" {
		return nil, fmt.Errorf("todo %s has no reusable branch: run %s deleted its branch %q at teardown",
			ref, run.PromptRunID, wt.Branch)
	}
	return &wt, nil
}

// reusedWorktree is the step's worktree pointed at a previous run's branch:
// the run's kept worktree when its directory is still there, else the branch
// checked out as it is. Only mode, branch and path change; the step's keep and
// ignored settings carry over, while settings that describe branching a new
// worktree (prefix, base, uncommitted) do not apply to an existing branch.
func reusedWorktree(previous api.WorktreeState, step *shell.Worktree) *shell.Worktree {
	reused := &shell.Worktree{}
	if step != nil {
		reused.Keep, reused.Ignored = step.Keep, step.Ignored
	}
	if info, err := os.Stat(previous.Path); previous.Kept && err == nil && info.IsDir() {
		reused.Mode, reused.Path = shell.WorktreeExisting, previous.Path
		return reused
	}
	reused.Mode, reused.Branch = shell.WorktreeBranch, previous.Branch
	return reused
}

// reusedBranchNotice tells the agent which commits the branch it continues on
// already holds. It is phrased without position because codex folds an appended
// system prompt in after the user prompt rather than ahead of it.
func reusedBranchNotice(previous api.WorktreeState) string {
	return fmt.Sprintf("This run continues on branch `%s`, which already holds the previous attempt's commits (%s..%s). "+
		"Build on them, and address every unresolved review comment listed in the task.",
		previous.Branch, shortSHA(previous.Setup), shortSHA(previous.Head))
}

// applyReusedBranch makes the step continue on the todo's previous run branch:
// the resolved checkout's worktree is replaced as the top layer, and the system
// prompt is told which commits the branch already holds. The notice is run
// framing, not task text, so it rides the appended system prompt — which a
// project's `todos.run.prompt.user` override never replaces.
func (h *Host) applyReusedBranch(ctx context.Context, todo *types.TODO, step Step, prepared *preparedStep) error {
	if prepared.class != types.ModeRun {
		return fmt.Errorf("step %s: reusing a run's branch applies to a run step, not a %s step", step.Name, prepared.class)
	}
	previous, err := h.reusableWorktree(ctx, todo)
	if err != nil {
		return err
	}
	setup := shell.Setup{}
	if prepared.request.Setup != nil {
		setup = *prepared.request.Setup
	}
	checkout := shell.Checkout{Mode: shell.CheckoutLocal}
	if setup.Checkout != nil {
		checkout = *setup.Checkout
	}
	if checkout.Mode == shell.CheckoutRemote || strings.TrimSpace(checkout.URL) != "" {
		return fmt.Errorf("step %s: cannot reuse branch %s in a remote checkout", step.Name, previous.Branch)
	}
	checkout.Worktree = reusedWorktree(*previous, checkout.Worktree)
	setup.Checkout = &checkout
	prepared.request.Setup = &setup
	notice := reusedBranchNotice(*previous)
	if existing := strings.TrimSpace(prepared.request.Prompt.AppendSystem); existing != "" {
		notice = existing + "\n\n" + notice
	}
	prepared.request.Prompt.AppendSystem = notice
	prepared.provenance = recordRuntimeFields(prepared.provenance, runtimeFields{
		Name: "lifecycle reuse branch", Paths: []string{"/setup/checkout/worktree", "/prompt/appendSystem"}, Replace: true,
	})
	return nil
}
