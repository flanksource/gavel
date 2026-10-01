package fixturebenchtest

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flanksource/gavel/fixtures"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFixtureBenchmarkUI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fixture benchmark UI")
}

var _ = Describe("fixture benchmark snapshots", func() {
	It("publishes progress and the final report without ending a live run, then replays them", func() {
		server := testui.NewServer()
		server.BeginRun("initial")
		server.SetFixtureBenchmarkMode("benchmark")
		started := time.Now().UTC()
		progress := fixtures.ExecutionSnapshot{StartedAt: &started, State: fixtures.ExecutionRunning}
		server.SetFixtureBenchmarkProgress(progress)
		server.SetFixtureBenchmarkReport(&fixtures.BenchmarkReport{Version: 1, Mode: "benchmark", Status: fixtures.ExecutionPassed, DurationMS: 125}, "artifact.json")

		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/api/tests", nil))
		Expect(response.Code).To(Equal(200))
		var snapshot testui.Snapshot
		Expect(json.Unmarshal(response.Body.Bytes(), &snapshot)).To(Succeed())
		Expect(snapshot.Status.Running).To(BeTrue())
		Expect(snapshot.Performance).ToNot(BeNil())
		Expect(snapshot.Performance.Mode).To(Equal("benchmark"))
		Expect(snapshot.Performance.Status).To(Equal(fixtures.ExecutionPassed))
		Expect(snapshot.Performance.DurationMS).To(Equal(125.0))
		Expect(snapshot.Performance.ArtifactPath).To(Equal("artifact.json"))

		server.MarkDone()
		replayed := testui.NewServer()
		completed := server.Snapshot()
		completed.Error = "profile sample failed"
		replayed.LoadSnapshot(completed)
		Expect(replayed.Snapshot().Performance).To(Equal(server.Snapshot().Performance))
		Expect(replayed.Snapshot().Error).To(Equal("profile sample failed"))
		replayed.BeginRun("rerun")
		Expect(replayed.Snapshot().Performance).To(BeNil())
	})

	It("exports the complete fixture report and a readable comparison", func() {
		server := testui.NewServer()
		server.BeginRun("initial")
		server.SetFixtureBenchmarkMode("benchmark")
		report := &fixtures.BenchmarkReport{
			Version: 1, Mode: "benchmark", DurationMS: 125,
			Phases: []fixtures.BenchmarkEntry{{Key: "setup", Name: "Setup", Kind: "setup", DurationMS: 125}},
			Fixtures: []fixtures.BenchmarkEntry{{Key: "test", Name: "SQL fixture", Kind: "test", Status: fixtures.ExecutionFailed,
				SQLProfile: &fixtures.SQLProfile{QueryCount: 2, TotalDurationMS: 30, MaxQueryMS: 20,
					Statements: []fixtures.SQLProfileStatement{{SQL: "SELECT * FROM users WHERE id = ?", Params: []string{"42"}, DurationMS: 20, Rows: 1}}},
				Violations: []string{"SQL total 30.00ms exceeds 25.00ms"}}},
			Comparison: &fixtures.BenchmarkComparison{Baseline: "baseline.json", Deltas: []fixtures.BenchmarkDelta{{Key: "setup", Name: "Setup", BaselineMS: 100, CurrentMS: 125, DeltaMS: 25}}},
			Profile:    &fixtures.ProfileReport{PeakRSSBytes: 1024, Samples: []fixtures.ProfileSample{{RSSBytes: 1024}}},
		}
		server.LoadSnapshot(testui.Snapshot{Status: testui.SnapshotStatus{Running: false},
			Tests:       fixtures.BenchmarkReportToTests(report),
			Performance: fixtures.BenchmarkPerformanceFromReport(report, "report.json")})

		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/bench.json", nil))
		Expect(response.Code).To(Equal(200))
		Expect(response.Body.String()).To(ContainSubstring("\"performance\""))
		Expect(response.Body.String()).To(ContainSubstring("\"peak_rss_bytes\""))
		Expect(response.Body.String()).To(ContainSubstring("\"artifact_path\""))
		Expect(response.Body.String()).To(ContainSubstring("\"sql_profile\""))
		Expect(response.Body.String()).To(ContainSubstring("SELECT * FROM users WHERE id = ?"))
		var exported struct {
			Performance fixtures.BenchmarkPerformance `json:"performance"`
			Bench       []struct {
				SQLProfile *fixtures.SQLProfile `json:"sql_profile"`
			} `json:"bench"`
		}
		Expect(json.Unmarshal(response.Body.Bytes(), &exported)).To(Succeed())
		Expect(exported.Bench[1].SQLProfile.Statements[0].Params).To(Equal([]string{"42"}))
		Expect(response.Body.String()).To(ContainSubstring("SQL total 30.00ms exceeds 25.00ms"))

		response = httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/bench.md", nil))
		Expect(response.Code).To(Equal(200))
		Expect(response.Body.String()).To(ContainSubstring("Setup"))
		Expect(response.Body.String()).To(ContainSubstring("baseline.json"))
	})
})
