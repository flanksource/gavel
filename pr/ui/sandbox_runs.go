// Serves /api/sandbox-runs: a hosted organization's sandbox runs (push checks and agent runs) as
// the signed-in member may see them, and each run's stored log.
package ui

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
)

// sandboxRunsLimit is how many of the newest runs the list returns.
const sandboxRunsLimit = 200

// sandboxRun is one run as the runs tab shows it.
type sandboxRun struct {
	ID         uuid.UUID  `json:"id"`
	Kind       string     `json:"kind"`
	Status     string     `json:"status"`
	ExitCode   *int       `json:"exitCode,omitempty"`
	Ref        string     `json:"ref,omitempty"`
	SHA        string     `json:"sha,omitempty"`
	Repo       string     `json:"repo,omitempty"`
	User       string     `json:"user,omitempty"`
	TodoID     *uuid.UUID `json:"todoId,omitempty"`
	TodoTitle  string     `json:"todoTitle,omitempty"`
	Backend    string     `json:"backend"`
	HasLog     bool       `json:"hasLog"`
	CreatedAt  time.Time  `json:"createdAt"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
}

var errNoSandboxRuns = errors.New("sandbox runs exist only on a hosted dashboard with a sandbox (gavel serve --auth clerk --sandbox ...)")

// sandboxRunsQuery selects the organization's runs the context's user may see: a run on a todo
// is seen by those who see the todo.
func (s *Server) sandboxRunsQuery(r *http.Request, where string, args ...any) (string, []any) {
	visible, visibleArgs := native.VisibleIssue(r.Context(), "todo")
	query := `SELECT run.id, run.kind, run.status, run.exit_code, run.ref, run.sha, repo.name AS repo,
			COALESCE(NULLIF(u.name, ''), u.email, '') AS "user", run.todo_id, todo.title AS todo_title, run.backend,
			run.log_key <> '' AS has_log, run.created_at, run.started_at, run.finished_at
		FROM sandbox_runs AS run
		LEFT JOIN repos AS repo ON repo.id = run.repo_id
		LEFT JOIN users AS u ON u.id = run.user_id
		LEFT JOIN todo_issues AS todo ON todo.id = run.todo_id
		WHERE run.org_id = ? AND (run.todo_id IS NULL OR (todo.id IS NOT NULL AND ` + visible + `)) AND ` + where
	return query, append(append([]any{s.orgID()}, visibleArgs...), args...)
}

// sandboxRunsAvailable answers why there are no runs when this dashboard has none.
func (s *Server) sandboxRunsAvailable(w http.ResponseWriter, r *http.Request) bool {
	if !s.hosted || s.agentRuns == nil || tenant.FromContext(r.Context()).UserID == nil {
		writeJSONError(w, http.StatusNotFound, errNoSandboxRuns)
		return false
	}
	return true
}

func (s *Server) handleSandboxRuns(w http.ResponseWriter, r *http.Request) {
	if !s.sandboxRunsAvailable(w, r) {
		return
	}
	query, args := s.sandboxRunsQuery(r, "TRUE")
	runs := []sandboxRun{}
	if err := s.agentRuns.Dispatcher.DB.WithContext(r.Context()).Raw(query+` ORDER BY run.created_at DESC, run.id LIMIT ?`,
		append(args, sandboxRunsLimit)...).Scan(&runs).Error; err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Errorf("list sandbox runs: %w", err))
		return
	}
	writeJSON(w, runs)
}

func (s *Server) handleSandboxRunLog(w http.ResponseWriter, r *http.Request) {
	if !s.sandboxRunsAvailable(w, r) {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("%q is not a run ID", r.PathValue("id")))
		return
	}
	query, args := s.sandboxRunsQuery(r, "run.id = ?", id)
	var found []struct{ LogKey string }
	if err := s.agentRuns.Dispatcher.DB.WithContext(r.Context()).Raw(`SELECT run.log_key FROM (`+query+`) AS visible JOIN sandbox_runs AS run ON run.id = visible.id`,
		args...).Scan(&found).Error; err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Errorf("look up sandbox run %s: %w", id, err))
		return
	}
	if len(found) == 0 {
		writeJSONError(w, http.StatusNotFound, fmt.Errorf("no sandbox run %s in this organization", id))
		return
	}
	var store blob.Store
	if logs := s.agentRuns.Dispatcher.Logs; logs != nil {
		store = logs(s.orgID())
	}
	if found[0].LogKey == "" || store == nil {
		writeJSONError(w, http.StatusNotFound, fmt.Errorf("sandbox run %s has no stored log: logs are kept once a run ends, when serve has a blob store", id))
		return
	}
	log, err := store.Read(r.Context(), found[0].LogKey)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, fmt.Errorf("read the log of sandbox run %s: %w", id, err))
		return
	}
	defer log.Close()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.Copy(w, log)
}
