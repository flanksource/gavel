package headless

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"

	captainai "github.com/flanksource/captain/pkg/ai"
	captainprovider "github.com/flanksource/captain/pkg/ai/provider"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/commit"
	todopkg "github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// buildRequest augments the rendered template request with the run's
// budget/permissions/session plumbing. The template frontmatter's model and
// permission mode are the defaults; explicit user config overrides them. It
// returns the request, the fresh provider session id (cmux only), and the
// tool-permission broker.
func (e *Executor) buildRequest(ctx *todopkg.ExecutorContext, todosInGroup []*types.TODO, base captainai.Request, resume bool) (captainai.Request, string, captainai.PermissionFunc) {
	workDir := groupWorkDir(e.config.WorkDir, todosInGroup)
	req := base
	req.Budget = api.Budget{MaxTurns: e.config.MaxTurns}
	req.SetCwd(workDir)
	// claude conveys effort through the prompt directive (rendered by
	// todos/prompt); codex takes a real reasoning-effort flag.
	if e.config.Agent == "codex" {
		req.Effort = api.Effort(e.config.Effort)
	}
	// Explicit user model/backend overrides win over the template frontmatter.
	if m := strings.TrimSpace(e.config.Model); m != "" {
		req.Name = m
	}
	if b := strings.TrimSpace(e.config.Backend); b != "" {
		req.Backend = captainai.Backend(b)
	}

	templateMode := base.Permissions.Mode
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
		req.Permissions.Mode = e.resolveMode(templateMode, api.PermissionDefault)
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
		if envMap := cmuxRunEnv(todosInGroup, firstNonEmpty(req.SessionID, providerSessionID)); len(envMap) > 0 {
			env := make([]string, 0, len(envMap))
			for k, v := range envMap {
				env = append(env, k+"="+v)
			}
			req.Setup.Env = env
		}
	} else {
		req.Permissions = api.Permissions{
			Mode:    e.resolveMode(templateMode, api.PermissionAcceptEdits), // acceptEdits so file edits are not blocked
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
		// continues that session); used by the feedback/answer turns.
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

// resolveMode picks the request's base permission posture: an explicit user
// PermissionMode wins; then the template frontmatter's mode (the plan template
// declares plan); then the backend's default.
func (e *Executor) resolveMode(fromTemplate, def api.PermissionMode) api.PermissionMode {
	if m := captainPermissionMode(e.config.PermissionMode); m != "" {
		return m
	}
	if fromTemplate != "" {
		return fromTemplate
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

// provider returns the streaming provider for a run: the test seam when set,
// otherwise a captain provider built from the request's resolved model/backend.
func (e *Executor) provider(req captainai.Request, canUseTool captainai.PermissionFunc, providerSessionID string) (captainai.StreamingProvider, error) {
	if e.config.Stream != nil {
		return seamProvider{fn: e.config.Stream, canUseTool: canUseTool, model: req.Name, backend: req.Backend}, nil
	}
	return e.newStreamer(canUseTool, providerSessionID, req.Name, string(req.Backend))
}

// seamProvider adapts the streamFunc test seam to captain's StreamingProvider.
type seamProvider struct {
	fn         streamFunc
	canUseTool captainai.PermissionFunc
	model      string
	backend    captainai.Backend
}

func (p seamProvider) ExecuteStream(ctx context.Context, req captainai.Request) (<-chan captainai.Event, error) {
	return p.fn(ctx, req, p.canUseTool)
}

func (p seamProvider) Execute(ctx context.Context, req captainai.Request) (*captainai.Response, error) {
	events, err := p.ExecuteStream(ctx, req)
	if err != nil {
		return nil, err
	}
	var text strings.Builder
	for ev := range events {
		if ev.Kind == captainai.EventText {
			text.WriteString(ev.Text)
		}
	}
	return &captainai.Response{Text: text.String()}, nil
}

func (p seamProvider) GetModel() string              { return p.model }
func (p seamProvider) GetBackend() captainai.Backend { return p.backend }

// newStreamer constructs the real captain provider for the resolved
// model/backend. cmux backends route to the interactive-TUI provider; headless
// claude/codex use the SDK backends.
func (e *Executor) newStreamer(canUseTool captainai.PermissionFunc, sessionID, model, backend string) (captainai.StreamingProvider, error) {
	model = strings.TrimSpace(model)
	backend = strings.TrimSpace(backend)
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
		if backend != "" && captainai.Backend(backend) != captainai.BackendCodexAgent {
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
// approval registry — the same one the dashboard shares — and maps the human's
// decision back onto the captain decision shape. It blocks until the dashboard
// resolves the request or the run's context is cancelled.
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
