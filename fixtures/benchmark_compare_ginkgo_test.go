package fixtures

import (
	"time"

	"github.com/flanksource/clicky/process"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture benchmark comparison", func() {
	It("keeps zero-baseline percentages undefined and cancelled work unmeasured", func() {
		baseline := &BenchmarkReport{
			Phases:   []BenchmarkEntry{{Key: "setup", Name: "Setup", DurationMS: 0, Status: ExecutionPassed}},
			Fixtures: []BenchmarkEntry{{Key: "command", Name: "Command", DurationMS: 100, Status: ExecutionPassed}},
			Profile:  &ProfileReport{Samples: []ProfileSample{{}}, PeakCPUPercent: 10, PeakRSSBytes: 100},
		}
		current := &BenchmarkReport{
			Phases: []BenchmarkEntry{{Key: "setup", Name: "Setup", DurationMS: 2, Status: ExecutionPassed}},
			Fixtures: []BenchmarkEntry{
				{Key: "command", Name: "Command", DurationMS: 0, Status: ExecutionCancelled},
				{Key: "new-command", Name: "New command", DurationMS: 0, Status: ExecutionCancelled},
			},
			Profile: &ProfileReport{Samples: []ProfileSample{{}}, PeakCPUPercent: 15, PeakRSSBytes: 150},
		}
		comparison := compareBenchmarkReports("baseline.json", baseline, current)
		Expect(comparison.Deltas).To(HaveLen(1))
		Expect(comparison.Deltas[0].Key).To(Equal("setup"))
		Expect(comparison.Deltas[0].DeltaPct).To(BeNil())
		Expect(comparison.Unmeasured).To(Equal([]string{"command", "new-command"}))
		Expect(comparison.Added).To(BeEmpty())
		Expect(comparison.PeakCPUPercentDelta).To(HaveValue(BeNumerically("==", 5)))
		Expect(comparison.PeakRSSBytesDelta).To(HaveValue(BeNumerically("==", 50)))
	})
})

var _ = Describe("fixture process profile", func() {
	It("parses file-level Go profile output declarations", func() {
		root, err := ParseMarkdownDocument("profiles.fixture.md", "---\nexec: sh\ngoProfiles:\n  cpu:\n    file: cpu.pprof\n    env: CPU_PROFILE\n---\n# Profile\n```sh\necho ok\n```\n", ".")
		Expect(err).NotTo(HaveOccurred())
		fixture := firstFixtureTest(root)
		Expect(fixture).NotTo(BeNil())
		Expect(fixture.ExecBase().GoProfiles).To(Equal(map[string]GoProfileOutput{
			"cpu": {File: "cpu.pprof", Env: "CPU_PROFILE"},
		}))
	})

	It("lets a command block replace file-level Go profile outputs", func() {
		root, err := ParseMarkdownDocument("profiles.fixture.md", "---\ngoProfiles:\n  cpu:\n    file: cpu.pprof\n    env: CPU_PROFILE\n---\n# Profile\n```exec\ncontent: echo ok\ngoProfiles:\n  heap:\n    file: heap.pprof\n    env: HEAP_PROFILE\n```\n", ".")
		Expect(err).NotTo(HaveOccurred())
		fixture := firstFixtureTest(root)
		Expect(fixture).NotTo(BeNil())
		Expect(fixture.ExecBase().GoProfiles).To(Equal(map[string]GoProfileOutput{
			"heap": {File: "heap.pprof", Env: "HEAP_PROFILE"},
		}))
	})

	It("keeps a child's disk bytes after exit on Darwin", func() {
		started := time.Date(2026, time.January, 1, 0, 0, 0, 500_000_000, time.UTC)
		processStarted := started.Truncate(time.Second)
		finished := started.Add(3 * time.Second)
		samples := []ProfileSample{
			{SampledAt: started.Add(time.Second), Processes: []process.Process{
				{PID: 100, StartedAt: &processStarted, IO: &process.ProcessIO{DiskWriteBytes: 5}},
				{PID: 101, PPID: 100, StartedAt: &processStarted, IO: &process.ProcessIO{DiskWriteBytes: 10}},
			}},
			{SampledAt: started.Add(2 * time.Second), Processes: []process.Process{
				{PID: 100, StartedAt: &processStarted, IO: &process.ProcessIO{DiskWriteBytes: 7}},
			}},
		}
		Expect(profileDiskIO(samples, &ExecutionNode{StartedAt: &started, FinishedAt: &finished}, 100, "darwin")).To(Equal(&ProfileDiskIO{DiskWriteBytes: 17, SampleCount: 2}))
	})

	It("does not count waited-for child bytes twice on Linux", func() {
		started := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		finished := started.Add(3 * time.Second)
		samples := []ProfileSample{
			{SampledAt: started.Add(time.Second), Processes: []process.Process{
				{PID: 100, StartedAt: &started, IO: &process.ProcessIO{DiskWriteBytes: 5}},
				{PID: 101, PPID: 100, StartedAt: &started, IO: &process.ProcessIO{DiskWriteBytes: 10}},
			}},
			{SampledAt: started.Add(2 * time.Second), Processes: []process.Process{
				{PID: 100, StartedAt: &started, IO: &process.ProcessIO{DiskWriteBytes: 17}},
			}},
		}
		Expect(profileDiskIO(samples, &ExecutionNode{StartedAt: &started, FinishedAt: &finished}, 100, "linux")).To(Equal(&ProfileDiskIO{DiskWriteBytes: 17, SampleCount: 2}))
	})

	It("attributes command descendants within the test window and excludes other processes", func() {
		started := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		finished := started.Add(2 * time.Second)
		node := &ExecutionNode{StartedAt: &started, FinishedAt: &finished}
		samples := []ProfileSample{
			{SampledAt: started.Add(-time.Millisecond), Processes: []process.Process{{PID: 100, RSSBytes: 1000}}},
			{SampledAt: started.Add(time.Second), CPUPercent: 105, MemoryPercent: 10.5, RSSBytes: 1050, Processes: []process.Process{
				{PID: 100, CPUPercent: 10, MemoryPercent: 1, RSSBytes: 100},
				{PID: 101, PPID: 100, CPUPercent: 5, MemoryPercent: 0.5, RSSBytes: 50},
				{PID: 200, CPUPercent: 90, MemoryPercent: 9, RSSBytes: 900},
			}},
			{SampledAt: finished.Add(time.Millisecond), Processes: []process.Process{{PID: 100, RSSBytes: 1000}}},
		}

		Expect(profileFixtureNode(samples, node, 100)).To(Equal(&FixtureProfile{
			Scope: "process_tree", PID: 100, SampleCount: 1,
			PeakCPUPercent: 15, PeakMemoryPercent: 1.5, PeakRSSBytes: 150,
		}))
		Expect(profileFixtureNode(samples, node, 0)).To(Equal(&FixtureProfile{
			Scope: "run_tree", SampleCount: 1,
			PeakCPUPercent: 105, PeakMemoryPercent: 10.5, PeakRSSBytes: 1050,
		}))
	})
})
