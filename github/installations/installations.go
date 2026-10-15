// Package installations records which GitHub App installations each organization has
// linked, keeps the repositories they grant in repos, and mints the org's GitHub tokens.
package installations

import (
	"context"
	"errors"
	"fmt"

	"github.com/flanksource/gavel/github"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrLinkedToAnotherOrg means the installation already belongs to a different organization.
	ErrLinkedToAnotherOrg = errors.New("the GitHub App installation is linked to another organization")
	// ErrNotLinked means the organization never linked the installation it named.
	ErrNotLinked = errors.New("the GitHub App installation isn't linked to this organization")
	// ErrNoInstallation means the organization hasn't installed the GitHub App yet.
	ErrNoInstallation = errors.New("connect GitHub to this organization: install the gavel GitHub App")
)

// Store reads and writes organizations' installations and repositories.
type Store struct {
	DB  *gorm.DB
	App *github.App
}

// Link records that org installed the App as installation, on behalf of user. Linking an
// installation the org already has is a no-op; one linked to another org is refused.
func (s Store) Link(ctx context.Context, org, user uuid.UUID, installation github.Installation) error {
	db := s.DB.WithContext(ctx)
	if err := db.Exec(`
		INSERT INTO github_installations (id, org_id, account_login, account_type, created_by_user_id)
		VALUES (?, ?, ?, ?, ?) ON CONFLICT (id) DO NOTHING`,
		installation.ID, org, installation.Account.Login, installation.Account.Type, user).Error; err != nil {
		return fmt.Errorf("record GitHub App installation %d: %w", installation.ID, err)
	}
	var owner struct{ OrgID uuid.UUID }
	if err := db.Raw(`SELECT org_id FROM github_installations WHERE id = ?`, installation.ID).Scan(&owner).Error; err != nil {
		return fmt.Errorf("read GitHub App installation %d: %w", installation.ID, err)
	}
	if owner.OrgID != org {
		return fmt.Errorf("%w: %s (installation %d)", ErrLinkedToAnotherOrg, installation.Account.Login, installation.ID)
	}
	return nil
}

// Installations lists org's active installations, oldest first.
func (s Store) Installations(ctx context.Context, org uuid.UUID) ([]int64, error) {
	var ids []int64
	if err := s.DB.WithContext(ctx).Raw(`
		SELECT id FROM github_installations WHERE org_id = ? AND suspended_at IS NULL ORDER BY created_at, id`, org).
		Scan(&ids).Error; err != nil {
		return nil, fmt.Errorf("read the organization's GitHub App installations: %w", err)
	}
	return ids, nil
}

// SyncRepos makes repos match what installationID grants org. A repository the
// installation no longer grants is deleted, unless a sandbox run refers to it; then it
// is kept without its installation.
func (s Store) SyncRepos(ctx context.Context, org uuid.UUID, installationID int64) error {
	var linked int64
	if err := s.DB.WithContext(ctx).Raw(`SELECT count(*) FROM github_installations WHERE id = ? AND org_id = ?`,
		installationID, org).Scan(&linked).Error; err != nil {
		return fmt.Errorf("read GitHub App installation %d: %w", installationID, err)
	}
	if linked == 0 {
		return fmt.Errorf("%w (installation %d)", ErrNotLinked, installationID)
	}
	granted, err := s.App.InstallationRepos(ctx, installationID)
	if err != nil {
		return err
	}
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ids := make([]int64, 0, len(granted))
		for _, repo := range granted {
			ids = append(ids, repo.ID)
			if err := tx.Exec(`
				INSERT INTO repos (org_id, name, source, github_repo_id, installation_id, default_branch)
				VALUES (?, ?, 'github', ?, ?, ?)
				ON CONFLICT (github_repo_id) WHERE github_repo_id IS NOT NULL DO UPDATE SET
					name = EXCLUDED.name, installation_id = EXCLUDED.installation_id,
					default_branch = EXCLUDED.default_branch, updated_at = now()
				WHERE repos.org_id = EXCLUDED.org_id`,
				org, repo.FullName, repo.ID, installationID, repo.DefaultBranch).Error; err != nil {
				return fmt.Errorf("record repository %s: %w", repo.FullName, err)
			}
		}
		gone := tx.Where("org_id = ? AND installation_id = ?", org, installationID)
		if len(ids) > 0 {
			gone = gone.Where("github_repo_id NOT IN ?", ids)
		}
		if err := gone.Session(&gorm.Session{}).
			Where("NOT EXISTS (SELECT 1 FROM sandbox_runs WHERE sandbox_runs.repo_id = repos.id)").
			Delete(&repoRow{}).Error; err != nil {
			return fmt.Errorf("remove repositories the installation no longer grants: %w", err)
		}
		if err := gone.Session(&gorm.Session{}).Model(&repoRow{}).
			Updates(map[string]any{"installation_id": nil, "updated_at": gorm.Expr("now()")}).Error; err != nil {
			return fmt.Errorf("unlink repositories the installation no longer grants: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("sync the repositories of GitHub App installation %d: %w", installationID, err)
	}
	return nil
}

// SyncAll syncs the repositories of every installation org linked.
func (s Store) SyncAll(ctx context.Context, org uuid.UUID) error {
	ids, err := s.Installations(ctx, org)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := s.SyncRepos(ctx, org, id); err != nil {
			return err
		}
	}
	return nil
}

// repoRow names the repos table for gorm's query builder.
type repoRow struct{}

func (repoRow) TableName() string { return "repos" }

// Repos lists the full names of the GitHub repositories org's installations grant.
func (s Store) Repos(ctx context.Context, org uuid.UUID) ([]string, error) {
	var names []string
	if err := s.DB.WithContext(ctx).Raw(`
		SELECT name FROM repos WHERE org_id = ? AND source = 'github' AND installation_id IS NOT NULL ORDER BY name`, org).
		Scan(&names).Error; err != nil {
		return nil, fmt.Errorf("read the organization's repositories: %w", err)
	}
	return names, nil
}

// InstallationRepos is the GitHub repositories one installation grants.
type InstallationRepos struct {
	Installation int64
	Repos        []string
}

// ReposByInstallation groups org's GitHub repositories by the installation that grants
// them, oldest installation first, so each group is searched with its own token.
func (s Store) ReposByInstallation(ctx context.Context, org uuid.UUID) ([]InstallationRepos, error) {
	var rows []struct {
		InstallationID int64
		Name           string
	}
	if err := s.DB.WithContext(ctx).Raw(`
		SELECT r.installation_id, r.name FROM repos r
		JOIN github_installations i ON i.id = r.installation_id
		WHERE r.org_id = ? AND r.source = 'github' AND i.suspended_at IS NULL
		ORDER BY i.created_at, i.id, r.name`, org).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("read the organization's repositories: %w", err)
	}
	var groups []InstallationRepos
	for _, row := range rows {
		if len(groups) == 0 || groups[len(groups)-1].Installation != row.InstallationID {
			groups = append(groups, InstallationRepos{Installation: row.InstallationID})
		}
		last := &groups[len(groups)-1]
		last.Repos = append(last.Repos, row.Name)
	}
	return groups, nil
}

// InstallationTokenSource mints tokens for one installation, for github.Options.TokenSource.
func (s Store) InstallationTokenSource(installationID int64) func() (string, error) {
	return func() (string, error) {
		return s.App.InstallationToken(context.Background(), installationID)
	}
}

// TokenSource mints tokens for org's first installation, for github.Options.TokenSource.
func (s Store) TokenSource(org uuid.UUID) func() (string, error) {
	return func() (string, error) {
		ctx := context.Background()
		ids, err := s.Installations(ctx, org)
		if err != nil {
			return "", err
		}
		if len(ids) == 0 {
			return "", ErrNoInstallation
		}
		return s.App.InstallationToken(ctx, ids[0])
	}
}
