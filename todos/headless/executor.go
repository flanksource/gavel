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
	"fmt"
	"path/filepath"
	"strings"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	gavelai "github.com/flanksource/gavel/ai"
	todopkg "github.com/flanksource/gavel/todos"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
)

const (
	defaultTimeout        = 30 * time.Minute
	defaultCheckIters     = 3
	finalResultTimeout    = 5 * time.Minute
	finalResultMaxTurns   = 4
	defaultToolsAllowlist = "Read,Edit,Write,Bash,Glob,Grep"
)

// streamFunc opens a captain event stream for a request. It is the seam tests
// inject a fake stream through; production builds it from the resolved
// model/backend. canUseTool is the tool-permission broker, which lives on the
// provider Config (baked in at construction) rather than on the request — it is
// passed here so tests that bypass provider construction can still observe it.
type streamFunc func(ctx context.Context, req captainai.Request, canUseTool captainai.PermissionFunc) (<-chan captainai.Event, error)

type Config struct {
	WorkDir string
	Agent   string // "claude" or "codex"
	// Model / Backend are explicit user overrides; empty defers to the mode's
	// .prompt frontmatter (captain resolves the provider from the model name).
	Model    string
	Backend  string
	Effort   string
	MaxTurns int
	Tools    []string
	Timeout  time.Duration
	// Mode selects the built-in prompt (run or plan); empty means run. The plan
	// template's frontmatter carries the plan permission posture.
	Mode types.RunMode
	// ExistingPlan is the current content of the todo's recorded plan file
	// (plan mode); empty means no prior plan.
	ExistingPlan string
	// PromptOverride, when set, replaces the rendered prompt body — the
	// dashboard's editable prompt. The envelope schema instruction is still
	// appended so the structured-output contract holds.
	PromptOverride string
	// Verifiers gate each run-mode iteration: a failing verdict's feedback is
	// sent back to the same session for another attempt (captain agent.Runner
	// drives the loop). Plan runs register none.
	Verifiers []agent.Plugin
	// MaxIterations bounds the verify-feedback loop (only meaningful with
	// Verifiers); 0 defaults to 3.
	MaxIterations int
	// Resume / SessionID: Resume continues the todo's prior session; SessionID is
	// the pre-generated claude session id for a fresh cmux run (the dashboard
	// knows it up front to follow the session log live).
	Resume    bool
	SessionID string
	// ToolModes is the per-tool exposure (tool name → enabled/ask/disabled) and
	// PermissionMode the base permission posture (a clicky ClaudePermissionMode).
	// Together they shape the request's api.Permissions: enabled tools are
	// allow-listed, disabled ones denied, and ask ones brokered via CanUseTool.
	ToolModes      map[string]string
	PermissionMode string
	// Approvals brokers tool permissions over the can_use_tool control protocol:
	// each tool the agent wants to run that is not auto-approved is surfaced to the
	// process-wide approval registry (the dashboard resolves it). Off by default so
	// CLI runs with no resolver keep the auto-approve behaviour instead of blocking.
	Approvals bool
	// Stream overrides the captain provider; nil uses the real claude/codex CLI.
	Stream streamFunc
}

type Executor struct {
	config Config
}

func NewExecutor(config Config) *Executor {
	if config.Agent == "" {
		config.Agent = "claude"
	}
	if len(config.Tools) == 0 {
		config.Tools = strings.Split(defaultToolsAllowlist, ",")
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.Mode == "" {
		config.Mode = types.ModeRun
	}
	return &Executor{config: config}
}

func (e *Executor) Name() string {
	if e.isCmuxBackend() {
		return "cmux-" + e.config.Agent
	}
	return "headless-" + e.config.Agent
}

func (e *Executor) Execute(ctx *todopkg.ExecutorContext, todo *types.TODO) (*todopkg.ExecutionResult, error) {
	return e.ExecuteGroup(ctx, []*types.TODO{todo})
}

func (e *Executor) ExecuteGroup(ctx *todopkg.ExecutorContext, todosInGroup []*types.TODO) (*todopkg.ExecutionResult, error) {
	start := time.Now()
	if len(todosInGroup) == 0 {
		return nil, fmt.Errorf("no todos supplied")
	}
	workDir := groupWorkDir(e.config.WorkDir, todosInGroup)
	tmpl, err := todoprompt.ResolveTemplate(workDir, e.config.Mode)
	if err != nil {
		return nil, err
	}
	rendered, _, err := todoprompt.Render(todosInGroup, todoprompt.Options{
		WorkDir:      workDir,
		Mode:         e.config.Mode,
		Effort:       e.config.Effort,
		Template:     tmpl,
		ExistingPlan: e.config.ExistingPlan,
		BodyOverride: e.config.PromptOverride,
	})
	if err != nil {
		return nil, err
	}
	req, providerSessionID, canUseTool := e.buildRequest(ctx, todosInGroup, rendered, e.config.Resume)
	return e.run(ctx, start, req, canUseTool, providerSessionID, todosInGroup)
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
	schema, err := todoprompt.EnvelopeSchemaJSON(e.config.Mode)
	if err != nil {
		return e.failed(start, err), err
	}
	base := captainai.Request{Prompt: api.Prompt{
		User:   feedback + "\n\n" + gavelai.SchemaInstruction(schema),
		Source: "todos.feedback",
	}}
	req, providerSessionID, canUseTool := e.buildRequest(ctx, todosInGroup, base, true)
	return e.run(ctx, start, req, canUseTool, providerSessionID, todosInGroup)
}

// run drives the request through captain's agent.Runner (verify plugins gate
// run-mode iterations; SessionReuse carries feedback turns), then captures the
// structured result envelope — from the final iteration's text, or by resuming
// the session with a final-result request when the text has none or the run
// timed out.
func (e *Executor) run(ctx *todopkg.ExecutorContext, start time.Time, req captainai.Request, canUseTool captainai.PermissionFunc, providerSessionID string, todosInGroup []*types.TODO) (*todopkg.ExecutionResult, error) {
	provider, err := e.provider(req, canUseTool, providerSessionID)
	if err != nil {
		return e.failed(start, err), err
	}

	ctx.Logger.Infof("%s: dispatching %d TODO(s) in %s", e.Name(), len(todosInGroup), req.Cwd())
	gavelai.NormalizeEnv()

	streamCtx, cancel := context.WithTimeout(ctx, e.config.Timeout)
	defer cancel()

	result := &todopkg.ExecutionResult{ExecutorName: e.Name(), Transcript: ctx.GetTranscript()}
	var sawResult bool

	plugins := e.config.Verifiers
	if e.config.Mode != types.ModeRun {
		plugins = nil // plan runs are read-only; nothing to verify
	}
	maxIter := 1
	if len(plugins) > 0 {
		maxIter = e.config.MaxIterations
		if maxIter <= 0 {
			maxIter = defaultCheckIters
		}
	}

	runner := &agent.Runner{
		Provider: provider,
		Repo:     req.Cwd(),
		Cwd:      req.Cwd(),
		Plugins:  plugins,
		Loop: captainai.LoopOptions{
			MaxIterations: maxIter,
			SessionReuse:  true,
			OnEvent: func(_ int, ev captainai.Event) {
				e.handleEvent(ctx, ev, result, todosInGroup, &sawResult)
			},
		},
		Build: func(_ *agent.RunContext, iter int, _ *captainai.LoopIteration, feedback string) captainai.Request {
			if iter == 0 {
				return req
			}
			next := req
			next.Prompt = api.Prompt{User: feedback, Source: "todos.check"}
			next.SessionID = "" // SessionReuse fills the prior iteration's session
			return next
		},
	}

	rres, runErr := runner.Run(streamCtx)
	result.Duration = time.Since(start)
	timedOut := streamCtx.Err() != nil && !sawResult

	if runErr != nil && !timedOut {
		result.ErrorMessage = runErr.Error()
		return result, fmt.Errorf("%s: %w", e.Name(), runErr)
	}

	// Definition-of-done verdict: when the run had verifiers, the loop stops
	// "condition-met" only if every verifier passed; any other stop reason
	// (budget/cost exhausted, timeout) means a check is still red. This drives
	// the verified/unverified terminal status in applyOutcome.
	if len(plugins) > 0 {
		passed := rres != nil && rres.Loop != nil && rres.Loop.StopReason == "condition-met"
		result.DoD = &todopkg.DoDOutcome{Ran: true, Passed: passed}
	}

	envErr := e.captureEnvelope(ctx, provider, result, rres, todosInGroup, timedOut)
	return e.classify(result, sawResult, timedOut, envErr)
}

// classify folds the transport outcome and the envelope into the final verdict.
// The envelope's endStatus drives it when present; otherwise the transport
// result decides as before (plan mode already hard-failed in captureEnvelope
// when no envelope could be obtained).
func (e *Executor) classify(result *todopkg.ExecutionResult, sawResult, timedOut bool, envErr error) (*todopkg.ExecutionResult, error) {
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
		// completed / ask: the envelope is authoritative — a timed-out run that
		// still produced a final result via the resume turn counts.
		result.Success = true
		return result, nil
	case timedOut:
		err := fmt.Errorf("%s run did not complete within %s", e.Name(), e.config.Timeout)
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

func (e *Executor) handleEvent(ctx *todopkg.ExecutorContext, ev captainai.Event, result *todopkg.ExecutionResult, todosInGroup []*types.TODO, sawResult *bool) {
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
// provider (vs the headless SDK backends).
func (e *Executor) isCmuxBackend() bool {
	switch captainai.Backend(strings.TrimSpace(e.config.Backend)) {
	case captainai.BackendClaudeCmux, captainai.BackendCodexCmux:
		return true
	default:
		return false
	}
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
