// The connection a local gavel reaches a GitLab project's API through: the git connection owning
// it, else GITLAB_TOKEN for a gitlab.com project. A self-managed GitLab needs a connection.
package gitlab

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/forge"
)

// FindConnection, when set, is the git connection that owns a GitLab project, or nil when none
// does. A local gavel sets it to read its connection files and catalog.
var FindConnection func(context.Context, git.RepoURL) (*models.Connection, error)

// dotCom is the GitLab host GITLAB_TOKEN reaches.
const dotCom = "gitlab.com"

// LocalConnection is the connection a local gavel reaches repo through: the git connection that
// owns it; else, on gitlab.com, one carrying GITLAB_TOKEN.
func LocalConnection(ctx context.Context, repo git.RepoURL) (models.Connection, error) {
	described := fmt.Sprintf("GitLab project https://%s/%s", repo.Host, repo.FullName())
	owner, err := forge.OwningConnection(ctx, FindConnection, repo, described)
	if err != nil {
		return models.Connection{}, err
	}
	if owner != nil {
		return *owner, nil
	}
	if !strings.EqualFold(repo.Host, dotCom) {
		return models.Connection{}, fmt.Errorf("no git connection owns %s: a self-managed GitLab is reached only through a gitlab connection for %s", described, repo.Host)
	}
	token := strings.TrimSpace(os.Getenv("GITLAB_TOKEN"))
	if token == "" {
		return models.Connection{}, fmt.Errorf("no git connection owns %s and GITLAB_TOKEN is not set", described)
	}
	return models.Connection{Name: "GITLAB_TOKEN", Type: models.ConnectionTypeGitlab, URL: "https://" + dotCom, Password: token}, nil
}
