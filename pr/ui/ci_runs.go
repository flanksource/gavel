// Lists a hosted repository's CI runs, the ones GitHub Actions jobs upload to /api/ci/results,
// and serves one run's snapshot, so a member can watch a run while its job streams it.
package ui

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/blob"
)

// ciRunView is one CI run in a repository's list: its summary row in test_run_caches.
type ciRunView struct {
	RunID   string     `json:"runId"`
	SHA     string     `json:"sha,omitempty"`
	Started *time.Time `json:"started,omitempty"`
	// Ended is nil while the run streams, and stays nil for a run whose job was killed.
	Ended *time.Time `json:"ended,omitempty"`
	// Updated is when the last upload arrived, so a run that stopped streaming shows as stale.
	Updated time.Time `json:"updated"`
	Passed  int       `json:"passed"`
	Failed  int       `json:"failed"`
	Skipped int       `json:"skipped"`
	Total   int       `json:"total"`
}

// ciRunsShown bounds how many of a repository's newest CI runs are listed.
const ciRunsShown = 50

// handleCIRuns lists a hosted repository's newest CI runs.
func (s *Server) handleCIRuns(w http.ResponseWriter, r *http.Request) {
	repo, ok := s.hostedProjectRepo(w, r)
	if !ok {
		return
	}
	rows, err := s.ghOpts.Cache.ListTestRunsForDir(r.Context(), RepoWorkspace(repo.ID))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	runs := make([]ciRunView, 0, min(len(rows), ciRunsShown))
	for _, row := range rows[:min(len(rows), ciRunsShown)] {
		runs = append(runs, ciRunView{
			RunID: row.RunID, SHA: row.SHA, Started: fromNanos(row.StartedTS), Ended: fromNanos(row.EndedTS), Updated: row.SyncedAt,
			Passed: row.Passed, Failed: row.Failed, Skipped: row.Skipped, Total: row.Total,
		})
	}
	respondJSON(w, http.StatusOK, runs)
}

// handleCIRun serves one CI run's snapshot as its job last uploaded it.
func (s *Server) handleCIRun(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("run")
	if !ciRunID.MatchString(runID) {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("%q is not a gavel run ID", runID))
		return
	}
	repo, ok := s.hostedProjectRepo(w, r)
	if !ok {
		return
	}
	blobs := s.ghOpts.Cache.OrgBlobs(s.orgID())
	if blobs == nil {
		writeJSONError(w, http.StatusNotFound, errors.New("this server stores no CI runs"))
		return
	}
	reader, err := blob.WithPrefix(blobs, RepoRunsPrefix(repo.ID)).Read(r.Context(), runID+".json")
	if errors.Is(err, blob.ErrNotFound) {
		respondError(w, http.StatusNotFound, "run not found")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	defer reader.Close() //nolint:errcheck
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if _, err := io.Copy(w, reader); err != nil {
		logger.Warnf("serve CI run %s of %s: %v", runID, repo.Name, err)
	}
}

// fromNanos is a unix-nanosecond timestamp as a time, nil for 0.
func fromNanos(ts int64) *time.Time {
	if ts == 0 {
		return nil
	}
	t := time.Unix(0, ts).UTC()
	return &t
}
