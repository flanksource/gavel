package fixtures

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
)

type benchmarkSnapshotMetadata struct {
	Version string    `json:"version,omitempty"`
	Started time.Time `json:"started,omitempty"`
	Ended   time.Time `json:"ended,omitempty"`
	Kind    string    `json:"kind,omitempty"`
}

type benchmarkSnapshotStatus struct {
	Running bool `json:"running"`
}

type benchmarkSnapshot struct {
	Metadata    benchmarkSnapshotMetadata `json:"metadata"`
	Status      benchmarkSnapshotStatus   `json:"status"`
	Tests       []parsers.Test            `json:"tests"`
	Performance *BenchmarkPerformance     `json:"performance"`
}

func loadSnapshotBenchmarkReport(data []byte, path string) (*BenchmarkReport, error) {
	var snapshot benchmarkSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("decode benchmark snapshot %s: %w", path, err)
	}
	if snapshot.Performance == nil || snapshot.Performance.Mode != "benchmark" {
		return nil, fmt.Errorf("%s is not a benchmark snapshot", path)
	}
	if snapshot.Status.Running {
		return nil, fmt.Errorf("%s is an unfinished benchmark snapshot", path)
	}
	report := &BenchmarkReport{
		Version: benchmarkReportVersion, Mode: snapshot.Performance.Mode,
		StartedAt: snapshot.Metadata.Started, FinishedAt: snapshot.Metadata.Ended,
		DurationMS: snapshot.Performance.DurationMS, Status: snapshot.Performance.Status,
		Profile: snapshot.Performance.Profile, Comparison: snapshot.Performance.Comparison,
		Limits: snapshot.Performance.Limits, Violations: snapshot.Performance.Violations, Path: path,
		Files: []string{}, Phases: []BenchmarkEntry{}, Fixtures: []BenchmarkEntry{},
	}
	files := make(map[string]bool)
	var visit func([]parsers.Test)
	visit = func(tests []parsers.Test) {
		for _, test := range tests {
			if test.Fixture != nil {
				entry := testToBenchmarkEntry(test)
				switch entry.Kind {
				case "setup", "build", "daemon", "daemon-stop", "cleanup":
					report.Phases = append(report.Phases, entry)
				default:
					report.Fixtures = append(report.Fixtures, entry)
					if test.File != "" {
						files[test.File] = true
					}
				}
			}
			visit(test.Children)
		}
	}
	visit(snapshot.Tests)
	for file := range files {
		report.Files = append(report.Files, file)
	}
	sort.Strings(report.Files)
	return report, nil
}

func testToBenchmarkEntry(test parsers.Test) BenchmarkEntry {
	entry := BenchmarkEntry{Key: test.Fixture.Key, Name: test.Name, Kind: test.Fixture.Kind,
		Status: ExecutionState(test.Fixture.State), DurationMS: milliseconds(test.Duration),
		CommandMS: test.Fixture.CommandMS, Violations: test.Fixture.Violations}
	if test.File != "" {
		entry.Origin = &FixtureOrigin{File: test.File, Line: test.Line}
	}
	if test.FixtureProfile != nil {
		p := test.FixtureProfile
		entry.Profile = &FixtureProfile{Scope: p.Scope, PID: p.PID, SampleCount: p.SampleCount,
			PeakCPUPercent: p.PeakCPUPercent, PeakMemoryPercent: p.PeakMemoryPercent, PeakRSSBytes: p.PeakRSSBytes}
		if p.DiskIO != nil {
			entry.Profile.DiskIO = &ProfileDiskIO{DiskReadBytes: p.DiskIO.DiskReadBytes,
				DiskWriteBytes: p.DiskIO.DiskWriteBytes, SampleCount: p.DiskIO.SampleCount,
				MissingProcesses: p.DiskIO.MissingProcesses}
		}
	}
	if test.SQLProfile != nil {
		p := test.SQLProfile
		entry.SQLProfile = &SQLProfile{Path: p.Path, QueryCount: p.QueryCount,
			SlowQueryCount: p.SlowQueryCount, TotalDurationMS: p.TotalDurationMS, MaxQueryMS: p.MaxQueryMS}
	}
	return entry
}

type BenchmarkPerformance struct {
	Mode         string               `json:"mode"`
	Status       ExecutionState       `json:"status"`
	DurationMS   float64              `json:"duration_ms"`
	Profile      *ProfileReport       `json:"profile,omitempty"`
	Comparison   *BenchmarkComparison `json:"comparison,omitempty"`
	Limits       *BenchmarkLimits     `json:"limits,omitempty"`
	Violations   []BenchmarkViolation `json:"violations,omitempty"`
	ArtifactPath string               `json:"artifact_path,omitempty"`
}

func BenchmarkPerformanceFromReport(report *BenchmarkReport, path string) *BenchmarkPerformance {
	if report == nil {
		return nil
	}
	return &BenchmarkPerformance{
		Mode: report.Mode, Status: report.Status, DurationMS: report.DurationMS,
		Profile: report.Profile, Comparison: report.Comparison, Limits: report.Limits,
		Violations: report.Violations, ArtifactPath: path,
	}
}

func BenchmarkReportToTests(report *BenchmarkReport) []parsers.Test {
	if report == nil {
		return nil
	}
	tests := make([]parsers.Test, 0, len(report.Files)+1)
	if len(report.Phases) > 0 {
		lifecycle := parsers.Test{Name: "Fixture lifecycle", Framework: parsers.Fixture}
		for _, phase := range report.Phases {
			lifecycle.Children = append(lifecycle.Children, benchmarkEntryToTest(phase))
		}
		tests = append(tests, lifecycle)
	}
	files := make(map[string]int)
	for _, entry := range report.Fixtures {
		file := "Fixtures"
		if entry.Origin != nil && entry.Origin.File != "" {
			file = entry.Origin.File
		}
		index, ok := files[file]
		if !ok {
			index = len(tests)
			files[file] = index
			tests = append(tests, parsers.Test{Name: file, File: file, Framework: parsers.Fixture})
		}
		tests[index].Children = append(tests[index].Children, benchmarkEntryToTest(entry))
	}
	return tests
}

func benchmarkEntryToTest(entry BenchmarkEntry) parsers.Test {
	test := parsers.Test{
		Name: entry.Name, Framework: parsers.Fixture, TaskID: entry.Key,
		Duration: time.Duration(entry.DurationMS * float64(time.Millisecond)),
		Fixture: &parsers.FixtureExecution{Key: entry.Key, Kind: entry.Kind, State: string(entry.Status),
			CommandMS: entry.CommandMS, Violations: entry.Violations},
	}
	if entry.Origin != nil {
		test.File, test.Line = entry.Origin.File, entry.Origin.Line
	}
	if entry.Profile != nil {
		test.FixtureProfile = &parsers.FixtureProfile{
			Scope: entry.Profile.Scope, PID: entry.Profile.PID, SampleCount: entry.Profile.SampleCount,
			PeakCPUPercent: entry.Profile.PeakCPUPercent, PeakMemoryPercent: entry.Profile.PeakMemoryPercent,
			PeakRSSBytes: entry.Profile.PeakRSSBytes,
		}
		if entry.Profile.DiskIO != nil {
			disk := entry.Profile.DiskIO
			test.FixtureProfile.DiskIO = &parsers.ProfileDiskIO{DiskReadBytes: disk.DiskReadBytes,
				DiskWriteBytes: disk.DiskWriteBytes, SampleCount: disk.SampleCount, MissingProcesses: disk.MissingProcesses}
		}
	}
	for _, artifact := range entry.GoProfiles {
		test.GoProfiles = append(test.GoProfiles, parsers.GoProfileArtifact{
			Name: artifact.Name, ID: artifact.ID, Path: artifact.Path, Status: artifact.Status,
			Bytes: artifact.Bytes, SampleTypes: artifact.SampleTypes, Error: artifact.Error,
		})
	}
	if entry.SQLProfile != nil {
		profile := entry.SQLProfile
		test.SQLProfile = &parsers.SQLProfile{Path: profile.Path, QueryCount: profile.QueryCount,
			SlowQueryCount: profile.SlowQueryCount, TotalDurationMS: profile.TotalDurationMS,
			MaxQueryMS: profile.MaxQueryMS}
		for _, statement := range profile.Statements {
			test.SQLProfile.Statements = append(test.SQLProfile.Statements, parsers.SQLProfileStatement{
				SQL: statement.SQL, Params: statement.Params, DurationMS: statement.DurationMS,
				Rows: statement.Rows, Slow: statement.Slow, Error: statement.Error,
			})
		}
	}
	switch entry.Status {
	case ExecutionQueued:
		test.Pending = true
	case ExecutionRunning:
		test.Running = true
	case ExecutionPassed:
		test.Passed = true
	case ExecutionFailed, ExecutionErrored, ExecutionTimedOut:
		test.Failed = true
		test.TimedOut = entry.Status == ExecutionTimedOut
	case ExecutionWarned:
		test.Warned = true
	case ExecutionSkipped, ExecutionCancelled:
		test.Skipped = true
	}
	if len(entry.Violations) > 0 {
		test.Failed = true
		test.Passed = false
		test.Message = entry.Violations[0]
	}
	return test
}
