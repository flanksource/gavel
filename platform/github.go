// Reads which GitHub App installations the platform's organizations have linked, and the
// repositories each brought, for the platform dashboard's GitHub panel.
package platform

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// LinkedInstallation is a GitHub App installation an organization linked.
type LinkedInstallation struct {
	ID           int64     `json:"id"`
	OrgID        uuid.UUID `json:"orgId"`
	OrgSlug      string    `json:"orgSlug"`
	OrgName      string    `json:"orgName"`
	AccountLogin string    `json:"accountLogin"`
	Repos        int       `json:"repos"`
	LinkedAt     time.Time `json:"linkedAt"`
}

// LinkedInstallations lists the installations the Clerk instance's organizations linked.
func (s Store) LinkedInstallations(ctx context.Context) ([]LinkedInstallation, error) {
	linked := []LinkedInstallation{}
	if err := s.DB.WithContext(ctx).Raw(`
		SELECT gi.id, gi.org_id, COALESCE(o.slug, '') AS org_slug, o.name AS org_name, gi.account_login,
			gi.created_at AS linked_at,
			(SELECT count(*) FROM repos r WHERE r.installation_id = gi.id) AS repos
		FROM github_installations gi JOIN orgs o ON o.id = gi.org_id
		WHERE o.issuer = ? ORDER BY gi.id`, s.Issuer).Scan(&linked).Error; err != nil {
		return nil, fmt.Errorf("read the platform's linked GitHub App installations: %w", err)
	}
	return linked, nil
}
