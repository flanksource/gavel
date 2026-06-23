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
	WorkDir               string
	Model                 string
	Effort                string
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

	// Pre-generate the Claude session id so we can launch with --session-id and
	// tail the resulting session history log for structured progress. codex
	// manages its own sessions, so it keeps the screen-idle detection path.
	sessionID := ""
	if agent == "claude" {
		sessionID = uuid.NewString()
		recordSessionID(todosInGroup, sessionID)
	}
	agentCommand := AgentCommand(agent, model, sessionID)

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

	if err := e.sendSurfaceText(ctx, ref.String(), ref.SurfaceID, "agent command", agentCommand); err != nil {
		return failedResult(e.Name(), start, err), err
	}

	ctx.Logger.Infof("cmux: waiting for %s screen to stabilize after launch", agent)
	beforePromptScreen, err := e.waitForScreenIdle(ctx, ref, "after agent launch", timeout, beforeAgentScreen, true)
	if err != nil {
		return failedResult(e.Name(), start, err), err
	}

	prompt := BuildPrompt(todosInGroup, workDir, e.config.Effort)
	promptPath, err := WritePromptFile(workDir, todosInGroup, prompt)
	if err != nil {
		return failedResult(e.Name(), start, err), err
	}
	ctx.Logger.Infof("cmux: wrote initial prompt to %s", promptPath)
	ctx.Logger.V(2).Infof("cmux prompt body:\n%s", prompt)

	instruction := fmt.Sprintf("Read %s and implement all TODOs described there. When finished, stop and wait for verification.", promptPath)
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
		logPath, completed, serr := e.awaitSessionCompletion(ctx, sessionID, workDir, timeout)
		if logPath != "" {
			actions = append(actions, "claude session "+logPath)
		}
		switch {
		case errors.Is(serr, errSessionLogNotFound):
			ctx.Logger.Warnf("cmux: claude session log never appeared; falling back to screen-idle detection")
			if _, err := e.waitForScreenIdle(ctx, ref, "after prompt dispatch", timeout, beforePromptScreen, true); err != nil {
				return failedResult(e.Name(), start, err), err
			}
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

func AgentCommand(agent, model, sessionID string) string {
	switch agent {
	case "codex":
		if model != "" {
			return "codex -m " + model
		}
		return "codex"
	default:
		cmd := "claude"
		if sessionID != "" {
			cmd += " --session-id " + sessionID
		}
		if model != "" {
			cmd += " --model " + model
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
