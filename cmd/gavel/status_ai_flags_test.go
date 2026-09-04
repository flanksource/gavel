package main

import (
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// `gavel status --ai --ai-model api:haiku` used to run on agent mode: init()
// bound the --ai-* flags into a local AgentConfig that runStatus never saw, and
// runStatus built its agent from a fresh DefaultConfig(). The explicit mode was
// dropped, so the provider fell back to its DefaultMode (agent) — every --ai-*
// flag was silently inert.
func TestStatusAIFlagsReachTheAgentConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	defer func() { statusAI = clickyai.DefaultConfig() }()

	cmd, _, err := rootCmd.Find([]string{"status"})
	require.NoError(t, err)
	require.NoError(t, cmd.Flags().Parse([]string{
		"--ai-model", "api:haiku",
		"--ai-max-concurrent", "7",
		"--ai-max-tokens", "512",
	}))

	resolved, err := captainai.Resolve(statusAI.Model)
	require.NoError(t, err)

	assert.Equal(t, "claude-haiku-4-5", resolved.Name)
	assert.EqualValues(t, "api", resolved.Mode, "an explicit api: prefix must beat the provider's agent default")
	assert.Equal(t, 7, statusAI.MaxConcurrent)
	assert.Equal(t, 512, statusAI.Budget.MaxTokens)
}

// A bare model keeps the provider default, so the prefix is what carries the
// override rather than the flag merely being present.
func TestStatusAIModelWithoutPrefixKeepsProviderDefault(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	defer func() { statusAI = clickyai.DefaultConfig() }()

	cmd, _, err := rootCmd.Find([]string{"status"})
	require.NoError(t, err)
	require.NoError(t, cmd.Flags().Parse([]string{"--ai-model", "haiku"}))

	resolved, err := captainai.Resolve(statusAI.Model)
	require.NoError(t, err)

	assert.EqualValues(t, "agent", resolved.Mode)
}
