package testui_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/commons-db/tracing"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/traces"
)

// keptRun is a finished run, runID, that traced one request to path through an
// http_proxy capture into its store under workDir.
func keptRun(workDir, runID, path string) (testui.TraceRun, tracing.TraceRef) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "ok") }))
	defer target.Close()
	host, err := traces.Open(workDir, runID)
	Expect(err).ToNot(HaveOccurred())
	capture, err := host.Start("http_proxy", nil)
	Expect(err).ToNot(HaveOccurred())
	proxy, err := url.Parse(capture.Exports()["HTTP_PROXY"])
	Expect(err).ToNot(HaveOccurred())
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}, Timeout: 5 * time.Second}
	response, err := client.Get(target.URL + path)
	Expect(err).ToNot(HaveOccurred())
	Expect(response.Body.Close()).To(Succeed())
	end, err := capture.Stop(context.Background())
	Expect(err).ToNot(HaveOccurred())
	Expect(host.Close(context.Background())).To(Succeed())
	return testui.TraceRun{RunID: host.RunID(), WorkDir: workDir}, capture.Ref(0, end)
}

// replay serves a finished run's snapshot with the stores of runs reopened.
func replay(runs ...testui.TraceRun) (*testui.Server, error) {
	srv := testui.NewServer()
	srv.LoadSnapshot(testui.Snapshot{Metadata: &testui.SnapshotMetadata{RunID: runs[len(runs)-1].RunID}})
	err := srv.ReopenTraces(runs...)
	DeferCleanup(srv.CloseTraces)
	return srv, err
}

func get(handler http.Handler, target string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
	return response
}

// exportedURLs downloads ref's window from the store mounted at mount as HAR
// and returns the URLs it recorded.
func exportedURLs(handler http.Handler, mount string, ref tracing.TraceRef) []string {
	window := url.Values{"format": {"har"}, "stream": {ref.Stream}, "toSeq": {fmt.Sprint(ref.ToSeq)}}
	response := get(handler, mount+"/api/v1/traces/http_proxy/export?"+window.Encode())
	Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
	var har struct {
		Log struct {
			Entries []struct {
				Request struct{ URL string } `json:"request"`
			} `json:"entries"`
		} `json:"log"`
	}
	Expect(json.Unmarshal(response.Body.Bytes(), &har)).To(Succeed())
	urls := []string{}
	for _, entry := range har.Log.Entries {
		parsed, err := url.Parse(entry.Request.URL)
		Expect(err).ToNot(HaveOccurred())
		urls = append(urls, parsed.Path)
	}
	return urls
}

var _ = Describe("replayed runs' trace stores", func() {
	It("serves each run's own store under the query mount, by run id", func() {
		workDir := GinkgoT().TempDir()
		first, firstRef := keptRun(workDir, "run-first", "/first")
		second, secondRef := keptRun(workDir, "run-second", "/second")
		srv, err := replay(first, second)
		Expect(err).ToNot(HaveOccurred())

		Expect(srv.Snapshot().Status.Traces).To(Equal(&testui.TracesStatus{Runs: map[string]testui.RunTraces{
			"run-first":  {Mount: testui.QueryMount + "/run-first"},
			"run-second": {Mount: testui.QueryMount + "/run-second"},
		}}))
		Expect(exportedURLs(srv.Handler(), testui.QueryMount+"/run-first", firstRef)).To(Equal([]string{"/first"}))
		Expect(exportedURLs(srv.Handler(), testui.QueryMount+"/run-second", secondRef)).To(Equal([]string{"/second"}))
		Expect(get(srv.Handler(), testui.QueryMount+"/run-other/api/v1/traces/kinds").Code).To(Equal(http.StatusNotFound))
	})

	It("reports a pruned store gone, not failed, and serves the run without it", func() {
		workDir := GinkgoT().TempDir()
		srv, err := replay(testui.TraceRun{RunID: "run-pruned", WorkDir: workDir})
		Expect(err).ToNot(HaveOccurred())
		run := srv.Snapshot().Status.Traces.Runs["run-pruned"]
		Expect(run.Mount).To(BeEmpty())
		Expect(run.Unavailable).To(ContainSubstring(filepath.Join(workDir, traces.Dir, "run-pruned") + " is gone"))
		Expect(get(srv.Handler(), testui.QueryMount+"/run-pruned/api/v1/traces/kinds").Code).To(Equal(http.StatusNotFound))
	})

	It("fails a replay whose store is there but does not open", func() {
		workDir := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(workDir, traces.Dir), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, traces.Dir, "run-file"), []byte("not a store"), 0o600)).To(Succeed())
		_, err := replay(testui.TraceRun{RunID: "run-file", WorkDir: workDir})
		Expect(err).To(MatchError(ContainSubstring("is not a directory")))
	})

	It("stops serving a run's store once its traces are closed", func() {
		run, _ := keptRun(GinkgoT().TempDir(), "run-closed", "/closed")
		srv, err := replay(run)
		Expect(err).ToNot(HaveOccurred())
		Expect(srv.CloseTraces()).To(Succeed())
		Expect(srv.Snapshot().Status.Traces.Runs["run-closed"].Mount).To(BeEmpty())
		Expect(get(srv.Handler(), testui.QueryMount+"/run-closed/api/v1/traces/kinds").Code).To(Equal(http.StatusNotFound))
	})

	It("reads the run a snapshot names from its metadata", func() {
		Expect(testui.TraceRunOf(&testui.SnapshotMetadata{RunID: "run-a", StateRoot: "/repo", Args: map[string]any{"work_dir": "/repo/pkg"}})).
			To(Equal(&testui.TraceRun{RunID: "run-a", WorkDir: "/repo"}), "state_root locates the store, not the run's work_dir")
		Expect(testui.TraceRunOf(&testui.SnapshotMetadata{RunID: "run-a", Args: map[string]any{"work_dir": "/w"}})).
			To(Equal(&testui.TraceRun{RunID: "run-a", WorkDir: "/w"}), "snapshots saved before state_root fall back to work_dir")
		Expect(testui.TraceRunOf(&testui.SnapshotMetadata{})).To(BeNil())
		_, err := testui.TraceRunOf(&testui.SnapshotMetadata{RunID: "run-a"})
		Expect(err).To(MatchError(ContainSubstring("records neither state_root nor work_dir")))
	})
})

var _ = Describe("a multi-run server evicting a run", func() {
	It("closes the evicted run's trace stores", func() {
		run, _ := keptRun(GinkgoT().TempDir(), "run-evicted", "/evicted")
		multi := testui.NewMultiServer()
		multi.SetLimits(1, time.Hour)
		evicted := multi.BeginRun("a", "initial")
		evicted.LoadSnapshot(testui.Snapshot{Metadata: &testui.SnapshotMetadata{RunID: run.RunID}})
		Expect(evicted.ReopenTraces(run)).To(Succeed())
		multi.BeginRun("b", "initial")
		Expect(evicted.Snapshot().Status.Traces.Runs[run.RunID].Mount).To(BeEmpty(), "eviction closed its store")
	})
})
