package runtime

import (
	"testing"

	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/runtime/runtimetest"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A run step that verified its own work and then asked is listed by the phase
// index under both run and verify: one waiting run, two phases. Resolving the
// asking session must name the run step that link was dispatched for, which is
// the step an answer in that session resumes.
func TestSessionAttemptResolvesRunStepThatReachedVerification(t *testing.T) {
	provider := newTestProvider(t, "gavel_todo_session_attempt")

	const providerSessionID = "019fa17d-622a-7ef3-b8ad-d8b1d7cd3837"
	todo, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Ask after verifying", Body: "Verify, then ask", Status: types.StatusPending,
	})
	require.NoError(t, err)
	node := api.VerifyNode{Name: "go test ./...", Failed: true}
	report := api.NewNodeReport(api.VerifyKindFixture, "fixture", node)
	report.Ran, report.Iteration = true, 1
	asked := &todos.ExecutionResult{
		Success: true, ExecutorName: "headless-codex", EndStatus: types.EndAsk, Summary: "which fixture should pass?",
		OutputJSON: map[string]any{"summary": "which fixture should pass?", "endStatus": "ask"},
	}
	preparation, err := runtimetest.Admit(t.Context(), provider, todo, todos.RunPreparation{
		Mode: types.ModeRun, ExecutorName: "headless-codex",
		Requested: captaindb.PromptRunRuntimeSelection{Provider: "openai", Mode: "agent", Model: "gpt-5.6-sol", Effort: "high"},
	}, promptrun.Completed{
		Runtime: api.RuntimeOf(api.OpenAI, api.ModeAgent), Model: "gpt-5.6-sol", ProviderSessionID: providerSessionID,
		Iterations: promptrun.IterationRecords(promptrun.Result{
			Verdicts: []agent.VerifyResult{{Iteration: 1, Report: &report}}, Report: &report,
		}, false),
		Outcome: lifecycle.RunOutcome(asked, false),
	})
	require.NoError(t, err)
	t.Cleanup(func() { provider.ownership.stop(preparation.PromptRunID) })
	todo.Attempts = 1
	require.NoError(t, provider.SaveAttempt(t.Context(), todo, asked))

	issueID := mustUUID(t, todo.ID)
	phaseRuns, err := provider.Repository().ListIssuePhaseRuns(t.Context(), provider.workspace.ID)
	require.NoError(t, err)
	activePhases := map[native.StepKind]string{}
	for _, row := range phaseRuns {
		if row.IssueID == issueID && row.Active {
			activePhases[row.Phase] = row.State
		}
	}
	waiting := string(captaindb.PromptRunStateWaiting)
	require.Equal(t, map[native.StepKind]string{native.StepRun: waiting, native.StepVerify: waiting}, activePhases,
		"the phase index lists the one waiting run under both run and verify")

	want := run.SessionAttempt{PromptRunID: preparation.PromptRunID, Step: string(native.StepRun)}
	for _, sessionID := range []string{providerSessionID, preparation.PromptRunID.String(), preparation.SessionID} {
		got, err := provider.SessionAttempt(t.Context(), todo, sessionID)
		require.NoError(t, err, "session %s", sessionID)
		assert.Equal(t, want, got, "session %s resolves to the run step's attempt", sessionID)
	}

	_, err = provider.SessionAttempt(t.Context(), todo, "019fa17d-0000-7000-8000-000000000000")
	assert.ErrorIs(t, err, native.ErrNotFound, "a session of no attempt of this todo")
	_, err = provider.SessionAttempt(t.Context(), todo, "")
	assert.ErrorIs(t, err, native.ErrInvalidInput)
}
