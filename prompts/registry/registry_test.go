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

// registeredPromptCount is the number of surviving prompt-driven operations
// (lint.fix, pr.fix, commit.message/summary/grouping, pr.content,
// todos.run/plan/triage, status.summary, test.outlineSummary).
const registeredPromptCount = 11

func TestAllRegisteredPromptsAreUnique(t *testing.T) {
	all := All()
	require.Len(t, all, registeredPromptCount)
	seen := map[string]bool{}
	for _, desc := range all {
		assert.NotEmpty(t, desc.ID)
		assert.Equal(t, desc.ID, desc.ConfigPath, "ConfigPath must equal ID")
		assert.False(t, seen[desc.ID], desc.ID)
		seen[desc.ID] = true
	}
}

// The operation spec's model wins over both the base ai: spec and the built-in
// default prompt.
func TestResolveOperationModelWins(t *testing.T) {
	cfg := verify.GavelConfig{
		AI:     api.Spec{Model: api.Model{Name: "ai-base-model"}},
		Commit: verify.CommitConfig{Grouping: verify.PromptSpec{Spec: api.Spec{Model: api.Model{Name: "op-group-model"}}}},
	}
	items, err := Resolve(ResolveOptions{Preview: true, Trace: verify.GavelConfigTrace{TargetDir: t.TempDir(), Merged: cfg}})
	require.NoError(t, err)
	require.Len(t, items, registeredPromptCount)

	grouping := resolvedByID(t, items, prompts.CommitGrouping)
	assert.Equal(t, "op-group-model", grouping.Effective.Name)
}

// A built-in default that pins a model (todos.run → "claude") wins over the base
// ai: spec when the operation has no override.
func TestResolvePromptDefaultModelWins(t *testing.T) {
	cfg := verify.GavelConfig{AI: api.Spec{Model: api.Model{Name: "ai-base-model"}}}
	items, err := Resolve(ResolveOptions{Preview: true, Trace: verify.GavelConfigTrace{TargetDir: t.TempDir(), Merged: cfg}})
	require.NoError(t, err)

	run := resolvedByID(t, items, prompts.TodosRun)
	assert.Equal(t, "claude", run.Declared.Model.Name)
	assert.Equal(t, "claude", run.Effective.Name)
}

// A prompt whose default pins no model inherits the base ai: spec model.
func TestResolveAIBaseModelInherited(t *testing.T) {
	cfg := verify.GavelConfig{AI: api.Spec{Model: api.Model{Name: "ai-base-model"}}}
	items, err := Resolve(ResolveOptions{Preview: true, Trace: verify.GavelConfigTrace{TargetDir: t.TempDir(), Merged: cfg}})
	require.NoError(t, err)

	message := resolvedByID(t, items, prompts.CommitMessage)
	assert.Equal(t, "ai-base-model", message.Effective.Name)
}

func TestResolveLintFixUsesDedicatedDefaultModel(t *testing.T) {
	cfg := verify.GavelConfig{AI: api.Spec{Model: api.Model{Name: "commit-message-model"}}}
	items, err := Resolve(ResolveOptions{Preview: true, Trace: verify.GavelConfigTrace{TargetDir: t.TempDir(), Merged: cfg}})
	require.NoError(t, err)

	fix := resolvedByID(t, items, prompts.LintFix)
	assert.Equal(t, "agent:gpt-5.6-sol:medium,agent:opus:medium", fix.Effective.Name)
	assert.NotEqual(t, cfg.AI.Model.Name, fix.Effective.Name)
}

// With no ai: base, no default model, and no override the model is chosen at
// runtime (empty).
func TestResolveRuntimeModel(t *testing.T) {
	items, err := Resolve(ResolveOptions{Preview: true, Trace: verify.GavelConfigTrace{TargetDir: t.TempDir(), Merged: verify.GavelConfig{}}})
	require.NoError(t, err)

	message := resolvedByID(t, items, prompts.CommitMessage)
	assert.Empty(t, message.Effective.Name)
}

// A structured inline override carries model/effort/system on the PromptSpec's
// embedded api.Spec; the effective model reflects them.
func TestResolveStructuredInlinePrompt(t *testing.T) {
	cfg := verify.GavelConfig{
		Todos: verify.TodosConfig{Plan: verify.PromptSpec{Spec: api.Spec{
			Model:  api.Model{Name: "claude-sonnet-5", Mode: api.ModeAPI, Effort: api.EffortHigh},
			Prompt: api.Prompt{User: "Plan {{body}}", System: "Be exact"},
		}}},
	}
	items, err := Resolve(ResolveOptions{Preview: true, Trace: verify.GavelConfigTrace{TargetDir: t.TempDir(), Merged: cfg}})
	require.NoError(t, err)

	plan := resolvedByID(t, items, prompts.TodosPlan)
	assert.Equal(t, "inline", plan.Source)
	assert.Equal(t, "Plan {{body}}", plan.Body)
	assert.Equal(t, "Be exact", plan.Declared.Prompt.System)
	assert.Equal(t, "claude-sonnet-5", plan.Effective.Name)
	assert.Equal(t, api.EffortHigh, plan.Effective.Effort)
	assert.Equal(t, api.ModeAPI, plan.Effective.Mode)
}

func TestResolveFileUsesDeclaringLayer(t *testing.T) {
	configDir := t.TempDir()
	targetDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "prompts"), 0o755))
	promptPath := filepath.Join(configDir, "prompts", "status.prompt")
	require.NoError(t, os.WriteFile(promptPath, []byte("Summarize {{details}}"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, ".gavel.yaml"), []byte("status:\n  summary:\n    file: prompts/status.prompt\n"), 0o644))

	cfg, err := verify.LoadSingleGavelConfig(filepath.Join(configDir, ".gavel.yaml"))
	require.NoError(t, err)
	items, err := Resolve(ResolveOptions{Preview: true, Trace: verify.GavelConfigTrace{TargetDir: targetDir, Merged: cfg}})
	require.NoError(t, err)

	statusPrompt := resolvedByID(t, items, prompts.StatusSummary)
	assert.Equal(t, "file", statusPrompt.Source)
	assert.Equal(t, promptPath, statusPrompt.Path)
	assert.Equal(t, "Summarize {{details}}", statusPrompt.Body)
}

func TestResolveMissingPromptFileFails(t *testing.T) {
	cfg := verify.GavelConfig{
		Status: verify.StatusConfig{Summary: verify.PromptSpec{File: "missing.prompt"}},
	}
	_, err := Resolve(ResolveOptions{Preview: true, Trace: verify.GavelConfigTrace{TargetDir: t.TempDir(), Merged: cfg}})
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
