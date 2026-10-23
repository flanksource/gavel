package runs_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/testrunner/ui/runs"
	"github.com/flanksource/gavel/traces"
)

func TestRuns(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run history")
}

var (
	testStarted = time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	lintStarted = time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	testRunID   = traces.RunID(testStarted)
	lintRunID   = traces.RunID(lintStarted)
)

const latestBody = "latest view"

var _ = Describe("Handler", func() {
	var handler http.Handler

	get := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec
	}
	listAt := func(path string) runs.Page {
		rec := get(path)
		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		var page runs.Page
		Expect(json.Unmarshal(rec.Body.Bytes(), &page)).To(Succeed())
		return page
	}
	runIDs := func(page runs.Page) []string {
		var ids []string
		for _, run := range page.Runs {
			ids = append(ids, run.RunID)
		}
		return ids
	}

	BeforeEach(func() {
		workDir := GinkgoT().TempDir()
		_, err := snapshots.SavePerRun(workDir, &testui.Snapshot{
			Metadata: &testui.SnapshotMetadata{Started: testStarted, Ended: testStarted.Add(time.Minute)},
			Tests:    []parsers.Test{{Name: "TestPasses", Passed: true}, {Name: "TestFails", Failed: true}},
		}, testStarted, "")
		Expect(err).NotTo(HaveOccurred())
		_, err = snapshots.SavePerRun(workDir, &testui.Snapshot{
			Metadata: &testui.SnapshotMetadata{Started: lintStarted, Ended: lintStarted.Add(time.Minute)},
			Status:   testui.SnapshotStatus{LintRun: true},
			Lint:     []*linters.LinterResult{{Linter: "golangci-lint", Violations: make([]models.Violation, 3)}},
		}, lintStarted, "")
		Expect(err).NotTo(HaveOccurred())

		handler = runs.Handler(runs.Options{
			WorkDir:      workDir,
			LatestRunIDs: []string{lintRunID},
			Latest: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, latestBody)
			}),
		})
	})

	It("lists runs newest first with links to their views", func() {
		page := listAt("/api/runs")
		Expect(page.Current).To(BeEmpty())
		Expect(page.Latest).To(Equal([]string{lintRunID}))
		Expect(runIDs(page)).To(Equal([]string{lintRunID, testRunID}))
		Expect(page.Runs[0].Href).To(Equal("/runs/" + lintRunID + "/"))
		Expect(page.Runs[0].LintViolations).To(Equal(3))
		Expect(page.Runs[1].Failed).To(Equal(1))
	})

	It("pages with a cursor at the oldest run of the previous page", func() {
		first := listAt("/api/runs?limit=1")
		Expect(runIDs(first)).To(Equal([]string{lintRunID}))
		Expect(first.Next).To(Equal(lintRunID))
		second := listAt("/api/runs?limit=1&before=" + first.Next)
		Expect(runIDs(second)).To(Equal([]string{testRunID}))
	})

	It("serves a stored run's results under its prefix", func() {
		rec := get("/runs/" + testRunID + "/api/tests")
		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		var snap testui.Snapshot
		Expect(json.Unmarshal(rec.Body.Bytes(), &snap)).To(Succeed())
		Expect(snap.Tests).To(HaveLen(2))
		Expect(snap.Status.Running).To(BeFalse())
	})

	It("marks the viewed run as current in its listing", func() {
		Expect(listAt("/runs/" + testRunID + "/api/runs").Current).To(Equal(testRunID))
	})

	It("serves the page with the run's base path", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/runs/"+testRunID+"/history", nil)
		req.Header.Set("Accept", "text/html")
		handler.ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring(`window.__gavelBasePath = "/runs/` + testRunID + `";`))
	})

	It("redirects the bare run prefix to its page", func() {
		rec := get("/runs/" + testRunID)
		Expect(rec.Code).To(Equal(http.StatusMovedPermanently))
		Expect(rec.Header().Get("Location")).To(Equal("/runs/" + testRunID + "/"))
	})

	DescribeTable("rejects run ids that do not name a stored run",
		func(path string) { Expect(get(path).Code).To(Equal(http.StatusNotFound)) },
		Entry("unknown run", "/runs/run-2020-01-01T00-00-00Z/api/tests"),
		Entry("not a run file", "/runs/last/api/tests"),
		Entry("path traversal", "/runs/run-..%2F..%2Fsecret/api/tests"),
	)

	It("passes every other path to the latest view", func() {
		Expect(get("/tests").Body.String()).To(Equal(latestBody))
		Expect(get("/api/tests").Body.String()).To(Equal(latestBody))
	})
})
