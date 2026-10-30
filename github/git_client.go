package github

import (
	"context"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
)

// GitClient shares dashboard authentication with the provider-neutral git API.
func GitClient(ctx context.Context, options Options) (*git.Client, error) {
	token, err := options.token()
	if err != nil {
		return nil, err
	}
	return git.FromConnection(dbcontext.NewContext(ctx), models.Connection{Type: models.ConnectionTypeGithub, URL: githubAPIBase(), Password: token})
}
