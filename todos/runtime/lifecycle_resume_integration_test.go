package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newResumeTestProvider(t *testing.T, database_ string) *Provider {
	t.Helper()
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir: filepath.Join(t.TempDir(), "postgres"), Database: database_,
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })

	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })

	root := filepath.Join(t.TempDir(), "workspace")
	require.NoError(t, os.MkdirAll(root, 0o755))
	provider, err := New(t.Context(), opened.Gorm(), WorkspaceOptions{
		Name: "Resume", RootPath: root, Repositories: []string{"example/resume"},
	})
	require.NoError(t, err)
	return provider
}

// countSessions reports how many Captain sessions share a provider session id.
// The resume path used to fork a second agent row whenever it could not name the
// provider, because the session identity key is (source, provider, host, id).
func countSessions(t *testing.T, provider *Provider, providerSessionID string) int64 {
	t.Helper()
	var count int64
	require.NoError(t, provider.db.Table("captain_sessions").
		Where("provider_session_id = ?", providerSessionID).Count(&count).Error)
	return count
}

// TestRecordRunStartResumeKeepsExecutionSession is the regression oracle for the
// "Send & resume" dead end: a resumed turn reports only the session id, mode and
// executor name, so RecordRunStart used to look up the agent session under a
// blank provider, miss the bound one, and then lose the optimistic update to the
// execution-session guard — failing the resume with a prompt-run conflict before
// any feedback was ever sent.
func TestRecordRunStartResumeKeepsExecutionSession(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native runtime tests")
	}
	provider := newResumeTestProvider(t, "gavel_todo_resume")

	const providerSessionID = "019fa17d-622a-7ef3-b8ad-d8b1d7cd3836"
	todo, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Resume after ask", Body: "Answer the agent's question", Status: types.StatusPending,
	})
	require.NoError(t, err)

	require.NoError(t, provider.PrepareRun(t.Context(), todo, todos.RunPreparation{
		Mode: types.ModeRun, ExecutorName: "headless-codex",
		Requested: captaindb.PromptRunRuntimeSelection{
			Provider: "openai", Backend: "codex-agent", Model: "gpt-5.6-sol", Effort: "high",
		},
	}))
	require.NoError(t, provider.RecordRunStart(t.Context(), todo, todos.RunStartMetadata{
		SessionID: providerSessionID, Mode: "run", Driver: "headless-codex", Agent: "codex",
		Provider: "openai", Backend: "codex-agent", ResolvedModel: "gpt-5.6-sol", Effort: "high",
	}))

	issue, err := provider.Repository().GetIssue(t.Context(), mustUUID(t, todo.ID))
	require.NoError(t, err)
	require.NotNil(t, issue.ActivePromptRunID)
	firstRun, err := provider.Captain().GetPromptRun(t.Context(), *issue.ActivePromptRunID)
	require.NoError(t, err)
	require.NotNil(t, firstRun.ExecutionSessionID)
	boundExecutionSession := *firstRun.ExecutionSessionID
	sessionsAfterStart := countSessions(t, provider, providerSessionID)

	// The agent finished its turn with questions: the prompt run parks at
	// waiting, which is what the dashboard projects as StatusAsk.
	todo.Attempts = 1
	require.NoError(t, provider.SaveAttempt(t.Context(), todo, &todos.ExecutionResult{
		Success: true, ExecutorName: "headless-codex", EndStatus: types.EndAsk,
		Summary: "which database should the migration target?",
	}))
	waiting, err := provider.Captain().GetPromptRun(t.Context(), firstRun.ID)
	require.NoError(t, err)
	require.Equal(t, captaindb.PromptRunStateWaiting, waiting.State)

	// The user answers. A resumed turn continues an already-resolved thread, so
	// it reports only the session and the mode — driver, provider, backend, model
	// and effort are all blank.
	require.NoError(t, provider.PrepareRun(t.Context(), todo, todos.RunPreparation{
		Mode: types.ModeRun, ExecutorName: "headless-codex", Resume: true,
		Spec: api.Spec{Prompt: api.Prompt{User: "Target the staging database."}},
	}))
	require.NoError(t, provider.RecordRunStart(t.Context(), todo, todos.RunStartMetadata{
		SessionID: providerSessionID, Mode: "run",
	}), "a resumed turn must not be rejected for failing to re-name its provider")

	resumed, err := provider.Captain().GetPromptRun(t.Context(), firstRun.ID)
	require.NoError(t, err)
	assert.Equal(t, captaindb.PromptRunStateRunning, resumed.State, "resume must reactivate the parked prompt run")
	require.NotNil(t, resumed.ExecutionSessionID)
	assert.Equal(t, boundExecutionSession, *resumed.ExecutionSessionID, "a prompt run's execution thread is bound once")
	assert.Equal(t, sessionsAfterStart, countSessions(t, provider, providerSessionID),
		"resume must reuse the bound agent session instead of forking a provider-less duplicate")

	assert.Equal(t, "headless-codex", resumed.Runtime.Driver, "resume must not erase the recorded driver")
	assert.Equal(t, captaindb.PromptRunRuntimeSelection{
		Provider: "openai", Backend: "codex-agent", Model: "gpt-5.6-sol", Effort: "high",
	}, resumed.Runtime.Resolved, "resume must not erase the resolved runtime")

	require.NoError(t, provider.reloadTODO(t.Context(), todo, todo.CWD))
	assert.Equal(t, types.StatusInProgress, todo.Status, "a resumed todo leaves ask")
	assert.Equal(t, string(native.ExecutionRunning), todo.ExecutionState)
}
