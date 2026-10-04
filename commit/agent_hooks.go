package commit

import (
	"errors"

	"github.com/flanksource/captain/pkg/ai/agent"
	capcommit "github.com/flanksource/captain/pkg/ai/agent/commit"
	"github.com/flanksource/captain/pkg/api"
)

// AgentHooksOptions describes the commit hooks one agent run wants.
type AgentHooksOptions struct {
	// Commits are the run spec's declared policies (api.Spec.Workflow.Commits).
	// Empty produces no hooks and makes no commit.
	Commits []api.Commit
	Meta    AgentRunMetadata
	// Push pushes the branch after each commit the hooks cut. Only for loops whose
	// verification depends on the remote seeing the change (a PR's CI re-running);
	// it makes every commit a fixup chain must not be, so pair it with
	// `mode: commit`.
	Push bool
	// PushBranch, with Push, is the remote branch every commit is pushed to (see
	// Options.PushBranch).
	PushBranch string
}

// AgentHooks builds captain commit hooks from a run's spec, one per declared
// policy. Captain owns the lifecycle — which phase fires, tracking the anchor,
// collapsing the chain at the end of the run — while the Do callback routes the
// commit itself through gavel's pipeline, so the generated message, pre-commit
// gates and Gavel-Issue-Id / session trailers are identical to `gavel commit`.
func AgentHooks(opts AgentHooksOptions) []any {
	if len(opts.Commits) == 0 {
		return nil
	}
	hooks := make([]any, 0, len(opts.Commits))
	for _, spec := range opts.Commits {
		hook := capcommit.New(spec)
		hook.Do = agentCommitter(opts, spec.Message != "")
		hooks = append(hooks, hook)
	}
	return hooks
}

// agentCommitter adapts one captain commit Plan onto gavel's commit pipeline.
// keepSubject forwards captain's subject only when the spec named one: left
// empty the pipeline generates the message, which is the reason to route
// through it rather than committing with plain git.
func agentCommitter(opts AgentHooksOptions, keepSubject bool) func(*agent.HookContext, capcommit.Plan) (string, error) {
	return func(hc *agent.HookContext, plan capcommit.Plan) (string, error) {
		meta := opts.Meta
		if meta.SessionID == "" {
			// A provider that generates its own session id only reports it once the
			// first turn is underway; the trailer needs whatever it settled on. The
			// workspace holds it at every phase — hc.Turn is nil after the turn
			// ends, which left a post-run commit without its session trailer.
			meta.SessionID = hc.Workspace().SessionID
		}
		run := AgentRun{
			WorkDir:    plan.Dir,
			Meta:       meta,
			Fixup:      plan.Anchor,
			DryRun:     plan.DryRun,
			Push:       opts.Push,
			PushBranch: opts.PushBranch,
			// Captain selected the path set for this boundary; staging exactly it
			// keeps unrelated working-tree changes out of the commit. A fixup
			// resolves its own paths from the anchor, and the two are mutually
			// exclusive.
			Files: filesUnlessFixup(plan),
			// Captain already ran the cheap gates over the exact path set it
			// selected; only `gates: full` asks for gavel's fuller pipeline
			// (hooks, lint, linked-deps, tidy) on top.
			SkipGates: plan.Gates != api.CommitGatesFull,
		}
		if keepSubject {
			run.Message = plan.Subject
		}
		result, err := RunAfterAgent(hc.Context, run)
		// A turn that edited nothing stageable is a normal outcome, not a
		// failure: the hook records no commit and the run continues. So is a turn
		// that left a merge half-resolved — the merge stays for the next turn,
		// and the run's verify commands feed back which paths remain.
		if errors.Is(err, ErrMergeUnresolved) {
			hc.Notify("[post-%s] %v — not committing", plan.Phase, err)
			return "", nil
		}
		if errors.Is(err, ErrNothingStaged) || errors.Is(err, ErrSessionNoFiles) {
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

func filesUnlessFixup(plan capcommit.Plan) []string {
	if plan.Anchor != "" {
		return nil
	}
	return plan.Paths
}
