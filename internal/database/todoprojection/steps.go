package todoprojection

import (
	"fmt"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
)

// stepKind names one gavel-owned projection function (schema/110). Each reads
// Captain rows and advances only todo_issues.updated_at.
type stepKind string

const (
	// projectPromptRun carries the run's updated_at onto its active TODO.
	projectPromptRun stepKind = "project_prompt_run"
	// projectSession maps a session through its family's provider identity to
	// the admission root whose run is some TODO's active run.
	projectSession stepKind = "project_session"
	// projectTurnRequest carries a request's created/resolved time onto the
	// TODO whose active run it belongs to.
	projectTurnRequest stepKind = "project_turn_request"
	// projectRootSession carries a session's activity onto the TODO whose root
	// session (the TODO's own id) heads its tree.
	projectRootSession stepKind = "project_root_session"
	// touchPromptRun and touchIssue are for deletes: the row is gone, so its
	// timestamps are too, and the change is dated when the listener observes it.
	touchPromptRun stepKind = "touch_prompt_run"
	touchIssue     stepKind = "touch_issue"
)

var stepSQL = map[stepKind]string{
	projectPromptRun:   `SELECT public.gavel_project_todo_prompt_run(?)`,
	projectSession:     `SELECT public.gavel_project_todo_session(?)`,
	projectTurnRequest: `SELECT public.gavel_project_todo_turn_request(?)`,
	projectRootSession: `SELECT public.gavel_project_todo_root_session(?)`,
	touchPromptRun:     `SELECT public.gavel_touch_todo_prompt_run(?, clock_timestamp())`,
	touchIssue:         `SELECT public.gavel_touch_todo_issue(?, clock_timestamp())`,
}

type step struct {
	kind stepKind
	id   uuid.UUID
}

// stepsFor maps one Captain row change onto the projections that can move a
// TODO. The payload is identity only, so a written row is re-read by its
// projection function and a deleted one can only be addressed by the ids the
// payload still carries. A payload missing an id Captain's contract promises
// (85_row_change_notify.sql) is an error, never a silently skipped change.
func stepsFor(change captaindb.RowChange) ([]step, error) {
	deleted := change.Op == captaindb.RowChangeDelete
	switch change.Table {
	case captaindb.RowChangeSessions:
		root, err := required(change, "rootSessionId", change.RootSessionID)
		if err != nil {
			return nil, err
		}
		if !deleted {
			return []step{{projectSession, change.ID}, {projectRootSession, change.ID}}, nil
		}
		if root == change.ID {
			return []step{{touchIssue, root}}, nil
		}
		return []step{{projectSession, root}, {touchIssue, root}}, nil
	case captaindb.RowChangePromptRuns:
		return []step{{projectPromptRun, change.ID}}, nil
	case captaindb.RowChangeTurnRequests:
		session, err := required(change, "sessionId", change.SessionID)
		if err != nil {
			return nil, err
		}
		return turnRequestSteps(change, session, deleted), nil
	case captaindb.RowChangePromptRunIterations:
		run, err := required(change, "promptRunId", change.PromptRunID)
		if err != nil {
			return nil, err
		}
		if deleted {
			return []step{{touchPromptRun, run}}, nil
		}
		return []step{{projectPromptRun, run}}, nil
	default:
		return nil, fmt.Errorf("project Captain row change: table %q is not part of the row-change contract", change.Table)
	}
}

func turnRequestSteps(change captaindb.RowChange, session uuid.UUID, deleted bool) []step {
	var steps []step
	switch {
	case deleted && change.PromptRunID != nil:
		steps = append(steps, step{touchPromptRun, *change.PromptRunID})
	case !deleted:
		steps = append(steps, step{projectTurnRequest, change.ID})
		if change.PromptRunID != nil {
			steps = append(steps, step{projectPromptRun, *change.PromptRunID})
		}
	}
	return append(steps, step{projectSession, session})
}

func required(change captaindb.RowChange, field string, id *uuid.UUID) (uuid.UUID, error) {
	if id == nil || *id == uuid.Nil {
		return uuid.Nil, fmt.Errorf("project Captain row change: %s %s %s has no %s",
			change.Op, change.Table, change.ID, field)
	}
	return *id, nil
}
