// Holds an organization's agent runs to its token budget on a live Substrate cluster when
// GAVEL_SANDBOX_SUBSTRATE_TEST=1: a plan step runs from the organization's own template and its
// tokens count, a run that uses the budget up is stopped, and the next is refused.
package ui

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/gavel/budget"
	"github.com/flanksource/gavel/modelkeys"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/sandbox/substratetest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOrgTokenBudgetsHoldAgentRunsOnSubstrate(t *testing.T) {
	cluster := substratetest.FromEnv(t)
	useModelKeyKeeper(t)
	client, err := substrate.Dial(substrate.Options{EnvAPI: cluster.EnvAPI, Control: cluster.Control, ControlTokenFile: cluster.TokenFile})
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	var dispatcher *sandbox.Dispatcher
	h := serveHostedRuns(t, sandbox.SubstrateBackend(substrate.Options{
		EnvAPI: cluster.EnvAPI, Control: cluster.Control, ControlTokenFile: cluster.TokenFile,
		Template: cluster.Template, TrustBundle: substratetest.TrustBundle,
	}, nil), func(d *sandbox.Dispatcher) {
		d.Keys = modelkeys.Store{DB: d.DB}
		d.CredentialURI = func(uuid.UUID, string) string { return cluster.AgentCredentialURI }
		d.Limits.Timeout = 10 * time.Minute
		d.Templates = &sandbox.OrgTemplates{DB: d.DB, Client: client, Atespace: "ate-env", Base: cluster.Template}
		d.Admission = &sandbox.Admission{Budgets: budget.Store{DB: d.DB}, RunLifetime: d.RunLifetime(), Wait: time.Minute, Recheck: 2 * time.Second}
		dispatcher = d
	})
	// The key's value stays here: the gateway injects the Secret the credential URI names.
	require.NoError(t, modelkeys.Store{DB: h.db}.Set(t.Context(), h.org, nil, "claude-code-oauth-token", "sk-ant-oat-unused-by-the-gateway"))
	t.Cleanup(func() {
		if t.Failed() {
			reportEnvironment(t, h)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	template, err := dispatcher.Templates.Resolve(ctx, h.org)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, client.DeleteTemplate(context.Background(), template.Name)) })

	budgets := dispatcher.Admission.Budgets
	setTokens := func(amount float64) {
		_, err := budgets.Set(ctx, h.org, nil, budget.Budget{Resource: budget.Tokens, Kind: budget.Cumulative, Amount: amount, Window: "now/d"})
		require.NoError(t, err)
	}
	usedTokens := func() float64 {
		statuses, err := budgets.Status(ctx, h.org, dispatcher.RunLifetime())
		require.NoError(t, err)
		for _, status := range statuses {
			if status.Resource == budget.Tokens {
				return status.Used
			}
		}
		t.Fatal("the organization has no token budget")
		return 0
	}
	t.Run("an agent run within its budget runs from the organization's template, and its tokens count", func(t *testing.T) {
		setTokens(1e9)
		runsPlanStep(t, h, "Add a CONTRIBUTING.md saying that changes go through pull requests", 10*time.Minute)
		var recorded struct {
			Template      string
			CPUMillicores int64
			MemoryBytes   int64
		}
		require.NoError(t, h.db.Raw(`SELECT template, cpu_millicores, memory_bytes FROM sandbox_runs WHERE org_id = ?`, h.org).Scan(&recorded).Error)
		assert.Equal(t, template.Name, recorded.Template)
		assert.Equal(t, template.Size.CPUMillicores, recorded.CPUMillicores)
		assert.Equal(t, template.Size.MemoryBytes, recorded.MemoryBytes)
		assert.Positive(t, usedTokens(), "the run's model calls count towards the organization's tokens")
	})

	t.Run("an agent run that uses its token budget up is stopped", func(t *testing.T) {
		setTokens(usedTokens() + 1)
		todo, promptRunID := startPlanStep(t, h, "Add a SECURITY.md saying how to report a vulnerability")
		status, message := endedAgentRun(t, h.db, todo)
		assert.Equal(t, "cancelled", status)
		assert.Contains(t, message, "stopped: budget used up: this organization has used ")
		var state string
		require.Eventually(t, func() bool {
			require.NoError(t, h.db.Raw(`SELECT state FROM captain_prompt_runs WHERE id = ?`, promptRunID).Scan(&state).Error)
			return state == "succeeded" || state == "failed" || state == "cancelled"
		}, time.Minute, time.Second, "Captain recorded the run's outcome")
		assert.NotEqual(t, "succeeded", state, "the stopped run didn't finish its step")
		assert.Eventually(t, func() bool { return !actorFrom(t, client, template.Name) }, 2*time.Minute, time.Second, "its environment is deleted")
	})

	t.Run("the next agent run is refused, saying why", func(t *testing.T) {
		status, created := hostedCall(t, h.base, http.MethodPost, "/api/todos", h.ada, "org:member", map[string]any{"dir": h.workspace, "title": "Add a CODE_OF_CONDUCT.md"})
		require.Equal(t, http.StatusCreated, status, created)
		status, refused := hostedCall(t, h.base, http.MethodPost, "/api/todos/run", h.ada, "org:member", map[string]any{"dir": h.workspace, "ref": created["id"], "step": "plan"})
		assert.Equal(t, http.StatusTooManyRequests, status, refused)
		assert.Regexp(t, `budget used up: this organization has used \d+ tokens of its \d+ tokens budget since `, refused["error"])
		run, message := endedAgentRun(t, h.db, created["id"].(string))
		assert.Equal(t, "failed", run)
		assert.Contains(t, message, "budget used up: this organization has used ")
		assert.False(t, actorFrom(t, client, template.Name), "it got no environment")
	})
}

// endedAgentRun waits for the sandbox run of todo to end, returning how it ended and why.
func endedAgentRun(t *testing.T, db *gorm.DB, todo string) (status, message string) {
	t.Helper()
	var row struct{ Status, StatusMessage string }
	require.Eventually(t, func() bool {
		require.NoError(t, db.Raw(`SELECT status, status_message FROM sandbox_runs WHERE todo_id = ?`, todo).Scan(&row).Error)
		return row.Status == "succeeded" || row.Status == "failed" || row.Status == "cancelled"
	}, 10*time.Minute, time.Second, "the agent run ended")
	return row.Status, row.StatusMessage
}
