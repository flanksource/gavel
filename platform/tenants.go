// Lists every organization on the platform with its members, repositories, GitHub App
// installations and sandbox runs, for the platform dashboard's tenant panel.
package platform

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Tenant is one organization and how much of gavel it uses.
type Tenant struct {
	ID                  uuid.UUID `json:"id"`
	ClerkOrgID          string    `json:"clerkOrgId"`
	Slug                string    `json:"slug"`
	Name                string    `json:"name"`
	CreatedAt           time.Time `json:"createdAt"`
	Members             int       `json:"members"`
	Repos               int       `json:"repos"`
	PushedRepos         int       `json:"pushedRepos"`
	GitHubInstallations int       `json:"githubInstallations" gorm:"column:github_installations"`
	// SandboxRuns counts the runs created since the window Tenants was asked for.
	SandboxRuns int        `json:"sandboxRuns"`
	LastRunAt   *time.Time `json:"lastRunAt,omitempty"`
}

// Tenants lists the Clerk instance's organizations by name, counting runs created since since.
func (s Store) Tenants(ctx context.Context, since time.Time) ([]Tenant, error) {
	tenants := []Tenant{}
	err := s.DB.WithContext(ctx).Raw(`
		SELECT o.id, o.clerk_org_id, COALESCE(o.slug, '') AS slug, o.name, o.created_at,
			(SELECT count(*) FROM org_memberships m WHERE m.org_id = o.id) AS members,
			(SELECT count(*) FROM repos r WHERE r.org_id = o.id) AS repos,
			(SELECT count(*) FROM repos r WHERE r.org_id = o.id AND r.source = 'push') AS pushed_repos,
			(SELECT count(*) FROM github_installations gi WHERE gi.org_id = o.id) AS github_installations,
			(SELECT count(*) FROM sandbox_runs run WHERE run.org_id = o.id AND run.created_at >= ?) AS sandbox_runs,
			(SELECT max(run.created_at) FROM sandbox_runs run WHERE run.org_id = o.id) AS last_run_at
		FROM orgs o WHERE o.issuer = ? ORDER BY o.name, o.clerk_org_id`, since, s.Issuer).Scan(&tenants).Error
	if err != nil {
		return nil, fmt.Errorf("read the platform's organizations: %w", err)
	}
	return tenants, nil
}
