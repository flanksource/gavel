package fixtures_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/types"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture benchmarking", func() {
	It("marks a fixture failed and persists its exceeded time limit", func() {
		workDir := GinkgoT().TempDir()
		path := filepath.Join(workDir, "limit.fixture.md")
		Expect(os.WriteFile(path, []byte("# Limit\n\n```bash\nsleep 0.02\n```\n"), 0o600)).To(Succeed())
		runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{path}, WorkDir: workDir, Benchmark: true,
			Limits: fixtures.BenchmarkLimits{MaxDurationMS: 1}})
		Expect(err).NotTo(HaveOccurred())
		tree, err := runner.Run()
		Expect(err).To(MatchError(ContainSubstring("fixture benchmark limit")))
		Expect(tree.Stats.Failed).To(Equal(1))
		report := runner.BenchmarkReport()
		Expect(report.Status).To(Equal(fixtures.ExecutionFailed))
		Expect(report.Fixtures[0].Violations).To(ContainElement(ContainSubstring("time")))
		Expect(report.Path).To(BeAnExistingFile())
		var saved testui.Snapshot
		data, readErr := os.ReadFile(report.Path)
		Expect(readErr).NotTo(HaveOccurred())
		Expect(json.Unmarshal(data, &saved)).To(Succeed())
		Expect(saved.Status.Running).To(BeFalse())
		Expect(saved.Performance.Status).To(Equal(fixtures.ExecutionFailed))
	})

	It("collects SQL export from a fixture and applies its limit", func() {
		workDir := GinkgoT().TempDir()
		path := filepath.Join(workDir, "sql-limit.fixture.md")
		Expect(os.WriteFile(path, []byte("# SQL\n\n```bash\nprintf '{\"duration_ns\":30000000,\"rows\":1}\n' > \"$SQL_PROFILE_FILE\"\n```\n"), 0o600)).To(Succeed())
		runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{path}, WorkDir: workDir, Benchmark: true,
			Limits: fixtures.BenchmarkLimits{MaxSQLDurationMS: 20}})
		Expect(err).NotTo(HaveOccurred())
		_, err = runner.Run()
		Expect(err).To(MatchError(ContainSubstring("fixture benchmark limit")))
		Expect(runner.BenchmarkReport().Fixtures[0].SQLProfile.TotalDurationMS).To(BeNumerically("==", 30))
	})

	It("fails a SQL budget when the command never initializes the SQL logger", func() {
		workDir := GinkgoT().TempDir()
		path := filepath.Join(workDir, "missing-sql.fixture.md")
		Expect(os.WriteFile(path, []byte("# Missing SQL\n\n```bash\nprintf ready\n```\n"), 0o600)).To(Succeed())
		runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{path}, WorkDir: workDir, Benchmark: true,
			Limits: fixtures.BenchmarkLimits{MaxSQLDurationMS: 20}})
		Expect(err).NotTo(HaveOccurred())
		_, err = runner.Run()
		Expect(err).To(MatchError(ContainSubstring("fixture benchmark limit")))
		Expect(runner.BenchmarkReport().Fixtures[0].Violations).To(ContainElement("SQL total unmeasured"))
	})

	It("fails a fixture whose duration regresses beyond its baseline percentage", func() {
		workDir := GinkgoT().TempDir()
		path := filepath.Join(workDir, "deviation.fixture.md")
		Expect(os.WriteFile(path, []byte("# Deviation\n\n```bash\nsleep 0.01\n```\n"), 0o600)).To(Succeed())
		baseline, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{path}, WorkDir: workDir, Benchmark: true})
		Expect(err).NotTo(HaveOccurred())
		_, err = baseline.Run()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(path, []byte("# Deviation\n\n```bash\nsleep 0.1\n```\n"), 0o600)).To(Succeed())
		zero := 0.0
		current, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{path}, WorkDir: workDir, Benchmark: true,
			Baseline: baseline.BenchmarkReport().Path, Limits: fixtures.BenchmarkLimits{MaxDeviationPct: &zero}})
		Expect(err).NotTo(HaveOccurred())
		_, err = current.Run()
		Expect(err).To(MatchError(ContainSubstring("fixture benchmark limit")))
		Expect(current.BenchmarkReport().Fixtures[0].Violations).To(ContainElement(ContainSubstring("time baseline deviation")))
	})

	It("records passing fixtures and prerequisite durations even when passing results are hidden", func() {
		workDir := GinkgoT().TempDir()
		path := filepath.Join(workDir, "benchmark.fixture.md")
		Expect(os.WriteFile(path, []byte("---\nsetup:\n  envVars:\n    - name: BENCH_SETUP\n      value: ready\nbuild: printf ready\n---\n\n# Bench\n\n```bash\nsleep 0.02\nprintf passed\n```\n\n- contains: passed\n"), 0o600)).To(Succeed())
		runner, err := fixtures.NewRunner(fixtures.RunnerOptions{
			Paths: []string{path}, WorkDir: workDir, Benchmark: true,
			Display: &fixtures.DisplayOptions{ShowPassed: false},
		})
		Expect(err).NotTo(HaveOccurred())
		tree, err := runner.Run()
		Expect(err).NotTo(HaveOccurred())
		var visible int
		tree.Walk(func(node *fixtures.FixtureNode) {
			if node.Results != nil {
				visible++
			}
		})
		Expect(visible).To(Equal(1))
		report := runner.BenchmarkReport()
		Expect(report).NotTo(BeNil())
		Expect(report.Path).To(BeAnExistingFile())
		Expect(report.Phases).To(ContainElement(HaveField("Kind", "build")))
		Expect(report.Phases).To(ContainElement(HaveField("Kind", "setup")))
		Expect(report.Phases).To(ContainElement(HaveField("Kind", "cleanup")))
		Expect(report.Fixtures).To(HaveLen(1))
		Expect(report.Fixtures[0].DurationMS).To(BeNumerically(">", 0))
		Expect(report.Fixtures[0].CommandMS).NotTo(BeNil())
		var saved testui.Snapshot
		data, err := os.ReadFile(report.Path)
		Expect(err).NotTo(HaveOccurred())
		Expect(json.Unmarshal(data, &saved)).To(Succeed())
		Expect(saved.Tests).NotTo(BeEmpty())
		Expect(saved.Performance).NotTo(BeNil())
	})

	It("compares the same fixture with a saved baseline", func() {
		workDir := GinkgoT().TempDir()
		path := filepath.Join(workDir, "comparison.fixture.md")
		write := func(delay string) {
			GinkgoHelper()
			content := "# Timing\n\n```bash\nsleep " + delay + "\n```\n"
			Expect(os.WriteFile(path, []byte(content), 0o600)).To(Succeed())
		}
		write("0.02")
		baseline, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{path}, WorkDir: workDir, Benchmark: true})
		Expect(err).NotTo(HaveOccurred())
		_, err = baseline.Run()
		Expect(err).NotTo(HaveOccurred())
		write("0.06")
		current, err := fixtures.NewRunner(fixtures.RunnerOptions{
			Paths: []string{path}, WorkDir: workDir, Benchmark: true,
			Baseline: filepath.Join(".gavel", "benchmarks", "fixtures", filepath.Base(baseline.BenchmarkReport().Path)),
		})
		Expect(err).NotTo(HaveOccurred())
		_, err = current.Run()
		Expect(err).NotTo(HaveOccurred())
		comparison := current.BenchmarkReport().Comparison
		Expect(comparison).NotTo(BeNil())
		Expect(comparison.Deltas).To(HaveLen(1))
		Expect(comparison.Deltas[0].DeltaMS).To(BeNumerically(">", 0))
		Expect(comparison.Added).To(BeEmpty())
		Expect(comparison.Missing).To(BeEmpty())
		data, err := os.ReadFile(current.BenchmarkReport().Path)
		Expect(err).NotTo(HaveOccurred())
		var saved testui.Snapshot
		Expect(json.Unmarshal(data, &saved)).To(Succeed())
		Expect(saved.Performance.Comparison).NotTo(BeNil())
		Expect(saved.Performance.Comparison.Deltas).To(HaveLen(1))
	})

	It("records Clicky process samples for a running child", func() {
		workDir := GinkgoT().TempDir()
		path := filepath.Join(workDir, "profile.fixture.md")
		Expect(os.WriteFile(path, []byte("# Profile\n\n```bash\nsleep 0.8\n```\n"), 0o600)).To(Succeed())
		for _, benchmark := range []bool{false, true} {
			runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{path}, WorkDir: workDir, Benchmark: benchmark, Profile: true})
			Expect(err).NotTo(HaveOccurred())
			tree, err := runner.Run()
			Expect(err).NotTo(HaveOccurred())
			report := runner.BenchmarkReport()
			Expect(report.Profile).NotTo(BeNil())
			Expect(report.Path).To(BeAnExistingFile())
			Expect(report.Profile.Samples).NotTo(BeEmpty())
			Expect(report.Fixtures).To(HaveLen(1))
			Expect(report.Fixtures[0].ProfileSamples).NotTo(BeNil())
			Expect(*report.Fixtures[0].ProfileSamples).To(BeNumerically(">", 0))
			Expect(report.Fixtures[0].Profile).NotTo(BeNil())
			Expect(report.Fixtures[0].Profile.Scope).To(Equal("process_tree"))
			Expect(report.Fixtures[0].Profile.SampleCount).To(BeNumerically(">", 0))
			Expect(report.Fixtures[0].Profile.PeakRSSBytes).To(BeNumerically(">", 0))
			Expect(report.Fixtures[0].Profile.PID).To(BeNumerically(">", 0))
			Expect(report.Fixtures[0].Profile.PeakRSSBytes).To(BeNumerically("<", report.Profile.PeakRSSBytes))
			var resultProfile *fixtures.FixtureProfile
			tree.Walk(func(node *fixtures.FixtureNode) {
				if node.Results != nil && node.Results.Profile != nil {
					resultProfile = node.Results.Profile
				}
			})
			Expect(resultProfile).To(Equal(report.Fixtures[0].Profile))
			var sawChild bool
			for _, sample := range report.Profile.Samples {
				for _, child := range sample.Processes {
					sawChild = sawChild || strings.Contains(child.Command, "sleep 0.8")
				}
			}
			Expect(sawChild).To(BeTrue(), "profile should include the fixture child process")
			data, readErr := os.ReadFile(report.Path)
			Expect(readErr).NotTo(HaveOccurred())
			var saved testui.Snapshot
			Expect(json.Unmarshal(data, &saved)).To(Succeed())
			Expect(saved.Performance.Profile).NotTo(BeNil())
			Expect(saved.Performance.Profile.Samples).NotTo(BeEmpty())
			Expect(saved.Tests[0].Children[0].FixtureProfile.PeakRSSBytes).To(Equal(report.Fixtures[0].Profile.PeakRSSBytes))
			Expect(string(data)).NotTo(ContainSubstring("\"environment\""))
		}
	})

	It("saves timings and failure state when a prerequisite fails", func() {
		workDir := GinkgoT().TempDir()
		path := filepath.Join(workDir, "failure.fixture.md")
		Expect(os.WriteFile(path, []byte("---\nbuild: printf private-marker; exit 7\n---\n\n# Check\n\n```bash\nprintf skipped\n```\n"), 0o600)).To(Succeed())
		runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{path}, WorkDir: workDir, Benchmark: true})
		Expect(err).NotTo(HaveOccurred())
		tree, err := runner.Run()
		Expect(err).To(HaveOccurred())
		Expect(tree).NotTo(BeNil())
		report := runner.BenchmarkReport()
		Expect(report.Path).To(BeAnExistingFile())
		Expect(report.Phases).To(ContainElement(And(HaveField("Kind", "build"), HaveField("Status", fixtures.ExecutionErrored))))
		Expect(report.Fixtures).To(ContainElement(And(HaveField("Kind", "command"), HaveField("Status", fixtures.ExecutionCancelled))))
		data, readErr := os.ReadFile(report.Path)
		Expect(readErr).NotTo(HaveOccurred())
		Expect(string(data)).NotTo(ContainSubstring("private-marker"))
	})

	It("rejects a baseline without benchmark mode", func() {
		_, err := fixtures.NewRunner(fixtures.RunnerOptions{Baseline: "previous.json"})
		Expect(err).To(MatchError(ContainSubstring("--baseline requires --benchmark")))
	})
})
