// Tests that native todos are isolated per organization: another org can neither see nor
// change a workspace, issue, label, relationship or Captain link it doesn't own.
package native_test

import (
	"errors"
	"testing"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// requireHidden passes when a read of another org's record either reports
// ErrNotFound or returns nothing; both reveal no data.
func requireHidden[T any](t *testing.T, items []T, err error, what string) {
	t.Helper()
	if errors.Is(err, native.ErrNotFound) {
		return
	}
	require.NoError(t, err, what)
	assert.Empty(t, items, what)
}

func openTenancyRepositories(t *testing.T) (local, other *native.Repository, db *gorm.DB, otherOrg uuid.UUID) {
	t.Helper()
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_native_tenancy"}).DSN())
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	otherOrg = uuid.New()
	require.NoError(t, opened.Gorm().Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Other')`, otherOrg, otherOrg.String()).Error)
	local, err = native.NewRepository(opened.Gorm())
	require.NoError(t, err)
	return local, local.ForOrg(otherOrg), opened.Gorm(), otherOrg
}

func TestRepositoryIsolatesOrganizations(t *testing.T) {
	local, other, db, _ := openTenancyRepositories(t)
	ctx := t.Context()
	const actor = "tenancy-test"

	// The local org's state: a workspace with two related issues, an alias, a
	// workspace label and a global label, and a linked prompt run and plan.
	workspace, err := local.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/a/b", RootPath: "/src/local"})
	require.NoError(t, err)
	issue, err := local.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Local issue", Labels: []string{"bug"}, Aliases: []native.AliasInput{{Alias: "alpha-ticket-1", Kind: "ticket"}}})
	require.NoError(t, err)
	target := createIssue(t, local, workspace.ID, "Local dependency")
	issue, err = local.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	_, err = local.AddRelationship(ctx, issue.ID, target.ID, native.RelationshipDependsOn, issue.Version, actor)
	require.NoError(t, err)
	_, err = local.SetLabelDefinition(ctx, native.LabelDefinitionInput{WorkspaceID: &workspace.ID, Name: "bug", Color: "red"})
	require.NoError(t, err)
	_, err = local.SetLabelDefinition(ctx, native.LabelDefinitionInput{Name: "shared", Color: "red"})
	require.NoError(t, err)
	promptRunID := insertCaptainPromptRun(t, db)
	issue, err = local.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	_, err = local.LinkPromptRun(ctx, issue.ID, promptRunID, native.StepRun, 0, issue.Version, actor)
	require.NoError(t, err)
	planID := insertCaptainPlan(t, db, promptRunID)
	issue, err = local.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	_, err = local.LinkPlan(ctx, issue.ID, planID, 0, issue.Version, actor)
	require.NoError(t, err)
	owner, err := native.LocalOwner()
	require.NoError(t, err)
	require.NoError(t, local.ClaimPromptRunOwner(ctx, promptRunID, owner))
	issue, err = local.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	var sessionID string
	require.NoError(t, db.Raw(`SELECT session_id FROM captain_prompt_runs WHERE id = ?`, promptRunID).Scan(&sessionID).Error)
	comment, err := local.AddComment(ctx, issue.ID, issue.Version, native.CommentInput{Actor: actor, Body: "a comment"})
	require.NoError(t, err)
	_, err = local.RecordLanding(ctx, issue.ID, native.RunLanding{PromptRunID: promptRunID, Via: native.LandingMerge, TargetBranch: "main", LandedSHA: "abc123", CommitCount: 1}, actor)
	require.NoError(t, err)
	issue, err = local.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	const prURL = "https://github.com/a/b/pull/1"
	_, err = local.LinkIssuePR(ctx, issue.ID, issue.Version, native.PRLinkInput{URL: prURL, Relation: types.PRRelationRelated, Source: types.PRLinkSourceManual, Actor: actor})
	require.NoError(t, err)
	issue, err = local.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	captain, err := captaindb.Use(db)
	require.NoError(t, err)
	otherCoordinator, err := native.NewLaunchCoordinator(captain, other)
	require.NoError(t, err)

	// Another org can hold the same repo key and global label name.
	otherWorkspace, err := other.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/a/b", RootPath: "/src/other"})
	require.NoError(t, err)
	otherIssue := createIssue(t, other, otherWorkspace.ID, "Other issue")
	_, err = other.SetLabelDefinition(ctx, native.LabelDefinitionInput{Name: "shared", Color: "blue"})
	require.NoError(t, err)

	t.Run("reads", func(t *testing.T) {
		_, err := other.GetWorkspace(ctx, workspace.ID)
		require.ErrorIs(t, err, native.ErrNotFound)
		byKey, err := other.GetWorkspaceByRepoKey(ctx, "github.com/a/b")
		require.NoError(t, err)
		assert.Equal(t, otherWorkspace.ID, byKey.ID)
		_, err = other.GetWorkspaceByPath(ctx, "/src/local")
		require.ErrorIs(t, err, native.ErrNotFound)
		paths, err := other.ListWorkspacePaths(ctx, workspace.ID)
		requireHidden(t, paths, err, "ListWorkspacePaths")
		lookup, err := other.LookupWorkspaces(ctx, []string{"/src/local"}, []string{"github.com/a/b"})
		require.NoError(t, err)
		_, err = lookup.GetWorkspaceByPath(ctx, "/src/local")
		require.ErrorIs(t, err, native.ErrNotFound)
		byKey, err = lookup.GetWorkspaceByRepoKey(ctx, "github.com/a/b")
		require.NoError(t, err)
		assert.Equal(t, otherWorkspace.ID, byKey.ID)

		_, err = other.GetIssue(ctx, issue.ID)
		require.ErrorIs(t, err, native.ErrNotFound)
		issues, err := other.ListIssues(ctx, workspace.ID)
		requireHidden(t, issues, err, "ListIssues")
		_, err = other.GetIssueByRef(ctx, workspace.ID, "alpha-ticket-1")
		require.ErrorIs(t, err, native.ErrNotFound)
		for _, ref := range []string{"alpha-ticket-1", issue.ID.String(), issue.ID.String()[:8]} {
			_, err = other.GetIssueByGlobalRef(ctx, ref)
			require.ErrorIs(t, err, native.ErrNotFound, ref)
		}
		_, _, err = other.GetIssueBySessionRef(ctx, sessionID)
		require.ErrorIs(t, err, native.ErrNotFound)
		counts, err := other.CountIssuesByStatus(ctx, []uuid.UUID{workspace.ID})
		requireHidden(t, counts, err, "CountIssuesByStatus")
		phaseRuns, err := other.ListIssuePhaseRuns(ctx, workspace.ID)
		requireHidden(t, phaseRuns, err, "ListIssuePhaseRuns")

		aliases, err := other.ListAliases(ctx, issue.ID)
		requireHidden(t, aliases, err, "ListAliases")
		byKind, err := other.ListAliasesByKind(ctx, workspace.ID, "ticket")
		if !errors.Is(err, native.ErrNotFound) {
			require.NoError(t, err)
			assert.Empty(t, byKind)
		}
		events, err := other.ListEvents(ctx, issue.ID)
		requireHidden(t, events, err, "ListEvents")

		labels, err := other.ListLabelDefinitions(ctx, workspace.ID)
		if !errors.Is(err, native.ErrNotFound) {
			require.NoError(t, err)
			for _, label := range labels {
				assert.Nil(t, label.WorkspaceID, "only the org's own global labels: %s", label.Name)
				assert.Equal(t, "blue", label.Color, label.Name)
			}
		}
		globals, err := other.ListGlobalLabelDefinitions(ctx)
		require.NoError(t, err)
		require.Len(t, globals, 1)
		assert.Equal(t, "blue", globals[0].Color)
		byLabel, err := other.CountIssuesByLabel(ctx, workspace.ID)
		if !errors.Is(err, native.ErrNotFound) {
			require.NoError(t, err)
			assert.Empty(t, byLabel)
		}

		relationships, err := other.ListRelationships(ctx, issue.ID)
		requireHidden(t, relationships, err, "ListRelationships")
		dependents, err := other.ListDependents(ctx, target.ID)
		requireHidden(t, dependents, err, "ListDependents")
		unsatisfied, err := other.ListUnsatisfiedDependencies(ctx, issue.ID)
		requireHidden(t, unsatisfied, err, "ListUnsatisfiedDependencies")

		runs, err := other.ListPromptRuns(ctx, issue.ID)
		requireHidden(t, runs, err, "ListPromptRuns")
		runLinks, err := other.ListPromptRunLinks(ctx, workspace.ID)
		if !errors.Is(err, native.ErrNotFound) {
			require.NoError(t, err)
			assert.Empty(t, runLinks)
		}
		plans, err := other.ListPlans(ctx, issue.ID)
		requireHidden(t, plans, err, "ListPlans")
		history, err := other.ListIssueRunHistory(ctx, issue.ID, "run_outcome")
		requireHidden(t, history, err, "ListIssueRunHistory")
		_, err = other.PromptRunOwner(ctx, promptRunID)
		require.ErrorIs(t, err, native.ErrNotFound)
		landings, err := other.ListLandings(ctx, issue.ID)
		requireHidden(t, landings, err, "ListLandings")
		byRun, err := other.ListLandingsForRuns(ctx, []uuid.UUID{promptRunID})
		require.NoError(t, err)
		assert.Empty(t, byRun, "ListLandingsForRuns")
		prLinks, err := other.ListIssuePRLinks(ctx, issue.ID)
		requireHidden(t, prLinks, err, "ListIssuePRLinks")
		linkedIssues, err := other.ListPRLinkedIssues(ctx, native.PRLinkedIssuesQuery{URL: prURL, WorkspaceID: workspace.ID})
		requireHidden(t, linkedIssues, err, "ListPRLinkedIssues")
		allLinkedIssues, err := other.ListPRLinkedIssues(ctx, native.PRLinkedIssuesQuery{URL: prURL})
		requireHidden(t, allLinkedIssues, err, "ListPRLinkedIssues across workspaces")
	})

	t.Run("writes", func(t *testing.T) {
		title := "Hijacked"
		displayName := "Hijacked"
		for name, write := range map[string]func() error{
			"UpdateWorkspace": func() error {
				_, err := other.UpdateWorkspace(ctx, workspace.ID, native.UpdateWorkspaceInput{DisplayName: &displayName})
				return err
			},
			"DeleteWorkspace": func() error { return other.DeleteWorkspace(ctx, workspace.ID) },
			"CreateIssue": func() error {
				_, err := other.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: title})
				return err
			},
			"MoveIssueWorkspace out": func() error {
				_, err := other.MoveIssueWorkspace(ctx, issue.ID, otherWorkspace.ID, issue.Version, actor)
				return err
			},
			"MoveIssueWorkspace in": func() error {
				_, err := other.MoveIssueWorkspace(ctx, otherIssue.ID, workspace.ID, otherIssue.Version, actor)
				return err
			},
			"UpdateIssue": func() error {
				_, err := other.UpdateIssue(ctx, issue.ID, issue.Version, native.IssuePatch{Title: &title})
				return err
			},
			"DeleteIssue": func() error { return other.DeleteIssue(ctx, target.ID, target.Version) },
			"SetAliases": func() error {
				_, err := other.SetAliases(ctx, issue.ID, issue.Version, []native.AliasInput{{Alias: "stolen"}}, actor)
				return err
			},
			"AppendEvent": func() error {
				_, err := other.AppendEvent(ctx, issue.ID, issue.Version, native.EventInput{Kind: "comment", Body: title})
				return err
			},
			"AddComment": func() error {
				_, err := other.AddComment(ctx, issue.ID, issue.Version, native.CommentInput{Actor: actor, Body: title})
				return err
			},
			"ResolveComment": func() error {
				_, err := other.ResolveComment(ctx, issue.ID, issue.Version, actor, comment.ID, true)
				return err
			},
			"SetIssueParent": func() error {
				_, err := other.SetIssueParent(ctx, target.ID, &issue.ID, target.Version, actor)
				return err
			},
			"SetIssueParent to another org's issue": func() error {
				_, err := other.SetIssueParent(ctx, otherIssue.ID, &issue.ID, otherIssue.Version, actor)
				return err
			},
			"CreateIssue under another org's parent": func() error {
				_, err := other.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: otherWorkspace.ID, Title: title, ParentID: &issue.ID})
				return err
			},
			"RecordLanding": func() error {
				_, err := other.RecordLanding(ctx, issue.ID, native.RunLanding{PromptRunID: insertCaptainPromptRun(t, db), Via: native.LandingMerge, TargetBranch: "main", LandedSHA: "def456", CommitCount: 1}, actor)
				return err
			},
			"RecordAskAnswer": func() error {
				_, err := otherCoordinator.RecordAskAnswer(ctx, native.AskAnswerInput{IssueID: issue.ID, PromptRunID: promptRunID, ExpectedRunVersion: 1, Source: "test", Body: title})
				return err
			},
			"PersistAndSelectPlan": func() error {
				_, err := otherCoordinator.PersistAndSelectPlan(ctx, native.PersistPlanInput{
					Plan:       captaindb.CreatePlanInput{SourceSessionID: uuid.MustParse(sessionID), Title: title},
					Revision:   captaindb.AppendPlanRevisionInput{PlanMarkdown: "# Hijacked"},
					Attachment: native.PlanSelectionAttachment{IssueID: issue.ID, Ordinal: 1, ExpectedIssueVersion: issue.Version, Actor: actor},
				})
				return err
			},
			"ReviewPlan": func() error {
				_, err := otherCoordinator.ReviewPlan(ctx,
					captaindb.SetPlanReviewStateInput{PlanID: planID, State: captaindb.PlanApprovalRejected, Actor: actor},
					native.PlanSelectionAttachment{IssueID: issue.ID, ExpectedIssueVersion: issue.Version, Actor: actor})
				return err
			},
			"SetLabelDefinition": func() error {
				_, err := other.SetLabelDefinition(ctx, native.LabelDefinitionInput{WorkspaceID: &workspace.ID, Name: "bug", Color: "blue"})
				return err
			},
			"DeleteLabelDefinition": func() error {
				_, err := other.DeleteLabelDefinition(ctx, &workspace.ID, "bug")
				return err
			},
			"AddRelationship": func() error {
				_, err := other.AddRelationship(ctx, target.ID, issue.ID, native.RelationshipRelatedTo, target.Version, actor)
				return err
			},
			"AddRelationship across orgs": func() error {
				_, err := other.AddRelationship(ctx, otherIssue.ID, issue.ID, native.RelationshipRelatedTo, otherIssue.Version, actor)
				return err
			},
			"DeleteRelationship": func() error {
				return other.DeleteRelationship(ctx, issue.ID, target.ID, native.RelationshipDependsOn, issue.Version, actor)
			},
			"LinkPromptRun": func() error {
				_, err := other.LinkPromptRun(ctx, issue.ID, insertCaptainPromptRun(t, db), native.StepRun, 1, issue.Version, actor)
				return err
			},
			"SetActivePromptRun": func() error {
				_, err := other.SetActivePromptRun(ctx, issue.ID, &promptRunID, issue.Version, actor)
				return err
			},
			"UnlinkPromptRun": func() error {
				_, err := other.UnlinkPromptRun(ctx, issue.ID, promptRunID, issue.Version, actor)
				return err
			},
			"LinkPlan": func() error {
				_, err := other.LinkPlan(ctx, issue.ID, insertCaptainPlan(t, db, promptRunID), 1, issue.Version, actor)
				return err
			},
			"SelectPlan": func() error {
				_, err := other.SelectPlan(ctx, issue.ID, &planID, issue.Version, actor)
				return err
			},
			"UnlinkPlan": func() error {
				_, err := other.UnlinkPlan(ctx, issue.ID, planID, issue.Version, actor)
				return err
			},
			"ClaimPromptRunOwner": func() error {
				return other.ClaimPromptRunOwner(ctx, promptRunID, native.RunOwner{HostID: "other", PID: 1, StartedAt: time.Now(), Token: uuid.New()})
			},
			"ActivatePromptRun": func() error {
				execution, err := native.NewExecutionIntegration(other)
				require.NoError(t, err)
				_, err = execution.ActivatePromptRun(ctx, native.PromptRunAttachment{IssueID: issue.ID, PromptRunID: insertCaptainPromptRun(t, db), StepKind: native.StepRun, Ordinal: 2, ExpectedIssueVersion: issue.Version, Actor: actor})
				return err
			},
			"LinkIssuePR": func() error {
				_, err := other.LinkIssuePR(ctx, issue.ID, issue.Version, native.PRLinkInput{URL: "https://github.com/a/b/pull/2", Relation: types.PRRelationRelated, Source: types.PRLinkSourceManual, Actor: actor})
				return err
			},
			"UnlinkIssuePR": func() error {
				return other.UnlinkIssuePR(ctx, issue.ID, issue.Version, prURL, actor)
			},
			"ExecutionSelectPlan": func() error {
				execution, err := native.NewExecutionIntegration(other)
				require.NoError(t, err)
				_, err = execution.SelectPlan(ctx, native.PlanAttachment{IssueID: issue.ID, PlanID: planID, ExpectedIssueVersion: issue.Version, Actor: actor})
				return err
			},
		} {
			require.ErrorIs(t, write(), native.ErrNotFound, name)
		}
		touched, err := other.TouchPromptRunOwner(ctx, promptRunID, owner.Token)
		require.NoError(t, err)
		assert.False(t, touched, "another org cannot refresh this org's claim")
		require.NoError(t, other.ReleasePromptRunOwner(ctx, promptRunID, owner.Token))

		after, err := local.GetIssue(ctx, issue.ID)
		require.NoError(t, err)
		assert.Equal(t, issue.Version, after.Version, "another org's writes left the issue untouched")
		assert.Equal(t, "Local issue", after.Title)
		assert.Nil(t, after.ParentID)
		var hijackedPlans int64
		require.NoError(t, db.Raw(`SELECT COUNT(*) FROM captain_plans WHERE title = ?`, title).Scan(&hijackedPlans).Error)
		assert.Zero(t, hijackedPlans, "a rejected link rolls back Captain's plan write")
		var planState string
		require.NoError(t, db.Raw(`SELECT approval_state::text FROM captain_plans WHERE id = ?`, planID).Scan(&planState).Error)
		assert.NotEqual(t, string(captaindb.PlanApprovalRejected), planState, "a rejected link rolls back Captain's review")
		claim, err := local.PromptRunOwner(ctx, promptRunID)
		require.NoError(t, err)
		require.NotNil(t, claim, "another org cannot release this org's claim")
		assert.Equal(t, owner.Token, claim.Token)
		got, err := local.GetWorkspace(ctx, workspace.ID)
		require.NoError(t, err)
		assert.NotEqual(t, displayName, got.DisplayName)
		labels, err := local.ListLabelDefinitions(ctx, workspace.ID)
		require.NoError(t, err)
		require.Len(t, labels, 2, "the workspace label plus this org's global label")
		for _, label := range labels {
			assert.Equal(t, "red", label.Color, label.Name)
		}
		globals, err := local.ListGlobalLabelDefinitions(ctx)
		require.NoError(t, err)
		require.Len(t, globals, 1)
		assert.Equal(t, "red", globals[0].Color)
	})
}

func TestRepositoryAttributesWritesToTheRequestUser(t *testing.T) {
	local, _, db, _ := openTenancyRepositories(t)
	user := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, user, user.String()).Error)
	ctx := tenant.WithScope(t.Context(), tenant.Scope{OrgID: tenant.LocalOrgID, UserID: &user})
	const actor = "attribution-test"

	workspace, err := local.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/a/attributed"})
	require.NoError(t, err)
	issue, err := local.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Attributed"})
	require.NoError(t, err)
	_, err = local.AddComment(ctx, issue.ID, issue.Version, native.CommentInput{Actor: actor, Body: "a comment"})
	require.NoError(t, err)
	_, err = local.SetLabelDefinition(ctx, native.LabelDefinitionInput{Name: "attributed", Color: "red"})
	require.NoError(t, err)
	promptRunID := insertCaptainPromptRun(t, db)
	issue, err = local.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	_, err = local.LinkPromptRun(ctx, issue.ID, promptRunID, native.StepRun, 0, issue.Version, actor)
	require.NoError(t, err)
	issue, err = local.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	_, err = local.LinkPlan(ctx, issue.ID, insertCaptainPlan(t, db, promptRunID), 0, issue.Version, actor)
	require.NoError(t, err)

	for query, args := range map[string][]any{
		`SELECT COALESCE(user_id::text, '') FROM todo_workspaces WHERE id = ?`:                           {workspace.ID},
		`SELECT COALESCE(created_by_user_id::text, '') FROM todo_issues WHERE id = ?`:                    {issue.ID},
		`SELECT COALESCE(updated_by_user_id::text, '') FROM todo_labels WHERE name = 'attributed'`:       nil,
		`SELECT COALESCE(requested_by_user_id::text, '') FROM todo_issue_prompt_runs WHERE issue_id = ?`: {issue.ID},
		`SELECT COALESCE(created_by_user_id::text, '') FROM todo_issue_plans WHERE issue_id = ?`:         {issue.ID},
		`SELECT DISTINCT COALESCE(actor_user_id::text, '') FROM todo_issue_events WHERE issue_id = ?`:    {issue.ID},
	} {
		var recorded []string
		require.NoError(t, db.Raw(query, args...).Scan(&recorded).Error, query)
		assert.Equal(t, []string{user.String()}, recorded, query)
	}

	// Without a user on the context (local mode, system jobs) attribution stays NULL.
	anonymous, err := local.CreateIssue(t.Context(), native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Anonymous"})
	require.NoError(t, err)
	var creators []string
	require.NoError(t, db.Raw(`SELECT COALESCE(created_by_user_id::text, 'NULL') FROM todo_issues WHERE id = ?`, anonymous.ID).Scan(&creators).Error)
	assert.Equal(t, []string{"NULL"}, creators)
}

func TestCaptainSessionsRecordTheirOwner(t *testing.T) {
	_, other, db, otherOrg := openTenancyRepositories(t)
	user := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, user, user.String()).Error)
	ctx := tenant.WithScope(t.Context(), tenant.Scope{OrgID: otherOrg, UserID: &user})
	workspace, err := other.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/a/sessions"})
	require.NoError(t, err)
	captain, err := captaindb.Use(db)
	require.NoError(t, err)
	coordinator, err := native.NewLaunchCoordinator(captain, other)
	require.NoError(t, err)

	issueID := uuid.New()
	created, err := coordinator.CreateIssueWithSession(ctx, native.CreateIssueSessionInput{
		Issue:       native.CreateIssueInput{ID: issueID, WorkspaceID: workspace.ID, Title: "With session"},
		RootSession: captaindb.CreateSessionInput{ID: issueID, Source: "gavel", Provider: "todos", Title: "With session"},
	})
	require.NoError(t, err)

	var owner struct {
		OrgID  string
		UserID string
	}
	require.NoError(t, db.Raw(`SELECT org_id::text AS org_id, COALESCE(user_id::text, '') AS user_id FROM session_owners WHERE session_id = ?`, created.Root.ID).Scan(&owner).Error)
	assert.Equal(t, otherOrg.String(), owner.OrgID)
	assert.Equal(t, user.String(), owner.UserID)

	operationID := uuid.New()
	renamed := "Renamed"
	_, err = coordinator.UpdateIssueWithSession(ctx, native.UpdateIssueSessionInput{
		IssueID: issueID, ExpectedIssueVersion: created.Issue.Version, Patch: native.IssuePatch{Title: &renamed},
		RootSession: captaindb.CreateSessionInput{ID: issueID, Source: "gavel", Provider: "todos", Title: "With session"},
		Session:     captaindb.CreateSessionInput{ID: operationID, ParentSessionID: &issueID, Source: "gavel", Provider: "todos", Title: "Rename"},
	})
	require.NoError(t, err, "recording an existing session's owner is a no-op")

	planSessionID := uuid.New()
	issue, err := other.GetIssue(ctx, issueID)
	require.NoError(t, err)
	_, err = coordinator.PersistAndSelectPlan(ctx, native.PersistPlanInput{
		Plan:        captaindb.CreatePlanInput{Title: "Plan"},
		Revision:    captaindb.AppendPlanRevisionInput{PlanMarkdown: "# Plan"},
		Attachment:  native.PlanSelectionAttachment{IssueID: issueID, ExpectedIssueVersion: issue.Version},
		RootSession: &captaindb.CreateSessionInput{ID: issueID, Source: "gavel", Provider: "todos", Title: "With session"},
		Session:     &captaindb.CreateSessionInput{ID: planSessionID, ParentSessionID: &issueID, Source: "gavel", Provider: "todos", Title: "Plan"},
	})
	require.NoError(t, err)

	var owners []string
	require.NoError(t, db.Raw(`SELECT org_id::text FROM session_owners WHERE session_id IN ?`, []uuid.UUID{issueID, operationID, planSessionID}).Scan(&owners).Error)
	assert.Equal(t, []string{otherOrg.String(), otherOrg.String(), otherOrg.String()}, owners)
}
