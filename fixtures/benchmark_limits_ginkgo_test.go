package fixtures

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture benchmark limits", func() {
	It("fails each exceeded absolute and baseline metric", func() {
		pct := 20.0
		entry := BenchmarkEntry{Key: "test", Name: "Test", Status: ExecutionPassed, DurationMS: 130,
			Profile:    &FixtureProfile{Scope: "process_tree", SampleCount: 1, PeakRSSBytes: 130, DiskIO: &ProfileDiskIO{DiskReadBytes: 130, DiskWriteBytes: 10}},
			SQLProfile: &SQLProfile{TotalDurationMS: 30, MaxQueryMS: 20}}
		baseline := BenchmarkEntry{Key: "test", DurationMS: 100,
			Profile:    &FixtureProfile{Scope: "process_tree", SampleCount: 1, PeakRSSBytes: 100, DiskIO: &ProfileDiskIO{DiskReadBytes: 100, DiskWriteBytes: 10}},
			SQLProfile: &SQLProfile{TotalDurationMS: 20, MaxQueryMS: 10}}
		limits := BenchmarkLimits{MaxDeviationPct: &pct, MaxDurationMS: 120, MaxRSSBytes: 120, MaxDiskReadBytes: 120, MaxSQLDurationMS: 25, MaxSQLQueryMS: 15}
		violations := checkBenchmarkLimits(entry, &baseline, limits)
		Expect(violations).To(ConsistOf(
			ContainSubstring("time"), ContainSubstring("memory"), ContainSubstring("disk read"),
			ContainSubstring("SQL total"), ContainSubstring("SQL query"),
			ContainSubstring("time baseline"), ContainSubstring("memory baseline"),
			ContainSubstring("disk read baseline"), ContainSubstring("SQL total baseline"), ContainSubstring("SQL query baseline"),
		))
	})

	It("fails when a required measurement is missing", func() {
		violations := checkBenchmarkLimits(BenchmarkEntry{Key: "test", Status: ExecutionPassed}, nil,
			BenchmarkLimits{MaxRSSBytes: 1, MaxDiskReadBytes: 1, MaxSQLDurationMS: 1})
		Expect(violations).To(ConsistOf(ContainSubstring("memory unmeasured"), ContainSubstring("disk read unmeasured"), ContainSubstring("SQL total unmeasured")))
	})

	It("enforces SQL query and slow-query counts, including a zero slow-query budget", func() {
		maxQueries, maxSlow := 2, 0
		violations := checkBenchmarkLimits(BenchmarkEntry{Key: "test", Status: ExecutionPassed,
			SQLProfile: &SQLProfile{QueryCount: 3, SlowQueryCount: 1}}, nil,
			BenchmarkLimits{MaxSQLQueries: &maxQueries, MaxSlowSQL: &maxSlow})
		Expect(violations).To(ConsistOf(ContainSubstring("SQL queries"), ContainSubstring("slow SQL queries")))
	})
})
