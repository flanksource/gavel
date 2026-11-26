// Summarizes every organization's sandbox runs for the platform dashboard's usage graphs:
// runs per day by kind and outcome, how long they ran and waited, and those still active.
package platform

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SandboxDay is one UTC day's runs of an organization with one kind, backend and status.
type SandboxDay struct {
	Day          string    `json:"day"` // YYYY-MM-DD
	Org          uuid.UUID `json:"org"`
	OrgName      string    `json:"orgName"`
	Kind         string    `json:"kind"`
	Backend      string    `json:"backend"`
	Status       string    `json:"status"`
	Runs         int       `json:"runs"`
	RunSeconds   float64   `json:"runSeconds"`   // started to finished, for finished runs
	QueueSeconds float64   `json:"queueSeconds"` // created to started, for started runs
}

// SandboxActive counts an organization's runs of one kind that are queued or running now.
type SandboxActive struct {
	Org     uuid.UUID `json:"org"`
	OrgName string    `json:"orgName"`
	Kind    string    `json:"kind"`
	Status  string    `json:"status"`
	Runs    int       `json:"runs"`
}

// SandboxUsage is the platform's sandbox runs since a window's start, and those active now.
type SandboxUsage struct {
	Days   []SandboxDay    `json:"days"`
	Active []SandboxActive `json:"active"`
}

// SandboxUsage summarizes the Clerk instance's organizations' runs created since since.
func (s Store) SandboxUsage(ctx context.Context, since time.Time) (SandboxUsage, error) {
	usage := SandboxUsage{Days: []SandboxDay{}, Active: []SandboxActive{}}
	db := s.DB.WithContext(ctx)
	if err := db.Raw(`
		SELECT to_char(run.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS day, run.org_id AS org, o.name AS org_name,
			run.kind, run.backend, run.status, count(*) AS runs,
			COALESCE(sum(EXTRACT(EPOCH FROM run.finished_at - run.started_at)), 0) AS run_seconds,
			COALESCE(sum(EXTRACT(EPOCH FROM run.started_at - run.created_at)), 0) AS queue_seconds
		FROM sandbox_runs run JOIN orgs o ON o.id = run.org_id
		WHERE o.issuer = ? AND run.created_at >= ?
		GROUP BY 1, 2, 3, 4, 5, 6 ORDER BY 1, 3, 4, 5, 6`, s.Issuer, since).Scan(&usage.Days).Error; err != nil {
		return SandboxUsage{}, fmt.Errorf("read the platform's sandbox runs: %w", err)
	}
	if err := db.Raw(`
		SELECT run.org_id AS org, o.name AS org_name, run.kind, run.status, count(*) AS runs
		FROM sandbox_runs run JOIN orgs o ON o.id = run.org_id
		WHERE o.issuer = ? AND run.status IN ('queued', 'provisioning', 'running')
		GROUP BY 1, 2, 3, 4 ORDER BY 2, 3, 4`, s.Issuer).Scan(&usage.Active).Error; err != nil {
		return SandboxUsage{}, fmt.Errorf("read the platform's active sandbox runs: %w", err)
	}
	return usage, nil
}
