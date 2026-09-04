package runtime

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderNativeLifecycleIntegration(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native runtime tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir: filepath.Join(t.TempDir(), "postgres"), Database: "gavel_todo_runtime",
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

	repository, err := native.NewRepository(opened.Gorm())
	require.NoError(t, err)
	oldRoot := filepath.Join(t.TempDir(), "workspace-old")
	newRoot := filepath.Join(t.TempDir(), "workspace-current")
	targetRoot := filepath.Join(t.TempDir(), "workspace-target")
	for _, root := range []string{oldRoot, newRoot, targetRoot} {
		require.NoError(t, os.MkdirAll(root, 0o755))
	}
	workspace, err := repository.CreateWorkspace(t.Context(), native.CreateWorkspaceInput{
		RepoKey: "github.com/example/runtime-source", RootPath: oldRoot, DisplayName: "Runtime source",
	})
	require.NoError(t, err)
	workspace, err = repository.UpdateWorkspace(t.Context(), workspace.ID, native.UpdateWorkspaceInput{RootPath: &newRoot})
	require.NoError(t, err)
	targetWorkspace, err := repository.CreateWorkspace(t.Context(), native.CreateWorkspaceInput{
		RepoKey: "github.com/example/runtime-target", RootPath: targetRoot, DisplayName: "Runtime target",
	})
	require.NoError(t, err)
	repoFallbackRoot := filepath.Join(t.TempDir(), "repo-fallback-primary")
	require.NoError(t, os.MkdirAll(repoFallbackRoot, 0o755))
	repoFallbackWorkspace, err := repository.CreateWorkspace(t.Context(), native.CreateWorkspaceInput{
		RepoKey: "github.com/flanksource/gavel", RootPath: repoFallbackRoot, DisplayName: "Repo fallback",
	})
	require.NoError(t, err)

	provider, err := New(t.Context(), opened.Gorm(), WorkspaceOptions{
		Name: "Runtime source", RootPath: oldRoot, Repositories: []string{"example/runtime-source"},
	})
	require.NoError(t, err, "retained workspace paths must resolve")
	targetProvider, err := New(t.Context(), opened.Gorm(), WorkspaceOptions{
		Name: "Runtime target", RootPath: targetRoot, Repositories: []string{"example/runtime-target"},
	})
	require.NoError(t, err)
	assert.Equal(t, workspace.ID, provider.Workspace().ID)
	assert.Equal(t, targetWorkspace.ID, targetProvider.Workspace().ID)
	assert.NotNil(t, provider.Repository())
	assert.NotNil(t, provider.Captain())
	assert.NotNil(t, provider.Coordinator())
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	repoFallbackProvider, err := New(t.Context(), opened.Gorm(), WorkspaceOptions{
		Name: "Repo fallback", RootPath: repoRoot, Repositories: []string{"flanksource/gavel"},
	})
	require.NoError(t, err, "GitHub repository identity must resolve after path lookup misses")
	assert.Equal(t, repoFallbackWorkspace.ID, repoFallbackProvider.Workspace().ID)

	missingRoot := filepath.Join(t.TempDir(), "missing-workspace")
	require.NoError(t, os.MkdirAll(missingRoot, 0o755))
	var workspaceCountBefore int64
	require.NoError(t, opened.Gorm().Table("todo_workspaces").Count(&workspaceCountBefore).Error)
	initialized, err := New(t.Context(), opened.Gorm(), WorkspaceOptions{
		Name: "Initialized", RootPath: missingRoot, Repositories: []string{"example/initialized"},
	})
	require.NoError(t, err)
	assert.Equal(t, missingRoot, initialized.Workspace().RootPath)
	var workspaceCountAfter int64
	require.NoError(t, opened.Gorm().Table("todo_workspaces").Count(&workspaceCountAfter).Error)
	assert.Equal(t, workspaceCountBefore+1, workspaceCountAfter, "opening a configured project must initialize one workspace")

	body := "Description\n\n## Verification\n\n```bash\ntrue\n```\n"
	created, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Runtime issue", Body: body, Priority: types.PriorityHigh,
		Status: types.StatusPending, Labels: []string{"Todos", "database", "todos"},
	})
	require.NoError(t, err)
	assert.Equal(t, todos.ProviderDB, created.Provider)
	assert.Equal(t, int64(1), created.Version)
	assert.Equal(t, []string{"database", "todos"}, created.Labels)

	stored, err := repository.GetIssue(t.Context(), mustUUID(t, created.ID))
	require.NoError(t, err)
	assert.Equal(t, "Description", stored.Body)
	assert.Equal(t, "```bash\ntrue\n```", stored.Verification)
	assert.Equal(t, "Description", created.MarkdownBody)
	assert.Equal(t, "```bash\ntrue\n```", created.VerificationMarkdown)

	alias := "e2a3b8c2d0f7c9a98b400dc78e8a94a5"
	stored, err = repository.SetAliases(t.Context(), stored.ID, stored.Version, []native.AliasInput{{
		Alias: alias, Kind: "external",
	}}, "integration-test")
	require.NoError(t, err)
	created, err = provider.Get(t.Context(), alias[:8])
	require.NoError(t, err)
	assert.Equal(t, stored.ID.String(), created.ID)

	updatedBody := "Edited\n\n## Verification\n\n```bash\nmake test\n```\n"
	versionBeforeEdit := created.Version
	require.NoError(t, provider.Edit(t.Context(), created, todos.EditRequest{Body: &updatedBody}))
	assert.Equal(t, versionBeforeEdit+1, created.Version, "body and verification must share one mutation")
	stored, err = repository.GetIssue(t.Context(), mustUUID(t, created.ID))
	require.NoError(t, err)
	assert.Equal(t, "Edited", stored.Body)
	assert.Equal(t, "```bash\nmake test\n```", stored.Verification)
	assert.Equal(t, "Edited", created.MarkdownBody)
	assert.Equal(t, "```bash\nmake test\n```", created.VerificationMarkdown)

	versionBeforeTransient := created.Version
	failed := types.StatusFailed
	require.NoError(t, provider.UpdateState(t.Context(), created, todos.StateUpdate{Status: &failed}))
	assert.Equal(t, types.StatusFailed, created.Status, "legacy executor may retain its in-memory status")
	stored, err = repository.GetIssue(t.Context(), mustUUID(t, created.ID))
	require.NoError(t, err)
	assert.Equal(t, versionBeforeTransient, stored.Version)
	assert.Equal(t, native.StatusOpen, stored.Status)
	assert.Equal(t, native.ExecutionIdle, stored.ExecutionState)

	created, err = provider.Get(t.Context(), alias)
	require.NoError(t, err)
	require.NoError(t, provider.Comment(t.Context(), created, "Native comment"))
	created.Attempts = 2
	require.NoError(t, provider.SaveAttempt(t.Context(), created, &todos.ExecutionResult{
		Success: true, ExecutorName: "codex", Duration: 2 * time.Second,
		TokensUsed: 42, CostUSD: 0.25, CommitSHA: "abc123",
	}))
	require.NoError(t, provider.UpdateLatestFailure(t.Context(), created, &types.TestResultInfo{
		Command: "go test ./...", CWD: oldRoot, Output: "failed", Duration: time.Second,
	}))
	verified := types.StatusVerified
	require.NoError(t, provider.UpdateState(t.Context(), created, todos.StateUpdate{Status: &verified}))
	stored, err = repository.GetIssue(t.Context(), mustUUID(t, created.ID))
	require.NoError(t, err)
	assert.Equal(t, native.StatusVerified, stored.Status)

	runTodo, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Runtime no-fixture run", Body: "Implement the runtime cutover", Status: types.StatusPending,
	})
	require.NoError(t, err)
	preparation, err := provider.PrepareRun(t.Context(), runTodo, todos.RunPreparation{
		Mode: types.ModeRun, ExecutorName: "codex",
		Requested: captaindb.PromptRunRuntimeSelection{Provider: "openai", Mode: "agent", Model: "gpt-requested", Effort: "high"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, preparation.SessionID)
	assert.Equal(t, types.StatusInProgress, runTodo.Status)
	assert.Equal(t, string(native.ExecutionRunning), runTodo.ExecutionState)
	require.NoError(t, provider.RecordRunStart(t.Context(), runTodo, todos.RunStartMetadata{
		SessionID: "runtime-run-1", Mode: "run", Driver: "codex", Provider: "openai",
		RuntimeMode: "agent", ResolvedModel: "gpt-runtime", Effort: "high",
	}))
	require.NotNil(t, runTodo.LLM)
	assert.Equal(t, "runtime-run-1", runTodo.LLM.SessionId)
	runningIssue, err := repository.GetIssue(t.Context(), mustUUID(t, runTodo.ID))
	require.NoError(t, err)
	require.NotNil(t, runningIssue.ActivePromptRunID)
	runningPromptRun, err := provider.Captain().GetPromptRun(t.Context(), *runningIssue.ActivePromptRunID)
	require.NoError(t, err)
	admissionSession, err := provider.Captain().GetSession(t.Context(), runningPromptRun.RootSessionID)
	require.NoError(t, err)
	assert.Equal(t, "gavel", admissionSession.Source)
	assert.Equal(t, captaindb.SessionLifecycleCreated, admissionSession.LifecycleStatus)
	agentSession, err := provider.Captain().GetSessionByIdentity(t.Context(), "runtime-run-1", "codex", "", captaindb.LocalHostID())
	require.NoError(t, err)
	assert.NotEqual(t, admissionSession.ID, agentSession.ID)
	assert.Equal(t, captaindb.SessionLifecycleCreated, agentSession.LifecycleStatus, "Gavel must not project an attempt state onto the provider thread")
	require.NotNil(t, runningPromptRun.ExecutionSessionID)
	assert.Equal(t, agentSession.ID, *runningPromptRun.ExecutionSessionID)
	assert.Equal(t, "run", runningPromptRun.Runtime.Mode)
	assert.Equal(t, "gpt-requested", runningPromptRun.Runtime.Requested.Model)
	assert.Equal(t, "gpt-runtime", runningPromptRun.Runtime.Resolved.Model)
	require.NoError(t, provider.SaveAttempt(t.Context(), runTodo, &todos.ExecutionResult{
		Success: true, ExecutorName: "codex", EndStatus: types.EndCompleted,
		Summary: "implementation finished without explicit verification",
	}))
	storedRun, err := repository.GetIssue(t.Context(), mustUUID(t, runTodo.ID))
	require.NoError(t, err)
	assert.Equal(t, native.StatusOpen, storedRun.Status, "no-fixture success must stay open")
	assert.Equal(t, native.ExecutionIdle, storedRun.ExecutionState)
	assert.Equal(t, types.StatusPending, runTodo.Status)
	require.NotNil(t, storedRun.ActivePromptRunID)
	captainRun, err := provider.Captain().GetPromptRun(t.Context(), *storedRun.ActivePromptRunID)
	require.NoError(t, err)
	assert.Equal(t, "succeeded", string(captainRun.State))
	assert.Equal(t, "finished", string(captainRun.Phase))
	agentSession, err = provider.Captain().GetSession(t.Context(), agentSession.ID)
	require.NoError(t, err)
	assert.Equal(t, captaindb.SessionLifecycleCreated, agentSession.LifecycleStatus, "attempt completion must not complete the provider thread")
	admissionSession, err = provider.Captain().GetSession(t.Context(), admissionSession.ID)
	require.NoError(t, err)
	assert.Equal(t, captaindb.SessionLifecycleCreated, admissionSession.LifecycleStatus)

	cancelledTodo, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Runtime cancelled run", Body: "Stop without recording a failure", Status: types.StatusPending,
	})
	require.NoError(t, err)
	preparation, err = provider.PrepareRun(t.Context(), cancelledTodo, todos.RunPreparation{
		Mode: types.ModeRun, ExecutorName: "headless-codex",
	})
	require.NoError(t, err)
	require.NotEmpty(t, preparation.SessionID)
	require.NoError(t, provider.RecordRunStart(t.Context(), cancelledTodo, todos.RunStartMetadata{
		SessionID: "runtime-cancelled-1", Mode: "run", Driver: "headless-codex", Provider: "openai",
		RuntimeMode: "agent", ResolvedModel: "gpt-runtime", Effort: "high",
	}))
	cancelledIssue, err := repository.GetIssue(t.Context(), mustUUID(t, cancelledTodo.ID))
	require.NoError(t, err)
	require.NotNil(t, cancelledIssue.ActivePromptRunID)
	cancelledRunID := *cancelledIssue.ActivePromptRunID
	require.NoError(t, provider.SaveAttempt(t.Context(), cancelledTodo, &todos.ExecutionResult{
		Cancelled: true, ExecutorName: "headless-codex", Summary: todos.ErrExecutionCancelled.Error(),
		ErrorMessage: todos.ErrExecutionCancelled.Error(),
	}))
	cancelledIssue, err = repository.GetIssue(t.Context(), mustUUID(t, cancelledTodo.ID))
	require.NoError(t, err)
	assert.Equal(t, native.StatusOpen, cancelledIssue.Status)
	assert.Equal(t, native.ExecutionIdle, cancelledIssue.ExecutionState)
	cancelledRun, err := provider.Captain().GetPromptRun(t.Context(), cancelledRunID)
	require.NoError(t, err)
	assert.Equal(t, captaindb.PromptRunStateCancelled, cancelledRun.State)
	assert.Equal(t, captaindb.PromptRunPhaseFinished, cancelledRun.Phase)

	tracingTodo, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Runtime exact prompt", Body: "The issue body is not the rendered agent prompt", Status: types.StatusPending,
	})
	require.NoError(t, err)
	// The admitted prompt run must carry the RENDERED prompt, not the issue body:
	// PrepareRun is what stores it, and it runs before anything is dispatched, so
	// a reader of the run row never sees a lossy reconstruction of what the agent
	// was actually asked.
	exactPrompt := "Rendered prompt with runtime instructions and the full issue envelope"
	tracingPreparation, err := provider.PrepareRun(t.Context(), tracingTodo, todos.RunPreparation{
		Mode: types.ModeRun, Prompt: "run", ExecutorName: "prompt-observer",
		Spec: api.Spec{Prompt: api.Prompt{User: exactPrompt}},
	})
	require.NoError(t, err)
	tracingRun, err := provider.Captain().GetPromptRun(t.Context(), tracingPreparation.PromptRunID)
	require.NoError(t, err)
	tracingSession, err := provider.Captain().GetSession(t.Context(), tracingRun.SessionID)
	require.NoError(t, err)
	assert.Equal(t, exactPrompt, tracingRun.PromptMarkdown, "Captain prompt run must be populated before external dispatch")
	assert.Equal(t, exactPrompt, tracingSession.InitialPrompt, "Captain session must retain the exact initial prompt")

	raceTodo, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Runtime admission race", Body: "Only one caller may dispatch", Status: types.StatusPending,
	})
	require.NoError(t, err)
	raceA, err := provider.Get(t.Context(), raceTodo.ID)
	require.NoError(t, err)
	contenderProvider, err := New(t.Context(), opened.Gorm(), WorkspaceOptions{
		Name: "Runtime source", RootPath: oldRoot, Repositories: []string{"example/runtime-source"},
	})
	require.NoError(t, err)
	raceB, err := contenderProvider.Get(t.Context(), raceTodo.ID)
	require.NoError(t, err)
	startRace := make(chan struct{})
	raceErrors := make(chan error, 2)
	go func() {
		<-startRace
		_, prepareErr := provider.PrepareRun(t.Context(), raceA, todos.RunPreparation{Mode: types.ModeRun, ExecutorName: "codex"})
		raceErrors <- prepareErr
	}()
	go func() {
		<-startRace
		_, prepareErr := contenderProvider.PrepareRun(t.Context(), raceB, todos.RunPreparation{Mode: types.ModeRun, ExecutorName: "codex"})
		raceErrors <- prepareErr
	}()
	close(startRace)
	firstRaceErr, secondRaceErr := <-raceErrors, <-raceErrors
	assert.Equal(t, 1, boolCount(firstRaceErr == nil, secondRaceErr == nil))
	assert.Equal(t, 1, boolCount(errors.Is(firstRaceErr, todos.ErrRunDispatchAlreadyClaimed), errors.Is(secondRaceErr, todos.ErrRunDispatchAlreadyClaimed)))
	raceLinks, err := repository.ListPromptRuns(t.Context(), mustUUID(t, raceTodo.ID))
	require.NoError(t, err)
	assert.Len(t, raceLinks, 1)

	resumeTodo, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Runtime resume mode", Body: "Do not cross step kinds", Status: types.StatusPending,
	})
	require.NoError(t, err)
	preparation, err = provider.PrepareRun(t.Context(), resumeTodo, todos.RunPreparation{Mode: types.ModePlan, ExecutorName: "claude"})
	require.NoError(t, err)
	require.NotEmpty(t, preparation.SessionID)
	resumeBefore, err := repository.GetIssue(t.Context(), mustUUID(t, resumeTodo.ID))
	require.NoError(t, err)
	_, err = provider.PrepareRun(t.Context(), resumeTodo, todos.RunPreparation{Mode: types.ModeRun, ExecutorName: "claude", Resume: true})
	require.ErrorIs(t, err, todos.ErrRunResumeModeMismatch)
	resumeAfter, err := repository.GetIssue(t.Context(), mustUUID(t, resumeTodo.ID))
	require.NoError(t, err)
	assert.Equal(t, resumeBefore.ActivePromptRunID, resumeAfter.ActivePromptRunID)

	failedPlanTodo, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Runtime failed plan", Body: "A failed result must not become executable", Status: types.StatusPending,
	})
	require.NoError(t, err)
	preparation, err = provider.PrepareRun(t.Context(), failedPlanTodo, todos.RunPreparation{Mode: types.ModePlan, ExecutorName: "claude"})
	require.NoError(t, err)
	require.NotEmpty(t, preparation.SessionID)
	require.NoError(t, provider.SaveAttempt(t.Context(), failedPlanTodo, &todos.ExecutionResult{
		Success: false, ExecutorName: "claude", EndStatus: types.EndFailed, ErrorMessage: "planning failed",
		Plan: &types.PlanResult{Status: types.PlanNew, Content: "# Partial and invalid plan"},
	}))
	failedPlanIssue, err := repository.GetIssue(t.Context(), mustUUID(t, failedPlanTodo.ID))
	require.NoError(t, err)
	assert.Nil(t, failedPlanIssue.SelectedPlanID)
	failedPlanLinks, err := repository.ListPlans(t.Context(), failedPlanIssue.ID)
	require.NoError(t, err)
	assert.Empty(t, failedPlanLinks)

	planTodo, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Runtime plan review", Body: "Design and implement the database plan flow", Status: types.StatusPending,
	})
	require.NoError(t, err)
	preparation, err = provider.PrepareRun(t.Context(), planTodo, todos.RunPreparation{
		Mode: types.ModePlan, ExecutorName: "claude",
	})
	require.NoError(t, err)
	require.NotEmpty(t, preparation.SessionID)
	require.NoError(t, provider.RecordRunStart(t.Context(), planTodo, todos.RunStartMetadata{
		SessionID: "runtime-plan-1", Mode: "plan", ResolvedModel: "sonnet-runtime",
	}))
	planMarkdown := "# Native plan\n\n1. Persist Captain revisions.\n2. Project Gavel state."
	require.NoError(t, provider.SaveAttempt(t.Context(), planTodo, &todos.ExecutionResult{
		Success: true, ExecutorName: "claude", EndStatus: types.EndCompleted,
		Summary: "plan ready for review",
		Plan:    &types.PlanResult{Status: types.PlanNew, Content: planMarkdown},
	}))
	assert.Equal(t, types.StatusReview, planTodo.Status)
	assert.Empty(t, planTodo.PlanPath)
	assert.True(t, todos.HasPlan(planTodo))
	latestPlan, err := provider.PlanMarkdown(t.Context(), planTodo, types.ModePlan)
	require.NoError(t, err)
	assert.Equal(t, planMarkdown, latestPlan)
	_, err = provider.PlanMarkdown(t.Context(), planTodo, types.ModeRun)
	require.ErrorContains(t, err, "approve an immutable revision")
	initialPlanIssue, err := repository.GetIssue(t.Context(), mustUUID(t, planTodo.ID))
	require.NoError(t, err)
	require.NotNil(t, initialPlanIssue.ActivePromptRunID)
	initialPlanRun, err := provider.Captain().GetPromptRun(t.Context(), *initialPlanIssue.ActivePromptRunID)
	require.NoError(t, err)

	planTodo, err = provider.RequestPlanRevision(t.Context(), planTodo, "reviewer", "add rollback steps")
	require.NoError(t, err)
	assert.Equal(t, types.StatusReview, planTodo.Status)
	preparation, err = provider.PrepareRun(t.Context(), planTodo, todos.RunPreparation{
		Mode: types.ModePlan, ExecutorName: "claude", Resume: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, preparation.SessionID)
	revisedPlanIssue, err := repository.GetIssue(t.Context(), mustUUID(t, planTodo.ID))
	require.NoError(t, err)
	require.NotNil(t, revisedPlanIssue.ActivePromptRunID)
	assert.NotEqual(t, *initialPlanIssue.ActivePromptRunID, *revisedPlanIssue.ActivePromptRunID)
	revisedPlanRun, err := provider.Captain().GetPromptRun(t.Context(), *revisedPlanIssue.ActivePromptRunID)
	require.NoError(t, err)
	assert.Equal(t, initialPlanRun.SessionID, revisedPlanRun.SessionID, "resume reuses the Captain session but creates a new prompt run")
	require.NoError(t, provider.RecordRunStart(t.Context(), planTodo, todos.RunStartMetadata{
		SessionID: "runtime-plan-1", Mode: "plan", ResolvedModel: "sonnet-runtime",
	}))
	revisedMarkdown := planMarkdown + "\n3. Add rollback steps."
	require.NoError(t, provider.SaveAttempt(t.Context(), planTodo, &todos.ExecutionResult{
		Success: true, ExecutorName: "claude", EndStatus: types.EndCompleted,
		Summary: "plan revised",
		Plan:    &types.PlanResult{Status: types.PlanUpdated, Content: revisedMarkdown},
	}))
	assert.Equal(t, types.StatusReview, planTodo.Status)

	planTodo, err = provider.ApprovePlan(t.Context(), planTodo, "reviewer", "looks good")
	require.NoError(t, err)
	assert.Equal(t, types.StatusPending, planTodo.Status)
	approvedPlan, err := provider.PlanMarkdown(t.Context(), planTodo, types.ModeRun)
	require.NoError(t, err)
	assert.Equal(t, revisedMarkdown, approvedPlan)
	humanEditedMarkdown := revisedMarkdown + "\n4. Verify rollback ownership."
	planTodo, err = provider.SavePlanRevision(t.Context(), planTodo, humanEditedMarkdown, "reviewer")
	require.NoError(t, err)
	assert.Equal(t, types.StatusReview, planTodo.Status, "a new revision resets prior approval")
	_, err = provider.PlanMarkdown(t.Context(), planTodo, types.ModeRun)
	require.ErrorContains(t, err, "pending")
	planTodo, err = provider.ApprovePlan(t.Context(), planTodo, "reviewer", "edited revision approved")
	require.NoError(t, err)
	approvedPlan, err = provider.PlanMarkdown(t.Context(), planTodo, types.ModeRun)
	require.NoError(t, err)
	assert.Equal(t, humanEditedMarkdown, approvedPlan)
	preparation, err = provider.PrepareRun(t.Context(), planTodo, todos.RunPreparation{
		Mode: types.ModeRun, ExecutorName: "claude",
	})
	require.NoError(t, err)
	require.NotEmpty(t, preparation.SessionID)
	currentPlanIssue, err := repository.GetIssue(t.Context(), mustUUID(t, planTodo.ID))
	require.NoError(t, err)
	require.NotNil(t, currentPlanIssue.ActivePromptRunID)
	implementationRun, err := provider.Captain().GetPromptRun(t.Context(), *currentPlanIssue.ActivePromptRunID)
	require.NoError(t, err)
	require.NotNil(t, implementationRun.InputPlanID)
	require.NotNil(t, implementationRun.InputPlanRevisionID)
	require.NoError(t, provider.RecordRunStart(t.Context(), planTodo, todos.RunStartMetadata{
		SessionID: "runtime-implement-1", Mode: "run", ResolvedModel: "sonnet-runtime",
	}))
	require.NoError(t, provider.SaveAttempt(t.Context(), planTodo, &todos.ExecutionResult{
		Success: true, ExecutorName: "claude", EndStatus: types.EndCompleted,
		Summary: "implemented approved plan",
	}))
	assert.Equal(t, types.StatusPending, planTodo.Status, "successful no-fixture implementation remains open")

	planTodo, err = provider.RejectPlan(t.Context(), planTodo, "reviewer", "superseded")
	require.NoError(t, err)
	assert.Equal(t, types.StatusPending, planTodo.Status)

	global, cwd, err := provider.GlobalGet(t.Context(), alias)
	require.NoError(t, err)
	assert.Equal(t, created.ID, global.ID)
	assert.Equal(t, newRoot, cwd, "global lookup returns the owning primary CWD")

	moved, err := provider.MoveTo(t.Context(), created, targetProvider)
	require.NoError(t, err)
	assert.Equal(t, targetWorkspace.ID.String(), moved.WorkspaceID)
	assert.Equal(t, created.ID, moved.ID)
	_, err = provider.Get(t.Context(), alias)
	assert.True(t, errors.Is(err, native.ErrNotFound))
	byAlias, err := targetProvider.Get(t.Context(), alias)
	require.NoError(t, err)
	assert.Equal(t, moved.ID, byAlias.ID)

	require.NoError(t, targetProvider.Delete(t.Context(), byAlias))
	stored, err = repository.GetIssue(t.Context(), mustUUID(t, byAlias.ID))
	require.NoError(t, err)
	assert.Equal(t, native.StatusCancelled, stored.Status, "delete must preserve the issue and history")
}

func boolCount(values ...bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}
	return count
}

func mustUUID(t *testing.T, value string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(value)
	require.NoError(t, err)
	return id
}
