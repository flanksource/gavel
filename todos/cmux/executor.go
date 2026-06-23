package cmux

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	todopkg "github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/claude"
	"github.com/flanksource/gavel/todos/types"
)

const defaultExecutorTimeout = 30 * time.Minute
const defaultSendAttempts = 3
const defaultSendRetryDelay = 2 * time.Second
const defaultScreenPollInterval = time.Second
const defaultScreenMaxPollInterval = 5 * time.Second
const defaultScreenStableDuration = 2 * time.Second
const defaultScreenLines = 120

type CmuxExecutorConfig struct {
	WorkDir string
	Model   string
	Effort  string
	// Plan launches the agent in plan-only mode: it investigates and proposes
	// an implementation plan without changing code (claude uses
	// --permission-mode plan; both agents are told to plan, not implement).
	Plan bool
	// Resume continues a todo's prior claude session (claude --resume <id>)
	// instead of starting a fresh one, so the agent keeps the earlier
	// conversation's context. When no prior session id is recorded it has no
	// effect and a new session starts. codex manages its own sessions and
	// ignores this. See ExecuteGroup.
	Resume bool
	// SessionID, when set, is the claude session id used for a fresh run instead
	// of a generated one. Callers set it so they know the id up front (e.g. the
	// dashboard returns it to follow the session log live). Ignored when
	// resuming, which reuses the prior id. Empty means generate one.
	SessionID             string
	Timeout               time.Duration
	Binary                string
	Runner                Runner
	SendAttempts          int
	SendRetryDelay        time.Duration
	ScreenPollInterval    time.Duration
	ScreenMaxPollInterval time.Duration
	ScreenStableDuration  time.Duration
	ScreenLines           int

	SessionLogPollInterval  time.Duration
	SessionLogAppearTimeout time.Duration
	SessionLogQuiescePeriod time.Duration
}

type CmuxExecutor struct {
	config CmuxExecutorConfig
	client *Client
}

func NewCmuxExecutor(config CmuxExecutorConfig) *CmuxExecutor {
	client := NewClient(config.Binary)
	client.Runner = config.Runner
	return &CmuxExecutor{
		config: config,
		client: client,
	}
}

func (e *CmuxExecutor) Name() string {
	agent, _ := ResolveAgent(e.config.Model)
	return "cmux-" + agent
}

func (e *CmuxExecutor) Execute(ctx *todopkg.ExecutorContext, todo *types.TODO) (*todopkg.ExecutionResult, error) {
	return e.ExecuteGroup(ctx, []*types.TODO{todo})
}

func (e *CmuxExecutor) ExecuteGroup(ctx *todopkg.ExecutorContext, todosInGroup []*types.TODO) (*todopkg.ExecutionResult, error) {
	start := time.Now()
	if len(todosInGroup) == 0 {
		return nil, fmt.Errorf("no todos supplied")
	}

	timeout := e.timeout()
	agent, model := ResolveAgent(e.config.Model)
	workDir := groupWorkDir(e.config.WorkDir, todosInGroup)

	// Resolve the Claude session id. By default we pre-generate one so we can
	// launch with --session-id and tail the resulting session history log for
	// structured progress. With Resume set and a prior session recorded, we
	// instead reuse that id and launch with --resume so the agent keeps the
	// earlier conversation's context. codex manages its own sessions, so it
	// keeps the screen-idle detection path.
	sessionID := ""
	resume := false
	prior := priorSessionID(todosInGroup)
	if agent == "claude" {
		switch {
		case e.config.Resume && prior != "":
			sessionID = prior
			resume = true
		case e.config.SessionID != "":
			sessionID = e.config.SessionID
			recordSessionID(todosInGroup, sessionID)
		default:
			sessionID = uuid.NewString()
			recordSessionID(todosInGroup, sessionID)
		}
	}
	agentCommand := AgentCommand(AgentCommandOpts{Agent: agent, Model: model, SessionID: sessionID, Resume: resume, Plan: e.config.Plan})

	ctx.Logger.Infof("cmux: dispatching %d TODO(s) with %s in %s", len(todosInGroup), agent, workDir)
	ctx.Logger.V(1).Infof("cmux command: cmux ping")
	if err := e.client.Available(ctx); err != nil {
		return failedResult(e.Name(), start, err), err
	}

	name := AgentWorkspaceName(workDir, agent)
	ctx.Logger.Infof("cmux: ensuring workspace %q for %s", name, agent)
	ctx.Logger.V(1).Infof("cmux command: cmux list-workspaces --json")
	ctx.Logger.V(1).Infof("cmux command: cmux new-workspace --cwd %q --name %q --focus true --id-format both (if missing)", workDir, name)
	workspace, reused, err := e.client.EnsureWorkspace(ctx, EnsureWorkspaceOpts{
		Cwd:         workDir,
		Name:        name,
		Description: fmt.Sprintf("gavel todos %s workspace for %s", agent, workDir),
		Focus:       true,
	})
	if err != nil {
		return failedResult(e.Name(), start, err), err
	}
	if reused {
		ctx.Logger.Infof("cmux: reusing workspace %s", workspace.String())
	} else {
		ctx.Logger.Infof("cmux: created workspace %s", workspace.String())
	}
	ctx.Logger.V(2).Infof("cmux workspace ref: raw=%q workspace=%q surface=%q", workspace.Raw, workspace.WorkspaceID, workspace.SurfaceID)

	ctx.Logger.Infof("cmux: creating %s terminal surface in workspace %s", agent, workspace.String())
	ctx.Logger.V(1).Infof("cmux command: cmux new-surface --type terminal --workspace %q --working-directory %q --focus true", workspace.String(), workDir)
	ref, err := e.client.NewSurface(ctx, NewSurfaceOpts{
		WorkspaceRef: workspace.String(),
		Cwd:          workDir,
		SurfaceType:  "terminal",
		Focus:        true,
	})
	if err != nil {
		return failedResult(e.Name(), start, err), err
	}
	ctx.Logger.V(2).Infof("cmux surface ref: raw=%q workspace=%q surface=%q", ref.Raw, ref.WorkspaceID, ref.SurfaceID)

	ctx.Logger.Infof("cmux: waiting for terminal surface to stabilize before launching %s", agent)
	beforeAgentScreen, err := e.waitForScreenIdle(ctx, ref, "before agent launch", timeout, "", false)
	if err != nil {
		return failedResult(e.Name(), start, err), err
	}

	// Persist the session id before launching the agent so an interrupted run
	// still records (and can resume) this session rather than the prior one.
	ctx.RecordSessionID(sessionID)

	if err := e.sendSurfaceText(ctx, ref.String(), ref.SurfaceID, "agent command", agentCommand); err != nil {
		return failedResult(e.Name(), start, err), err
	}

	ctx.Logger.Infof("cmux: waiting for %s screen to stabilize after launch", agent)
	beforePromptScreen, err := e.waitForScreenIdle(ctx, ref, "after agent launch", timeout, beforeAgentScreen, true)
	if err != nil {
		return failedResult(e.Name(), start, err), err
	}

	prompt := buildSessionPrompt(todosInGroup, workDir, e.config.Effort, resume, agent, prior)
	promptPath, err := WritePromptFile(workDir, todosInGroup, prompt)
	if err != nil {
		return failedResult(e.Name(), start, err), err
	}
	ctx.Logger.Infof("cmux: wrote initial prompt to %s", promptPath)
	ctx.Logger.V(2).Infof("cmux prompt body:\n%s", prompt)

	instruction := buildInstruction(todosInGroup, prompt, promptPath, e.config.Plan)
	if err := e.sendSurfaceText(ctx, ref.String(), ref.SurfaceID, "initial prompt", instruction); err != nil {
		return failedResult(e.Name(), start, err), err
	}

	actions := []string{
		"cmux workspace " + workspace.String(),
		"cmux new-surface " + ref.SurfaceID,
		"cmux agent " + agentCommand,
		"cmux prompt " + promptPath,
	}

	if sessionID != "" {
		// Register the run as a live in-progress session so the dashboard timer
		// reads token/cost totals straight from the tailer instead of re-reading
		// the growing log on every poll. Finish freezes the elapsed clock.
		acc := GlobalSessionStats().Begin(sessionID, agent, model, e.config.Effort, start)
		logPath, completed, serr := e.awaitSessionCompletion(ctx, sessionID, workDir, timeout, resume, acc)
		acc.Finish()
		if logPath != "" {
			actions = append(actions, "claude session "+logPath)
		}
		switch {
		case errors.Is(serr, errSessionLogNotFound):
			// A pre-generated claude session must produce its log; if it never
			// appears we fail the run loudly rather than inferring completion from
			// the terminal screen, which silently masks a broken agent launch.
			err := fmt.Errorf("claude session %s log %s did not appear within %s", sessionID, logPath, e.sessionLogAppearTimeout(timeout))
			return failedResult(e.Name(), start, err), err
		case serr != nil:
			return failedResult(e.Name(), start, serr), serr
		case !completed:
			err := fmt.Errorf("claude session %s did not complete within %s", sessionID, timeout)
			return failedResult(e.Name(), start, err), err
		default:
			ctx.Logger.Infof("cmux: claude session %s completed", sessionID)
		}
	} else {
		ctx.Logger.Infof("cmux: waiting for %s screen to change and stabilize after prompt dispatch", agent)
		if _, err := e.waitForScreenIdle(ctx, ref, "after prompt dispatch", timeout, beforePromptScreen, true); err != nil {
			return failedResult(e.Name(), start, err), err
		}
	}

	return &todopkg.ExecutionResult{
		Success:          true,
		ExecutorName:     e.Name(),
		Duration:         time.Since(start),
		ActionsPerformed: actions,
		Transcript:       ctx.GetTranscript(),
	}, nil
}

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
// todos, or "" when none has run before. It is the id resume reuses (and the one
// referenced in a fresh session's prompt for history).
func priorSessionID(todoList []*types.TODO) string {
	for _, t := range todoList {
		if t != nil && t.LLM != nil && t.LLM.SessionId != "" {
			return t.LLM.SessionId
		}
	}
	return ""
}

func (e *CmuxExecutor) sendSurfaceText(ctx *todopkg.ExecutorContext, workspaceRef, surfaceRef, label, text string) error {
	attempts := e.sendAttempts()
	delay := e.sendRetryDelay()
	text = strings.TrimRight(text, "\r\n")
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			ctx.Logger.V(1).Infof("cmux: waiting %s before retrying %s send", delay, label)
			if err := sleepContext(ctx, delay); err != nil {
				return err
			}
		}
		ctx.Logger.Infof("cmux: sending %s to workspace %s surface %s (attempt %d/%d)", label, workspaceRef, surfaceRef, attempt, attempts)
		ctx.Logger.V(1).Infof("cmux command: cmux send --workspace %q --surface %q -- <%s>", workspaceRef, surfaceRef, label)
		ctx.Logger.V(1).Infof("cmux command: cmux send-key --workspace %q --surface %q Enter", workspaceRef, surfaceRef)
		ctx.Logger.V(2).Infof("cmux send payload:\n%s", text)
		if err := e.client.SendSurface(ctx, workspaceRef, surfaceRef, text); err != nil {
			lastErr = err
			if ctx.Err() != nil {
				return err
			}
			if attempt < attempts {
				ctx.Logger.Warnf("cmux: %s send attempt %d/%d failed: %v; retrying in %s", label, attempt, attempts, err, delay)
			} else {
				ctx.Logger.Warnf("cmux: %s send attempt %d/%d failed: %v", label, attempt, attempts, err)
			}
			continue
		}
		if err := e.client.SendKeySurface(ctx, workspaceRef, surfaceRef, "Enter"); err != nil {
			lastErr = err
			if ctx.Err() != nil {
				return err
			}
			if attempt < attempts {
				ctx.Logger.Warnf("cmux: %s enter attempt %d/%d failed: %v; retrying in %s", label, attempt, attempts, err, delay)
			} else {
				ctx.Logger.Warnf("cmux: %s enter attempt %d/%d failed: %v", label, attempt, attempts, err)
			}
			continue
		}
		ctx.Logger.Infof("cmux: sent %s to workspace %s surface %s", label, workspaceRef, surfaceRef)
		return nil
	}
	return fmt.Errorf("send cmux %s after %d attempts: %w", label, attempts, lastErr)
}

func (e *CmuxExecutor) waitForScreenIdle(ctx *todopkg.ExecutorContext, ref WorkspaceRef, phase string, timeout time.Duration, baseline string, requireChange bool) (string, error) {
	workspaceRef := strings.TrimSpace(ref.String())
	if workspaceRef == "" {
		return "", fmt.Errorf("cmux workspace reference is required for read-screen")
	}
	surfaceRef := strings.TrimSpace(ref.SurfaceID)
	if surfaceRef == "" {
		return "", fmt.Errorf("cmux surface reference is required for read-screen")
	}
	if timeout <= 0 {
		timeout = defaultExecutorTimeout
	}
	poll := e.screenPollInterval()
	maxPoll := e.screenMaxPollInterval()
	stableFor := e.screenStableDuration()
	lines := e.screenLines()
	ctx.Logger.V(1).Infof("cmux wait: read-screen workspace=%q surface=%q phase=%q lines=%d poll=%s max-poll=%s stable=%s timeout=%s", workspaceRef, surfaceRef, phase, lines, poll, maxPoll, stableFor, timeout)

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	waitStart := time.Now()

	var (
		lastScreen    string
		lastChange    time.Time
		lastErr       error
		sawScreen     bool
		changedEnough = !requireChange
	)

	for {
		now := time.Now()
		screen, err := e.client.ReadScreen(waitCtx, ReadScreenOpts{
			WorkspaceRef: workspaceRef,
			SurfaceRef:   surfaceRef,
			Lines:        lines,
		})
		if err != nil {
			lastErr = err
			ctx.Logger.V(1).Infof("cmux read-screen failed while waiting for %s: %v", phase, err)
		} else {
			normalized := normalizeScreen(screen)
			if normalized != "" {
				sawScreen = true
				if lastScreen == "" || normalized != lastScreen {
					lastScreen = normalized
					lastChange = now
					if !changedEnough && normalizeScreen(baseline) != normalized {
						changedEnough = true
						ctx.Logger.V(1).Infof("cmux screen changed for %s", phase)
					}
					ctx.Logger.V(2).Infof("cmux read-screen %s changed (%d bytes):\n%s", phase, len(normalized), screenSnippet(normalized))
				}
				if changedEnough && !lastChange.IsZero() && now.Sub(lastChange) >= stableFor {
					ctx.Logger.V(1).Infof("cmux screen stable for %s during %s", now.Sub(lastChange).Round(time.Millisecond), phase)
					return normalized, nil
				}
			}
		}

		select {
		case <-waitCtx.Done():
			if lastErr != nil && !sawScreen {
				return "", fmt.Errorf("timed out waiting for cmux screen during %s: %w", phase, lastErr)
			}
			if requireChange && !changedEnough {
				return "", fmt.Errorf("timed out waiting for cmux screen to change during %s", phase)
			}
			return "", fmt.Errorf("timed out waiting for cmux screen to stabilize during %s", phase)
		default:
		}

		delay := screenPollDelay(waitStart, poll, maxPoll)
		ctx.Logger.V(2).Infof("cmux read-screen next poll for %s in %s", phase, delay)
		if err := sleepContext(waitCtx, delay); err != nil {
			if lastErr != nil && !sawScreen {
				return "", fmt.Errorf("timed out waiting for cmux screen during %s: %w", phase, lastErr)
			}
			if requireChange && !changedEnough {
				return "", fmt.Errorf("timed out waiting for cmux screen to change during %s", phase)
			}
			return "", fmt.Errorf("timed out waiting for cmux screen to stabilize during %s", phase)
		}
	}
}

func ResolveAgent(model string) (agent string, modelFlag string) {
	if model == "" {
		return "claude", ""
	}
	lower := strings.ToLower(model)
	if lower == "codex" || strings.HasPrefix(lower, "gpt-") || strings.HasPrefix(lower, "codex-") {
		if lower == "codex" {
			return "codex", ""
		}
		return "codex", model
	}
	if lower == "claude" {
		return "claude", ""
	}
	return "claude", model
}

type AgentCommandOpts struct {
	Agent     string
	Model     string
	SessionID string
	// Resume reuses SessionID as an existing conversation (claude --resume)
	// rather than creating a new one (claude --session-id). Ignored when
	// SessionID is empty or for codex.
	Resume bool
	// Plan starts claude in plan-only mode (--permission-mode plan). codex has
	// no equivalent flag, so plan there is enforced by the prompt instruction.
	Plan bool
}

func AgentCommand(opts AgentCommandOpts) string {
	switch opts.Agent {
	case "codex":
		if opts.Model != "" {
			return "codex -m " + opts.Model
		}
		return "codex"
	default:
		cmd := "claude"
		if opts.SessionID != "" {
			if opts.Resume {
				cmd += " --resume " + opts.SessionID
			} else {
				cmd += " --session-id " + opts.SessionID
			}
		}
		if opts.Plan {
			cmd += " --permission-mode plan"
		}
		if opts.Model != "" {
			cmd += " --model " + opts.Model
		}
		return cmd
	}
}

func BuildPrompt(todoList []*types.TODO, workDir, effort string) string {
	prompt := claude.BuildGroupPrompt(todoList, workDir)
	if directive := EffortDirective(effort); directive != "" {
		return directive + "\n\n" + prompt
	}
	return prompt
}

// buildSessionPrompt assembles the prompt body for a run: the effort-prefixed
// group prompt, plus a prior-session history note when a fresh claude session is
// started over a todo that already recorded one. Execute and PreviewInstruction
// share it so the dashboard preview matches what is actually sent.
func buildSessionPrompt(todoList []*types.TODO, workDir, effort string, resume bool, agent, prior string) string {
	prompt := BuildPrompt(todoList, workDir, effort)
	// When starting a fresh session despite a prior one existing, hand the agent
	// the previous session id so it can look up that history if it needs context
	// (the transcript lives in the session log, not the issue).
	if !resume && agent == "claude" && prior != "" {
		prompt = SessionHistoryDirective(prior) + "\n\n" + prompt
	}
	return prompt
}

// PreviewInstruction renders the exact text a cmux run would dispatch to the
// agent surface, without launching anything, so the dashboard's advanced run
// dialog can show the prompt before the run starts. It mirrors Execute's prompt
// assembly: the effort directive, an optional prior-session history note, the
// run's title header, and the implement/plan suffix.
func PreviewInstruction(todoList []*types.TODO, workDir, effort string, plan, resume bool, agent string) string {
	workDir = groupWorkDir(workDir, todoList)
	prompt := buildSessionPrompt(todoList, workDir, effort, resume, agent, priorSessionID(todoList))
	promptPath := filepath.Join(workDir, ".gavel", "cmux", promptFileName(todoList))
	return buildInstruction(todoList, prompt, promptPath, plan)
}

// SessionHistoryDirective tells a fresh agent session about the prior session
// that worked on the same todo(s), so it can read that history if it needs the
// earlier context. Returns "" for an empty id.
func SessionHistoryDirective(priorSessionID string) string {
	if priorSessionID == "" {
		return ""
	}
	return fmt.Sprintf("A previous agent session (id `%s`) already worked on this. Its transcript is available in the Claude session log; consult it for prior context before starting.", priorSessionID)
}

func EffortDirective(effort string) string {
	switch effort {
	case "low":
		return "Be concise."
	case "medium", "":
		return "Think carefully before implementing."
	case "high":
		return "Think hard and reason thoroughly; consider edge cases before implementing."
	default:
		return ""
	}
}

func WritePromptFile(workDir string, todoList []*types.TODO, prompt string) (string, error) {
	if workDir == "" {
		return "", fmt.Errorf("workDir is required")
	}
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(absWorkDir, ".gavel", "cmux")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, promptFileName(todoList))
	if err := os.WriteFile(path, []byte(strings.TrimSpace(prompt)+"\n"), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// maxInlinePromptBytes bounds how much of the prompt is dispatched directly to
// the agent surface. Prompts at or below this are inlined in full; larger ones
// are truncated to this size with a pointer to the prompt file for the rest.
const maxInlinePromptBytes = 10 * 1024

// buildInstruction renders the message dispatched to the agent surface. It leads
// with the todo title so the run is self-describing, inlines the prompt body, and
// only references the prompt file when the prompt is too large to send in full.
func buildInstruction(todoList []*types.TODO, prompt, promptPath string, plan bool) string {
	var b strings.Builder
	if title := promptTitle(todoList); title != "" {
		b.WriteString("# " + title + "\n\n")
	}
	body, truncated := truncatePrompt(prompt, maxInlinePromptBytes)
	b.WriteString(body)
	if truncated {
		fmt.Fprintf(&b, "\n\n... (prompt truncated — read %s for the full prompt)", promptPath)
	}
	if plan {
		b.WriteString("\n\nProduce a detailed implementation plan for all TODOs above. Investigate the codebase to understand each task, but do NOT make any code changes — only plan. When finished, present the plan and stop.")
	} else {
		b.WriteString("\n\nImplement all TODOs described above. When finished, stop and wait for verification.")
	}
	return b.String()
}

// promptTitle is the H1 at the top of the dispatched prompt. A single run uses
// the todo's title (falling back to its first path entry); a multi-todo run uses
// "N Todo Items" so the header stays readable instead of concatenating every
// title. Returns "" when no title can be derived.
func promptTitle(todoList []*types.TODO) string {
	if len(todoList) > 1 {
		return fmt.Sprintf("%d Todo Items", len(todoList))
	}
	for _, todo := range todoList {
		if todo == nil {
			continue
		}
		title := strings.TrimSpace(todo.Title)
		if title == "" && len(todo.Path) > 0 {
			title = strings.TrimSpace(todo.Path[0])
		}
		if title != "" {
			return title
		}
	}
	return ""
}

// truncatePrompt clamps prompt to max bytes, cutting on the last line boundary so
// the inlined body never ends mid-line. The bool reports whether truncation happened.
func truncatePrompt(prompt string, max int) (string, bool) {
	prompt = strings.TrimSpace(prompt)
	if len(prompt) <= max {
		return prompt, false
	}
	clipped := prompt[:max]
	if idx := strings.LastIndexByte(clipped, '\n'); idx > 0 {
		clipped = clipped[:idx]
	}
	return strings.TrimRight(clipped, "\n"), true
}

func (e *CmuxExecutor) timeout() time.Duration {
	if e.config.Timeout > 0 {
		return e.config.Timeout
	}
	return defaultExecutorTimeout
}

func (e *CmuxExecutor) sendAttempts() int {
	if e.config.SendAttempts > 0 {
		return e.config.SendAttempts
	}
	return defaultSendAttempts
}

func (e *CmuxExecutor) sendRetryDelay() time.Duration {
	if e.config.SendRetryDelay > 0 {
		return e.config.SendRetryDelay
	}
	return defaultSendRetryDelay
}

func (e *CmuxExecutor) screenPollInterval() time.Duration {
	if e.config.ScreenPollInterval > 0 {
		return e.config.ScreenPollInterval
	}
	return defaultScreenPollInterval
}

func (e *CmuxExecutor) screenMaxPollInterval() time.Duration {
	if e.config.ScreenMaxPollInterval > 0 {
		return e.config.ScreenMaxPollInterval
	}
	return defaultScreenMaxPollInterval
}

func (e *CmuxExecutor) screenStableDuration() time.Duration {
	if e.config.ScreenStableDuration > 0 {
		return e.config.ScreenStableDuration
	}
	return defaultScreenStableDuration
}

func (e *CmuxExecutor) screenLines() int {
	if e.config.ScreenLines > 0 {
		return e.config.ScreenLines
	}
	return defaultScreenLines
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func screenPollDelay(start time.Time, base, max time.Duration) time.Duration {
	if base <= 0 {
		base = defaultScreenPollInterval
	}
	if max <= 0 {
		max = defaultScreenMaxPollInterval
	}
	if max < base {
		max = base
	}
	elapsed := time.Since(start)
	steps := int(elapsed / (10 * time.Second))
	delay := base
	for i := 0; i < steps && delay < max; i++ {
		delay *= 2
		if delay > max {
			return max
		}
	}
	return delay
}

func normalizeScreen(screen string) string {
	return strings.TrimSpace(strings.ReplaceAll(screen, "\r\n", "\n"))
}

func screenSnippet(screen string) string {
	const max = 2000
	screen = normalizeScreen(screen)
	if len(screen) <= max {
		return screen
	}
	return screen[:max] + "\n... (truncated)"
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

func workspaceName(workDir string) string {
	name := filepath.Base(filepath.Clean(workDir))
	if name == "." || name == string(filepath.Separator) {
		return "gavel-todos"
	}
	return name
}

func AgentWorkspaceName(workDir, agent string) string {
	name := workspaceName(workDir)
	agent = strings.TrimSpace(agent)
	if agent == "" {
		return name
	}
	name = sanitizeName(name + "-" + agent)
	if name == "" {
		return "gavel-todos-" + sanitizeName(agent)
	}
	return name
}

func promptFileName(todoList []*types.TODO) string {
	name := "group"
	if len(todoList) == 1 && todoList[0] != nil {
		name = todoList[0].DisplayID()
		if name == "" {
			name = todoList[0].Filename()
		}
	} else if len(todoList) > 0 && todoList[0] != nil {
		name = todoList[0].DisplayID()
		if name == "" {
			name = todoList[0].Filename()
		}
		name += "-group"
	}
	name = sanitizeName(name)
	if name == "" {
		name = "group"
	}
	return "prompt-" + name + ".md"
}

var unsafePromptName = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func sanitizeName(name string) string {
	name = unsafePromptName.ReplaceAllString(strings.TrimSpace(name), "-")
	name = strings.Trim(name, "-._")
	if len(name) > 64 {
		name = name[:64]
	}
	return name
}

func failedResult(name string, start time.Time, err error) *todopkg.ExecutionResult {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return &todopkg.ExecutionResult{
		Success:      false,
		ExecutorName: name,
		Duration:     time.Since(start),
		ErrorMessage: msg,
	}
}

var _ todopkg.Executor = (*CmuxExecutor)(nil)
var _ todopkg.GroupExecutor = (*CmuxExecutor)(nil)
