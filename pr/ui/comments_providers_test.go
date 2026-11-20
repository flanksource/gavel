// Tests reaching the repository of a comment's Azure DevOps or GitLab pull request from a local
// project: only one the project's origin is, through the local connection for its provider.
package ui

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/comments"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// localProject registers a project named api whose origin is origin.
func localProject(t *testing.T, origin string) {
	t.Helper()
	dir := t.TempDir()
	out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput()
	require.NoError(t, err, "%s", out)
	out, err = exec.Command("git", "-C", dir, "remote", "add", "origin", origin).CombinedOutput()
	require.NoError(t, err, "%s", out)
	original := projectsPath
	projectsPath = filepath.Join(t.TempDir(), "projects.json")
	t.Cleanup(func() { projectsPath = original })
	data, err := json.Marshal([]Project{{Name: "api", Dir: dir}})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(projectsPath, data, 0o600))
}

func TestALocalProjectsGitLabMergeRequestIsReachedThroughGITLAB_TOKEN(t *testing.T) {
	localProject(t, "https://gitlab.com/acme/api.git")
	t.Setenv("GITLAB_TOKEN", "env-token")

	remote, err := commentConnection(context.Background(), comments.Principal{},
		comments.Target{URI: "https://gitlab.com/acme/api/-/merge_requests/5", ScopeURI: "gavel://project/api"})
	require.NoError(t, err)
	assert.Equal(t, models.ConnectionTypeGitlab, remote.Provider.Type())
	assert.Equal(t, "acme", remote.URL.Owner)
	assert.Equal(t, "api", remote.URL.Name)

	_, err = commentConnection(context.Background(), comments.Principal{},
		comments.Target{URI: "https://gitlab.com/acme/web/-/merge_requests/5", ScopeURI: "gavel://project/api"})
	require.ErrorIs(t, err, comments.ErrForbidden, "a merge request of a repository the project isn't")
	require.ErrorIs(t, (CommentAccess{}).Authorize(context.Background(), comments.Principal{}, "provider",
		comments.Target{URI: "https://gitlab.com/acme/web/-/merge_requests/5", ScopeURI: "gavel://project/api"}), comments.ErrForbidden)
}
