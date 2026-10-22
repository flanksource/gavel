package lifecycle

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/tools"
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
//  3. the run labels identifying the todo and workspace on the task row.
//
// Recording the run — its start, the spec setup produced, its outcome — is
// Captain's: promptrun appends its own recorder after setup.
type HookOptions struct {
	Meta todos.RunStartMetadata
	pr   *prExecution
	// tools is the run's tool state, where a PR turn reads the threads it
	// reported addressed.
	tools *tools.Run
}

func (h *Host) Hooks(todo *types.TODO, req captainai.Request, opts HookOptions) []any {
	return h.runHooks(todo, req, opts).build(opts.pr.forRun(opts.tools))
}

// RunHooks is the data gavel's hooks for one run are built from, so a run that executes in a
// sandbox rebuilds the same hooks there. It crosses to the sandbox as JSON.
type RunHooks struct {
	// Commits are the commit policies the resolved spec declares, and Meta the trailers the
	// commits carry.
	Commits []api.Commit            `json:"commits,omitempty"`
	Meta    commit.AgentRunMetadata `json:"meta"`
	Env     map[string]string       `json:"env,omitempty"`
	Labels  map[string]string       `json:"labels,omitempty"`
	// PR is the linked PR a step works on; nil for a todo linked to none.
	PR *PRHooks `json:"pr,omitempty"`
}

// PRHooks is the data the hook for a step on a linked PR is rebuilt from: the PR, its source
// branch at the head the step starts from, whether the step writes to it, the checkout the
// step works in, the step's envelope events a turn's result is held to, and the review threads
// a turn may report addressed.
type PRHooks struct {
	URL         string             `json:"url"`
	Number      int                `json:"number"`
	Target      provider.RunTarget `json:"target"`
	Write       bool               `json:"write"`
	WorkDir     string             `json:"workDir"`
	Events      []string           `json:"events,omitempty"`
	OpenThreads []string           `json:"openThreads,omitempty"`
}

func (h *Host) runHooks(todo *types.TODO, req captainai.Request, opts HookOptions) RunHooks {
	hooks := RunHooks{
		Env:    runEnvFor(todo, firstNonEmpty(req.SessionID, opts.Meta.SessionID)),
		Labels: runLabelsFor(todo, opts.Meta, h.WorkDir),
	}
	if req.Workflow != nil && len(req.Workflow.Commits) > 0 {
		hooks.Commits = req.Workflow.Commits
		hooks.Meta = commit.AgentRunMetadata{IssueID: todo.ID, SessionID: opts.Meta.SessionID}
	}
	if pr := opts.pr; pr != nil && todo.PR != nil {
		hooks.PR = &PRHooks{
			URL: todo.PR.URL, Number: pr.number, Target: *pr.target, Write: pr.write, WorkDir: pr.workDir,
			Events: pr.events, OpenThreads: slices.Sorted(maps.Keys(pr.openThreads)),
		}
	}
	return hooks
}

// Build makes the hooks, in the order Hooks documents, and the tools that must run beside them
// rather than where the run's other tools do: pr_thread_addressed records the threads the PR
// hook resolves after each turn's push, so in a sandbox job it runs in the job. A linked PR's
// client is resolved where the hooks are built: inside a sandbox job, that is serve's run API.
func (r RunHooks) Build() ([]any, []api.ToolDefinition, error) {
	if r.PR == nil {
		return r.build(nil), nil, nil
	}
	client, number, err := provider.ResolveURL(provider.Options{WorkDir: r.PR.WorkDir}, r.PR.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve linked PR %s: %w", r.PR.URL, err)
	}
	if number != r.PR.Number {
		return nil, nil, fmt.Errorf("linked PR %s is #%d, not #%d", r.PR.URL, number, r.PR.Number)
	}
	target := r.PR.Target
	open := make(map[string]bool, len(r.PR.OpenThreads))
	for _, id := range r.PR.OpenThreads {
		open[id] = true
	}
	run := &tools.Run{}
	pr := &prExecution{
		client: client, number: number, target: &target, workDir: r.PR.WorkDir, write: r.PR.Write,
		events: r.PR.Events, openThreads: open, tools: run,
	}
	return r.build(pr), tools.Sets()[tools.SetPR](tools.Options{Run: run, PRWrite: r.PR.Write}), nil
}

// build makes the hooks for a run, with pr's commit target and its own hook when the todo is
// linked to a PR.
func (r RunHooks) build(pr *prExecution) []any {
	var hooks []any
	if len(r.Commits) > 0 {
		options := commit.AgentHooksOptions{Commits: r.Commits, Meta: r.Meta}
		if pr != nil && pr.write {
			options.PRTarget, options.ShouldCommit = pr.target, pr.completedTurn
		}
		hooks = append(hooks, commit.AgentHooks(options)...)
	}
	if pr != nil {
		hooks = append(hooks, pr)
	}
	if len(r.Env) > 0 {
		hooks = append(hooks, &runEnv{env: r.Env})
	}
	if len(r.Labels) > 0 {
		hooks = append(hooks, &runLabels{labels: r.Labels})
	}
	return hooks
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

// runLabels tells the provider which todo this run is working, so the agent
// process it supervises names itself in the task list. Without it every
// concurrent agent renders as the same anonymous binary path and an operator
// watching /tasks cannot tell one run from another.
//
// A PreRun hook for the same reason runEnv is: the labels have to be on the
// request the provider is later built from.
type runLabels struct {
	labels map[string]string
}

func (h *runLabels) Name() string { return "gavel-run-labels" }

func (h *runLabels) PreRun(hc *agent.HookContext) error {
	if len(h.labels) == 0 || hc.Request == nil {
		return nil
	}
	if hc.Request.Labels == nil {
		hc.Request.Labels = map[string]string{}
	}
	for key, value := range h.labels {
		hc.Request.Labels[key] = value
	}
	return nil
}

// runLabelsFor is the identity captain cannot derive on its own: what the work
// is called, which todo it belongs to, where to read more, and which phase of
// the todo is running. Everything else on the task row — model, provider, mode,
// session — the provider fills in from what it already knows.
func runLabelsFor(todo *types.TODO, meta todos.RunStartMetadata, workDir string) map[string]string {
	if todo == nil || todo.ID == "" {
		return nil
	}
	labels := map[string]string{
		"todo": todo.ID,
		"href": "/todos/" + todo.ID,
	}
	if title := strings.TrimSpace(todo.Title); title != "" {
		labels["title"] = title
	}
	if meta.Mode != "" {
		labels["phase"] = meta.Mode
	}
	// The workspace the run belongs to, so a dashboard watching several can ask
	// the right one for this run's transcript.
	if workDir != "" {
		labels["dir"] = workDir
	}
	return labels
}

// admittedHook runs once Captain has committed the run's admission — the first
// gavel code a recorded run executes after it — and before setup relocates the
// tree: the dashboard learns the run's identity, and the runtime starts
// driving it.
type admittedHook struct {
	admitted func() error
}

func (h *admittedHook) Name() string { return "gavel-run-admitted" }

func (h *admittedHook) PreRun(*agent.HookContext) error { return h.admitted() }

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
