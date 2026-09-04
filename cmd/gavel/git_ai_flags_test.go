package main

import (
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resolveAnalyzeModel mirrors what newAnalyzeAgent does, against a config whose
// ai: base is the built-in one, so these assertions are about precedence rather
// than about whatever .gavel.yaml the test happens to run under.
func resolveAnalyzeModel(t *testing.T, override api.Model) api.Model {
	t.Helper()
	cfg := verify.DefaultGavelConfig()
	resolved, err := captainai.Resolve(cfg.ModelFor(cfg.Commit.Message, override))
	require.NoError(t, err)
	return resolved
}

// `git analyze --ai --ai-model api:haiku` ran on agent mode no matter what was
// asked: AnalyzeCommitHistory built its own agent from ai.DefaultConfig() deep in
// the git package, so the flag reached only the --summary pass.
func TestGitAnalyzeAIFlagsReachTheAgentConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	defer func() { analyzeAI = clickyai.DefaultConfig() }()

	cmd, _, err := rootCmd.Find([]string{"git", "analyze"})
	require.NoError(t, err)
	require.NoError(t, cmd.Flags().Parse([]string{"--ai", "--ai-model", "api:haiku"}))

	resolved := resolveAnalyzeModel(t, analyzeAI.Model)
	assert.Equal(t, "claude-haiku-4-5", resolved.Name)
	assert.EqualValues(t, "api", resolved.Mode, "an explicit api: prefix must beat the provider's agent default")
}

// amendAI was declared after the closure that should have read it, so every
// --ai-* flag on `git amend-commits` parsed into a struct nothing consumed.
func TestGitAmendAIFlagsReachTheAgentConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	defer func() { amendAI = clickyai.DefaultConfig() }()

	cmd, _, err := rootCmd.Find([]string{"git", "amend-commits"})
	require.NoError(t, err)
	require.NoError(t, cmd.Flags().Parse([]string{"--ai-model", "api:haiku", "--ai-max-tokens", "512"}))

	assert.EqualValues(t, "api", resolveAnalyzeModel(t, amendAI.Model).Mode)
	assert.Equal(t, 512, amendAI.Budget.MaxTokens)
}

// Without a prefix the configured base decides, so the prefix is what carries the
// override rather than the flag merely being present.
func TestGitAnalyzeBareModelKeepsConfiguredMode(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	assert.EqualValues(t, "agent", resolveAnalyzeModel(t, api.Model{Name: "haiku"}).Mode)
}

// `git analyze` no longer carries a --model flag: it was advertised in the README
// and MANUAL, read by nothing but Pretty(), and so printed one model while running
// another. --ai-model is the single model flag.
func TestGitAnalyzeHasNoDeadModelFlag(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"git", "analyze"})
	require.NoError(t, err)

	assert.Nil(t, cmd.Flags().Lookup("model"), "--model was inert; --ai-model is the model flag")
	assert.NotNil(t, cmd.Flags().Lookup("ai-model"))
}

// DefaultConfig used to pin claude-haiku-4-5, which made the bug class possible
// twice: an agent could be built without anyone choosing a model, and because
// BindFlags merges --ai-model onto this struct, the hardcoded name outranked
// whatever .gavel.yaml configured.
func TestDefaultConfigCarriesNoModel(t *testing.T) {
	assert.Empty(t, clickyai.DefaultConfig().Model.Name, "DefaultConfig must not choose a model on the user's behalf")
}
