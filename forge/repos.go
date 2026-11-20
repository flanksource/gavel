// Syncs the repositories an organization's git connections bring into its repos table: every
// repository in each connection's namespace, followed across renames by the provider's id.
package forge

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/flanksource/commons-db/connection/catalog"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/connections"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SyncRepos makes scope's repositories from git connections match what each connection lists.
// A connection is used as stored, never hydrated, so a hosted organization's connection can't
// name the server's own secrets; options configures the provider clients, which a hosted
// server makes Defensive. One connection failing is reported, joined, without stopping the rest.
func SyncRepos(ctx context.Context, db *gorm.DB, scope tenant.Scope, options git.ClientOptions) error {
	var records []catalog.Record
	if err := db.WithContext(ctx).Where("org_id = ? AND type IN ?", scope.OrgID, GitTypes).Order("namespace, name").Find(&records).Error; err != nil {
		return fmt.Errorf("list the organization's git connections: %w", err)
	}
	if len(records) == 0 {
		return nil
	}
	store, closeStore, err := connections.OpenStore(ctx, connections.Options{DB: db, Scope: scope})
	if err != nil {
		return fmt.Errorf("open the organization's connection catalog: %w", err)
	}
	var errs []error
	for _, record := range records {
		connection, err := store.ResolveConnection(dbcontext.NewContext(ctx), record.ID.String())
		if err != nil {
			errs = append(errs, fmt.Errorf("read connection %q: %w", record.Name, err))
			continue
		}
		if err := syncConnection(ctx, db, scope.OrgID, *connection, options); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(append(errs, closeStore())...)
}

// syncConnection records the repositories connection lists and prunes those it no longer does:
// a repository a sandbox run refers to, or its installation also grants, is kept without it.
func syncConnection(ctx context.Context, db *gorm.DB, org uuid.UUID, connection models.Connection, options git.ClientOptions) error {
	client, err := git.FromConnection(dbcontext.NewContext(ctx), connection, options)
	if err != nil {
		return fmt.Errorf("connection %q: %w", connection.Name, err)
	}
	listed, err := client.Provider.ListRepositories(ctx, git.ListRepositoriesOptions{})
	if err != nil {
		return fmt.Errorf("list the repositories of connection %q: %w", connection.Name, err)
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ids := make([]string, 0, len(listed))
		for _, repo := range listed {
			id := cmp.Or(repo.ID, repo.CloneURL)
			ids = append(ids, id)
			if err := recordRepo(tx, org, connection, id, repo); err != nil {
				return fmt.Errorf("record repository %s of connection %q: %w", repo.FullName, connection.Name, err)
			}
		}
		gone := tx.Model(&repoRow{}).Where("org_id = ? AND connection_id = ?", org, connection.ID)
		if len(ids) > 0 {
			gone = gone.Where("provider_repo_id NOT IN ?", ids)
		}
		if err := releaseRepos(gone); err != nil {
			return fmt.Errorf("release the repositories connection %q no longer lists: %w", connection.Name, err)
		}
		return nil
	})
}

// Remote is the remote at remoteURL of a repository scope's connection brings, with that
// connection's credentials, as stored and never hydrated. Another organization's connection is
// not found, and the connection reaches only the repositories its scope owns.
func Remote(ctx context.Context, db *gorm.DB, scope tenant.Scope, connectionID uuid.UUID, remoteURL string, options git.ClientOptions) (*git.Repository, error) {
	store, closeStore, err := connections.OpenStore(ctx, connections.Options{DB: db, Scope: scope})
	if err != nil {
		return nil, fmt.Errorf("open the organization's connection catalog: %w", err)
	}
	client, err := connectionClient(ctx, store, connectionID, options)
	if err := errors.Join(err, closeStore()); err != nil {
		return nil, fmt.Errorf("read the connection of %s: %w", remoteURL, err)
	}
	remote, err := client.Repository(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("connection %q: %w", client.Name, err)
	}
	return remote, nil
}

// namedClient is a connection's provider client, with the connection's name for errors.
type namedClient struct {
	*git.Client
	Name string
}

// connectionClient is the provider client of store's connection id, as stored and never hydrated.
func connectionClient(ctx context.Context, store *catalog.Store, id uuid.UUID, options git.ClientOptions) (namedClient, error) {
	connection, err := store.ResolveConnection(dbcontext.NewContext(ctx), id.String())
	if err != nil {
		return namedClient{}, err
	}
	client, err := git.FromConnection(dbcontext.NewContext(ctx), *connection, options)
	if err != nil {
		return namedClient{}, fmt.Errorf("connection %q: %w", connection.Name, err)
	}
	return namedClient{Client: client, Name: connection.Name}, nil
}

// DeleteConnection deletes scope's connection reference and, in the same transaction, the
// repositories it brings: one a sandbox run refers to, or its installation grants, is kept
// without the connection.
func DeleteConnection(ctx context.Context, db *gorm.DB, scope tenant.Scope, reference string) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		store, closeStore, err := connections.OpenStore(ctx, connections.Options{DB: tx, Scope: scope})
		if err != nil {
			return fmt.Errorf("open the organization's connection catalog: %w", err)
		}
		connection, err := store.ResolveConnection(dbcontext.NewContext(ctx), reference)
		if err != nil {
			return errors.Join(err, closeStore())
		}
		if err := releaseRepos(tx.Model(&repoRow{}).Where("org_id = ? AND connection_id = ?", scope.OrgID, connection.ID)); err != nil {
			return errors.Join(fmt.Errorf("release the repositories of connection %q: %w", connection.Name, err), closeStore())
		}
		if err := store.DeleteConnection(ctx, connection.ID.String()); err != nil {
			return errors.Join(fmt.Errorf("delete connection %q: %w", connection.Name, err), closeStore())
		}
		return closeStore()
	})
}

// releaseRepos lets go of the repositories selected: one a sandbox run refers to, or its
// installation grants, is kept without its connection; the rest are deleted.
func releaseRepos(selected *gorm.DB) error {
	if err := selected.Session(&gorm.Session{}).Where("installation_id IS NULL").
		Where("NOT EXISTS (SELECT 1 FROM sandbox_runs WHERE sandbox_runs.repo_id = repos.id)").
		Delete(&repoRow{}).Error; err != nil {
		return fmt.Errorf("remove them: %w", err)
	}
	if err := selected.Session(&gorm.Session{}).Updates(map[string]any{"connection_id": nil, "updated_at": gorm.Expr("now()")}).Error; err != nil {
		return fmt.Errorf("unlink the ones kept: %w", err)
	}
	return nil
}

// repoRow names the repos table for gorm's query builder.
type repoRow struct{}

func (repoRow) TableName() string { return "repos" }

// recordRepo upserts one listed repository. A GitHub one shares the organization's row for its
// GitHub id with the App's installation; any other is keyed by its connection and provider id.
func recordRepo(tx *gorm.DB, org uuid.UUID, connection models.Connection, id string, repo git.RepositoryInfo) error {
	branch := cmp.Or(repo.DefaultBranch, "main")
	if connection.Type == models.ConnectionTypeGithub {
		githubID, err := strconv.ParseInt(repo.ID, 10, 64)
		if err != nil {
			return fmt.Errorf("GitHub repository id %q: %w", repo.ID, err)
		}
		return tx.Exec(`
			INSERT INTO repos (org_id, name, source, github_repo_id, connection_id, remote_url, provider_repo_id, default_branch)
			VALUES (?, ?, 'github', ?, ?, ?, ?, ?)
			ON CONFLICT (org_id, github_repo_id) WHERE github_repo_id IS NOT NULL DO UPDATE SET
				name = EXCLUDED.name, connection_id = EXCLUDED.connection_id, remote_url = EXCLUDED.remote_url,
				provider_repo_id = EXCLUDED.provider_repo_id, default_branch = EXCLUDED.default_branch, updated_at = now()`,
			org, repo.FullName, githubID, connection.ID, repo.CloneURL, id, branch).Error
	}
	return tx.Exec(`
		INSERT INTO repos (org_id, name, source, connection_id, remote_url, provider_repo_id, default_branch)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (org_id, connection_id, provider_repo_id) WHERE connection_id IS NOT NULL DO UPDATE SET
			name = EXCLUDED.name, remote_url = EXCLUDED.remote_url, default_branch = EXCLUDED.default_branch, updated_at = now()`,
		org, repo.FullName, connection.Type, connection.ID, repo.CloneURL, id, branch).Error
}
