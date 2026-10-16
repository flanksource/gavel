package ui

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/traces"
)

// testRunTracesPrefix is where the Tests tab's runs have their trace stores
// served: testRunTracesPrefix/<store key>/query/<run id>/..., the query app and
// API over the store a run left at .gavel/traces/<run id>, reopened read-only.
const testRunTracesPrefix = "/api/tests/traces"

// testRunTraceStores holds the stores the Tests tab has reopened, one replay
// Server each, keyed by workspace and run. The MultiServer evicts idle and
// surplus ones and closes their stores, so a long-lived dashboard never holds
// more open sqlite handles than its cap.
type testRunTraceStores struct {
	mu   sync.Mutex
	runs *testui.MultiServer
}

func (s *Server) testRunTraceStores() *testRunTraceStores {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.testRunStores == nil {
		s.testRunStores = &testRunTraceStores{runs: testui.NewMultiServer()}
	}
	return s.testRunStores
}

// testRunTraces serves the reopened stores under testRunTracesPrefix.
func (s *Server) testRunTraces() http.Handler {
	return http.StripPrefix(testRunTracesPrefix, s.testRunTraceStores().runs.Handler())
}

// closeTestRunTraces closes every store the Tests tab reopened.
func (s *Server) closeTestRunTraces() error {
	return s.testRunTraceStores().runs.CloseTraces()
}

// status reopens run's store, or reuses the one open, and returns its traces
// status with the mounts as the dashboard serves them. process is the capture
// the run's snapshot names (nil when it was not watched); the status carries it
// so the Processes view reads the tree and samples back from the reopened
// store. A store that is there
// but does not open is an error; a pruned one is served without, saying so.
func (st *testRunTraceStores) status(run testui.TraceRun, process *traces.ProcessRefs) (*testui.TracesStatus, error) {
	key := traceStoreKey(run)
	st.mu.Lock()
	defer st.mu.Unlock()
	srv, ok := st.runs.Get(key)
	if !ok {
		srv = testui.NewServer()
		srv.LoadSnapshot(testui.Snapshot{Metadata: &testui.SnapshotMetadata{RunID: run.RunID, ProcessTrace: process}})
		if err := srv.ReopenTraces(run); err != nil {
			return nil, err
		}
		st.runs.Put(key, srv)
	}
	status := srv.Snapshot().Status.Traces
	for id, served := range status.Runs {
		if served.Mount != "" {
			served.Mount = testRunTracesPrefix + "/" + key + served.Mount
			status.Runs[id] = served
		}
	}
	return status, nil
}

// traceStoreKey names a run's store in a URL without exposing its path.
func traceStoreKey(run testui.TraceRun) string {
	sum := sha256.Sum256([]byte(run.WorkDir + "\x00" + run.RunID))
	return hex.EncodeToString(sum[:8])
}

// withTestRunTraces sets the status.traces of a run snapshot read from disk
// to its store as this dashboard serves it, replacing the status the run saved
// with (its own server's mounts). A snapshot that kept no store is returned as
// read.
func (s *Server) withTestRunTraces(data []byte) ([]byte, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("decode run snapshot: %w", err)
	}
	var meta *testui.SnapshotMetadata
	if raw, ok := doc["metadata"]; ok {
		if err := json.Unmarshal(raw, &meta); err != nil {
			return nil, fmt.Errorf("decode run snapshot metadata: %w", err)
		}
	}
	run, err := testui.TraceRunOf(meta)
	if err != nil || run == nil {
		return data, err
	}
	served, err := s.testRunTraceStores().status(*run, meta.ProcessTrace)
	if err != nil {
		return nil, err
	}
	status := map[string]json.RawMessage{}
	if raw, ok := doc["status"]; ok {
		if err := json.Unmarshal(raw, &status); err != nil {
			return nil, fmt.Errorf("decode run snapshot status: %w", err)
		}
	}
	if status["traces"], err = json.Marshal(served); err != nil {
		return nil, err
	}
	if doc["status"], err = json.Marshal(status); err != nil {
		return nil, err
	}
	return json.Marshal(doc)
}
