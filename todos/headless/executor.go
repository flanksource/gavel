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
	"strings"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	captainprovider "github.com/flanksource/captain/pkg/ai/provider"
	captainclaudeagent "github.com/flanksource/captain/pkg/ai/provider/claudeagent"
	gavelai "github.com/flanksource/gavel/ai"
	todopkg "github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/cmux"
	"github.com/flanksource/gavel/todos/types"
)

const defaultTimeout = 30 * time.Minute

// streamFunc opens a captain event stream for a request. It is the seam tests
// inject a fake stream through; production builds it from the agent + model.
type streamFunc func(ctx context.Context, req captainai.Request) (<-chan captainai.Event, error)

type Config struct {
	WorkDir  string
	Agent    string // "claude" or "codex"
	Model    string
	Effort   string
	MaxTurns int
	Tools    []string
	Timeout  time.Duration
	// PromptOverride, when set, is used verbatim instead of cmux.BuildPrompt.
	PromptOverride string
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

func (e *Executor) Name() string { return "headless-" + e.config.Agent }

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
		prompt = cmux.BuildPrompt(todosInGroup, workDir, e.config.Effort)
	}

	req := captainai.Request{
		Prompt:         prompt,
		Cwd:            workDir,
		Verbose:        true, // required for claude stream-json
		Edit:           true, // acceptEdits so file edits are not blocked
		AllowedTools:   e.config.Tools,
		PermissionMode: "acceptEdits",
		MaxTurns:       e.config.MaxTurns,
	}
	// claude conveys effort through the prompt directive (cmux.BuildPrompt);
	// codex takes a real reasoning-effort flag.
	if e.config.Agent == "codex" {
		req.ReasoningEffort = e.config.Effort
	}
	if e.config.Approvals {
		req.CanUseTool = e.buildCanUseTool(ctx)
		// Allow-listed tools skip the can_use_tool callback, so drop Bash from the
		// allowlist to route command execution through approval. acceptEdits still
		// auto-approves file edits and the read/search tools stay allow-listed.
		req.AllowedTools = withoutBash(e.config.Tools)
	}

	stream := e.config.Stream
	if stream == nil {
		provider, err := e.newStreamer()
		if err != nil {
			return e.failed(start, err), err
		}
		stream = provider.ExecuteStream
	}

	ctx.Logger.Infof("%s: dispatching %d TODO(s) in %s", e.Name(), len(todosInGroup), workDir)
	gavelai.NormalizeEnv()

	streamCtx, cancel := context.WithTimeout(ctx, e.config.Timeout)
	defer cancel()

	events, err := stream(streamCtx, req)
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

func (e *Executor) newStreamer() (captainai.StreamingProvider, error) {
	model := strings.TrimSpace(e.config.Model)
	switch e.config.Agent {
	case "codex":
		if model == "codex" {
			model = ""
		}
		return captainprovider.NewCodexAppServer(model)
	case "claude":
		if model == "" || model == "claude" {
			model = "sonnet"
		}
		return captainclaudeagent.New(captainai.Config{Model: model})
	default:
		return nil, fmt.Errorf("headless: unsupported agent %q", e.config.Agent)
	}
}

// buildCanUseTool returns the permission callback the captain provider invokes on
// a can_use_tool control request. It routes each request to the process-wide
// approval registry — the same one the cmux driver and the dashboard share — and
// maps the human's decision back onto the captain decision shape. It blocks until
// the dashboard resolves the request or the run's context is cancelled.
func (e *Executor) buildCanUseTool(ctx *todopkg.ExecutorContext) captainai.PermissionFunc {
	return func(callCtx context.Context, preq captainai.PermissionRequest) (captainai.PermissionDecision, error) {
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
	_ todopkg.Executor      = (*Executor)(nil)
	_ todopkg.GroupExecutor = (*Executor)(nil)
)
