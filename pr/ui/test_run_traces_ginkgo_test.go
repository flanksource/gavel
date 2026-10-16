package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/commons-db/tracing"
	"github.com/flanksource/commons-db/tracing/process"
	"github.com/flanksource/commons-db/tracing/stacks"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/traces"
)

// keptTestRun writes the per-run snapshot of a finished run, runID, of the
// workspace dir; withStore also leaves its trace store, as a run that was not
// pruned since.
func keptTestRun(dir, runID string, withStore bool) {
	if withStore {
		host, err := traces.Open(dir, runID)
		Expect(err).ToNot(HaveOccurred())
		Expect(host.Close(context.Background())).To(Succeed())
	}
	writeTestRunSnapshot(dir, runID, nil)
}

// keptWatchedTestRun is keptTestRun for a run watched in a UI: its store
// holds a process capture of this process with every facet armed and one
// stacks checkpoint, which its snapshot names.
func keptWatchedTestRun(dir, runID string) *traces.ProcessRefs {
	host, err := traces.Open(dir, runID)
	Expect(err).ToNot(HaveOccurred())
	trace, err := host.StartProcessTrace(traces.ProcessTraceOptions{RootPID: os.Getpid(), Facets: true})
	Expect(err).ToNot(HaveOccurred())
	Expect(trace.CheckpointStacks(context.Background(), stacks.CheckpointParams{Reason: "manual"})).To(Succeed())
	Eventually(func() ([]process.Sample, error) {
		readings, err := trace.Readings(context.Background(), traces.ProcessMark{})
		return readings.Samples, err
	}).WithTimeout(5 * time.Second).Should(ContainElement(HaveField("OpenFiles", Not(BeNil()))))
	refs := trace.Refs()
	Expect(trace.Stop(context.Background())).To(Succeed())
	Expect(host.Close(context.Background())).To(Succeed())
	writeTestRunSnapshot(dir, runID, &refs)
	return &refs
}

func writeTestRunSnapshot(dir, runID string, process *traces.ProcessRefs) {
	snapshot := testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{RunID: runID, Args: map[string]any{"work_dir": dir}, ProcessTrace: process},
		// The status the run saved with: its own server's mount, stale here.
		Status: testui.SnapshotStatus{Traces: &testui.TracesStatus{Runs: map[string]testui.RunTraces{runID: {Mount: "/query/" + runID}}}},
	}
	data, err := json.Marshal(snapshot)
	Expect(err).ToNot(HaveOccurred())
	Expect(os.MkdirAll(filepath.Join(dir, ".gavel"), 0o755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(dir, ".gavel", runID+".json"), data, 0o600)).To(Succeed())
}

var _ = Describe("a test run's trace store in the Tests tab", func() {
	var server *Server
	var dir string

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		server = &Server{}
		DeferCleanup(func() { Expect(server.closeTestRunTraces()).To(Succeed()) })
	})

	run := func(runID string) (int, testui.Snapshot) {
		response := httptest.NewRecorder()
		server.handleTestRun(response, httptest.NewRequest(http.MethodGet, "/api/tests/run?dir="+dir+"&runId="+runID, nil))
		var snapshot testui.Snapshot
		if response.Code == http.StatusOK {
			Expect(json.Unmarshal(response.Body.Bytes(), &snapshot)).To(Succeed(), response.Body.String())
		}
		return response.Code, snapshot
	}

	It("is reopened and served under the run's own mount, which the run's status names", func() {
		const runID = "run-2026-10-12T09-00-00Z"
		keptTestRun(dir, runID, true)
		code, snapshot := run(runID)
		Expect(code).To(Equal(http.StatusOK))
		mount := snapshot.Status.Traces.Runs[runID].Mount
		Expect(mount).To(HavePrefix(testRunTracesPrefix + "/"))
		Expect(mount).To(HaveSuffix(testui.QueryMount + "/" + runID))

		response := httptest.NewRecorder()
		server.testRunTraces().ServeHTTP(response, httptest.NewRequest(http.MethodGet, mount+"/api/v1/traces/kinds", nil))
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(response.Body.String()).To(ContainSubstring("http_proxy"))

		_, again := run(runID)
		Expect(again.Status.Traces.Runs[runID].Mount).To(Equal(mount), "a second read reuses the open store")
	})

	It("serves a watched run's stacks and facets back from the reopened store", func() {
		const runID = "run-2026-10-12T09-30-00Z"
		process := keptWatchedTestRun(dir, runID)
		code, snapshot := run(runID)
		Expect(code).To(Equal(http.StatusOK))
		Expect(snapshot.Status.Traces.ProcessRun).To(Equal(runID))
		Expect(snapshot.Status.Traces.Process).To(Equal(process))
		Expect(snapshot.Status.Traces.Sessions).To(BeNil(), "the capture has ended, so it cannot be checkpointed")

		mount := snapshot.Status.Traces.Runs[runID].Mount
		export := func(ref tracing.TraceRef) string {
			response := httptest.NewRecorder()
			query := url.Values{"format": {"ndjson"}, "stream": {ref.Stream}}
			server.testRunTraces().ServeHTTP(response, httptest.NewRequest(http.MethodGet, mount+"/api/v1/traces/"+ref.Kind+"/export?"+query.Encode(), nil))
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
			return response.Body.String()
		}
		Expect(export(process.Stacks)).To(And(ContainSubstring(`"reason":"manual"`), ContainSubstring(`"transport":"inprocess"`)))
		Expect(export(process.Samples)).To(ContainSubstring(`"openFileList":[{`))
	})

	It("does not name a process capture for a run that was not watched", func() {
		const runID = "run-2026-10-12T09-40-00Z"
		keptTestRun(dir, runID, true)
		_, snapshot := run(runID)
		Expect(snapshot.Status.Traces.ProcessRun).To(BeEmpty())
		Expect(snapshot.Status.Traces.Process).To(BeNil())
	})

	It("serves a pruned run without its store, saying why", func() {
		const runID = "run-2026-10-12T10-00-00Z"
		keptTestRun(dir, runID, false)
		code, snapshot := run(runID)
		Expect(code).To(Equal(http.StatusOK))
		Expect(snapshot.Status.Traces.Runs[runID].Mount).To(BeEmpty())
		Expect(snapshot.Status.Traces.Runs[runID].Unavailable).To(ContainSubstring("is gone"))
	})

	It("fails the read of a run whose store is there but does not open", func() {
		const runID = "run-2026-10-12T11-00-00Z"
		keptTestRun(dir, runID, false)
		Expect(os.MkdirAll(filepath.Join(dir, traces.Dir), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, traces.Dir, runID), []byte("not a store"), 0o600)).To(Succeed())
		code, _ := run(runID)
		Expect(code).To(Equal(http.StatusInternalServerError))
	})

	It("closes a run's store when its traces are evicted", func() {
		const runID = "run-2026-10-12T12-00-00Z"
		keptTestRun(dir, runID, true)
		_, snapshot := run(runID)
		mount := snapshot.Status.Traces.Runs[runID].Mount
		Expect(server.closeTestRunTraces()).To(Succeed())
		response := httptest.NewRecorder()
		server.testRunTraces().ServeHTTP(response, httptest.NewRequest(http.MethodGet, mount+"/api/v1/traces/kinds", nil))
		Expect(response.Code).To(Equal(http.StatusNotFound))
		Expect(strings.TrimSpace(response.Body.String())).ToNot(BeEmpty())
	})
})
