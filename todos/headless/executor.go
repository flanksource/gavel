// Package headless drives an AI coding agent non-interactively via captain's
// streaming providers, composed through captain's agent.Runner: the mode's
// dotprompt template renders the initial request (frontmatter declares model
// and permissions), verify plugins drive re-run-with-feedback iterations, and
// the run's structured result envelope is captured when the session ends or
// times out. The cmux backends route through the same executor; there is no
// screen-scraping or session-log tailing.
package headless

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	capsetup "github.com/flanksource/captain/pkg/ai/agent/setup"
	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	captainapi "github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	gavelai "github.com/flanksource/gavel/ai"
	todopkg "github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/claude"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/utils"
)

const (
	defaultTimeout    = 30 * time.Minute
	defaultCheckIters = 3
)

// streamFunc opens a captain event stream for a request. It is the seam tests
// inject a fake stream through; production builds it from the resolved
// model/backend. canUseTool is the tool-permission broker, which lives on the
// provider Config (baked in at construction) rather than on the request — it is
// passed here so tests that bypass provider construction can still observe it.
type streamFunc func(ctx context.Context, req captainai.Request, canUseTool captainai.PermissionFunc) (<-chan captainai.Event, error)

type Executor struct {
	config todopkg.AgentRunConfig
	agent  string
	stream streamFunc
}

var _ todopkg.RunSpecProvider = (*Executor)(nil)
var _ todopkg.RunRuntimeProvider = (*Executor)(nil)

type option func(*Executor)

func withStream(stream streamFunc) option {
	return func(executor *Executor) {
		executor.stream = stream
	}
}

func NewExecutor(config todopkg.AgentRunConfig, options ...option) *Executor {
	if config.Mode == "" {
		config.Mode = types.ModeRun
	}
	agentName, _ := claude.ResolveAgent(config.Spec.Name)
	executor := &Executor{config: config, agent: agentName}
	for _, option := range options {
		option(executor)
	}
	return executor
}

func (e *Executor) Name() string {
	return e.driver() + "-" + e.agent
}

// RunPromptName implements todos.RunPromptProvider: it reports which named
// prompt this executor renders, so the run's durable identity distinguishes two
// prompts of the same behaviour class.
func (e *Executor) RunPromptName() string {
	if name := strings.TrimSpace(e.config.Prompt); name != "" {
		return name
	}
	return string(e.config.Mode)
}

func (e *Executor) driver() string {
	return string(e.config.Spec.Mode)
}

func (e *Executor) RunRuntimeSelection() captaindb.PromptRunRuntimeSelection {
	return captaindb.PromptRunRuntimeSelection{
		Provider: captainapi.RuntimeOf(e.config.Spec.Provider, e.config.Spec.Mode).Provider,
		Mode:     string(e.config.Spec.Mode),
		Model:    e.config.Spec.Name,
		Effort:   string(e.config.Spec.Effort),
	}
}

func (e *Executor) Execute(ctx *todopkg.ExecutorContext, todo *types.TODO) (*todopkg.ExecutionResult, error) {
	return e.ExecuteGroup(ctx, []*types.TODO{todo})
}

// RenderRunSpec implements todos.RunSpecProvider. It shares the exact renderer
// used by ExecuteGroup, so native admission persists Captain's real request —
// prompt included — before the external provider is dispatched. What setup then
// does to that request is reported separately, once the run is under way.
func (e *Executor) RenderRunSpec(_ *todopkg.ExecutorContext, todo *types.TODO) (captainai.Request, error) {
	return e.renderInitialRequest([]*types.TODO{todo})
}

func (e *Executor) ExecuteGroup(ctx *todopkg.ExecutorContext, todosInGroup []*types.TODO) (*todopkg.ExecutionResult, error) {
	start := time.Now()
	if len(todosInGroup) == 0 {
		return nil, fmt.Errorf("no todos supplied")
	}
	rendered, err := e.renderInitialRequest(todosInGroup)
	if err != nil {
		return nil, err
	}
	req, providerSessionID, canUseTool, err := e.buildRequest(ctx, todosInGroup, rendered, e.config.Resume)
	if err != nil {
		return nil, err
	}
	return e.run(ctx, start, req, canUseTool, providerSessionID, todosInGroup)
}

func (e *Executor) renderInitialRequest(todosInGroup []*types.TODO) (captainai.Request, error) {
	if len(todosInGroup) == 0 {
		return captainai.Request{}, fmt.Errorf("no todos supplied")
	}
	workDir := groupWorkDir(e.config.WorkDir, todosInGroup)
	rendered, _, err := todoprompt.Render(todosInGroup, e.config.PromptOptions(workDir))
	if err != nil {
		return captainai.Request{}, err
	}
	return rendered, nil
}

// SendFeedback resumes the group's prior agent session with a user message (an
// answer to the agent's questions, or ad-hoc feedback) and waits for the next
// turn, capturing a fresh result envelope. It implements todos.FeedbackExecutor
// for every captain backend (cmux and the headless SDK), which all resume via
// req.SessionID. It requires a recorded prior session.
func (e *Executor) SendFeedback(ctx *todopkg.ExecutorContext, todosInGroup []*types.TODO, feedback string) (*todopkg.ExecutionResult, error) {
	start := time.Now()
	feedback = strings.TrimSpace(feedback)
	if feedback == "" {
		err := fmt.Errorf("%s: empty feedback", e.Name())
		return e.failed(start, err), err
	}
	if priorSessionID(todosInGroup) == "" {
		err := fmt.Errorf("%s: no prior session to resume for feedback", e.Name())
		return e.failed(start, err), err
	}
	schema, err := todoprompt.EnvelopeSchemaJSON(e.envelope())
	if err != nil {
		return e.failed(start, err), err
	}
	// A resumed turn keeps the mode template's spec — workflow (commit hooks),
	// permissions, model and effort — and swaps only the user message, so an
	// answered run can still auto-commit exactly like the turn it continues.
	base, err := e.renderInitialRequest(todosInGroup)
	if err != nil {
		return e.failed(start, err), err
	}
	base.Prompt.User = feedback
	base.Prompt.SchemaJSON = schema
	base.Prompt.Source = "todos.feedback"
	req, providerSessionID, canUseTool, err := e.buildRequest(ctx, todosInGroup, base, true)
	if err != nil {
		return e.failed(start, err), err
	}
	return e.run(ctx, start, req, canUseTool, providerSessionID, todosInGroup)
}

// run drives the request through captain's agent.Runner (Verify hooks gate
// run-mode iterations; a failing verdict's Retry request carries the session
// id forward explicitly), then resolves the response contract in precedence
// order: native terminal outcome, structured data, response text.
func (e *Executor) run(ctx *todopkg.ExecutorContext, start time.Time, req captainai.Request, canUseTool captainai.PermissionFunc, providerSessionID string, todosInGroup []*types.TODO) (*todopkg.ExecutionResult, error) {
	workDir := groupWorkDir(e.config.WorkDir, todosInGroup)
	provider, err := e.provider(req, canUseTool, providerSessionID)
	if err != nil {
		return e.failed(start, err), err
	}
	defer func() {
		if err := gavelai.CloseProvider(provider); err != nil {
			ctx.Logger.Warnf("%s: failed to close AI provider: %v", e.Name(), err)
		}
	}()
	runMeta := todopkg.RunStartMetadata{
		SessionID:     firstNonEmpty(firstNonEmpty(req.SessionID, providerSessionID), priorSessionID(todosInGroup)),
		Mode:          string(e.config.Mode),
		Driver:        e.driver(),
		Agent:         e.agent,
		Provider:      provider.GetRuntime().Provider,
		RuntimeMode:   string(provider.GetRuntime().Mode),
		ResolvedModel: provider.GetModel(),
		Effort:        string(req.Effort),
	}
	ctx.RecordRunStart(runMeta)

	modelSource := "runtime-default"
	if strings.TrimSpace(req.Name) != "" {
		modelSource = "todos." + string(e.config.Mode) + "-prompt"
	}
	if strings.TrimSpace(e.config.Spec.Name) != "" {
		modelSource = "run-option"
	}
	ctx.Logger.Infof(
		"Resolved TODO runtime: driver=%s agent=%s provider=%s mode=%s model=%s effort=%s model_source=%s; dispatching %d TODO(s) cwd=%s",
		runMeta.Driver, runMeta.Agent, firstNonEmpty(runMeta.Provider, "unknown"),
		firstNonEmpty(runMeta.RuntimeMode, "default"), firstNonEmpty(runMeta.ResolvedModel, "default"),
		firstNonEmpty(runMeta.Effort, "default"), modelSource, len(todosInGroup), workDir,
	)
	gavelai.NormalizeEnv()

	timeout, err := e.timeout()
	if err != nil {
		return e.failed(start, err), err
	}
	streamCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := &todopkg.ExecutionResult{ExecutorName: e.Name(), Runtime: runMeta, Transcript: ctx.GetTranscript()}
	var sawResult bool

	fixtureVerifiers := e.config.Verifiers
	if e.config.Mode != types.ModeRun {
		fixtureVerifiers = nil // plan runs are read-only; nothing to verify
	}
	var verifyHooks []any
	if e.config.Mode == types.ModeRun {
		verifyHooks = append(verifyHooks, capverify.HooksForWorkflow(req.Workflow)...)
		for _, verifier := range fixtureVerifiers {
			verifyHooks = append(verifyHooks, verifier)
		}
	}
	maxIter := capverify.MaxIterationsForWorkflow(req.Workflow)
	if len(fixtureVerifiers) > 0 {
		maxIter = e.config.MaxIterations
		if maxIter <= 0 {
			maxIter = defaultCheckIters
		}
	}

	scope := capverify.ScopeForWorkflow(req.Workflow)
	if len(fixtureVerifiers) > 0 && (req.Workflow == nil || req.Workflow.Verify == nil || req.Workflow.Verify.Scope == "") {
		scope = agent.ScopeChanged
	}
	// Commit hooks lead the list: at PhaseRun they collapse the run's fixup
	// chain before any teardown hook can act on the result. A plan run declares
	// no commits — it is read-only by construction.
	var hooks []any
	if e.config.Mode == types.ModeRun {
		hooks = append(hooks, commitHooks(req, todosInGroup, runMeta.SessionID)...)
	}
	hooks = append(hooks, verifyHooks...)
	// Setup trails the list because the runner dispatches Post in hook order:
	// its teardown at PhaseRun must come after the commit hooks above have cut
	// their commits, or it would remove the worktree they commit from. PreRun
	// order does not matter — no other hook here declares one. The plugin
	// materialises the checkout and rewrites the spec to describe where the work
	// landed; the group's work dir is what relative setup paths anchor to, and
	// it seeds the workspace the plugin then repoints at the prepared tree.
	hooks = append(hooks, &capsetup.Plugin{BaseDir: workDir})
	// Trails setup so it reports the transformed spec, not the requested one.
	hooks = append(hooks, &specRecorder{meta: runMeta, report: ctx.RecordRunStart})

	runner := &agent.Runner[string]{
		Provider:      provider,
		Request:       req,
		Hooks:         hooks,
		MaxIterations: maxIter,
		// Repo is the root of the tree workDir sits in, not workDir itself: a
		// todo carrying a subdirectory CWD still has its edits recorded relative
		// to the root, which is the namespace the commit hooks compare against.
		Repo:  utils.GitRoot(workDir),
		Cwd:   workDir,
		Scope: scope,
		OnEvent: func(_ int, ev captainai.Event) {
			e.handleEvent(ctx, ev, result, todosInGroup, &sawResult, runMeta)
		},
	}

	rres, runErr := runner.Run(streamCtx)
	result.Duration = time.Since(start)
	// Flushed here rather than from the hooks that produced them: a hook firing
	// mid-turn cannot know the transcript session's id, because that row only
	// exists once the monitor has ingested the provider's log. Before the
	// cancellation and error returns below — a run's commits are just as worth
	// narrating when it failed.
	if rres.Response != nil && rres.Response.Workspace != nil {
		ctx.RecordNotices(runMeta.SessionID, rres.Response.Workspace.Notices)
	}
	if errors.Is(context.Cause(streamCtx), todopkg.ErrExecutionCancelled) {
		result.Cancelled = true
		result.ErrorMessage = todopkg.ErrExecutionCancelled.Error()
		result.Summary = todopkg.ErrExecutionCancelled.Error()
		return result, context.Canceled
	}
	timedOut := errors.Is(streamCtx.Err(), context.DeadlineExceeded) && !sawResult

	// Report what the commit hooks cut before the error paths return: the work
	// is on the branch either way, and CommitSHA is what stops the dashboard's
	// tail auto-commit from committing the same change set a second time.
	result.CommitSHA = lastCommitSHA(rres.Response)

	if runErr != nil && !timedOut {
		result.ErrorMessage = runErr.Error()
		return result, fmt.Errorf("%s: %w", e.Name(), runErr)
	}

	// Definition-of-done verdict: when the run had verifiers, the loop stops
	// "condition-met" only if every verifier passed; any other stop reason
	// (budget/cost exhausted, timeout) means a check is still red. This drives
	// the verified/unverified terminal status in applyOutcome.
	if len(verifyHooks) > 0 {
		passed := rres.Loop != nil && rres.Loop.StopReason == "condition-met"
		var output *types.VerificationOutput
		for _, verdict := range rres.Verdicts {
			switch value := verdict.Output.(type) {
			case types.VerificationOutput:
				copy := value
				output = &copy
			case *types.VerificationOutput:
				output = value
			}
		}
		result.DoD = &todopkg.DoDOutcome{Ran: true, Passed: passed, Output: output}
	}
	envErr := e.captureEnvelope(result, rres.Response)
	return e.classify(result, sawResult, timedOut, timeout, envErr)
}

// classify folds the transport outcome and the envelope into the final verdict.
// The envelope's endStatus drives it when present; otherwise the transport
// result decides as before. A missing response contract is always an error.
func (e *Executor) classify(result *todopkg.ExecutionResult, sawResult, timedOut bool, timeout time.Duration, envErr error) (*todopkg.ExecutionResult, error) {
	if envErr != nil {
		result.ErrorMessage = envErr.Error()
		return result, fmt.Errorf("%s: %w", e.Name(), envErr)
	}
	switch {
	case result.EndStatus == types.EndFailed:
		result.Success = false
		if result.ErrorMessage == "" {
			result.ErrorMessage = result.Summary
		}
		return result, fmt.Errorf("%s: agent reported failure: %s", e.Name(), result.Summary)
	case result.EndStatus != "":
		// completed / ask: the response contract is authoritative.
		result.Success = true
		return result, nil
	case timedOut:
		err := fmt.Errorf("%s run did not complete within %s", e.Name(), timeout)
		result.ErrorMessage = err.Error()
		return result, err
	case !sawResult:
		err := fmt.Errorf("%s stream ended without a result event", e.Name())
		result.ErrorMessage = err.Error()
		return result, err
	case !result.Success:
		if result.ErrorMessage == "" {
			result.ErrorMessage = "agent reported failure"
		}
		return result, fmt.Errorf("%s: %s", e.Name(), result.ErrorMessage)
	default:
		return result, nil
	}
}

func (e *Executor) handleEvent(ctx *todopkg.ExecutorContext, ev captainai.Event, result *todopkg.ExecutionResult, todosInGroup []*types.TODO, sawResult *bool, runMeta todopkg.RunStartMetadata) {
	transcript := ctx.GetTranscript()
	switch ev.Kind {
	case captainai.EventText:
		if ev.Text == "" {
			return
		}
		transcript.AddExecutorMessage(truncate(ev.Text, 200), todopkg.EntryText, nil)
		ctx.Notify(todopkg.Notification{Type: todopkg.NotifyProgress, Message: truncate(ev.Text, 100)})
	case captainai.EventThinking:
		transcript.AddExecutorMessage(ev.Text, todopkg.EntryThinking, nil)
		ctx.Notify(todopkg.Notification{Type: todopkg.NotifyThinking, Message: truncate(ev.Text, 100)})
	case captainai.EventToolUse:
		action := toolSummary(ev)
		transcript.AddExecutorMessage(action, todopkg.EntryAction, map[string]any{"tool": ev.Tool})
		ctx.Notify(todopkg.Notification{Type: todopkg.NotifyAction, Message: action})
	case captainai.EventPermission:
		action := toolSummary(ev)
		transcript.AddExecutorMessage("awaiting approval: "+action, todopkg.EntryAction, map[string]any{"tool": ev.Tool})
		ctx.Notify(todopkg.Notification{Type: todopkg.NotifyApproval, Message: action})
	case captainai.EventSystem:
		if ev.SessionID != "" {
			recordSessionID(todosInGroup, ev.SessionID)
			ctx.RecordSessionID(ev.SessionID)
			runMeta.SessionID = ev.SessionID
			result.Runtime.SessionID = ev.SessionID
			ctx.RecordRunStart(runMeta)
		}
		// A lifecycle hook narrating what it did between turns — the commit it
		// cut, the chain it squashed. Recorded in the run's own voice so the
		// attempt transcript shows why the tree changed between two turns that
		// never mention it.
		if ev.Text != "" {
			transcript.AddExecutorMessage(ev.Text, todopkg.EntryAction, map[string]any{"role": "system"})
			ctx.Notify(todopkg.Notification{Type: todopkg.NotifyAction, Message: ev.Text})
		}
	case captainai.EventResult:
		*sawResult = true
		result.Success = ev.Success
		if ev.Usage != nil {
			result.TokensUsed += ev.Usage.TotalTokens()
		}
		result.CostUSD += ev.CostUSD
		if !ev.Success && ev.Error != "" {
			result.ErrorMessage = ev.Error
		}
	case captainai.EventError:
		result.ErrorMessage = ev.Error
		ctx.Notify(todopkg.Notification{Type: todopkg.NotifyError, Message: ev.Error})
	}
}

func (e *Executor) failed(start time.Time, err error) *todopkg.ExecutionResult {
	return &todopkg.ExecutionResult{
		Success:      false,
		ExecutorName: e.Name(),
		Duration:     time.Since(start),
		ErrorMessage: err.Error(),
	}
}

// isCmuxBackend reports whether this executor drives the interactive cmux TUI
// provider (vs the headless agent and cli modes).
func (e *Executor) isCmuxBackend() bool {
	return e.config.Spec.Mode == captainapi.ModeCmux
}

func (e *Executor) timeout() (time.Duration, error) {
	if strings.TrimSpace(e.config.Spec.Budget.Timeout) == "" {
		return defaultTimeout, nil
	}
	timeout, err := time.ParseDuration(e.config.Spec.Budget.Timeout)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid timeout %q: %w", e.Name(), e.config.Spec.Budget.Timeout, err)
	}
	if timeout <= 0 {
		return 0, fmt.Errorf("%s: timeout must be greater than zero", e.Name())
	}
	return timeout, nil
}

func groupWorkDir(fallback string, todoList []*types.TODO) string {
	for _, todo := range todoList {
		if todo != nil && strings.TrimSpace(todo.CWD) != "" {
			if filepath.IsAbs(todo.CWD) {
				return filepath.Clean(todo.CWD)
			}
			if fallback != "" {
				return filepath.Clean(filepath.Join(fallback, todo.CWD))
			}
			return filepath.Clean(todo.CWD)
		}
	}
	if fallback != "" {
		return filepath.Clean(fallback)
	}
	return "."
}

var (
	_ todopkg.Executor         = (*Executor)(nil)
	_ todopkg.GroupExecutor    = (*Executor)(nil)
	_ todopkg.FeedbackExecutor = (*Executor)(nil)
)
