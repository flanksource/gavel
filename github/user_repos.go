// Reads the GitHub repositories a user token reaches directly, without a GitHub App: every one it
// can see, or one by name.
package github

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// RepoGroup is repositories searched together with one token, from the account Name describes.
type RepoGroup struct {
	Name  string
	Repos []string
	Token func() (string, error)
}

var userReposHTTP = &http.Client{Timeout: 20 * time.Second}

// UserRepos lists every repository token can see, page by page.
func UserRepos(ctx context.Context, token string) ([]InstallationRepo, error) {
	var repos []InstallationRepo
	for page := 1; ; page++ {
		var batch []InstallationRepo
		if err := callAPI(ctx, userReposHTTP, http.MethodGet, "/user/repos?per_page=100&page="+strconv.Itoa(page), token, &batch); err != nil {
			return nil, fmt.Errorf("list the repositories the token can see: %w", err)
		}
		if len(batch) == 0 {
			return repos, nil
		}
		repos = append(repos, batch...)
	}
}

// UserRepo reads the repository fullName (owner/name) with token.
func UserRepo(ctx context.Context, token, fullName string) (InstallationRepo, error) {
	var repo InstallationRepo
	if err := callAPI(ctx, userReposHTTP, http.MethodGet, "/repos/"+fullName, token, &repo); err != nil {
		return InstallationRepo{}, err
	}
	return repo, nil
}
