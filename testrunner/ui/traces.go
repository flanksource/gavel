package testui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/flanksource/commons-db/cmd/query/traceserve"
	"github.com/flanksource/gavel/traces"
)

// QueryMount is where the test UI serves its runs' trace stores through
// query, each at QueryMount/<run id>: the traces/<kind> result profiles, the
// trace exports, the checkpoint of a running capture and the query app
// (traceserve.Handler).
const QueryMount = "/query"

// TracesStatus is the trace stores of the runs the server shows, as the UI
// reaches them. A replay of several snapshots shows several runs, and a
// fixture's trace window names the run whose store holds it.
type TracesStatus struct {
	// Runs are the runs' stores by run id: where each is served, or why not.
	Runs map[string]RunTraces `json:"runs"`
	// ProcessRun is the run whose process capture the Processes tab reads,
	// Process its streams; both empty when that run was not watched or its
	// store is not served.
	ProcessRun string              `json:"process_run,omitempty"`
	Process    *traces.ProcessRefs `json:"process,omitempty"`
	// ProcessSession is the running process capture's session, which a
	// checkpoint names; empty once the run has ended.
	ProcessSession string `json:"process_session,omitempty"`
}

// RunTraces is one run's store as the UI reaches it.
type RunTraces struct {
	// Mount is QueryMount/<run id> while the store is served.
	Mount string `json:"mount,omitempty"`
	// Unavailable says why a store is not served, such as a store Prune
	// removed since the run was saved.
	Unavailable string `json:"unavailable,omitempty"`
}

// TraceRun names a finished run's store: .gavel/traces/<RunID> under WorkDir.
type TraceRun struct {
	RunID   string
	WorkDir string
}

// servedTraces is one run's store on the server: its handler, or why it has
// none, and the close of a store the server reopened itself.
type servedTraces struct {
	handler     http.Handler
	unavailable string
	close       func() error
}

// TraceRunOf is the run whose store a snapshot names, nil when it kept none.
func TraceRunOf(meta *SnapshotMetadata) (*TraceRun, error) {
	if meta == nil || meta.RunID == "" {
		return nil, nil
	}
	workDir, ok := meta.Args["work_dir"].(string)
	if !ok || workDir == "" {
		return nil, fmt.Errorf("snapshot of run %s records no work_dir to find its trace store under", meta.RunID)
	}
	return &TraceRun{RunID: meta.RunID, WorkDir: workDir}, nil
}

// SetTraceRun records the run's trace store and its process capture, which
// a saved snapshot keeps so the store can be reopened.
func (s *Server) SetTraceRun(runID string, process *traces.ProcessRefs) {
	s.mu.Lock()
	defer s.mu.Unlock()
	meta := s.ensureMetadataLocked()
	meta.RunID = runID
	meta.ProcessTrace = process
}

// ServeTraces mounts handler, run runID's store served through query, at
// QueryMount/<runID>; its owner closes the store. processSession names the
// running process capture, empty when there is none.
func (s *Server) ServeTraces(runID string, handler http.Handler, processSession string) {
	s.setTraces(runID, &servedTraces{handler: handler}, processSession)
}

// TracesUnavailable records why run runID's store is not served.
func (s *Server) TracesUnavailable(runID, reason string) {
	s.setTraces(runID, &servedTraces{unavailable: reason}, "")
}

func (s *Server) setTraces(runID string, served *servedTraces, processSession string) {
	s.mu.Lock()
	if s.traceRuns == nil {
		s.traceRuns = map[string]*servedTraces{}
	}
	s.traceRuns[runID] = served
	s.processSession = processSession
	s.mu.Unlock()
	s.notify()
}

// ReopenTraces reopens the stores finished runs left, read-only, and serves
// each at QueryMount/<run id> until CloseTraces. A store pruned (or never
// copied) since is not an error: the run is served without it and the UI says
// so. A store that is there but does not open is, and closes the ones opened.
func (s *Server) ReopenTraces(runs ...TraceRun) error {
	for _, run := range runs {
		served, err := reopenTraces(run)
		if err != nil {
			return errors.Join(err, s.CloseTraces())
		}
		s.setTraces(run.RunID, served, "")
	}
	return nil
}

func reopenTraces(run TraceRun) (*servedTraces, error) {
	dir := traces.StoreDir(run.WorkDir, run.RunID)
	runtime, closeStore, err := traces.Reopen(dir)
	if errors.Is(err, os.ErrNotExist) {
		return &servedTraces{unavailable: fmt.Sprintf("trace store %s is gone: pruned, or not copied with the snapshot", dir)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reopen the trace store of run %s: %w", run.RunID, err)
	}
	handler, err := traceserve.Handler(traceserve.Options{Runtime: runtime, Authorize: traceserve.ReadOnly})
	if err != nil {
		return nil, errors.Join(err, closeStore())
	}
	return &servedTraces{handler: handler, close: closeStore}, nil
}

// CloseTraces closes the stores ReopenTraces opened and stops serving them;
// a MultiServer calls it as it evicts the run.
func (s *Server) CloseTraces() error {
	s.mu.Lock()
	var errs []error
	for runID, served := range s.traceRuns {
		if served.close == nil {
			continue
		}
		errs = append(errs, served.close())
		s.traceRuns[runID] = &servedTraces{unavailable: "this run's trace store was closed"}
	}
	s.mu.Unlock()
	s.notify()
	return errors.Join(errs...)
}

// tracesStatusLocked is nil when no run the server shows kept a store.
func (s *Server) tracesStatusLocked() *TracesStatus {
	primary := ""
	if s.metadata != nil {
		primary = s.metadata.RunID
	}
	if primary == "" && len(s.traceRuns) == 0 {
		return nil
	}
	status := &TracesStatus{Runs: map[string]RunTraces{}}
	if primary != "" {
		status.Runs[primary] = RunTraces{}
	}
	for runID, served := range s.traceRuns {
		run := RunTraces{Unavailable: served.unavailable}
		if served.handler != nil {
			run.Mount = QueryMount + "/" + runID
		}
		status.Runs[runID] = run
	}
	if status.Runs[primary].Mount != "" && s.metadata.ProcessTrace != nil {
		status.ProcessRun, status.Process, status.ProcessSession = primary, s.metadata.ProcessTrace, s.processSession
	}
	return status
}

// handleTraces serves QueryMount/<run id>/... from that run's store.
func (s *Server) handleTraces(w http.ResponseWriter, r *http.Request) {
	runID, rest, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, QueryMount+"/"), "/")
	s.mu.RLock()
	served := s.traceRuns[runID]
	s.mu.RUnlock()
	if runID == "" || served == nil || served.handler == nil {
		http.Error(w, tracesNotServed(runID, served), http.StatusNotFound)
		return
	}
	if rest == "" && !strings.HasSuffix(r.URL.Path, "/") {
		// Relative, so it holds under a host's prefix too (pr/ui mounts runs).
		w.Header().Set("Location", runID+"/")
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}
	http.StripPrefix(QueryMount+"/"+runID, served.handler).ServeHTTP(w, r)
}

func tracesNotServed(runID string, served *servedTraces) string {
	if served != nil && served.unavailable != "" {
		return served.unavailable
	}
	if runID == "" {
		return "name a run: " + QueryMount + "/<run id>/..."
	}
	return fmt.Sprintf("the trace store of run %s is not served", runID)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(value) //nolint:errcheck
}
