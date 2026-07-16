package main

import (
	"testing"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildAIFixRequestUsesOperationModelIndependently(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	operation := api.Spec{
		Model:  api.Model{Name: "agent:sonnet", Effort: api.EffortHigh},
		Budget: api.Budget{Cost: 2, MaxTokens: 8192, MaxTurns: 20},
	}

	cfg, req, err := buildAIFixRequest(captaincli.AIRuntimeOptions{}, operation)
	require.NoError(t, err)
	assert.Equal(t, "claude-sonnet-5", cfg.Model.Name)
	assert.Equal(t, api.BackendClaudeAgent, cfg.Model.Backend)
	assert.Equal(t, api.EffortHigh, req.Model.Effort)
	assert.Equal(t, 2.0, cfg.Budget.Cost)
	assert.Equal(t, 8192, req.Budget.MaxTokens)
	assert.Equal(t, 20, req.Budget.MaxTurns)
}

func TestBuildAIFixRequestCLIModelOverridesOperation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	operation := api.Spec{Model: api.Model{Name: "agent:sonnet", Effort: api.EffortHigh}}
	opts := captaincli.AIRuntimeOptions{
		AIProviderOptions: captaincli.AIProviderOptions{Model: "agent:opus"},
		Effort:            "medium",
	}

	cfg, req, err := buildAIFixRequest(opts, operation)
	require.NoError(t, err)
	assert.Equal(t, "claude-opus-4-8", cfg.Model.Name)
	assert.Equal(t, api.EffortMedium, req.Model.Effort)
}
