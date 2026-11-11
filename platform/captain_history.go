// Summarizes Captain's history across every organization for the platform dashboard: prompt runs
// per day by origin and state, the latest failures, and the sessions live now by health.
package platform

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CaptainRunDay is one UTC day's prompt runs of an organization with one origin and state. Org is
// nil for runs no organization of the Clerk instance owns.
type CaptainRunDay struct {
	Day        string     `json:"day"` // YYYY-MM-DD
	Org        *uuid.UUID `json:"org,omitempty"`
	OrgName    string     `json:"orgName,omitempty"`
	Origin     string     `json:"origin"`
	State      string     `json:"state"`
	Runs       int        `json:"runs"`
	RunSeconds float64    `json:"runSeconds"` // started to finished, for finished runs
}

// CaptainRunFailure is one failed prompt run and why it failed.
type CaptainRunFailure struct {
	ID         uuid.UUID  `json:"id"`
	Org        *uuid.UUID `json:"org,omitempty"`
	OrgName    string     `json:"orgName,omitempty"`
	Origin     string     `json:"origin"`
	Error      string     `json:"error"`
	CreatedAt  time.Time  `json:"createdAt"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
}

// CaptainLiveSessions counts an organization's top-level sessions created or running now, by health.
type CaptainLiveSessions struct {
	Org      *uuid.UUID `json:"org,omitempty"`
	OrgName  string     `json:"orgName,omitempty"`
	Health   string     `json:"health"`
	Sessions int        `json:"sessions"`
}

// CaptainHistory is Captain's prompt runs since a window's start, and its sessions live now.
type CaptainHistory struct {
	Days     []CaptainRunDay       `json:"days"`
	Failures []CaptainRunFailure   `json:"failures"`
	Live     []CaptainLiveSessions `json:"live"`
}

// captainFailureLimit is how many of the latest failures CaptainHistory lists.
const captainFailureLimit = 20

// runOwner joins orgs as o to the prompt runs aliased run: the owner of the run's session, or of
// its root session when the session has none, and only the Clerk instance's organizations.
const runOwner = `
	LEFT JOIN session_owners own ON own.session_id = run.session_id
	LEFT JOIN session_owners root ON root.session_id = run.root_session_id
	LEFT JOIN orgs o ON o.id = COALESCE(own.org_id, root.org_id) AND o.issuer = ?`

// CaptainHistory summarizes the prompt runs created since since, the latest of them that failed,
// and the top-level sessions live now.
func (s Store) CaptainHistory(ctx context.Context, since time.Time) (CaptainHistory, error) {
	history := CaptainHistory{Days: []CaptainRunDay{}, Failures: []CaptainRunFailure{}, Live: []CaptainLiveSessions{}}
	db := s.DB.WithContext(ctx)
	if err := db.Raw(`
		SELECT to_char(run.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS day, o.id AS org, o.name AS org_name,
			COALESCE(run.origin, '') AS origin, run.state, count(*) AS runs,
			COALESCE(sum(EXTRACT(EPOCH FROM run.finished_at - run.started_at)), 0) AS run_seconds
		FROM captain_prompt_runs run`+runOwner+`
		WHERE run.created_at >= ?
		GROUP BY 1, 2, 3, 4, 5 ORDER BY 1, 3 NULLS LAST, 4, 5`, s.Issuer, since).Scan(&history.Days).Error; err != nil {
		return CaptainHistory{}, fmt.Errorf("read Captain's prompt runs: %w", err)
	}
	if err := db.Raw(`
		SELECT run.id, o.id AS org, o.name AS org_name, COALESCE(run.origin, '') AS origin,
			COALESCE(run.error, '') AS error, run.created_at, run.finished_at
		FROM captain_prompt_runs run`+runOwner+`
		WHERE run.created_at >= ? AND run.state = 'failed'
		ORDER BY run.created_at DESC LIMIT ?`, s.Issuer, since, captainFailureLimit).Scan(&history.Failures).Error; err != nil {
		return CaptainHistory{}, fmt.Errorf("read Captain's failed prompt runs: %w", err)
	}
	if err := db.Raw(`
		SELECT o.id AS org, o.name AS org_name, session.health_state AS health, count(*) AS sessions
		FROM captain_sessions session
		LEFT JOIN session_owners own ON own.session_id = session.id
		LEFT JOIN orgs o ON o.id = own.org_id AND o.issuer = ?
		WHERE session.parent_session_id IS NULL AND session.lifecycle_status IN ('created', 'running')
		GROUP BY 1, 2, 3 ORDER BY 2 NULLS LAST, 3`, s.Issuer).Scan(&history.Live).Error; err != nil {
		return CaptainHistory{}, fmt.Errorf("read Captain's live sessions: %w", err)
	}
	return history, nil
}
