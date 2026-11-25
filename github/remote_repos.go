package github

import (
	"errors"
	"fmt"
	"slices"

	"github.com/go-git/go-git/v5"
)

// prRemotes are the remotes whose GitHub repos receive a checkout's pull
// requests: upstream first, because a fork's PRs are opened against it.
var prRemotes = []string{"upstream", "origin"}

// ReposFromRemotes returns the GitHub owner/repo of dir's upstream and origin
// remotes. Remotes that are not on GitHub are skipped. A directory outside any
// git repository has no repos; any other failure to read the repository is an
// error.
func ReposFromRemotes(dir string) ([]string, error) {
	repo, err := git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{DetectDotGit: true, EnableDotGitCommonDir: true})
	if errors.Is(err, git.ErrRepositoryNotExists) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open git repository at %s: %w", dir, err)
	}
	var repos []string
	for _, name := range prRemotes {
		remote, err := repo.Remote(name)
		if errors.Is(err, git.ErrRemoteNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read remote %q of %s: %w", name, dir, err)
		}
		for _, url := range remote.Config().URLs {
			if parsed, err := parseGitHubRepo(url); err == nil && !slices.Contains(repos, parsed) {
				repos = append(repos, parsed)
			}
		}
	}
	return repos, nil
}
