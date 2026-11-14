package editor

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/sdk/status"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Recorded problems", func() {
	var dir, guide, config string
	var store *Store
	var server *Server

	lintState := func() LintRun {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/api/lint", nil))
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		var run LintRun
		Expect(json.Unmarshal(response.Body.Bytes(), &run)).To(Succeed())
		return run
	}
	idOf := func(path string) string {
		for id, p := range store.Paths() {
			if p == path {
				return id
			}
		}
		Fail("no file " + path)
		return ""
	}
	recordRun := func(ended time.Time, snap testui.Snapshot) {
		if snap.Metadata == nil {
			snap.Metadata = &testui.SnapshotMetadata{}
		}
		snap.Metadata.Started, snap.Metadata.Ended = ended.Add(-time.Minute), ended
		_, err := snapshots.SavePerRun(dir, &snap, snap.Metadata.Started, "")
		Expect(err).NotTo(HaveOccurred())
	}

	BeforeEach(func() {
		var err error
		dir, err = filepath.EvalSymlinks(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		guide, config = filepath.Join(dir, "guide.md"), filepath.Join(dir, "config.json")
		Expect(os.WriteFile(guide, []byte("# Guide\n\nstep\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(config, []byte(`{"name":"initial"}`), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "unopened.md"), []byte("# Other\n"), 0o600)).To(Succeed())
		store, err = NewStore([]string{guide, config})
		Expect(err).NotTo(HaveOccurred())
		server = &Server{Store: store, Context: context.Background()}
	})

	It("loads the recorded lint violations and failed tests of the edited files only", func() {
		recordRun(time.Now().Add(time.Hour), testui.Snapshot{
			Git:    &testui.SnapshotGit{Root: dir},
			Status: testui.SnapshotStatus{LintRun: true},
			Lint: []*linters.LinterResult{{Linter: "markdown", WorkDir: dir, Success: true, Violations: []models.Violation{
				{File: "guide.md", Line: 3, Message: new("heading level")},
				{File: "unopened.md", Line: 1, Message: new("elsewhere")},
			}}},
			Tests: []parsers.Test{{Name: "guide step", File: "guide.md", Line: 3, Failed: true}},
		})

		Expect(server.Reload()).To(Succeed())
		run := lintState()

		Expect(run.Results).To(HaveLen(1))
		Expect(run.Results[0].Violations).To(ConsistOf(HaveField("Message", HaveValue(Equal("heading level")))))
		Expect(run.Tests).To(ConsistOf(SatisfyAll(HaveField("Name", "guide step"), HaveField("File", guide), HaveField("WorkDir", dir))))
		Expect(run.Stale).To(BeFalse())
		Expect(run.Recorded).NotTo(BeNil())
	})

	It("places recorded problems on the path a file was opened through a symlinked directory", func() {
		link := filepath.Join(GinkgoT().TempDir(), "checkout")
		Expect(os.Symlink(dir, link)).To(Succeed())
		opened := filepath.Join(link, "guide.md")
		var err error
		store, err = NewStore([]string{opened})
		Expect(err).NotTo(HaveOccurred())
		server = &Server{Store: store, Context: context.Background()}
		recordRun(time.Now().Add(time.Hour), testui.Snapshot{
			Git:    &testui.SnapshotGit{Root: dir},
			Status: testui.SnapshotStatus{LintRun: true},
			Lint: []*linters.LinterResult{{Linter: "markdown", WorkDir: dir, Success: true,
				Violations: []models.Violation{{File: "guide.md", Line: 3}}}},
			Tests: []parsers.Test{{Name: "guide step", File: "guide.md", Line: 3, Failed: true}},
		})

		Expect(server.Reload()).To(Succeed())
		run := lintState()

		Expect(run.Results).To(ConsistOf(HaveField("WorkDir", link)))
		Expect(run.Tests).To(ConsistOf(SatisfyAll(HaveField("WorkDir", link), HaveField("File", opened))))
	})

	It("marks a file stale when it changed after the run that checked it, or is edited after loading", func() {
		recordRun(time.Now().Add(-time.Hour), testui.Snapshot{
			Metadata: &testui.SnapshotMetadata{LintFiles: []string{guide}},
			Status:   testui.SnapshotStatus{LintRun: true},
			Lint: []*linters.LinterResult{{Linter: "markdown", WorkDir: dir, Success: true,
				Violations: []models.Violation{{File: "guide.md", Line: 3}}}},
		})
		Expect(server.Reload()).To(Succeed())
		Expect(lintState().StaleFiles).To(Equal([]string{idOf(guide)}))

		doc, err := store.Read(idOf(config))
		Expect(err).NotTo(HaveOccurred())
		_, err = store.Write(Document{File: File{ID: doc.ID}, Source: `{"name":"edited"}`, Revision: doc.Revision})
		Expect(err).NotTo(HaveOccurred())

		Expect(lintState().StaleFiles).To(ConsistOf(idOf(guide), idOf(config)))
	})

	It("saves an editor lint run limited to the linted files and reloads from it", func() {
		Expect(os.WriteFile(config, []byte(`{"name":`), 0o600)).To(Succeed())
		Expect(server.Reload()).To(Succeed())
		wake, unsubscribe := server.updated.Subscribe()
		DeferCleanup(unsubscribe)

		request := httptest.NewRequest("POST", "/api/lint", bytes.NewBufferString(`{"files":["`+idOf(config)+`"]}`))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusAccepted), response.Body.String())

		Eventually(wake).Should(Receive())
		Eventually(func() bool { return lintState().Running }, "60s").Should(BeFalse())
		Expect(lintState().Error).To(BeEmpty())
		runs, err := snapshots.LoadRuns(dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(runs).To(ConsistOf(HaveField("Snapshot.Metadata.LintFiles", Equal([]string{config}))))
	})

	It("shows a gavel run in an edited file's workspace while it runs, then reloads once it exits", func() {
		cache := GinkgoT().TempDir()
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		server.Context = ctx
		Expect(server.Reload()).To(Succeed())
		Expect(server.WatchRuns(cache)).To(Succeed())

		external, err := os.FindProcess(os.Getppid())
		Expect(err).NotTo(HaveOccurred())
		path := status.Path(cache, external.Pid)
		data, err := json.Marshal(status.Status{PID: external.Pid, Command: "gavel lint", WorkDir: dir, Started: time.Now(),
			Groups: []status.Group{{Name: "lint", Kind: status.GroupLint}}})
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(path, data, 0o600)).To(Succeed())

		Eventually(func() []ActiveRun { return lintState().Active }, "10s").Should(ConsistOf(HaveField("Command", "gavel lint")))

		recordRun(time.Now().Add(time.Hour), testui.Snapshot{
			Status: testui.SnapshotStatus{LintRun: true},
			Lint: []*linters.LinterResult{{Linter: "markdown", WorkDir: dir, Success: true,
				Violations: []models.Violation{{File: "guide.md", Line: 3}}}},
		})
		Expect(os.Remove(path)).To(Succeed())

		Eventually(func() LintRun { return lintState() }, "10s").Should(SatisfyAll(
			HaveField("Active", BeEmpty()),
			HaveField("Results", HaveLen(1)),
		))
	})
})
