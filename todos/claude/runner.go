package claude

import (
	"bufio"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

//go:embed agent.ts
var agentTS string

//go:embed package.json
var agentPackageJSON string

type ClaudeExecutorConfig struct {
	WorkDir      string
	SessionID    string
	MaxBudgetUsd float64
	MaxTurns     int
	Model        string
	SystemPrompt string
	Tools        []string
	Timeout      time.Duration
	Dirty        bool
}

type ClaudeExecutor struct {
	config ClaudeExecutorConfig
}

func NewClaudeExecutor(config ClaudeExecutorConfig) *ClaudeExecutor {
	if len(config.Tools) == 0 {
		config.Tools = []string{"Read", "Edit", "Write", "Bash", "Glob", "Grep"}
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Minute
	}
	return &ClaudeExecutor{config: config}
}

func (e *ClaudeExecutor) Name() string { return "claude-code" }

func (e *ClaudeExecutor) Execute(ctx *todos.ExecutorContext, todo *types.TODO) (*todos.ExecutionResult, error) {
	result := &todos.ExecutionResult{
		ExecutorName: e.Name(),
		Transcript:   ctx.GetTranscript(),
	}
	startTime := time.Now()

	// Checkout target branch if specified (uses --autostash internally)
	if todo.Branch != "" {
		restoreBranch, err := gitCheckoutBranch(e.config.WorkDir, todo.Branch)
		if err != nil {
			return result, fmt.Errorf("failed to checkout branch %s: %w", todo.Branch, err)
		}
		defer restoreBranch()
	}

	// Stash dirty working tree before Claude runs
	restore, err := gitStash(e.config.WorkDir, e.config.Dirty)
	if err != nil {
		return result, fmt.Errorf("failed to stash working tree: %w", err)
	}
	defer restore()

	ctx.Notify(todos.Notification{
		Type:    todos.NotifyProgress,
		Message: fmt.Sprintf("Starting %s session", e.Name()),
	})

	agentDir, err := prepareAgentDir()
	if err != nil {
		return result, fmt.Errorf("failed to prepare agent dir: %w", err)
	}

	if err := ensureDependencies(agentDir); err != nil {
		return result, fmt.Errorf("failed to ensure dependencies: %w", err)
	}

	prompt := BuildPrompt(todo, e.config.WorkDir)

	if err := e.runAgent(ctx, agentDir, prompt, todo, result); err != nil {
		result.Duration = time.Since(startTime)
		result.ErrorMessage = err.Error()
		return result, err
	}

	// Commit changes after successful execution
	if result.Success {
		sha, commitErr := gitCommitChanges(ctx.Context, e.config.WorkDir, todo)
		if commitErr != nil {
			ctx.Logger.Warnf("Failed to commit changes: %v", commitErr)
		} else {
			result.CommitSHA = sha
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

func (e *ClaudeExecutor) ExecuteGroup(ctx *todos.ExecutorContext, todosInGroup []*types.TODO) (*todos.ExecutionResult, error) {
	if len(todosInGroup) == 0 {
		return nil, fmt.Errorf("no TODOs in group")
	}

	// Validate all TODOs share the same branch
	branch := todosInGroup[0].Branch
	for _, t := range todosInGroup[1:] {
		if t.Branch != branch {
			return nil, fmt.Errorf("mixed branches in group: %q vs %q", branch, t.Branch)
		}
	}

	result := &todos.ExecutionResult{
		ExecutorName: e.Name(),
		Transcript:   ctx.GetTranscript(),
	}
	startTime := time.Now()

	if branch != "" {
		restoreBranch, err := gitCheckoutBranch(e.config.WorkDir, branch)
		if err != nil {
			return result, fmt.Errorf("failed to checkout branch %s: %w", branch, err)
		}
		defer restoreBranch()
	}

	restore, err := gitStash(e.config.WorkDir, e.config.Dirty)
	if err != nil {
		return result, fmt.Errorf("failed to stash working tree: %w", err)
	}
	defer restore()

	ctx.Notify(todos.Notification{
		Type:    todos.NotifyProgress,
		Message: fmt.Sprintf("Starting %s group session (%d TODOs)", e.Name(), len(todosInGroup)),
	})

	agentDir, err := prepareAgentDir()
	if err != nil {
		return result, fmt.Errorf("failed to prepare agent dir: %w", err)
	}
	if err := ensureDependencies(agentDir); err != nil {
		return result, fmt.Errorf("failed to ensure dependencies: %w", err)
	}

	prompt := BuildGroupPrompt(todosInGroup, e.config.WorkDir)
	if err := e.runAgent(ctx, agentDir, prompt, todosInGroup[0], result); err != nil {
		result.Duration = time.Since(startTime)
		result.ErrorMessage = err.Error()
		return result, err
	}

	// Store session ID on all TODOs
	for _, t := range todosInGroup {
		if t.LLM == nil {
			t.LLM = &types.LLM{}
		}
	}

	if result.Success {
		sha, commitErr := gitCommitGroupChanges(ctx.Context, e.config.WorkDir, todosInGroup)
		if commitErr != nil {
			ctx.Logger.Warnf("Failed to commit group changes: %v", commitErr)
		} else {
			result.CommitSHA = sha
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

func (e *ClaudeExecutor) runAgent(ctx *todos.ExecutorContext, agentDir, prompt string, todo *types.TODO, result *todos.ExecutionResult) error {
	promptFile, err := os.CreateTemp("", "gavel-prompt-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create prompt file: %w", err)
	}
	defer os.Remove(promptFile.Name())

	if _, err := promptFile.WriteString(prompt); err != nil {
		promptFile.Close()
		return fmt.Errorf("failed to write prompt: %w", err)
	}
	promptFile.Close()

	config := e.buildAgentConfig(todo)
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	tsxPath, err := findTsx(agentDir)
	if err != nil {
		return fmt.Errorf("tsx not found: %w", err)
	}

	agentTSPath := filepath.Join(agentDir, "agent.ts")
	cmdCtx, cancel := context.WithDeadline(ctx.Context, time.Now().Add(e.config.Timeout))
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, tsxPath, agentTSPath, promptFile.Name())
	cmd.Dir = agentDir
	cmd.Env = append(filterEnv(os.Environ()), "AGENT_CONFIG="+string(configJSON))

	// Create pipes manually so we control close behavior.
	// cmd.StdoutPipe() leaves the parent holding write ends, which prevents
	// EOF when grandchild processes (tsx → node → claude) inherit them.
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		stdoutR.Close()
		stdoutW.Close()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	if err := cmd.Start(); err != nil {
		stdoutR.Close()
		stdoutW.Close()
		stderrR.Close()
		stderrW.Close()
		return fmt.Errorf("failed to start agent: %w", err)
	}

	// Close write ends in parent immediately — without this, reads never get EOF
	stdoutW.Close()
	stderrW.Close()

	// Close read ends on context cancellation to unblock scanners
	go func() {
		<-cmdCtx.Done()
		stdoutR.Close()
		stderrR.Close()
	}()

	// Capture stderr in background
	var stderrOutput string
	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		scanner := bufio.NewScanner(stderrR)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			stderrOutput += scanner.Text() + "\n"
		}
	}()

	// Stream JSONL from stdout
	var gotResult bool
	scanner := bufio.NewScanner(stdoutR)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		msg, err := ParseLine(scanner.Bytes())
		if err != nil {
			ctx.Logger.Debugf("Failed to parse line: %v", err)
			continue
		}
		if msg == nil {
			continue
		}
		ProcessMessage(ctx, msg, result)

		if msg.Type == "result" {
			gotResult = true
			if msg.SessionID != "" && todo.LLM != nil {
				todo.LLM.SessionId = msg.SessionID
			}
		}
	}

	<-stderrDone

	if err := cmd.Wait(); err != nil {
		if gotResult {
			ctx.Logger.Debugf("Agent process exited with %v (result already received)", err)
			return nil
		}
		if stderrOutput != "" {
			return fmt.Errorf("agent exited with error: %w\nstderr: %s", err, stderrOutput)
		}
		return fmt.Errorf("agent exited with error: %w", err)
	}

	return nil
}

type agentConfig struct {
	CWD          string   `json:"cwd,omitempty"`
	SessionID    string   `json:"session_id,omitempty"`
	MaxBudgetUsd float64  `json:"max_budget_usd,omitempty"`
	MaxTurns     int      `json:"max_turns,omitempty"`
	Model        string   `json:"model,omitempty"`
	SystemPrompt string   `json:"system_prompt,omitempty"`
	Tools        []string `json:"tools,omitempty"`
}

func (e *ClaudeExecutor) buildAgentConfig(todo *types.TODO) agentConfig {
	cfg := agentConfig{
		CWD:          e.config.WorkDir,
		SessionID:    e.config.SessionID,
		MaxBudgetUsd: e.config.MaxBudgetUsd,
		MaxTurns:     e.config.MaxTurns,
		Model:        e.config.Model,
		SystemPrompt: e.config.SystemPrompt,
		Tools:        e.config.Tools,
	}
	if todo.LLM != nil {
		if todo.LLM.SessionId != "" && cfg.SessionID == "" {
			cfg.SessionID = todo.LLM.SessionId
		}
		if todo.LLM.MaxCost > 0 && cfg.MaxBudgetUsd == 0 {
			cfg.MaxBudgetUsd = todo.LLM.MaxCost
		}
		if todo.LLM.MaxTurns > 0 && cfg.MaxTurns == 0 {
			cfg.MaxTurns = todo.LLM.MaxTurns
		}
		if todo.LLM.Model != "" && cfg.Model == "" {
			cfg.Model = todo.LLM.Model
		}
	}
	return cfg
}

func prepareAgentDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get cache dir: %w", err)
	}

	agentDir := filepath.Join(cacheDir, "gavel", "claude-agent")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create agent dir: %w", err)
	}

	if err := writeIfChanged(filepath.Join(agentDir, "agent.ts"), agentTS); err != nil {
		return "", err
	}
	if err := writeIfChanged(filepath.Join(agentDir, "package.json"), agentPackageJSON); err != nil {
		return "", err
	}

	return agentDir, nil
}

func writeIfChanged(path, content string) error {
	existing, err := os.ReadFile(path)
	if err == nil && contentHash(existing) == contentHash([]byte(content)) {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func contentHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func ensureDependencies(agentDir string) error {
	sdkDir := filepath.Join(agentDir, "node_modules", "@anthropic-ai", "claude-agent-sdk")
	if _, err := os.Stat(sdkDir); err == nil {
		return nil
	}

	npmPath, err := exec.LookPath("npm")
	if err != nil {
		return fmt.Errorf("npm not found in PATH: %w", err)
	}

	cmd := exec.Command(npmPath, "install")
	cmd.Dir = agentDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func filterEnv(env []string) []string {
	var filtered []string
	for _, e := range env {
		// Remove CLAUDECODE to avoid "nested session" errors from the SDK
		if len(e) >= 10 && e[:10] == "CLAUDECODE" {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered
}

func findTsx(agentDir string) (string, error) {
	// Try local node_modules/.bin/tsx first
	localTsx := filepath.Join(agentDir, "node_modules", ".bin", "tsx")
	if _, err := os.Stat(localTsx); err == nil {
		return localTsx, nil
	}

	// Fall back to global tsx
	if path, err := exec.LookPath("tsx"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("tsx not found; install with: npm install -g tsx")
}
