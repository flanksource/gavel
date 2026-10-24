package testui

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/flanksource/gavel/traces"
)

// QueryMount is where the test UI serves its runs' trace stores through
// query, each at QueryMount/<run id>: the traces/<kind> result profiles, the
// trace exports and the checkpoint of a running capture under /api/
// (traceserve.Handler), and the query app over them when a host sets one
// (SetQueryApp).
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
	// Sessions are the running process capture's sessions, which a
	// checkpoint names (POST {mount}/api/v1/sessions/{id}/checkpoint): stacks
	// to collect a process's stacks, samples to refresh its facets. Nil once
	// the run has ended.
	Sessions *traces.ProcessSessions `json:"sessions,omitempty"`
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
// QueryMount/<runID>; its owner closes the store. sessions are the running
// process capture's, nil when there is none.
func (s *Server) ServeTraces(runID string, handler http.Handler, sessions *traces.ProcessSessions) {
	s.setTraces(runID, &servedTraces{handler: handler}, sessions)
}

// TracesUnavailable records why run runID's store is not served.
func (s *Server) TracesUnavailable(runID, reason string) {
	s.setTraces(runID, &servedTraces{unavailable: reason}, nil)
}

func (s *Server) setTraces(runID string, served *servedTraces, sessions *traces.ProcessSessions) {
	s.mu.Lock()
	if s.traceRuns == nil {
		s.traceRuns = map[string]*servedTraces{}
	}
	s.traceRuns[runID] = served
	s.processSessions = sessions
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
		s.setTraces(run.RunID, served, nil)
	}
	return nil
}

func reopenTraces(run TraceRun) (*servedTraces, error) {
	dir := traces.StoreDir(run.WorkDir, run.RunID)
	handler, closeStore, err := traces.Reopen(dir)
	if errors.Is(err, os.ErrNotExist) {
		return &servedTraces{unavailable: fmt.Sprintf("trace store %s is gone: pruned, or not copied with the snapshot", dir)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reopen the trace store of run %s: %w", run.RunID, err)
	}
	return &servedTraces{handler: handler, close: closeStore}, nil
}

// SetQueryApp serves app, the query app (commons-db www.Handler), at
// QueryMount/<run id>/ over each served store, for every path but /api/...,
// which the store answers. Without it a store serves its API only. The test
// UI's "Open in query" links (/traces/<kind>) need it; this package does not
// link the app itself, so a host that never opens it does not embed it.
func (s *Server) SetQueryApp(app http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queryApp = app
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
		status.ProcessRun, status.Process, status.Sessions = primary, s.metadata.ProcessTrace, s.processSessions
	}
	return status
}

// handleTraces serves QueryMount/<run id>/... from that run's store.
func (s *Server) handleTraces(w http.ResponseWriter, r *http.Request) {
	runID, rest, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, QueryMount+"/"), "/")
	s.mu.RLock()
	served, app := s.traceRuns[runID], s.queryApp
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
	handler := served.handler
	if app != nil && !strings.HasPrefix(rest, "api/") {
		handler = app
	}
	http.StripPrefix(QueryMount+"/"+runID, handler).ServeHTTP(w, r)
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
