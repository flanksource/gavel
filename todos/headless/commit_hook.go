package headless

import (
	"errors"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	capcommit "github.com/flanksource/captain/pkg/ai/agent/commit"
	"github.com/flanksource/captain/pkg/api"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos/types"
)

// commitHooks builds captain commit hooks from the run's spec, one per declared
// policy. Captain owns the lifecycle — which phase fires, tracking the anchor,
// collapsing the chain at the end of the run — while the Do callback routes the
// commit itself through gavel's pipeline, so the generated message, pre-commit
// gates and Gavel-Issue-Id / session trailers are identical to `gavel commit`.
//
// A spec with no commit policies returns no hooks, and the run falls back to
// maybeCommitAfterRun's single end-of-run commit.
func commitHooks(req captainai.Request, todosInGroup []*types.TODO, sessionID string) []any {
	if req.Workflow == nil || len(req.Workflow.Commits) == 0 {
		return nil
	}
	meta := commitpkg.AgentRunMetadata{IssueID: issueIDOf(todosInGroup), SessionID: sessionID}
	hooks := make([]any, 0, len(req.Workflow.Commits))
	for _, spec := range req.Workflow.Commits {
		hook := capcommit.New(spec)
		hook.Do = gavelCommitter(meta, spec.Message != "")
		hooks = append(hooks, hook)
	}
	return hooks
}

// gavelCommitter adapts one captain commit Plan onto gavel's commit pipeline.
// keepSubject forwards captain's subject only when the spec named one: left
// empty the pipeline generates the message, which is the reason to route
// through it rather than committing with plain git.
func gavelCommitter(meta commitpkg.AgentRunMetadata, keepSubject bool) func(*agent.HookContext, capcommit.Plan) (string, error) {
	return func(hc *agent.HookContext, plan capcommit.Plan) (string, error) {
		run := commitpkg.AgentRun{
			WorkDir: plan.Dir,
			Meta:    meta,
			Fixup:   plan.Anchor,
			DryRun:  plan.DryRun,
			// Captain already ran the cheap gates over the exact path set it
			// selected; only `gates: full` asks for gavel's fuller pipeline
			// (hooks, lint, linked-deps, tidy) on top.
			SkipGates: plan.Gates != api.CommitGatesFull,
		}
		if keepSubject {
			run.Message = plan.Subject
		}
		result, err := commitpkg.RunAfterAgent(hc.Context, run)
		// A turn that edited nothing stageable is a normal outcome, not a
		// failure: the hook records no commit and the run continues.
		if errors.Is(err, commitpkg.ErrNothingStaged) || errors.Is(err, commitpkg.ErrSessionNoFiles) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		if result == nil || len(result.Commits) == 0 {
			return "", nil
		}
		// The last hash is the one a following turn fixes up onto — the pipeline
		// may split a change set into several grouped commits.
		return result.Commits[len(result.Commits)-1].Hash, nil
	}
}

// lastCommitSHA returns the final commit a run's hooks recorded, which is the
// squashed one when the chain collapsed. Empty when the run declared no commit
// policy, which is what routes the dashboard to its tail auto-commit instead.
func lastCommitSHA(resp *captainai.Response) string {
	if resp == nil || resp.Workspace == nil || len(resp.Workspace.Commits) == 0 {
		return ""
	}
	return resp.Workspace.Commits[len(resp.Workspace.Commits)-1].SHA
}

func issueIDOf(todosInGroup []*types.TODO) string {
	for _, todo := range todosInGroup {
		if todo != nil {
			return todo.ID
		}
	}
	return ""
}
