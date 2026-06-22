package cmux

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	todopkg "github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/claude"
	"github.com/flanksource/gavel/todos/types"
)

const defaultExecutorTimeout = 30 * time.Minute

type CmuxExecutorConfig struct {
	WorkDir string
	Model   string
	Effort  string
	Timeout time.Duration
	Binary  string
	Runner  Runner
	Store   *SessionStore
}

type CmuxExecutor struct {
	config CmuxExecutorConfig
	client *Client
	store  *SessionStore
}

func NewCmuxExecutor(config CmuxExecutorConfig) *CmuxExecutor {
	client := NewClient(config.Binary)
	client.Runner = config.Runner
	store := config.Store
	if store == nil {
		store = DefaultSessionStore()
	}
	return &CmuxExecutor{
		config: config,
		client: client,
		store:  store,
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
	agentCommand := AgentCommand(agent, model)

	if err := e.client.Available(ctx); err != nil {
		return failedResult(e.Name(), start, err), err
	}

	ref, err := e.client.NewWorkspace(ctx, NewWorkspaceOpts{
		Cwd:     workDir,
		Name:    workspaceName(workDir),
		Command: agentCommand,
		Focus:   true,
	})
	if err != nil {
		return failedResult(e.Name(), start, err), err
	}

	if _, err := e.store.WaitForSession(ctx, agent, workDir, timeout); err != nil {
		return failedResult(e.Name(), start, err), err
	}

	prompt := BuildPrompt(todosInGroup, workDir, e.config.Effort)
	promptPath, err := WritePromptFile(workDir, todosInGroup, prompt)
	if err != nil {
		return failedResult(e.Name(), start, err), err
	}

	instruction := fmt.Sprintf("Read %s and implement all TODOs described there. When finished, stop and wait for verification.", promptPath)
	if err := e.client.Send(ctx, ref.String(), instruction+"\n"); err != nil {
		return failedResult(e.Name(), start, err), err
	}

	if _, err := e.store.WaitForIdle(ctx, agent, workDir, timeout); err != nil {
		return failedResult(e.Name(), start, err), err
	}

	return &todopkg.ExecutionResult{
		Success:      true,
		ExecutorName: e.Name(),
		Duration:     time.Since(start),
		ActionsPerformed: []string{
			"cmux new-workspace " + ref.String(),
			"cmux prompt " + promptPath,
		},
		Transcript: ctx.GetTranscript(),
	}, nil
}

func ResolveAgent(model string) (agent string, modelFlag string) {
	if model == "" {
		return "claude", ""
	}
	lower := strings.ToLower(model)
	if lower == "codex" || strings.HasPrefix(lower, "gpt-") || strings.HasPrefix(lower, "codex-") {
		return "codex", model
	}
	if lower == "claude" {
		return "claude", ""
	}
	return "claude", model
}

func AgentCommand(agent, model string) string {
	cmd := fmt.Sprintf("cmux %s-teams", agent)
	if model != "" {
		cmd += " --model " + model
	}
	return cmd
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
