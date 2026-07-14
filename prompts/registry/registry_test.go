package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllRegisteredPromptsAreUnique(t *testing.T) {
	all := All()
	require.Len(t, all, 11)
	seen := map[string]bool{}
	for _, desc := range all {
		assert.NotEmpty(t, desc.ModelPolicy, desc.ID)
		assert.False(t, seen[desc.ID], desc.ID)
		seen[desc.ID] = true
	}
}

func TestResolveBuiltinsAndModelPrecedence(t *testing.T) {
	cfg := verify.GavelConfig{
		Verify: verify.VerifyConfig{Model: "claude-code-sonnet"},
		Commit: verify.CommitConfig{Model: "claude-haiku-4-5", GroupModel: "claude-opus-4-8"},
	}
	items, err := Resolve(verify.GavelConfigTrace{TargetDir: t.TempDir(), Merged: cfg})
	require.NoError(t, err)
	require.Len(t, items, 11)

	verifyPrompt := resolvedByID(t, items, prompts.Verify)
	assert.Equal(t, "builtin", verifyPrompt.Source)
	assert.Equal(t, "claude-code-sonnet", verifyPrompt.EffectiveModel.Name)
	assert.Equal(t, "verify.model", verifyPrompt.ModelSource)
	assert.NotEmpty(t, verifyPrompt.Body)

	grouping := resolvedByID(t, items, prompts.CommitGrouping)
	assert.Equal(t, "claude-opus-4-8", grouping.EffectiveModel.Name)
	assert.Equal(t, "commit.groupModel", grouping.ModelSource)

	run := resolvedByID(t, items, prompts.TodosRun)
	assert.Equal(t, "claude", run.Declared.Model.Name)
	assert.Equal(t, "claude", run.EffectiveModel.Name)
	assert.Equal(t, "prompt spec", run.ModelSource)
}

func TestResolveStructuredInlinePrompt(t *testing.T) {
	override, err := verify.StructuredInlinePrompt(api.Spec{
		Model:  api.Model{Name: "claude-sonnet-5", Backend: api.BackendAnthropic, Effort: api.EffortHigh},
		Prompt: api.Prompt{User: "Plan {{body}}", System: "Be exact"},
	})
	require.NoError(t, err)
	cfg := verify.GavelConfig{
		Verify: verify.DefaultVerifyConfig(),
		Todos:  verify.TodosConfig{PlanPrompt: override},
	}
	items, err := Resolve(verify.GavelConfigTrace{TargetDir: t.TempDir(), Merged: cfg})
	require.NoError(t, err)

	plan := resolvedByID(t, items, prompts.TodosPlan)
	assert.Equal(t, "inline", plan.Source)
	assert.Equal(t, "Plan {{body}}", plan.Body)
	assert.Equal(t, "Be exact", plan.Declared.Prompt.System)
	assert.Equal(t, api.EffortHigh, plan.EffectiveModel.Effort)
	assert.Equal(t, api.BackendAnthropic, plan.EffectiveModel.Backend)
}

func TestResolveFileUsesDeclaringLayer(t *testing.T) {
	configDir := t.TempDir()
	targetDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "prompts"), 0o755))
	promptPath := filepath.Join(configDir, "prompts", "status.prompt")
	require.NoError(t, os.WriteFile(promptPath, []byte("Summarize {{details}}"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, ".gavel.yaml"), []byte("status:\n  summaryPrompt:\n    file: prompts/status.prompt\n"), 0o644))

	cfg, err := verify.LoadSingleGavelConfig(filepath.Join(configDir, ".gavel.yaml"))
	require.NoError(t, err)
	cfg.Verify = verify.DefaultVerifyConfig()
	items, err := Resolve(verify.GavelConfigTrace{TargetDir: targetDir, Merged: cfg})
	require.NoError(t, err)

	statusPrompt := resolvedByID(t, items, prompts.StatusSummary)
	assert.Equal(t, "file", statusPrompt.Source)
	assert.Equal(t, promptPath, statusPrompt.Path)
	assert.Equal(t, "Summarize {{details}}", statusPrompt.Body)
}

func TestResolveMissingPromptFileFails(t *testing.T) {
	cfg := verify.GavelConfig{
		Verify: verify.DefaultVerifyConfig(),
		Status: verify.StatusConfig{SummaryPrompt: verify.PromptOverride{File: "missing.prompt"}},
	}
	_, err := Resolve(verify.GavelConfigTrace{TargetDir: t.TempDir(), Merged: cfg})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status.summary")
}

func resolvedByID(t *testing.T, items []ResolvedPrompt, id string) ResolvedPrompt {
	t.Helper()
	for _, item := range items {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("resolved prompt %q not found", id)
	return ResolvedPrompt{}
}
