// Package headless drives an AI coding agent (claude or codex) non-interactively
// via captain's streaming providers (the Claude Agent SDK over JSON-RPC, and
// `codex app-server` over JSON-RPC). Unlike the cmux driver it does not automate
// a terminal: it consumes a structured ai.Event stream and completes on the
// terminal EventResult, so there is no screen-scraping or session-log tailing.
package headless

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	captainai "github.com/flanksource/captain/pkg/ai"
	captainprovider "github.com/flanksource/captain/pkg/ai/provider"
	"github.com/flanksource/captain/pkg/api"
	gavelai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/commit"
	todopkg "github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/claude"
	"github.com/flanksource/gavel/todos/types"
)

const defaultTimeout = 30 * time.Minute

// streamFunc opens a captain event stream for a request. It is the seam tests
// inject a fake stream through; production builds it from the agent + model.
// canUseTool is the tool-permission broker, which now lives on the provider
// Config (baked in at construction) rather than on the request — it is passed
// here so tests that bypass provider construction can still observe it.
type streamFunc func(ctx context.Context, req captainai.Request, canUseTool captainai.PermissionFunc) (<-chan captainai.Event, error)

type Config struct {
	WorkDir  string
	Agent    string // "claude" or "codex"
	Model    string
	Backend  string
	Effort   string
	MaxTurns int
	Tools    []string
	Timeout  time.Duration
	// PromptOverride, when set, is used verbatim instead of claude.BuildRunPrompt.
	PromptOverride string
	// Plan / Resume / SessionID are honoured by the cmux backend (the interactive
	// claude/codex TUI). Plan launches the agent in plan-only mode (permission mode
	// plan); Resume continues the todo's prior session; SessionID is the
	// pre-generated claude session id for a fresh run (the dashboard knows it up
	// front to follow the session log live). They are ignored by the headless SDK
	// backends, which carry no interactive session.
	Plan      bool
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
		config.Tools = []string{"Read", "Edit", "Write", "Bash", "Glob", "Grep"}
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
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
	prompt := e.config.PromptOverride
	if prompt == "" {
		prompt = claude.BuildRunPrompt(todosInGroup, workDir, e.config.Effort)
	}
	req, providerSessionID, canUseTool := e.buildRequest(ctx, todosInGroup, prompt, e.config.Resume)
	return e.runStream(ctx, start, req, canUseTool, providerSessionID, todosInGroup)
}

// SendFeedback resumes the group's prior agent session with feedback (the failing
// test/lint summary from the post-completion check loop) and waits for the next
// turn. It implements todos.FeedbackExecutor for every captain backend (cmux and
// the headless SDK), which all resume via req.SessionID. It requires a recorded
// prior session; without one it returns an error so the loop degrades to reporting
// the failures rather than iterating.
func (e *Executor) SendFeedback(ctx *todopkg.ExecutorContext, todosInGroup []*types.TODO, feedback string) (*todopkg.ExecutionResult, error) {
	start := time.Now()
	feedback = strings.TrimSpace(feedback)
	if feedback == "" {
		err := fmt.Errorf("%s: empty feedback", e.Name())
		return e.failed(start, err), err
	}
	if priorSessionID(todosInGroup) == "" {
		err := fmt.Errorf("%s: no prior session to resume for check feedback", e.Name())
		return e.failed(start, err), err
	}
	req, providerSessionID, canUseTool := e.buildRequest(ctx, todosInGroup, feedback, true)
	return e.runStream(ctx, start, req, canUseTool, providerSessionID, todosInGroup)
}

// buildRequest assembles the captain Request, the fresh provider session id (cmux
// only), and the tool-permission broker for a run. promptText is the verbatim
// prompt body; resume continues the group's prior session instead of starting a
// fresh one.
func (e *Executor) buildRequest(ctx *todopkg.ExecutorContext, todosInGroup []*types.TODO, promptText string, resume bool) (captainai.Request, string, captainai.PermissionFunc) {
	workDir := groupWorkDir(e.config.WorkDir, todosInGroup)
	req := captainai.Request{
		Prompt:   api.Prompt{User: promptText},
		Context:  api.Context{Dir: workDir},
		MaxTurns: e.config.MaxTurns,
	}
	// claude conveys effort through the prompt directive (claude.BuildRunPrompt);
	// codex takes a real reasoning-effort flag.
	if e.config.Agent == "codex" {
		req.Effort = api.Effort(e.config.Effort)
	}

	isCmux := e.isCmuxBackend()
	modes := e.config.ToolModes
	hasModes := len(modes) > 0
	// providerSessionID is the fresh claude session id handed to the cmux provider
	// (via Config.SessionID) so it launches `--session-id <id>` and the host can
	// follow the session log live. Empty for a resume (the provider takes the prior
	// id from req.SessionID) and for the non-cmux SDK backends.
	var providerSessionID string
	if isCmux {
		// cmux brokers tools through the terminal/approval round-trip. With no
		// per-tool preferences it carries no allowlist (every tool prompts); with
		// preferences, enabled→--allowedTools and disabled→--disallowedTools.
		req.Permissions.Mode = e.resolveMode(api.PermissionDefault)
		if hasModes {
			allow, deny := splitToolModes(modes)
			req.Permissions.Tools = api.Tools{Allow: allow, Deny: deny}
		}
		if prior := priorSessionID(todosInGroup); resume && prior != "" {
			req.SessionID = prior // provider resumes with --resume
		} else if e.config.Agent == "claude" {
			providerSessionID = e.config.SessionID
			if providerSessionID == "" {
				providerSessionID = uuid.NewString()
			}
			recordSessionID(todosInGroup, providerSessionID)
			ctx.RecordSessionID(providerSessionID)
		}
		// Stamp GAVEL_ISSUE_ID / GAVEL_SESSION_ID so a `gavel commit` the agent runs
		// itself writes the matching commit trailers.
		req.Context.Env = cmuxRunEnv(todosInGroup, firstNonEmpty(req.SessionID, providerSessionID))
	} else {
		req.Permissions = api.Permissions{
			Mode:    e.resolveMode(api.PermissionAcceptEdits), // acceptEdits so file edits are not blocked
			Presets: []api.Preset{api.PresetEdit},
			Tools:   api.Tools{Allow: e.config.Tools},
		}
		if hasModes {
			// Explicit per-tool preferences replace the edit preset's curated
			// allowlist: enabled→allow, disabled→deny, ask→brokered via CanUseTool.
			allow, deny := splitToolModes(modes)
			req.Permissions.Tools = api.Tools{Allow: allow, Deny: deny}
			req.Permissions.Presets = nil
		}
		// The headless SDK backends resume via req.SessionID (the claude-agent SDK
		// continues that session); used by the check-feedback loop.
		if prior := priorSessionID(todosInGroup); resume && prior != "" {
			req.SessionID = prior
		}
	}

	var canUseTool captainai.PermissionFunc
	if e.config.Approvals {
		canUseTool = e.buildCanUseTool(ctx, modes)
		// Allow-listed tools skip the can_use_tool callback, so for the SDK backends
		// with no explicit preferences drop Bash from the allowlist to route command
		// execution through approval. With explicit modes the allowlist already
		// reflects the user's enabled set; cmux carries no preset allowlist.
		if !isCmux && !hasModes {
			req.Permissions.Tools.Allow = withoutBash(e.config.Tools)
		}
	}
	return req, providerSessionID, canUseTool
}

// resolveMode picks the request's base permission posture: a plan run forces
// plan; otherwise an explicit user PermissionMode wins; otherwise the backend's
// default (acceptEdits for the SDK, default for cmux).
func (e *Executor) resolveMode(def api.PermissionMode) api.PermissionMode {
	if e.config.Plan {
		return api.PermissionPlan
	}
	if m := captainPermissionMode(e.config.PermissionMode); m != "" {
		return m
	}
	return def
}

// captainPermissionMode maps a clicky ClaudePermissionMode to captain's
// api.PermissionMode. captain has no "dontAsk"; it folds to the default posture
// (deny-by-default is then enforced per tool via the disabled modes). "" means no
// explicit mode was chosen.
func captainPermissionMode(s string) api.PermissionMode {
	switch s {
	case "plan":
		return api.PermissionPlan
	case "acceptEdits":
		return api.PermissionAcceptEdits
	case "auto":
		return api.PermissionAuto
	case "bypassPermissions":
		return api.PermissionBypass
	case "default", "dontAsk":
		return api.PermissionDefault
	default:
		return ""
	}
}

// splitToolModes partitions a tool-name→mode map into the allow list (enabled)
// and deny list (disabled); ask/unknown tools appear in neither and are brokered
// through CanUseTool.
func splitToolModes(modes map[string]string) (allow, deny []string) {
	for tool, mode := range modes {
		switch mode {
		case "enabled":
			allow = append(allow, tool)
		case "disabled":
			deny = append(deny, tool)
		}
	}
	sort.Strings(allow)
	sort.Strings(deny)
	return allow, deny
}

// runStream constructs the provider, streams the request's events into an
// ExecutionResult, and classifies the outcome (timeout / no-result / failure /
// success).
func (e *Executor) runStream(ctx *todopkg.ExecutorContext, start time.Time, req captainai.Request, canUseTool captainai.PermissionFunc, providerSessionID string, todosInGroup []*types.TODO) (*todopkg.ExecutionResult, error) {
	stream := e.config.Stream
	if stream == nil {
		provider, err := e.newStreamer(canUseTool, providerSessionID)
		if err != nil {
			return e.failed(start, err), err
		}
		// The provider already carries canUseTool on its Config, so the production
		// stream ignores the seam argument.
		stream = func(sctx context.Context, sreq captainai.Request, _ captainai.PermissionFunc) (<-chan captainai.Event, error) {
			return provider.ExecuteStream(sctx, sreq)
		}
	}

	ctx.Logger.Infof("%s: dispatching %d TODO(s) in %s", e.Name(), len(todosInGroup), req.Context.Dir)
	gavelai.NormalizeEnv()

	streamCtx, cancel := context.WithTimeout(ctx, e.config.Timeout)
	defer cancel()

	events, err := stream(streamCtx, req, canUseTool)
	if err != nil {
		return e.failed(start, err), err
	}

	result := &todopkg.ExecutionResult{ExecutorName: e.Name(), Transcript: ctx.GetTranscript()}
	var sawResult bool
	for ev := range events {
		e.handleEvent(ctx, ev, result, todosInGroup, &sawResult)
	}
	result.Duration = time.Since(start)

	switch {
	case !sawResult && streamCtx.Err() != nil:
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
		ctx.Logger.Infof("%s: completed", e.Name())
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
			result.TokensUsed = ev.Usage.TotalTokens()
		}
		result.CostUSD = ev.CostUSD
		if !ev.Success && ev.Error != "" {
			result.ErrorMessage = ev.Error
		}
	case captainai.EventError:
		result.ErrorMessage = ev.Error
		ctx.Notify(todopkg.Notification{Type: todopkg.NotifyError, Message: ev.Error})
	}
}

func (e *Executor) newStreamer(canUseTool captainai.PermissionFunc, sessionID string) (captainai.StreamingProvider, error) {
	model := strings.TrimSpace(e.config.Model)
	backend := strings.TrimSpace(e.config.Backend)
	// cmux backends route to the interactive-TUI provider regardless of agent; the
	// provider reads the agent from the backend and the fresh session id from
	// Config.SessionID (a resume takes the prior id from req.SessionID instead).
	if b := captainai.Backend(backend); b == captainai.BackendClaudeCmux || b == captainai.BackendCodexCmux {
		// NewProvider rejects an empty model name; a default cmux run carries none,
		// so fall back to the agent sentinel ("claude"/"codex"). The cmux provider's
		// modelFlag strips that sentinel back to "" (launch with no --model).
		name := model
		if name == "" {
			name = e.config.Agent
		}
		provider, err := captainai.NewProvider(captainai.Config{
			Model:      api.Model{Name: name, Backend: b},
			SessionID:  sessionID,
			CanUseTool: canUseTool,
		})
		if err != nil {
			return nil, err
		}
		streamer, ok := provider.(captainai.StreamingProvider)
		if !ok {
			return nil, fmt.Errorf("cmux backend %q does not support streaming", backend)
		}
		return streamer, nil
	}
	switch e.config.Agent {
	case "codex":
		if backend != "" && captainai.Backend(backend) != captainai.BackendCodexCLI {
			return nil, fmt.Errorf("headless codex does not support backend %q", backend)
		}
		if model == "codex" {
			model = ""
		}
		return captainprovider.NewCodexAppServer(model)
	case "claude":
		if backend == "" {
			backend = string(captainai.BackendClaudeAgent)
		}
		switch captainai.Backend(backend) {
		case captainai.BackendClaudeAgent, captainai.BackendClaudeCLI:
		default:
			return nil, fmt.Errorf("headless claude does not support backend %q", backend)
		}
		if model == "" || model == "claude" {
			model = "claude-agent-sonnet"
		}
		provider, err := captainai.NewProvider(captainai.Config{
			Model:      api.Model{Name: model, Backend: captainai.Backend(backend)},
			CanUseTool: canUseTool,
		})
		if err != nil {
			return nil, err
		}
		streamer, ok := provider.(captainai.StreamingProvider)
		if !ok {
			return nil, fmt.Errorf("headless claude backend %q does not support streaming", backend)
		}
		return streamer, nil
	default:
		return nil, fmt.Errorf("headless: unsupported agent %q", e.config.Agent)
	}
}

// buildCanUseTool returns the permission callback the captain provider invokes on
// a can_use_tool control request. It routes each request to the process-wide
// approval registry — the same one the cmux driver and the dashboard share — and
// maps the human's decision back onto the captain decision shape. It blocks until
// the dashboard resolves the request or the run's context is cancelled.
func (e *Executor) buildCanUseTool(ctx *todopkg.ExecutorContext, toolModes map[string]string) captainai.PermissionFunc {
	return func(callCtx context.Context, preq captainai.PermissionRequest) (captainai.PermissionDecision, error) {
		// A per-tool preference short-circuits the human round-trip: a disabled tool
		// is denied outright, an enabled one is auto-approved. Everything else
		// ("ask", or a tool with no recorded preference) is surfaced for approval.
		switch toolModes[preq.Tool] {
		case "disabled":
			ctx.Logger.Infof("%s: session %s denying disabled tool %s", e.Name(), preq.SessionID, preq.Tool)
			return captainai.PermissionDecision{Allow: false, Message: fmt.Sprintf("tool %s is disabled for this run", preq.Tool)}, nil
		case "enabled":
			return captainai.PermissionDecision{Allow: true}, nil
		}
		ctx.Logger.Infof("%s: session %s awaiting tool-permission approval: %s",
			e.Name(), preq.SessionID, preq.Tool)
		decision, err := todopkg.GlobalApprovals().Await(callCtx, todopkg.ApprovalRequest{
			SessionID: preq.SessionID,
			ToolUseID: preq.ToolUseID,
			Tool:      preq.Tool,
			Input:     preq.Input,
		})
		if err != nil {
			return captainai.PermissionDecision{}, err
		}
		return captainai.PermissionDecision{
			Allow:        decision.Allow,
			Message:      decision.Message,
			UpdatedInput: decision.UpdatedInput,
		}, nil
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

// recordSessionID stamps the agent's session id on each todo so the issue carries
// it and a later run can resume.
func recordSessionID(todoList []*types.TODO, sessionID string) {
	for _, t := range todoList {
		if t == nil {
			continue
		}
		if t.LLM == nil {
			t.LLM = &types.LLM{}
		}
		t.LLM.SessionId = sessionID
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

// priorSessionID returns the first recorded claude session id among the group's
// todos, or "" when none has run before — the id a resume reuses.
func priorSessionID(todoList []*types.TODO) string {
	for _, t := range todoList {
		if t != nil && t.LLM != nil && t.LLM.SessionId != "" {
			return t.LLM.SessionId
		}
	}
	return ""
}

// cmuxRunEnv builds the env the cmux provider exports to the agent process so a
// `gavel commit` the agent runs itself stamps the matching issue/session trailers.
func cmuxRunEnv(todoList []*types.TODO, sessionID string) map[string]string {
	env := map[string]string{}
	if id := joinIssueIDs(todoList); id != "" {
		env[commit.EnvIssueID] = id
	}
	if sessionID != "" {
		env[commit.EnvSessionID] = sessionID
	}
	if len(env) == 0 {
		return nil
	}
	return env
}

func joinIssueIDs(todoList []*types.TODO) string {
	var ids []string
	for _, t := range todoList {
		if t != nil && t.ID != "" {
			ids = append(ids, t.ID)
		}
	}
	return strings.Join(ids, ",")
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func toolSummary(ev captainai.Event) string {
	for _, key := range []string{"command", "file_path", "path", "pattern", "query"} {
		if v, ok := ev.Input[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return ev.Tool + ": " + truncate(s, 120)
			}
		}
	}
	return ev.Tool
}

// withoutBash returns tools without "Bash" so command execution is brokered via
// can_use_tool instead of auto-approved from the allowlist.
func withoutBash(tools []string) []string {
	out := make([]string, 0, len(tools))
	for _, t := range tools {
		if t == "Bash" {
			continue
		}
		out = append(out, t)
	}
	return out
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
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
