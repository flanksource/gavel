// Tests that a write refused as a conflict with an existing row is answered without the database
// refusing a statement, so the SQL log records no error for an outcome the caller expects.
package native_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestAConflictingWriteIsRefusedWithoutAFailingStatement(t *testing.T) {
	_, db, _ := openExecutionRepository(t)
	refused := &refusedStatements{Interface: gormlogger.Discard}
	repo, err := native.NewRepository(db.Session(&gorm.Session{Logger: refused}))
	require.NoError(t, err)
	integration, err := native.NewExecutionIntegration(repo)
	require.NoError(t, err)
	ctx := t.Context()
	conflict := func(err, sentinel error, write string) {
		t.Helper()
		require.ErrorIs(t, err, sentinel, write)
		assert.Empty(t, refused.take(), "%s is refused without a failing statement", write)
	}

	workspace, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/acme/conflicts", RootPath: "/workspace/conflicts"})
	require.NoError(t, err)
	other, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/acme/elsewhere", RootPath: "/workspace/elsewhere"})
	require.NoError(t, err)
	_, err = repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/acme/conflicts", RootPath: "/workspace/conflicts-2"})
	conflict(err, native.ErrWorkspaceConflict, "a workspace for a repository that has one")

	issue, err := repo.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Owner", Aliases: []native.AliasInput{{Alias: "jira-1"}}})
	require.NoError(t, err)
	_, err = repo.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Taker", Aliases: []native.AliasInput{{Alias: "jira-1"}}})
	conflict(err, native.ErrAliasConflict, "an issue with an alias its workspace uses")
	target := createIssue(t, repo, workspace.ID, "Target")
	_, err = repo.SetAliases(ctx, target.ID, target.Version, []native.AliasInput{{Alias: "jira-1"}}, "conflicts")
	conflict(err, native.ErrAliasConflict, "an alias its workspace uses")
	mover, err := repo.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: other.ID, Title: "Mover", Aliases: []native.AliasInput{{Alias: "jira-1"}}})
	require.NoError(t, err)
	_, err = repo.MoveIssueWorkspace(ctx, mover.ID, workspace.ID, mover.Version, "conflicts")
	conflict(err, native.ErrAliasConflict, "moving an issue to a workspace that uses its alias")

	issue = linkOnce(t, repo, issue, func(version int64) error {
		_, err := repo.AddRelationship(ctx, issue.ID, target.ID, native.RelationshipDependsOn, version, "conflicts")
		return err
	}, func(err error) { conflict(err, native.ErrRelationshipExists, "a relationship the issue has") })
	issue = linkOnce(t, repo, issue, func(version int64) error {
		_, err := repo.LinkIssuePR(ctx, issue.ID, version, native.PRLinkInput{URL: "https://github.com/acme/conflicts/pull/7",
			Relation: types.PRRelationRelated, Source: types.PRLinkSourceManual})
		return err
	}, func(err error) { conflict(err, native.ErrPRLinkExists, "a PR the issue links") })
	issue = linkOnce(t, repo, issue, func(version int64) error {
		_, err := repo.AppendEvent(ctx, issue.ID, version, native.EventInput{Kind: "migration_warning", Source: "portable-import", SourceID: "event-1"})
		return err
	}, func(err error) { conflict(err, native.ErrEventConflict, "an event its source recorded") })

	run := insertCaptainPromptRun(t, db)
	issue = linkOnce(t, repo, issue, func(version int64) error {
		_, err := repo.LinkPromptRun(ctx, issue.ID, run, native.StepRun, 0, version, "conflicts")
		return err
	}, func(err error) { conflict(err, native.ErrLinkConflict, "a prompt run the issue links") })
	plan := insertCaptainPlan(t, db, run)
	issue = linkOnce(t, repo, issue, func(version int64) error {
		_, err := repo.LinkPlan(ctx, issue.ID, plan, 0, version, "conflicts")
		return err
	}, func(err error) { conflict(err, native.ErrLinkConflict, "a plan the issue links") })

	owner, err := native.LocalOwner()
	require.NoError(t, err)
	_, err = integration.ActivatePromptRun(ctx, native.PromptRunAttachment{IssueID: issue.ID, PromptRunID: insertCaptainPromptRun(t, db),
		StepKind: native.StepRun, Ordinal: 0, ExpectedIssueVersion: issue.Version, Actor: "conflicts", Owner: &owner})
	conflict(err, native.ErrLinkConflict, "activating a prompt run at an ordinal the issue's step has")
	_, err = integration.SelectPlan(ctx, native.PlanAttachment{IssueID: issue.ID, PlanID: insertCaptainPlan(t, db, run),
		Ordinal: 0, ExpectedIssueVersion: issue.Version, Actor: "conflicts"})
	conflict(err, native.ErrLinkConflict, "selecting a plan at an ordinal the issue has")
}

// linkOnce makes the write once at issue's version, then again at the version that left, and
// hands the second attempt's error to refused. It returns issue as the first write left it.
func linkOnce(t *testing.T, repo *native.Repository, issue *native.Issue, write func(version int64) error, refused func(error)) *native.Issue {
	t.Helper()
	require.NoError(t, write(issue.Version))
	issue, err := repo.GetIssue(t.Context(), issue.ID)
	require.NoError(t, err)
	refused(write(issue.Version))
	return issue
}

// refusedStatements records the statements the database refused, passing every trace on.
type refusedStatements struct {
	gormlogger.Interface
	mu   sync.Mutex
	errs []error
}

func (r *refusedStatements) LogMode(gormlogger.LogLevel) gormlogger.Interface { return r }

func (r *refusedStatements) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		r.mu.Lock()
		r.errs = append(r.errs, err)
		r.mu.Unlock()
	}
	r.Interface.Trace(ctx, begin, fc, err)
}

// take returns the refusals recorded since it was last called.
func (r *refusedStatements) take() []error {
	r.mu.Lock()
	defer r.mu.Unlock()
	taken := r.errs
	r.errs = nil
	return taken
}
