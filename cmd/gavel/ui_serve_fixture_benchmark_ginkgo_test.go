package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("detached fixture benchmark snapshot", func() {
	It("returns lifecycle and fixture work in the test runner tree", func() {
		workDir := GinkgoT().TempDir()
		path := filepath.Join(workDir, "lifecycle.fixture.md")
		Expect(os.WriteFile(path, []byte("---\nbuild: printf ready\n---\n\n# Lifecycle\n\n```bash\nprintf passed\n```\n\n- contains: passed\n"), 0o600)).To(Succeed())
		result, err := testrunner.Run(testrunner.RunOptions{WorkDir: workDir, FixturesOnly: true, FixtureRunner: &fixtures.RunnerOptions{
			Paths: []string{path}, WorkDir: workDir, Benchmark: true,
		}})
		Expect(err).NotTo(HaveOccurred())
		tests, ok := result.([]parsers.Test)
		Expect(ok).To(BeTrue())
		Expect(tests).NotTo(BeEmpty())
		Expect(tests[0].Name).To(Equal("Fixture lifecycle"))
		Expect(tests[0].Children).To(HaveLen(1))
		Expect(tests[0].Children[0].Fixture.Kind).To(Equal("build"))
		Expect(tests[0].Children[0].Passed).To(BeTrue())
		var fixtureLeaf *parsers.Test
		var visit func([]parsers.Test)
		visit = func(nodes []parsers.Test) {
			for i := range nodes {
				if nodes[i].Fixture != nil && nodes[i].Fixture.Kind == "command" {
					fixtureLeaf = &nodes[i]
				}
				visit(nodes[i].Children)
			}
		}
		visit(tests)
		Expect(fixtureLeaf).NotTo(BeNil())
		Expect(fixtureLeaf.Stdout).To(ContainSubstring("passed"))
	})

	It("loads a historical benchmark file as a test snapshot", func() {
		path := filepath.Join(GinkgoT().TempDir(), "run.json")
		started := time.Date(2026, time.September, 27, 17, 16, 11, 0, time.UTC)
		data, err := json.Marshal(fixtures.BenchmarkReport{
			Version: 1, Mode: "benchmark", Status: fixtures.ExecutionFailed,
			StartedAt: started, FinishedAt: started.Add(125 * time.Millisecond), DurationMS: 125,
			Phases:   []fixtures.BenchmarkEntry{{Key: "build", Name: "Build", Kind: "build", Status: fixtures.ExecutionPassed, DurationMS: 125}},
			Fixtures: []fixtures.BenchmarkEntry{{Key: "fixture-1", Name: "Fixture 1", Kind: "test", Status: fixtures.ExecutionFailed, DurationMS: 50}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(path, data, 0o600)).To(Succeed())

		server := testui.NewServer()
		Expect(loadResults(server, path)).To(Succeed())
		snapshot := server.Snapshot()
		Expect(snapshot.Tests).NotTo(BeEmpty())
		Expect(snapshot.Tests[0].Name).To(Equal("Fixture lifecycle"))
		Expect(snapshot.Tests[0].Children[0].Name).To(Equal("Build"))
		Expect(snapshot.Metadata.Started).To(Equal(started))
		Expect(snapshot.Metadata.Ended).To(Equal(started.Add(125 * time.Millisecond)))
	})
	It("loads a newly persisted benchmark snapshot without conversion", func() {
		workDir := GinkgoT().TempDir()
		path := filepath.Join(workDir, "sample.fixture.md")
		Expect(os.WriteFile(path, []byte("---\nbuild: printf ready\n---\n\n# Sample\n\n```bash\nprintf passed\n```\n"), 0o600)).To(Succeed())
		runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{path}, WorkDir: workDir, Benchmark: true})
		Expect(err).NotTo(HaveOccurred())
		_, err = runner.Run()
		Expect(err).NotTo(HaveOccurred())

		server := testui.NewServer()
		Expect(loadResults(server, runner.BenchmarkReport().Path)).To(Succeed())
		snapshot := server.Snapshot()
		Expect(snapshot.Performance.Mode).To(Equal("benchmark"))
		Expect(snapshot.Tests[0].Name).To(Equal("Fixture lifecycle"))
		Expect(snapshot.Tests[0].Children[0].Fixture.Kind).To(Equal("build"))
		Expect(snapshot.Tests[1].Children).NotTo(BeEmpty())
	})

	It("preserves the report and failure when loading result files", func() {
		source := testui.Snapshot{
			FixtureBenchmark: &testui.FixtureBenchmarkState{Mode: "benchmark", Report: &fixtures.BenchmarkReport{Version: 1, DurationMS: 125}, ArtifactPath: "run.json"},
			Error:            "profile sample failed",
		}
		merged := mergeSnapshots(testui.Snapshot{}, source)
		Expect(merged.FixtureBenchmark).To(Equal(source.FixtureBenchmark))
		Expect(merged.Error).To(Equal(source.Error))
	})
})
