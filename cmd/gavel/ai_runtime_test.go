package main

import (
	"testing"

	"github.com/flanksource/captain/pkg/aiflags"
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ = Describe("AI fix request", func() {
	It("uses the requested working directory", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		workDir := GinkgoT().TempDir()
		operation := api.Spec{Model: api.Model{Name: "agent:sonnet"}}

		_, req, err := buildAIFixRequest(captaincli.AIRuntimeOptions{}, operation, workDir)

		Expect(err).NotTo(HaveOccurred())
		Expect(req.Cwd()).To(Equal(workDir))
	})
})

func TestBuildAIFixRequestUsesOperationModelIndependently(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	operation := api.Spec{
		Model:  api.Model{Name: "agent:sonnet", Effort: api.EffortHigh},
		Budget: api.Budget{Cost: 2, MaxTokens: 8192, MaxTurns: 20},
	}

	cfg, req, err := buildAIFixRequest(captaincli.AIRuntimeOptions{}, operation, t.TempDir())
	require.NoError(t, err)
	assert.Equal(t, "claude-sonnet-5", cfg.Model.Name)
	assert.Equal(t, api.BackendClaudeAgent, cfg.Model.Backend)
	assert.Equal(t, api.EffortHigh, req.Model.Effort)
	assert.Equal(t, 2.0, cfg.Budget.Cost)
	assert.Equal(t, 8192, req.Budget.MaxTokens)
	assert.Equal(t, 20, req.Budget.MaxTurns)
}

// The operation's prompt and workflow are what make a spec-driven aifix run
// possible at all: without them the request would carry no rendered prompt and no
// verify loop, and `--ai-fix` would be back to firing once with an empty prompt.
func TestBuildAIFixRequestCarriesOperationPromptWorkflowAndTimeout(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	operation := api.Spec{
		Model:  api.Model{Name: "agent:sonnet"},
		Budget: api.Budget{Timeout: "45m", MaxTurns: 40},
		Prompt: api.Prompt{System: "you fix PRs", User: "status snapshot"},
		Workflow: &api.Workflow{
			Verify: &api.Verify{Commands: []string{"gavel pr status --follow"}, MaxIterations: 3},
		},
	}

	_, req, err := buildAIFixRequest(captaincli.AIRuntimeOptions{}, operation, t.TempDir())
	require.NoError(t, err)
	assert.Equal(t, "you fix PRs", req.Prompt.System)
	assert.Equal(t, "status snapshot", req.Prompt.User)
	assert.Equal(t, "45m", req.Budget.Timeout)
	assert.Equal(t, 40, req.Budget.MaxTurns)
	require.NotNil(t, req.Workflow)
	require.NotNil(t, req.Workflow.Verify)
	assert.Equal(t, []string{"gavel pr status --follow"}, req.Workflow.Verify.Commands)
	assert.Equal(t, 3, req.Workflow.Verify.MaxIterations)
}

func TestBuildAIFixRequestCLIModelOverridesOperation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	operation := api.Spec{Model: api.Model{Name: "agent:sonnet", Effort: api.EffortHigh}}
	opts := captaincli.AIRuntimeOptions{
		AIProviderOptions: captaincli.AIProviderOptions{
			ModelFlags: aiflags.ModelFlags{Model: "agent:opus", Effort: "medium"},
		},
	}

	cfg, req, err := buildAIFixRequest(opts, operation, t.TempDir())
	require.NoError(t, err)
	assert.Equal(t, "claude-opus-5", cfg.Model.Name)
	assert.Equal(t, api.EffortMedium, req.Model.Effort)
}
