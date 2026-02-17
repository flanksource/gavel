package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareAgentDir(t *testing.T) {
	dir, err := prepareAgentDir()
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(dir, "agent.ts"))
	assert.FileExists(t, filepath.Join(dir, "package.json"))

	// Verify contents match embedded files
	ts, err := os.ReadFile(filepath.Join(dir, "agent.ts"))
	require.NoError(t, err)
	assert.Equal(t, agentTS, string(ts))

	pkg, err := os.ReadFile(filepath.Join(dir, "package.json"))
	require.NoError(t, err)
	assert.Equal(t, agentPackageJSON, string(pkg))
}

func TestWriteIfChanged_SkipsIdentical(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")

	require.NoError(t, writeIfChanged(path, "hello"))
	info1, _ := os.Stat(path)

	require.NoError(t, writeIfChanged(path, "hello"))
	info2, _ := os.Stat(path)

	assert.Equal(t, info1.ModTime(), info2.ModTime(), "file should not be rewritten when content is identical")
}

func TestWriteIfChanged_UpdatesOnChange(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")

	require.NoError(t, writeIfChanged(path, "v1"))

	require.NoError(t, writeIfChanged(path, "v2"))
	content, _ := os.ReadFile(path)
	assert.Equal(t, "v2", string(content))
}

func TestBuildAgentConfig(t *testing.T) {
	exec := NewClaudeExecutor(ClaudeExecutorConfig{
		WorkDir:      "/work",
		SessionID:    "cli-session",
		MaxBudgetUsd: 1.0,
		MaxTurns:     10,
		Model:        "opus",
		Tools:        []string{"Bash", "Read"},
	})

	todo := &types.TODO{}
	todo.LLM = &types.LLM{
		SessionId: "todo-session",
		MaxCost:   0.5,
		MaxTurns:  5,
		Model:     "sonnet",
	}

	cfg := exec.buildAgentConfig(todo)

	// CLI config takes precedence over TODO LLM config
	assert.Equal(t, "/work", cfg.CWD)
	assert.Equal(t, "cli-session", cfg.SessionID)
	assert.Equal(t, 1.0, cfg.MaxBudgetUsd)
	assert.Equal(t, 10, cfg.MaxTurns)
	assert.Equal(t, "opus", cfg.Model)
	assert.Equal(t, []string{"Bash", "Read"}, cfg.Tools)
}

func TestBuildAgentConfig_FallsBackToTODO(t *testing.T) {
	exec := NewClaudeExecutor(ClaudeExecutorConfig{
		WorkDir: "/work",
	})

	todo := &types.TODO{}
	todo.LLM = &types.LLM{
		SessionId: "todo-session",
		MaxCost:   0.5,
		MaxTurns:  5,
		Model:     "sonnet",
	}

	cfg := exec.buildAgentConfig(todo)

	assert.Equal(t, "todo-session", cfg.SessionID)
	assert.Equal(t, 0.5, cfg.MaxBudgetUsd)
	assert.Equal(t, 5, cfg.MaxTurns)
	assert.Equal(t, "sonnet", cfg.Model)
}

func TestBuildAgentConfig_JSON(t *testing.T) {
	exec := NewClaudeExecutor(ClaudeExecutorConfig{
		WorkDir: "/work",
		Model:   "opus",
		Tools:   []string{"Bash"},
	})

	cfg := exec.buildAgentConfig(&types.TODO{})
	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, "/work", parsed["cwd"])
	assert.Equal(t, "opus", parsed["model"])
}
