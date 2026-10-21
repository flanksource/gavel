package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/traces"
)

// savedRun writes the snapshot of a finished run, runID, whose trace store is
// under workDir, and returns its path.
func savedRun(workDir, runID string) string {
	host, err := traces.Open(workDir, runID)
	Expect(err).ToNot(HaveOccurred())
	Expect(host.Close(context.Background())).To(Succeed())
	data, err := json.Marshal(testui.Snapshot{Metadata: &testui.SnapshotMetadata{RunID: runID, Args: map[string]any{"work_dir": workDir}}})
	Expect(err).ToNot(HaveOccurred())
	path := filepath.Join(workDir, runID+".json")
	Expect(os.WriteFile(path, data, 0o600)).To(Succeed())
	return path
}

var _ = Describe("gavel ui serve over several snapshots", func() {
	It("serves the trace store of every run it loaded, each under its own run id", func() {
		workDir := GinkgoT().TempDir()
		srv := testui.NewServer()
		Expect(loadResults(srv, savedRun(workDir, "run-one"), savedRun(workDir, "run-two"))).To(Succeed())
		DeferCleanup(srv.CloseTraces)
		Expect(srv.Snapshot().Status.Traces.Runs).To(Equal(map[string]testui.RunTraces{
			"run-one": {Mount: testui.QueryMount + "/run-one"},
			"run-two": {Mount: testui.QueryMount + "/run-two"},
		}))
	})
})
