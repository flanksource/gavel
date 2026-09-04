package lifecycle

import (
	"sort"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// Hooks are gavel's own contributions to a step run, in the order promptrun
// dispatches them — after the workflow's checks, before setup, so a PhaseRun
// commit still sees a live worktree:
//
//  1. the commit pipeline, one hook per policy the resolved spec declares —
//     gavel's, not captain's, because a `gates: full` policy runs gavel's
//     pre-commit gates and stamps the issue/session trailers (promptrun is told
//     CallerOwnsCommits so it builds none of its own);
//  2. the run environment, so a `gavel commit` the agent runs itself writes the
//     same trailers;
//  3. the spec recorder, trailing setup so the persisted spec is the one that
//     actually ran.
func (h *Host) Hooks(todo *types.TODO, req captainai.Request, meta todos.RunStartMetadata, report func(todos.RunStartMetadata)) []any {
	var hooks []any
	if req.Workflow != nil && len(req.Workflow.Commits) > 0 {
		hooks = append(hooks, commit.AgentHooks(commit.AgentHooksOptions{
			Commits: req.Workflow.Commits,
			Meta:    commit.AgentRunMetadata{IssueID: todo.ID, SessionID: meta.SessionID},
		})...)
	}
	if env := runEnvFor(todo, firstNonEmpty(req.SessionID, meta.SessionID)); len(env) > 0 {
		hooks = append(hooks, &runEnv{env: env})
	}
	return append(hooks, &specRecorder{meta: meta, workflow: req.Workflow, report: report})
}

// runEnv stamps GAVEL_ISSUE_ID / GAVEL_SESSION_ID onto the setup's environment
// so a `gavel commit` the agent runs itself writes the matching commit trailers.
//
// It is a PreRun hook rather than a mutation of the dispatched request because
// promptrun hands PreRun the pre-setup request: the env has to be on the spec
// the setup plugin then materialises the checkout from.
type runEnv struct {
	env map[string]string
}

func (h *runEnv) Name() string { return "gavel-run-env" }

func (h *runEnv) PreRun(hc *agent.HookContext) error {
	if len(h.env) == 0 || hc.Request == nil {
		return nil
	}
	if hc.Request.Setup == nil {
		hc.Request.Setup = &shell.Setup{}
	}
	// Sorted so the dispatched spec stays byte-identical for equal inputs, which
	// is what the run's persisted record is compared on.
	keys := make([]string, 0, len(h.env))
	for key := range h.env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	env := append([]string(nil), hc.Request.Setup.Env...)
	for _, key := range keys {
		env = append(env, key+"="+h.env[key])
	}
	hc.Request.Setup.Env = env
	return nil
}

func runEnvFor(todo *types.TODO, sessionID string) map[string]string {
	env := map[string]string{}
	if todo != nil && todo.ID != "" {
		env[commit.EnvIssueID] = todo.ID
	}
	if sessionID != "" {
		env[commit.EnvSessionID] = sessionID
	}
	if len(env) == 0 {
		return nil
	}
	return env
}

// specRecorder reports the run's request once setup has transformed it — the
// checkout consumed, Cwd pointing at the tree the agent works in. That
// distinction is the whole point: replaying a persisted spec that still asked
// for a checkout would clone a second tree.
//
// It reports from Post at the first turn boundary rather than from PreRun,
// because promptrun dispatches every caller hook's PreRun BEFORE the setup
// plugin's — a PreRun here would record the spec the run was asked for, which
// is exactly the one that must not be replayed.
type specRecorder struct {
	meta     todos.RunStartMetadata
	workflow *api.Workflow
	report   func(todos.RunStartMetadata)
	reported bool
}

func (r *specRecorder) Name() string { return "gavel-spec-recorder" }

// Phases pairs the earliest boundary that can see the transformed request with
// the one that is guaranteed to fire. PhaseTurn reports as soon as a turn lands,
// which is what the dashboard wants; PhaseAgent is the backstop for a run whose
// loop ended without a turn boundary. Whichever arrives first reports, once.
func (r *specRecorder) Phases() []agent.Phase {
	return []agent.Phase{agent.PhaseTurn, agent.PhaseAgent}
}

func (r *specRecorder) Post(hc *agent.HookContext, _ agent.Phase) error {
	if r.reported || hc.Request == nil {
		return nil
	}
	r.reported = true
	meta := r.meta
	spec := *hc.Request
	spec.Workflow = r.workflow
	meta.Spec = &spec
	r.report(meta)
	return nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
