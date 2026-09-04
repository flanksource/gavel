package native

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// ListIssueRunHistory returns every prompt run linked to one issue, oldest
// first, with the status the lifecycle recorded for it. It is the per-issue
// counterpart of ListIssuePhaseRuns and applies the same two rules: a run step
// whose run verified its own work is listed a second time as `verify`, and the
// overview view is resolved by primary key through a fenced LATERAL join so its
// correlated aggregates run only for this issue's runs.
//
// outcomeEventKind names the issue event that records a run's outcome (its
// payload carries promptRunId and status). The lifecycle owns that constant;
// storage does not import the lifecycle.
func (r *Repository) ListIssueRunHistory(ctx context.Context, issueID uuid.UUID, outcomeEventKind string) ([]IssueRunRecord, error) {
	const columns = `
		link.issue_id,
		link.prompt_run_id,
		%s AS phase,
		link.ordinal,
		run.state::text AS state,
		run.queued_at,
		run.started_at,
		run.finished_at,
		COALESCE(outcome.status, '') AS outcome`

	const from = `
		FROM issue_links AS link
		JOIN issue_runs AS run ON run.id = link.prompt_run_id
		LEFT JOIN LATERAL (
			SELECT event.payload->>'status' AS status
			FROM todo_issue_events AS event
			WHERE event.issue_id = link.issue_id
			  AND event.kind = ?
			  AND event.payload->>'promptRunId' = link.prompt_run_id::text
			ORDER BY event.sequence DESC
			LIMIT 1
		) AS outcome ON true`

	var records []IssueRunRecord
	result := r.db.WithContext(ctx).Raw(`
		WITH issue_links AS MATERIALIZED (
			SELECT issue_id, prompt_run_id, step_kind::text AS step_kind, ordinal, created_at
			FROM todo_issue_prompt_runs
			WHERE issue_id = ?
		),
		issue_runs AS MATERIALIZED (
			SELECT run.id, run.state, run.phase, run.queued_at, run.started_at, run.finished_at,
			       run.latest_verification_result
			FROM (SELECT DISTINCT prompt_run_id FROM issue_links) AS ids
			JOIN LATERAL (
				SELECT * FROM captain_prompt_run_overview AS o
				WHERE o.id = ids.prompt_run_id
				OFFSET 0
			) AS run ON true
		),
		history AS (
			SELECT `+fmt.Sprintf(columns, "link.step_kind")+from+`
			UNION ALL
			SELECT `+fmt.Sprintf(columns, "'verify'")+from+`
			WHERE link.step_kind = 'run'
			  AND (run.phase = 'verify' OR run.latest_verification_result IS NOT NULL)
		)
		SELECT issue_id, prompt_run_id, phase, ordinal, state, queued_at, started_at, finished_at, outcome
		FROM history
		-- A run precedes its own verify listing; otherwise dispatch order.
		ORDER BY queued_at, ordinal, (phase = 'verify')`,
		issueID, outcomeEventKind, outcomeEventKind,
	).Scan(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	return records, nil
}
