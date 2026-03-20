package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConfig_CreatesDefaultFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))

	configPath, err := InitConfig(InitConfigOptions{
		Path:  dir,
		Model: "none",
	})
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, ".gitanalyze.yaml"), configPath)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "extends:")
	assert.Contains(t, string(data), "preset:bots")
	assert.Contains(t, string(data), "preset:noise")
}

func TestInitConfig_PreservesExistingFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))

	existing := "# my custom config\nignore_files:\n  - vendor/*\n"
	configPath := filepath.Join(dir, ".gitanalyze.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(existing), 0o644))

	resultPath, err := InitConfig(InitConfigOptions{
		Path:  dir,
		Model: "none",
	})
	require.NoError(t, err)
	assert.Equal(t, configPath, resultPath)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, existing, string(data))
}

func TestInitConfig_NotGitRepo(t *testing.T) {
	dir := t.TempDir()
	_, err := InitConfig(InitConfigOptions{
		Path:  dir,
		Model: "none",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a git repository")
}

func TestResolveAICLI(t *testing.T) {
	tests := []struct {
		model         string
		wantName      string
		wantModelFlag string
	}{
		{"claude", "claude", ""},
		{"gemini", "gemini", ""},
		{"codex", "codex", ""},
		{"claude-sonnet-4", "claude", "claude-sonnet-4"},
		{"gemini-2.5-flash", "gemini", "gemini-2.5-flash"},
		{"unknown-model", "claude", "unknown-model"},
	}
	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			name, modelFlag := resolveAICLI(tt.model)
			assert.Equal(t, tt.wantName, name)
			assert.Equal(t, tt.wantModelFlag, modelFlag)
		})
	}
}

func TestBuildInitPrompt(t *testing.T) {
	prompt := buildInitPrompt("/repo/.gitanalyze.yaml")
	assert.Contains(t, prompt, "/repo/.gitanalyze.yaml")
	assert.Contains(t, prompt, "git log")
	assert.Contains(t, prompt, "repository structure")
}
