// Tests the GitHub account a todo is pushed to issues through: a hosted organization's todos
// go through its GitHub App to their own repository, and only for the todo's owner.
package ui

import (
	"context"
	"net/http"
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostedTodosPushThroughTheOrgsApp(t *testing.T) {
	creator, member := uuid.New(), uuid.New()
	todo := &types.TODO{ID: uuid.NewString(), CreatedBy: creator.String()}
	appToken := func() (string, error) { return "ghs_app", nil }
	srv := &Server{ghOpts: github.Options{TokenSource: appToken, Repo: "acme/dashboard-search"}}
	actingAs := func(user uuid.UUID) context.Context {
		return tenant.WithScope(context.Background(), tenant.Scope{OrgID: uuid.New(), UserID: &user, Role: "org:member"})
	}
	linked := todoSource{Dir: "/srv/workspaces/org/app", Hosted: &HostedWorkspace{Repo: sandbox.Repo{Name: "acme/app", Source: "github"}}}

	account, err := srv.todoGitHub(actingAs(creator), linked, todo)
	require.NoError(t, err)
	assert.Equal(t, "acme/app", account.Repo, "the workspace's own repository, not its checkout's origin")
	token, err := account.TokenSource()
	require.NoError(t, err)
	assert.Equal(t, "ghs_app", token)

	_, err = srv.todoGitHub(actingAs(member), linked, todo)
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, todoErrorStatus(http.StatusInternalServerError, err), err.Error())

	pushed := todoSource{Dir: "/srv/workspaces/org/tool", Hosted: &HostedWorkspace{Repo: sandbox.Repo{Name: "tool", Source: "push"}}}
	_, err = srv.todoGitHub(actingAs(creator), pushed, todo)
	require.EqualError(t, err, "repository tool was pushed to gavel, not linked from GitHub, so its todos have no GitHub issues")

	local, err := srv.todoGitHub(context.Background(), todoSource{Dir: "/work/app"}, todo)
	require.NoError(t, err)
	assert.Equal(t, "/work/app", local.WorkDir)
	assert.Empty(t, local.Repo, "a local todo's repository is its workspace's origin")
}
