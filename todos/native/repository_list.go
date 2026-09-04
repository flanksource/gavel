package native

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// CountIssuesByStatus groups a workspace's issues by the four columns the
// derived TODO status is a function of. Callers fold the groups through the
// same derivation List uses, so counting never has to materialize issue bodies.
func (r *Repository) CountIssuesByStatus(ctx context.Context, workspaceID uuid.UUID) ([]IssueStatusCount, error) {
	var counts []IssueStatusCount
	result := r.db.WithContext(ctx).Raw(`
		SELECT issue.status,
		       COALESCE(runtime.execution_state, 'idle') AS execution_state,
		       COALESCE(active_link.step_kind::text, '') AS step_kind,
		       COALESCE(plan.approval_state::text, '')   AS approval_state,
		       COUNT(*)                                  AS count
		FROM `+issueFrom+`
		LEFT JOIN todo_issue_prompt_runs AS active_link
		  ON active_link.issue_id = issue.id
		 AND active_link.prompt_run_id = issue.active_prompt_run_id
		LEFT JOIN captain_plans AS plan ON plan.id = issue.selected_plan_id
		WHERE issue.workspace_id = ?
		GROUP BY 1, 2, 3, 4`,
		workspaceID,
	).Scan(&counts)
	if result.Error != nil {
		return nil, result.Error
	}
	return counts, nil
}

// ListIssuePhaseRuns returns the latest prompt run per (issue, phase) for a
// whole workspace in one query, so a backlog can render per-phase status,
// progress and elapsed time without a lookup per row. Resolving a todo's phases
// as it is rendered is the N+1 that made /api/projects take 46 seconds.
//
// Two things the SQL does that the phase model does not make obvious:
//
//   - The verify phase has two sources. A standalone step_kind='verify' run is
//     the rare case; most verification happens as phase='verify' INSIDE a
//     step_kind='run' run, which is why the second arm of the UNION re-emits
//     those runs as verify. gavel_todo_issue_execution_state already folds the
//     same two sources (110_todo_projection_functions.sql), and a verify column
//     that skipped this would read empty for almost every todo.
//   - It reads captain_prompt_run_overview rather than captain_prompt_runs. The
//     view already computes duration, iteration counts and cost that
//     ListPromptRunOverviews does not return.
//   - The workspace's runs are resolved by primary key through a LATERAL join
//     fenced with OFFSET 0, not by joining the view directly. The view computes
//     four correlated aggregates per row, and nothing pushes the workspace
//     predicate into it: a plain join makes the planner evaluate every prompt
//     run in the database — twice, once per UNION arm — and then discard all but
//     this workspace's. The fence blocks subquery pull-up so the aggregates run
//     only for the ids this workspace actually links. Removing OFFSET 0 silently
//     restores the full scan; measured on a developer database it cost 24.8ms
//     and 7752 shared buffers against 3.8ms and 1192 for this shape.
func (r *Repository) ListIssuePhaseRuns(ctx context.Context, workspaceID uuid.UUID) ([]IssuePhaseRun, error) {
	const phaseColumns = `
		link.issue_id,
		%s AS phase,
		link.ordinal,
		run.state::text                                   AS state,
		run.phase::text                                   AS run_phase,
		run.started_at,
		run.finished_at,
		run.queued_at,
		run.duration_seconds,
		run.iteration_count                               AS iterations,
		run.succeeded_iteration_count                     AS succeeded,
		run.failed_iteration_count                        AS failed,
		COALESCE(run.latest_verification_result::text, '') AS verification_result,
		run.cost_usd,
		(run.id = link.active_prompt_run_id)              AS active`

	const phaseFrom = `
		FROM workspace_links AS link
		JOIN workspace_runs AS run ON run.id = link.prompt_run_id`

	var records []IssuePhaseRun
	result := r.db.WithContext(ctx).Raw(`
		WITH workspace_links AS MATERIALIZED (
			SELECT link.issue_id, link.step_kind::text AS step_kind, link.ordinal,
			       link.prompt_run_id, issue.active_prompt_run_id
			FROM todo_issue_prompt_runs AS link
			JOIN todo_issues AS issue ON issue.id = link.issue_id
			WHERE issue.workspace_id = ?
		),
		workspace_runs AS MATERIALIZED (
			SELECT run.id, run.state, run.phase, run.started_at, run.finished_at,
			       run.queued_at, run.duration_seconds, run.iteration_count,
			       run.succeeded_iteration_count, run.failed_iteration_count,
			       run.latest_verification_result, run.cost_usd
			FROM (SELECT DISTINCT prompt_run_id FROM workspace_links) AS ids
			JOIN LATERAL (
				SELECT * FROM captain_prompt_run_overview AS o
				WHERE o.id = ids.prompt_run_id
				OFFSET 0
			) AS run ON true
		),
		phase_runs AS (
			SELECT `+fmt.Sprintf(phaseColumns, "link.step_kind")+phaseFrom+`
			UNION ALL
			SELECT `+fmt.Sprintf(phaseColumns, "'verify'")+phaseFrom+`
			WHERE link.step_kind = 'run'
			  AND (run.phase = 'verify' OR run.latest_verification_result IS NOT NULL)
		)
		SELECT DISTINCT ON (issue_id, phase)
		       issue_id, phase, state, run_phase, started_at, finished_at,
		       duration_seconds, iterations, succeeded, failed,
		       verification_result, cost_usd, active
		FROM phase_runs
		-- queued_at leads, not ordinal: ordinals are numbered per step_kind, so
		-- for the verify phase — whose rows come from two different kinds — a
		-- standalone verify at ordinal 0 would otherwise lose to an older run at
		-- ordinal 3. Within one kind the two agree anyway.
		ORDER BY issue_id, phase, queued_at DESC NULLS LAST, ordinal DESC`,
		workspaceID,
	).Scan(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	return records, nil
}

func (r *Repository) ListIssues(ctx context.Context, workspaceID uuid.UUID) ([]Issue, error) {
	var records []issueRecord
	result := r.db.WithContext(ctx).Raw(
		`SELECT `+issueColumns+` FROM `+issueFrom+` WHERE issue.workspace_id = ? ORDER BY issue.updated_at DESC, issue.id`,
		workspaceID,
	).Scan(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	issues := make([]Issue, 0, len(records))
	for _, record := range records {
		issues = append(issues, *record.issue())
	}
	return issues, nil
}

// GetIssueByRef resolves exact workspace aliases before UUIDs so imported
// UUID-shaped aliases remain authoritative. It then accepts native UUID and
// unambiguous UUID/alias prefixes of at least MinShortReferenceLength.
func (r *Repository) GetIssueByRef(ctx context.Context, workspaceID uuid.UUID, ref string) (*Issue, error) {
	ref = normalizeToken(ref)
	if ref == "" {
		return nil, fmt.Errorf("%w: issue reference is required", ErrInvalidInput)
	}

	var exactAlias struct{ IssueID uuid.UUID }
	aliasResult := r.db.WithContext(ctx).Raw(`
		SELECT issue_id FROM todo_issue_aliases
		WHERE workspace_id = ? AND alias = ?`, workspaceID, ref,
	).Scan(&exactAlias)
	if aliasResult.Error != nil {
		return nil, aliasResult.Error
	}
	if aliasResult.RowsAffected > 0 {
		return getIssue(r.db.WithContext(ctx), `workspace_id = ? AND id = ?`, workspaceID, exactAlias.IssueID)
	}

	if id, err := uuid.Parse(ref); err == nil {
		issue, err := getIssue(r.db.WithContext(ctx), `workspace_id = ? AND id = ?`, workspaceID, id)
		if err == nil || !errors.Is(err, ErrNotFound) {
			return issue, err
		}
	}

	compact := strings.ReplaceAll(ref, "-", "")
	if len(compact) < MinShortReferenceLength {
		return nil, fmt.Errorf("%w: short issue reference must contain at least %d characters", ErrInvalidInput, MinShortReferenceLength)
	}
	prefix := escapeLike(ref) + "%"
	compactPrefix := escapeLike(compact) + "%"
	var matches []struct{ IssueID uuid.UUID }
	result := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT issue_id
		FROM (
			SELECT id AS issue_id
			FROM todo_issues
			WHERE workspace_id = ?
			  AND replace(id::text, '-', '') LIKE ? ESCAPE '\'
			UNION
			SELECT issue_id
			FROM todo_issue_aliases
			WHERE workspace_id = ? AND alias LIKE ? ESCAPE '\'
		) AS matched_refs
		LIMIT 2`, workspaceID, compactPrefix, workspaceID, prefix,
	).Scan(&matches)
	if result.Error != nil {
		return nil, result.Error
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("%w: issue reference %q", ErrNotFound, ref)
	case 1:
		return getIssue(r.db.WithContext(ctx), `workspace_id = ? AND id = ?`, workspaceID, matches[0].IssueID)
	default:
		return nil, fmt.Errorf("%w: issue reference %q", ErrAmbiguousReference, ref)
	}
}

// GetIssueByGlobalRef resolves issue references without consulting a runtime
// provider or requiring a workspace. Exact aliases remain authoritative over
// UUIDs, including when an alias happens to be a valid native UUID. References
// that identify issues in multiple workspaces are rejected as ambiguous.
func (r *Repository) GetIssueByGlobalRef(ctx context.Context, ref string) (*Issue, error) {
	ref = normalizeToken(ref)
	if ref == "" {
		return nil, fmt.Errorf("%w: issue reference is required", ErrInvalidInput)
	}

	var exactAliases []struct{ IssueID uuid.UUID }
	aliasResult := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT issue_id
		FROM todo_issue_aliases
		WHERE alias = ?
		LIMIT 2`, ref,
	).Scan(&exactAliases)
	if aliasResult.Error != nil {
		return nil, aliasResult.Error
	}
	switch len(exactAliases) {
	case 1:
		return getIssue(r.db.WithContext(ctx), `id = ?`, exactAliases[0].IssueID)
	case 2:
		return nil, fmt.Errorf("%w: issue reference %q", ErrAmbiguousReference, ref)
	}

	if id, err := uuid.Parse(ref); err == nil {
		issue, err := getIssue(r.db.WithContext(ctx), `id = ?`, id)
		if err == nil || !errors.Is(err, ErrNotFound) {
			return issue, err
		}
	}

	compact := strings.ReplaceAll(ref, "-", "")
	if len(compact) < MinShortReferenceLength {
		return nil, fmt.Errorf("%w: short issue reference must contain at least %d characters", ErrInvalidInput, MinShortReferenceLength)
	}
	prefix := escapeLike(ref) + "%"
	compactPrefix := escapeLike(compact) + "%"
	var matches []struct{ IssueID uuid.UUID }
	result := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT issue_id
		FROM (
			SELECT id AS issue_id
			FROM todo_issues
			WHERE replace(id::text, '-', '') LIKE ? ESCAPE '\'
			UNION
			SELECT issue_id
			FROM todo_issue_aliases
			WHERE alias LIKE ? ESCAPE '\'
		) AS matched_refs
		LIMIT 2`, compactPrefix, prefix,
	).Scan(&matches)
	if result.Error != nil {
		return nil, result.Error
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("%w: issue reference %q", ErrNotFound, ref)
	case 1:
		return getIssue(r.db.WithContext(ctx), `id = ?`, matches[0].IssueID)
	default:
		return nil, fmt.Errorf("%w: issue reference %q", ErrAmbiguousReference, ref)
	}
}

// GetIssueBySessionRef resolves either a Captain-native session UUID or a
// provider session UUID to the one native issue whose prompt run is linked to
// that session. A provider session can have separate Gavel orchestration and
// Claude/Codex transcript rows; matching the shared provider identity bridges
// those rows without making Captain own Gavel's issue links.
func (r *Repository) GetIssueBySessionRef(ctx context.Context, ref string) (*Issue, string, error) {
	ref = normalizeToken(ref)
	parsed, err := uuid.Parse(ref)
	if err != nil {
		return nil, "", fmt.Errorf("%w: session reference must be a UUID", ErrInvalidInput)
	}
	ref = parsed.String()

	var matches []struct {
		IssueID         uuid.UUID
		SessionIdentity string
	}
	result := r.db.WithContext(ctx).Raw(`
		WITH target_sessions AS (
			SELECT id, NULLIF(btrim(provider_session_id), '') AS provider_session_id
			FROM captain_sessions
			WHERE id = ? OR provider_session_id = ?
		), candidate_sessions AS (
			SELECT DISTINCT
				session.id,
				COALESCE(NULLIF(btrim(session.provider_session_id), ''), session.id::text) AS session_identity
			FROM captain_sessions AS session
			WHERE session.id IN (SELECT id FROM target_sessions)
			   OR NULLIF(btrim(session.provider_session_id), '') IN (
				SELECT provider_session_id
				FROM target_sessions
				WHERE provider_session_id IS NOT NULL
			   )
		)
		SELECT
			link.issue_id,
			MIN(candidate.session_identity) AS session_identity
		FROM todo_issue_prompt_runs AS link
		JOIN captain_prompt_runs AS run ON run.id = link.prompt_run_id
		JOIN candidate_sessions AS candidate ON candidate.id = run.session_id
		GROUP BY link.issue_id
		LIMIT 2`, parsed, ref,
	).Scan(&matches)
	if result.Error != nil {
		return nil, "", result.Error
	}
	switch len(matches) {
	case 0:
		return nil, "", fmt.Errorf("%w: session reference %q", ErrNotFound, ref)
	case 1:
		issue, err := getIssue(r.db.WithContext(ctx), `id = ?`, matches[0].IssueID)
		return issue, matches[0].SessionIdentity, err
	default:
		return nil, "", fmt.Errorf("%w: session reference %q", ErrAmbiguousReference, ref)
	}
}
