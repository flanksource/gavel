package headless

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	captainai "github.com/flanksource/captain/pkg/ai"
	captainprovider "github.com/flanksource/captain/pkg/ai/provider"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/commit"
	todopkg "github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// buildRequest adds only Gavel's runtime session, permission defaults, and
// metadata environment to the already-rendered canonical Spec.
func (e *Executor) buildRequest(ctx *todopkg.ExecutorContext, todosInGroup []*types.TODO, base captainai.Request, resume bool) (captainai.Request, string, captainai.PermissionFunc, error) {
	workDir := groupWorkDir(e.config.WorkDir, todosInGroup)
	req := base
	if req.Setup == nil {
		req.Setup = &shell.Setup{}
	}
	if req.Setup.Cwd == "" {
		req.Setup.Cwd = workDir
	}

	isCmux := e.isCmuxBackend()
	explicitPolicies := req.Permissions.Tools.Policies()
	req.Permissions = permissionDefaults(req.Permissions, isCmux)
	if e.config.Approvals && !isCmux && len(explicitPolicies) == 0 {
		req.Permissions.Tools = api.ToolsFromLists(withoutBash(todopkg.DefaultAgentTools()), nil)
		req.Permissions.Tools["Bash"] = api.ToolPolicyAsk
	}

	// providerSessionID is the fresh claude session id handed to the cmux provider
	// (via Config.SessionID) so it launches `--session-id <id>` and the host can
	// follow the session log live. Empty for a resume (the provider takes the prior
	// id from req.SessionID) and for the non-cmux SDK backends.
	var providerSessionID string
	requestedSessionID := req.SessionID
	req.SessionID = ""
	if resume {
		req.SessionID = firstNonEmpty(priorSessionID(todosInGroup), requestedSessionID)
	}
	if isCmux {
		if !resume && e.agent == "claude" {
			providerSessionID = requestedSessionID
			if providerSessionID == "" {
				providerSessionID = uuid.NewString()
			}
			recordSessionID(todosInGroup, providerSessionID)
			ctx.RecordSessionID(providerSessionID)
		}
		// Stamp GAVEL_ISSUE_ID / GAVEL_SESSION_ID so a `gavel commit` the agent runs
		// itself writes the matching commit trailers.
		if envMap := cmuxRunEnv(todosInGroup, firstNonEmpty(req.SessionID, providerSessionID)); len(envMap) > 0 {
			env := append([]string(nil), req.Setup.Env...)
			for k, v := range envMap {
				env = append(env, k+"="+v)
			}
			req.Setup.Env = env
		}
	}

	var canUseTool captainai.PermissionFunc
	if e.config.Approvals {
		canUseTool = e.buildCanUseTool(ctx, req.Permissions.Tools.Policies())
	}
	return req, providerSessionID, canUseTool, nil
}

// permissionDefaults supplies a posture for a request that states none.
//
// Capability is all-or-nothing. Filling the fields in one at a time treats
// silence on presets and tools as "nothing was asked for", when a request that
// named a mode has already described how it wants to run — which is how a plan
// run, declaring `mode: plan` and nothing else, ended up carrying the edit
// preset and the full edit toolset. A request that states any part of its
// posture owns the capability half of it.
//
// The mode is the exception and is always filled in, because downstream an
// unset mode is not neutral: claudeagent reads "" as bypassPermissions whenever
// no approval broker is attached. Leaving it to the provider would trade a
// too-permissive toolset for a too-permissive mode.
func permissionDefaults(permissions api.Permissions, cmux bool) api.Permissions {
	stated := permissions.Mode != "" || len(permissions.Presets) > 0 || len(permissions.Tools.Policies()) > 0
	if permissions.Mode == "" {
		if cmux {
			permissions.Mode = api.PermissionDefault
		} else {
			permissions.Mode = api.PermissionAcceptEdits
		}
	}
	// cmux drives the agent's own TUI, which owns its tool posture; only the
	// permission mode crosses over.
	if stated || cmux {
		return permissions
	}
	permissions.Presets = []api.Preset{api.PresetEdit}
	permissions.Tools = api.ToolsFromLists(todopkg.DefaultAgentTools(), nil)
	return permissions
}

// provider returns the streaming provider for a run: the test seam when set,
// otherwise a captain provider built from the request's resolved model/backend.
func (e *Executor) provider(req captainai.Request, canUseTool captainai.PermissionFunc, providerSessionID string) (captainai.StreamingProvider, error) {
	if e.stream != nil {
		return seamProvider{fn: e.stream, canUseTool: canUseTool, model: req.Name, backend: req.Backend}, nil
	}
	return e.newStreamer(req, canUseTool, providerSessionID)
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
func (e *Executor) newStreamer(req captainai.Request, canUseTool captainai.PermissionFunc, sessionID string) (captainai.StreamingProvider, error) {
	model := req.Name
	backend := string(req.Backend)
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
			name = e.agent
		}
		runtimeModel := req.Model
		runtimeModel.Name = name
		runtimeModel.Backend = b
		provider, err := captainai.NewProvider(captainai.Config{
			Model:      runtimeModel,
			Budget:     req.Budget,
			NoCache:    req.NoCache,
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
	switch e.agent {
	case "codex":
		if backend != "" && captainai.Backend(backend) != captainai.BackendCodexAgent {
			return nil, fmt.Errorf("headless codex does not support backend %q", backend)
		}
		if model == "codex" {
			model = ""
		}
		runtimeModel := req.Model
		runtimeModel.Name = model
		runtimeModel.Backend = captainai.BackendCodexAgent
		return captainprovider.NewCodexAppServer(captainai.Config{
			Model:      runtimeModel,
			Budget:     req.Budget,
			NoCache:    req.NoCache,
			CanUseTool: canUseTool,
		})
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
		runtimeModel := req.Model
		runtimeModel.Name = model
		runtimeModel.Backend = captainai.Backend(backend)
		provider, err := captainai.NewProvider(captainai.Config{
			Model:      runtimeModel,
			Budget:     req.Budget,
			NoCache:    req.NoCache,
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
		return nil, fmt.Errorf("headless: unsupported agent %q", e.agent)
	}
}

// buildCanUseTool returns the permission callback the captain provider invokes on
// a can_use_tool control request. It routes each request to the process-wide
// approval registry — the same one the dashboard shares — and maps the human's
// decision back onto the captain decision shape. It blocks until the dashboard
// resolves the request or the run's context is cancelled.
func (e *Executor) buildCanUseTool(ctx *todopkg.ExecutorContext, policies map[string]api.ToolPolicy) captainai.PermissionFunc {
	return func(callCtx context.Context, preq captainai.PermissionRequest) (captainai.PermissionDecision, error) {
		switch policies[preq.Tool] {
		case api.ToolPolicyDeny:
			ctx.Logger.Infof("%s: session %s denying disabled tool %s", e.Name(), preq.SessionID, preq.Tool)
			return captainai.PermissionDecision{Allow: false, Message: fmt.Sprintf("tool %s is disabled for this run", preq.Tool)}, nil
		case api.ToolPolicyAllow, api.ToolPolicyAuto:
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
