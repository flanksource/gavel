package verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodosConfigValidateRejectsBuiltinStepUnderSteps(t *testing.T) {
	for _, name := range []string{"run", "plan", "triage", "verify"} {
		cfg := TodosConfig{Steps: map[string]api.Spec{name: {Budget: api.Budget{MaxTurns: 1}}}}
		err := cfg.Validate()
		require.Errorf(t, err, "todos.steps.%s must be rejected", name)
		assert.Contains(t, err.Error(), "todos.steps."+name)
		assert.Contains(t, err.Error(), "todos."+name)
	}
}

func TestTodosConfigValidateAcceptsACustomStep(t *testing.T) {
	cfg := TodosConfig{Steps: map[string]api.Spec{"handoff": {Budget: api.Budget{MaxTurns: 1}}}}
	require.NoError(t, cfg.Validate())
}

func TestTodosConfigValidateRejectsAStepNameThatIsNotAnIdentifier(t *testing.T) {
	err := TodosConfig{Steps: map[string]api.Spec{"Hand Off": {}}}.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"Hand Off"`)
}

func TestLifecycleConfigValidateRejectsFileWithInlineDefinition(t *testing.T) {
	cases := map[string]LifecycleConfig{
		"name":    {File: "lifecycle.yaml", Name: "acme"},
		"subject": {File: "lifecycle.yaml", Subject: map[string]string{"owner": "string"}},
		"steps":   {File: "lifecycle.yaml", Steps: []map[string]any{{"name": "run"}}},
	}
	for field, cfg := range cases {
		err := cfg.Validate()
		require.Errorf(t, err, "file plus inline %s must be rejected", field)
		assert.Contains(t, err.Error(), "mutually exclusive")
		assert.Contains(t, err.Error(), "lifecycle.yaml")
	}
	require.NoError(t, LifecycleConfig{File: "lifecycle.yaml"}.Validate())
	require.NoError(t, LifecycleConfig{Name: "acme"}.Validate())
}

// A .gavel.yaml that configures one step from two places is refused when it is
// read, naming the file, so the layer never reaches a run.
func TestLoadGavelConfigRejectsBuiltinStepUnderTodosSteps(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	doc := "todos:\n  steps:\n    run:\n      budget:\n        maxTurns: 3\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte(doc), 0o644))

	_, err := LoadGavelConfig(dir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), filepath.Join(dir, ".gavel.yaml"))
	assert.Contains(t, err.Error(), "todos.steps.run")
}

func TestLoadGavelConfigReadsACustomStepSpec(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	doc := "todos:\n  steps:\n    handoff:\n      model: sonnet\n      budget:\n        maxTurns: 3\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte(doc), 0o644))

	cfg, err := LoadGavelConfig(dir)

	require.NoError(t, err)
	require.Contains(t, cfg.Todos.Steps, "handoff")
	assert.Equal(t, "sonnet", cfg.Todos.Steps["handoff"].Model.Name)
	assert.Equal(t, 3, cfg.Todos.Steps["handoff"].Budget.MaxTurns)
}
