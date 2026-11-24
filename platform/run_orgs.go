// Finds the organizations of sandbox runs, so the platform dashboard can say whose run each of
// substrate's actors is.
package platform

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// RunOrg is the organization a sandbox run belongs to, and the run's kind.
type RunOrg struct {
	Org     uuid.UUID `json:"org"`
	OrgName string    `json:"orgName"`
	Kind    string    `json:"kind"`
}

// RunOrgs finds the organization of each of runs; runs gavel doesn't know, or of another Clerk
// instance's organizations, are left out.
func (s Store) RunOrgs(ctx context.Context, runs []uuid.UUID) (map[uuid.UUID]RunOrg, error) {
	orgs := map[uuid.UUID]RunOrg{}
	if len(runs) == 0 {
		return orgs, nil
	}
	var rows []struct {
		Run uuid.UUID
		RunOrg
	}
	if err := s.DB.WithContext(ctx).Raw(`
		SELECT run.id AS run, run.org_id AS org, o.name AS org_name, run.kind
		FROM sandbox_runs run JOIN orgs o ON o.id = run.org_id
		WHERE o.issuer = ? AND run.id IN ?`, s.Issuer, runs).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("read the organizations of %d sandbox runs: %w", len(runs), err)
	}
	for _, row := range rows {
		orgs[row.Run] = row.RunOrg
	}
	return orgs, nil
}
