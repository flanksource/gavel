// Tests hosted todo visibility: a todo is private to its creator by default, a public todo
// is seen by the whole organization, and only its creator or an org admin may change it.
package native_test

import (
	"context"
	"testing"

	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func insertUser(t *testing.T, db *gorm.DB, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
	return id
}

func actingAs(ctx context.Context, org uuid.UUID, user uuid.UUID, role string) context.Context {
	return tenant.WithScope(ctx, tenant.Scope{OrgID: org, UserID: &user, Role: role})
}

func issueIDs(issues []native.Issue) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(issues))
	for _, issue := range issues {
		ids = append(ids, issue.ID)
	}
	return ids
}

func TestHostedTodoVisibility(t *testing.T) {
	_, repo, db, org := openTenancyRepositories(t)
	creator, member, admin := insertUser(t, db, "ada"), insertUser(t, db, "grace"), insertUser(t, db, "linus")
	asCreator := actingAs(t.Context(), org, creator, "org:member")
	asMember := actingAs(t.Context(), org, member, "org:member")
	asAdmin := actingAs(t.Context(), org, admin, "org:admin")

	workspace, err := repo.CreateWorkspace(asCreator, native.CreateWorkspaceInput{RepoKey: "github.com/acme/app", RootPath: "/srv/gavel/acme/app"})
	require.NoError(t, err)
	private, err := repo.CreateIssue(asCreator, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Private by default"})
	require.NoError(t, err)
	public, err := repo.CreateIssue(asCreator, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Shared", Visibility: native.VisibilityPublic})
	require.NoError(t, err)

	assert.Equal(t, native.VisibilityPrivate, private.Visibility)
	assert.Equal(t, native.VisibilityPublic, public.Visibility)
	require.NotNil(t, private.CreatedBy)
	assert.Equal(t, creator, *private.CreatedBy)

	t.Run("the creator sees both", func(t *testing.T) {
		issues, err := repo.ListIssues(asCreator, workspace.ID)
		require.NoError(t, err)
		assert.ElementsMatch(t, []uuid.UUID{private.ID, public.ID}, issueIDs(issues))
	})

	t.Run("another member and an admin see only the public one", func(t *testing.T) {
		for name, ctx := range map[string]context.Context{"member": asMember, "admin": asAdmin} {
			issues, err := repo.ListIssues(ctx, workspace.ID)
			require.NoError(t, err, name)
			assert.Equal(t, []uuid.UUID{public.ID}, issueIDs(issues), name)

			_, err = repo.GetIssue(ctx, private.ID)
			assert.ErrorIs(t, err, native.ErrNotFound, name)
			_, err = repo.GetIssueByGlobalRef(ctx, private.ID.String())
			assert.ErrorIs(t, err, native.ErrNotFound, name)

			counts, err := repo.CountIssuesByStatus(ctx, []uuid.UUID{workspace.ID})
			require.NoError(t, err, name)
			total := 0
			for _, count := range counts {
				total += count.Count
			}
			assert.Equal(t, 1, total, "%s counts only what it can see", name)
		}
	})

	t.Run("lists that name other issues leave out the ones a member can't see", func(t *testing.T) {
		const prURL = "https://github.com/acme/app/pull/7"
		linkedPrivate, err := repo.CreateIssue(asCreator, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Private, linked", Labels: []string{"secret"}, PR: &types.PR{Number: 7, URL: prURL}})
		require.NoError(t, err)
		linkedPublic, err := repo.CreateIssue(asCreator, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Public, linked", Visibility: native.VisibilityPublic, PR: &types.PR{Number: 7, URL: prURL}})
		require.NoError(t, err)
		_, err = repo.AddRelationship(asCreator, linkedPrivate.ID, linkedPublic.ID, native.RelationshipDependsOn, linkedPrivate.Version, "test")
		require.NoError(t, err)

		linked, err := repo.ListPRLinkedIssues(asMember, native.PRLinkedIssuesQuery{URL: prURL, WorkspaceID: workspace.ID})
		require.NoError(t, err)
		require.Len(t, linked, 1)
		assert.Equal(t, linkedPublic.ID, linked[0].Issue.ID)

		dependents, err := repo.ListDependents(asMember, linkedPublic.ID)
		require.NoError(t, err)
		assert.Empty(t, dependents)
		dependents, err = repo.ListDependents(asCreator, linkedPublic.ID)
		require.NoError(t, err)
		assert.Equal(t, []uuid.UUID{linkedPrivate.ID}, issueIDs(dependents))

		labelCounts, err := repo.CountIssuesByLabel(asMember, workspace.ID)
		require.NoError(t, err)
		assert.NotContains(t, labelCounts, "secret")
	})

	t.Run("only the creator or an admin changes a public todo", func(t *testing.T) {
		title := "Edited by a member"
		_, err := repo.UpdateIssue(asMember, public.ID, public.Version, native.IssuePatch{Title: &title})
		require.ErrorIs(t, err, native.ErrForbidden)
		assert.EqualError(t, err, "native todo change forbidden: issue "+public.ID.String()+" belongs to another user; only its creator or an org admin can change it")
		require.ErrorIs(t, repo.DeleteIssue(asMember, public.ID, public.Version), native.ErrForbidden)

		title = "Edited by an admin"
		updated, err := repo.UpdateIssue(asAdmin, public.ID, public.Version, native.IssuePatch{Title: &title})
		require.NoError(t, err)
		assert.Equal(t, title, updated.Title)
	})

	t.Run("the creator makes a private todo public", func(t *testing.T) {
		visibility := native.VisibilityPublic
		updated, err := repo.UpdateIssue(asCreator, private.ID, private.Version, native.IssuePatch{Visibility: &visibility})
		require.NoError(t, err)
		assert.Equal(t, native.VisibilityPublic, updated.Visibility)

		seen, err := repo.GetIssue(asMember, private.ID)
		require.NoError(t, err)
		assert.Equal(t, private.ID, seen.ID)
	})

	t.Run("an unknown visibility is refused", func(t *testing.T) {
		_, err := repo.CreateIssue(asCreator, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Odd", Visibility: "secret"})
		require.ErrorIs(t, err, native.ErrInvalidInput)
		assert.EqualError(t, err, `invalid native todo input: visibility "secret" is not private or public`)
	})

	t.Run("a scope without a user, as in local mode and system jobs, sees and changes everything", func(t *testing.T) {
		system := tenant.WithScope(t.Context(), tenant.Scope{OrgID: org})
		created, err := repo.CreateIssue(asCreator, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Still private"})
		require.NoError(t, err)
		issues, err := repo.ListIssues(system, workspace.ID)
		require.NoError(t, err)
		assert.Contains(t, issueIDs(issues), created.ID)
		title := "Changed by a system job"
		_, err = repo.UpdateIssue(system, created.ID, created.Version, native.IssuePatch{Title: &title})
		require.NoError(t, err)
	})
}
