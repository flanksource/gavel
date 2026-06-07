package testui

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
)

// fetchSnapshot GETs /<runID>/api/tests through the MultiServer handler and
// decodes the Snapshot.
func fetchSnapshot(t *testing.T, h http.Handler, runID string) Snapshot {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/"+runID+"/api/tests", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /%s/api/tests = %d, want 200", runID, rec.Code)
	}
	var snap Snapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	return snap
}

func testNames(tests []parsers.Test) []string {
	out := make([]string, len(tests))
	for i, test := range tests {
		out[i] = test.Name
	}
	return out
}

func TestMultiServerIsolatesConcurrentRunsByID(t *testing.T) {
	m := NewMultiServer()
	srvA := m.BeginRun("run-a", "fixtures")
	srvB := m.BeginRun("run-b", "fixtures")

	srvA.SetResults([]parsers.Test{{Name: "alpha", Passed: true}})
	srvB.SetResults([]parsers.Test{{Name: "bravo", Failed: true}})

	h := m.Handler()
	snapA := fetchSnapshot(t, h, "run-a")
	snapB := fetchSnapshot(t, h, "run-b")

	if got := testNames(snapA.Tests); len(got) != 1 || got[0] != "alpha" {
		t.Fatalf("run-a tests = %v, want [alpha]", got)
	}
	if got := testNames(snapB.Tests); len(got) != 1 || got[0] != "bravo" {
		t.Fatalf("run-b tests = %v, want [bravo]", got)
	}
}

func TestMultiServerSSEStreamScopedToRun(t *testing.T) {
	m := NewMultiServer()
	srvA := m.BeginRun("run-a", "fixtures")
	m.BeginRun("run-b", "fixtures")

	httpSrv := httptest.NewServer(m.Handler())
	defer httpSrv.Close()

	// Connect an SSE client to run-a, then drive run-a to a terminal state with
	// a distinctive tree. The stream must only ever carry run-a's payload.
	req, err := http.NewRequest("GET", httpSrv.URL+"/run-a/api/tests/stream", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("content-type = %q, want text/event-stream", ct)
	}

	srvA.SetResults([]parsers.Test{{Name: "alpha-only", Passed: true}})

	scanner := bufio.NewScanner(resp.Body)
	sawAlpha := false
	deadline := time.After(3 * time.Second)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "bravo") {
				t.Errorf("run-a stream leaked run-b payload: %q", line)
				return
			}
			if strings.Contains(line, "alpha-only") {
				sawAlpha = true
			}
			if strings.HasPrefix(line, "event: done") {
				return
			}
		}
	}()
	select {
	case <-done:
	case <-deadline:
		t.Fatalf("timed out waiting for run-a stream to finish")
	}
	if !sawAlpha {
		t.Fatalf("run-a stream never carried its own tree")
	}
}

func TestMultiServerUnknownRunIs404(t *testing.T) {
	m := NewMultiServer()
	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/missing/api/tests", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown run id = %d, want 404", rec.Code)
	}
}

func TestMultiServerEvictsIdleDoneRunsOnly(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	m := NewMultiServer()
	m.timeNow = func() time.Time { return now }
	m.SetLimits(0, time.Minute) // keep default cap, 1m idle TTL

	// A finished, idle run is reclaimed; a still-running idle run is kept.
	doneSrv := m.BeginRun("done-run", "fixtures")
	doneSrv.MarkDone()
	m.BeginRun("live-run", "fixtures") // never MarkDone → running

	now = now.Add(2 * time.Minute) // both now idle past TTL
	m.BeginRun("trigger", "fixtures")

	if _, ok := m.Get("done-run"); ok {
		t.Fatalf("idle finished run was not evicted")
	}
	if _, ok := m.Get("live-run"); !ok {
		t.Fatalf("running run was evicted despite being live")
	}
}

func TestMultiServerCapEvictsLeastRecentDoneRun(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	m := NewMultiServer()
	m.timeNow = func() time.Time { return now }
	m.SetLimits(2, time.Hour) // cap 2, long TTL so only the cap evicts

	a := m.BeginRun("a", "fixtures")
	a.MarkDone()
	now = now.Add(time.Second)
	b := m.BeginRun("b", "fixtures")
	b.MarkDone()
	now = now.Add(time.Second)
	m.BeginRun("c", "fixtures") // over cap → evict oldest done (a)

	if _, ok := m.Get("a"); ok {
		t.Fatalf("oldest done run a was not evicted at cap")
	}
	if _, ok := m.Get("b"); !ok {
		t.Fatalf("run b should be retained")
	}
	if _, ok := m.Get("c"); !ok {
		t.Fatalf("newest run c should be retained")
	}
}
