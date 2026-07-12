package native_test

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRepositoryLifecycle(t *testing.T) {
	repo, db := openRepository(t)
	ctx := t.Context()

	workspace, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{
		RepoKey:     " GitHub.com/Flanksource/Gavel ",
		RootPath:    " /workspace/./gavel ",
		DisplayName: " Gavel ",
	})
	require.NoError(t, err)
	assert.Equal(t, "github.com/flanksource/gavel", workspace.RepoKey)
	assert.Equal(t, "Gavel", workspace.DisplayName)

	newRoot := "/workspace/gavel-2"
	workspace, err = repo.UpdateWorkspace(ctx, workspace.ID, native.UpdateWorkspaceInput{RootPath: &newRoot})
	require.NoError(t, err)
	assert.Equal(t, newRoot, workspace.RootPath)
	byRepo, err := repo.GetWorkspaceByRepoKey(ctx, " GITHUB.COM/FLANKSOURCE/GAVEL ")
	require.NoError(t, err)
	assert.Equal(t, workspace.ID, byRepo.ID)
	for name, path := range map[string]string{
		"current": newRoot,
		"old":     " /workspace/gavel/../gavel ",
	} {
		t.Run("workspace path lookup "+name, func(t *testing.T) {
			byPath, err := repo.GetWorkspaceByPath(ctx, path)
			require.NoError(t, err)
			assert.Equal(t, workspace.ID, byPath.ID)
			assert.Equal(t, newRoot, byPath.RootPath)
		})
	}
	paths, err := repo.ListWorkspacePaths(ctx, workspace.ID)
	require.NoError(t, err)
	require.Len(t, paths, 2)
	assert.True(t, paths[0].IsPrimary)
	assert.Equal(t, newRoot, paths[0].Path)
	assert.False(t, paths[1].IsPrimary)
	assert.Equal(t, "/workspace/gavel", paths[1].Path)

	nonGit, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{
		RootPath:    "/workspace/non-git",
		DisplayName: "Non-Git",
	})
	require.NoError(t, err)
	assert.Empty(t, nonGit.RepoKey)
	byPath, err := repo.GetWorkspaceByPath(ctx, "/workspace/non-git")
	require.NoError(t, err)
	assert.Equal(t, nonGit.ID, byPath.ID)
	conflictingPath := "/workspace/non-git"
	_, err = repo.UpdateWorkspace(ctx, workspace.ID, native.UpdateWorkspaceInput{RootPath: &conflictingPath})
	require.ErrorIs(t, err, native.ErrWorkspaceConflict)
	workspaceAfterConflict, err := repo.GetWorkspace(ctx, workspace.ID)
	require.NoError(t, err)
	assert.Equal(t, newRoot, workspaceAfterConflict.RootPath, "a path collision must roll back the primary-path move")
	_, err = repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{DisplayName: "No identity"})
	require.ErrorIs(t, err, native.ErrInvalidInput)
	_, err = repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{
		RepoKey:  "github.com/example/path-conflict",
		RootPath: "/workspace/non-git",
	})
	require.ErrorIs(t, err, native.ErrWorkspaceConflict)

	legacyRef := "E2A3B8C2D0F7C9A98B400DC78E8A94A5"
	issue, err := repo.CreateIssue(ctx, native.CreateIssueInput{
		WorkspaceID:  workspace.ID,
		Aliases:      []native.AliasInput{{Alias: legacyRef, Kind: "GRITE"}},
		Title:        " Native issue ",
		Body:         "## Acceptance Criteria\n- [ ] works",
		Verification: "## Verification\n```bash\ntrue\n```",
		Labels:       []string{" Todos ", "DATABASE", "todos"},
		Priority:     native.PriorityHigh,
		Actor:        "integration-test",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), issue.Version)
	assert.Equal(t, "Native issue", issue.Title)
	assert.Equal(t, []string{"database", "todos"}, issue.Labels)
	assert.Equal(t, native.StatusOpen, issue.Status)
	assert.Equal(t, native.ExecutionIdle, issue.ExecutionState)

	for name, ref := range map[string]string{
		"uuid":         issue.ID.String(),
		"uuid short":   issue.ID.String()[:8],
		"legacy":       legacyRef,
		"legacy short": legacyRef[:8],
	} {
		t.Run("lookup "+name, func(t *testing.T) {
			resolved, err := repo.GetIssueByRef(ctx, workspace.ID, ref)
			require.NoError(t, err)
			assert.Equal(t, issue.ID, resolved.ID)
		})
	}

	title := "Updated native issue"
	labels := []string{"Zeta", "alpha", "ALPHA"}
	issue, err = repo.UpdateIssue(ctx, issue.ID, issue.Version, native.IssuePatch{
		Title:  &title,
		Labels: &labels,
		Actor:  "integration-test",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), issue.Version)
	assert.Equal(t, []string{"alpha", "zeta"}, issue.Labels)

	_, err = repo.UpdateIssue(ctx, issue.ID, 1, native.IssuePatch{Title: &title})
	require.ErrorIs(t, err, native.ErrVersionConflict)

	comment, err := repo.AddComment(ctx, issue.ID, issue.Version, "reviewer", "Looks good")
	require.NoError(t, err)
	assert.Equal(t, int64(3), comment.Sequence)
	assert.Equal(t, "comment", comment.Kind)
	issue, err = repo.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), issue.Version)

	issue, err = repo.SetAliases(ctx, issue.ID, issue.Version, []native.AliasInput{
		{Alias: legacyRef, Kind: "grite"},
		{Alias: "JIRA-42", Kind: "external"},
	}, "integration-test")
	require.NoError(t, err)
	assert.Equal(t, int64(4), issue.Version)
	aliases, err := repo.ListAliases(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, aliases, 2)
	assert.Equal(t, "jira-42", aliases[1].Alias)

	target := createIssue(t, repo, workspace.ID, "Dependency target")
	issues, err := repo.ListIssues(ctx, workspace.ID)
	require.NoError(t, err)
	require.Len(t, issues, 2)
	_, err = repo.SetAliases(ctx, target.ID, target.Version, []native.AliasInput{{Alias: "jira-42"}}, "integration-test")
	require.ErrorIs(t, err, native.ErrAliasConflict)
	unchangedTarget, err := repo.GetIssue(ctx, target.ID)
	require.NoError(t, err)
	assert.Equal(t, target.Version, unchangedTarget.Version, "alias conflict must roll back the version increment")

	_, err = repo.AddRelationship(ctx, issue.ID, target.ID, native.RelationshipDependsOn, issue.Version, "integration-test")
	require.NoError(t, err)
	issue, err = repo.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(5), issue.Version)
	require.ErrorIs(t, repo.DeleteIssue(ctx, target.ID, target.Version), native.ErrIssueHasRelationships)
	unchangedTarget, err = repo.GetIssue(ctx, target.ID)
	require.NoError(t, err)
	assert.Equal(t, target.Version, unchangedTarget.Version)

	dependents, err := repo.ListDependents(ctx, target.ID)
	require.NoError(t, err)
	require.Len(t, dependents, 1)
	assert.Equal(t, issue.ID, dependents[0].ID)
	unsatisfied, err := repo.ListUnsatisfiedDependencies(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, unsatisfied, 1)
	assert.Equal(t, target.ID, unsatisfied[0].ID)
	targetRelationships, err := repo.ListRelationships(ctx, target.ID)
	require.NoError(t, err)
	require.Len(t, targetRelationships, 1)
	assert.Equal(t, target.ID, targetRelationships[0].IssueID)
	assert.Equal(t, issue.ID, targetRelationships[0].TargetIssueID)
	assert.Equal(t, native.RelationshipBlocks, targetRelationships[0].Relation)

	verified := native.StatusVerified
	target, err = repo.UpdateIssue(ctx, target.ID, target.Version, native.IssuePatch{Status: &verified})
	require.NoError(t, err)
	unsatisfied, err = repo.ListUnsatisfiedDependencies(ctx, issue.ID)
	require.NoError(t, err)
	assert.Empty(t, unsatisfied)
	open := native.StatusOpen
	target, err = repo.UpdateIssue(ctx, target.ID, target.Version, native.IssuePatch{Status: &open})
	require.NoError(t, err)

	_, err = repo.AddRelationship(ctx, issue.ID, target.ID, native.RelationshipDependsOn, issue.Version, "integration-test")
	require.ErrorIs(t, err, native.ErrRelationshipExists)
	_, err = repo.AddRelationship(ctx, issue.ID, issue.ID, native.RelationshipDependsOn, issue.Version, "integration-test")
	require.ErrorIs(t, err, native.ErrSelfRelationship)
	_, err = repo.AddRelationship(ctx, target.ID, issue.ID, native.RelationshipDependsOn, target.Version, "integration-test")
	require.ErrorIs(t, err, native.ErrRelationshipCycle)

	_, err = repo.AddRelationship(ctx, issue.ID, target.ID, native.RelationshipRelatedTo, issue.Version, "integration-test")
	require.NoError(t, err)
	issue, err = repo.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(6), issue.Version)
	_, err = repo.AddRelationship(ctx, target.ID, issue.ID, native.RelationshipRelatedTo, target.Version, "integration-test")
	require.ErrorIs(t, err, native.ErrRelationshipExists)
	relationships, err := repo.ListRelationships(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, relationships, 2)
	for _, relationship := range relationships {
		assert.Equal(t, issue.ID, relationship.IssueID)
		assert.Equal(t, target.ID, relationship.TargetIssueID)
	}

	targetVersion := target.Version
	require.NoError(t, repo.DeleteRelationship(ctx, target.ID, issue.ID, native.RelationshipRelatedTo, target.Version, "integration-test"))
	target, err = repo.GetIssue(ctx, target.ID)
	require.NoError(t, err)
	assert.Equal(t, targetVersion+1, target.Version)
	require.NoError(t, repo.DeleteRelationship(ctx, issue.ID, target.ID, native.RelationshipDependsOn, issue.Version, "integration-test"))
	issue, err = repo.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(7), issue.Version)

	planSourceRunID := insertCaptainPromptRun(t, db)
	planID := insertCaptainPlan(t, db, planSourceRunID)
	_, err = repo.LinkPlan(ctx, issue.ID, planID, 0, issue.Version, "integration-test")
	require.NoError(t, err)
	issue, err = repo.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	plans, err := repo.ListPlans(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, plans, 1)
	assert.Equal(t, planID, plans[0].PlanID)
	issue, err = repo.SelectPlan(ctx, issue.ID, &planID, issue.Version, "integration-test")
	require.NoError(t, err)
	require.NotNil(t, issue.SelectedPlanID)
	assert.Equal(t, planID, *issue.SelectedPlanID)
	selectedVersion := issue.Version
	eventsBeforeReplay, err := repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	issue, err = repo.SelectPlan(ctx, issue.ID, &planID, issue.Version, "integration-test")
	require.NoError(t, err)
	assert.Equal(t, selectedVersion, issue.Version)
	eventsAfterReplay, err := repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	assert.Len(t, eventsAfterReplay, len(eventsBeforeReplay))
	missingPlan := uuid.New()
	_, err = repo.SelectPlan(ctx, issue.ID, &missingPlan, issue.Version, "integration-test")
	require.ErrorIs(t, err, native.ErrLinkConflict)

	promptRunID := insertCaptainPromptRun(t, db)
	_, err = repo.LinkPromptRun(ctx, issue.ID, promptRunID, native.StepRun, 0, issue.Version, "integration-test")
	require.NoError(t, err)
	issue, err = repo.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	promptRuns, err := repo.ListPromptRuns(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, promptRuns, 1)
	assert.Equal(t, promptRunID, promptRuns[0].PromptRunID)
	issue, err = repo.SetActivePromptRun(ctx, issue.ID, &promptRunID, issue.Version, "integration-test")
	require.NoError(t, err)
	require.NotNil(t, issue.ActivePromptRunID)
	assert.Equal(t, promptRunID, *issue.ActivePromptRunID)
	activeVersion := issue.Version
	eventsBeforeReplay, err = repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	issue, err = repo.SetActivePromptRun(ctx, issue.ID, &promptRunID, issue.Version, "integration-test")
	require.NoError(t, err)
	assert.Equal(t, activeVersion, issue.Version)
	eventsAfterReplay, err = repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	assert.Len(t, eventsAfterReplay, len(eventsBeforeReplay))
	issue, err = repo.UnlinkPromptRun(ctx, issue.ID, promptRunID, issue.Version, "integration-test")
	require.NoError(t, err)
	assert.Nil(t, issue.ActivePromptRunID)
	issue, err = repo.UnlinkPlan(ctx, issue.ID, planID, issue.Version, "integration-test")
	require.NoError(t, err)
	assert.Nil(t, issue.SelectedPlanID)

	events, err := repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	require.Equal(t, int(issue.Version), len(events))
	for i, event := range events {
		assert.Equal(t, int64(i+1), event.Sequence)
	}

	importEvent, err := repo.AppendEvent(ctx, target.ID, target.Version, native.EventInput{
		Kind:     " MIGRATION_WARNING ",
		Source:   " GRITE-IMPORT ",
		SourceID: "legacy-event-1",
		Payload:  map[string]any{"warning": "missing run link"},
	})
	require.NoError(t, err)
	assert.Equal(t, "grite-import", importEvent.Source)
	target, err = repo.GetIssue(ctx, target.ID)
	require.NoError(t, err)
	_, err = repo.AppendEvent(ctx, target.ID, target.Version, native.EventInput{
		Kind:     "migration_warning",
		Source:   "grite-import",
		SourceID: "legacy-event-1",
	})
	require.ErrorIs(t, err, native.ErrEventConflict)
	afterConflict, err := repo.GetIssue(ctx, target.ID)
	require.NoError(t, err)
	assert.Equal(t, target.Version, afterConflict.Version, "duplicate source event must roll back the version increment")

	temporary := createIssue(t, repo, workspace.ID, "Temporary")
	require.NoError(t, repo.DeleteIssue(ctx, temporary.ID, temporary.Version))
	_, err = repo.GetIssue(ctx, temporary.ID)
	require.ErrorIs(t, err, native.ErrNotFound)

	temporaryWorkspace, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/example/temporary"})
	require.NoError(t, err)
	temporaryPaths, err := repo.ListWorkspacePaths(ctx, temporaryWorkspace.ID)
	require.NoError(t, err)
	assert.Empty(t, temporaryPaths)
	require.NoError(t, repo.DeleteWorkspace(ctx, temporaryWorkspace.ID))
	_, err = repo.GetWorkspace(ctx, temporaryWorkspace.ID)
	require.ErrorIs(t, err, native.ErrNotFound)
}

func TestRepositoryConcurrentInvariants(t *testing.T) {
	repo, _ := openRepository(t)
	ctx := t.Context()
	workspace, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/example/concurrency"})
	require.NoError(t, err)

	t.Run("alias uniqueness", func(t *testing.T) {
		first := createIssue(t, repo, workspace.ID, "First alias owner")
		second := createIssue(t, repo, workspace.ID, "Second alias owner")
		errs := runConcurrently(
			func() error {
				_, err := repo.SetAliases(ctx, first.ID, first.Version, []native.AliasInput{{Alias: "shared-ref"}}, "test")
				return err
			},
			func() error {
				_, err := repo.SetAliases(ctx, second.ID, second.Version, []native.AliasInput{{Alias: "shared-ref"}}, "test")
				return err
			},
		)
		assertExactlyOne(t, errs, nil, native.ErrAliasConflict)
	})

	t.Run("dependency cycle", func(t *testing.T) {
		first := createIssue(t, repo, workspace.ID, "First dependency")
		second := createIssue(t, repo, workspace.ID, "Second dependency")
		errs := runConcurrently(
			func() error {
				_, err := repo.AddRelationship(ctx, first.ID, second.ID, native.RelationshipDependsOn, first.Version, "test")
				return err
			},
			func() error {
				_, err := repo.AddRelationship(ctx, second.ID, first.ID, native.RelationshipDependsOn, second.Version, "test")
				return err
			},
		)
		assertExactlyOne(t, errs, nil, native.ErrRelationshipCycle)
	})
}

func openRepository(t *testing.T) (*native.Repository, *gorm.DB) {
	t.Helper()
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native repository tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir:  filepath.Join(t.TempDir(), "postgres"),
		Database: "gavel_native_todos",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })

	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	db, err := database.Open(t.Context())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	repo, err := native.NewRepository(db.Gorm())
	require.NoError(t, err)
	return repo, db.Gorm()
}

func createIssue(t *testing.T, repo *native.Repository, workspaceID uuid.UUID, title string) *native.Issue {
	t.Helper()
	issue, err := repo.CreateIssue(t.Context(), native.CreateIssueInput{WorkspaceID: workspaceID, Title: title})
	require.NoError(t, err)
	return issue
}

func runConcurrently(functions ...func() error) []error {
	start := make(chan struct{})
	errs := make([]error, len(functions))
	var wg sync.WaitGroup
	for i, function := range functions {
		wg.Add(1)
		go func(index int, run func() error) {
			defer wg.Done()
			<-start
			errs[index] = run()
		}(i, function)
	}
	close(start)
	wg.Wait()
	return errs
}

func assertExactlyOne(t *testing.T, errs []error, success, expected error) {
	t.Helper()
	require.Len(t, errs, 2)
	var successes, matches int
	for _, err := range errs {
		if err == success {
			successes++
		}
		if errors.Is(err, expected) {
			matches++
		}
	}
	assert.Equal(t, 1, successes, "errors: %v", errs)
	assert.Equal(t, 1, matches, "errors: %v", errs)
}
