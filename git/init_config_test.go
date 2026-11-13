package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/verify"
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

func TestBuildInitPrompt(t *testing.T) {
	layers, err := (verify.PromptSpec{}).Layers(verify.PromptResolveOptions{DefaultPrompt: initConfigPrompt, Data: map[string]any{"workDir": "/repo", "configPath": "/repo/.gitanalyze.yaml"}})
	require.NoError(t, err)
	assert.Contains(t, layers[1].Spec.Prompt.User, "/repo/.gitanalyze.yaml")
	assert.Contains(t, layers[1].Spec.Prompt.User, "git log")
	assert.Contains(t, layers[1].Spec.Prompt.User, "repository structure")
}
