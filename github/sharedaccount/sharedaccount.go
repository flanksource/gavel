// Package sharedaccount serves every hosted organization from one GitHub login, the server's gh
// CLI, with the repositories the server names: a development stand-in for the GitHub App.
package sharedaccount

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/service"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AllRepos, as the only repository name, serves every repository the login can see.
const AllRepos = "all"

// ErrNoLogin is a gh CLI that is missing or not logged in.
var ErrNoLogin = errors.New("the gh CLI has no GitHub login; run gh auth login on the server")

// GHToken is the gh CLI's GitHub token.
func GHToken() (string, error) {
	token, err := service.GHCLIToken()
	if err != nil {
		return "", fmt.Errorf("run gh auth token: %w", err)
	}
	if token == "" {
		return "", ErrNoLogin
	}
	return token, nil
}

// Account is one GitHub login every organization shares.
type Account struct {
	db    *gorm.DB
	names []string
	token func() (string, error)

	mu     sync.Mutex
	served []string // the repositories the last sync found, owner/name
}

// New serves names (owner/name, or just AllRepos) to every organization with token.
func New(db *gorm.DB, names []string, token func() (string, error)) *Account {
	return &Account{db: db, names: names, token: token}
}

// TokenSource is the login's token for every organization.
func (a *Account) TokenSource(uuid.UUID) func() (string, error) { return a.token }

// RepoToken is the login's token for any repository, which no installation grants.
func (a *Account) RepoToken(context.Context, int64) (string, error) { return a.token() }

// SearchGroups is the served repositories, searched together with the login's token.
func (a *Account) SearchGroups(context.Context, uuid.UUID) ([]github.RepoGroup, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.served) == 0 {
		return nil, nil
	}
	return []github.RepoGroup{{Name: "the gh CLI's login", Repos: append([]string(nil), a.served...), Token: a.token}}, nil
}

// SyncAll makes org's repositories without an installation match the served ones. One the
// server no longer serves is deleted, unless a sandbox run refers to it or a git connection
// brings it.
func (a *Account) SyncAll(ctx context.Context, org uuid.UUID) error {
	repos, err := a.resolve(ctx)
	if err != nil {
		return err
	}
	err = a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ids := make([]int64, 0, len(repos))
		for _, repo := range repos {
			ids = append(ids, repo.ID)
			if err := tx.Exec(`
				INSERT INTO repos (org_id, name, source, github_repo_id, default_branch)
				VALUES (?, ?, 'github', ?, ?)
				ON CONFLICT (org_id, github_repo_id) WHERE github_repo_id IS NOT NULL DO UPDATE SET
					name = EXCLUDED.name, default_branch = EXCLUDED.default_branch, updated_at = now()`,
				org, repo.FullName, repo.ID, repo.DefaultBranch).Error; err != nil {
				return fmt.Errorf("record repository %s: %w", repo.FullName, err)
			}
		}
		// A repository a git connection brings is the connection's to remove.
		gone := tx.Where("org_id = ? AND source = 'github' AND installation_id IS NULL AND connection_id IS NULL", org).
			Where("NOT EXISTS (SELECT 1 FROM sandbox_runs WHERE sandbox_runs.repo_id = repos.id)")
		if len(ids) > 0 {
			gone = gone.Where("github_repo_id NOT IN ?", ids)
		}
		if err := gone.Delete(&repoRow{}).Error; err != nil {
			return fmt.Errorf("remove repositories the server no longer serves: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("sync the gh CLI login's repositories into org %s: %w", org, err)
	}
	served := make([]string, 0, len(repos))
	for _, repo := range repos {
		served = append(served, repo.FullName)
	}
	a.mu.Lock()
	a.served = served
	a.mu.Unlock()
	return nil
}

// resolve reads the served repositories from GitHub with the login's token.
func (a *Account) resolve(ctx context.Context) ([]github.InstallationRepo, error) {
	if len(a.names) == 1 && a.names[0] == AllRepos {
		token, err := a.token()
		if err != nil {
			return nil, fmt.Errorf("list the gh CLI login's repositories: %w", err)
		}
		return github.UserRepos(ctx, token)
	}
	repos := make([]github.InstallationRepo, 0, len(a.names))
	for _, name := range a.names {
		repo, err := a.repo(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("read GitHub repository %s with the gh CLI's login: %w", name, err)
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

func (a *Account) repo(ctx context.Context, name string) (github.InstallationRepo, error) {
	token, err := a.token()
	if err != nil {
		return github.InstallationRepo{}, err
	}
	return github.UserRepo(ctx, token, name)
}

// repoRow names the repos table for gorm's query builder.
type repoRow struct{}

func (repoRow) TableName() string { return "repos" }
