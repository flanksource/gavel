package fixtures

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	clickyexec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/process"
	"github.com/flanksource/commons/logger"
)

const benchmarkReportVersion = 1

type BenchmarkEntry struct {
	Key            string              `json:"key"`
	Name           string              `json:"name"`
	Kind           string              `json:"kind"`
	Status         ExecutionState      `json:"status"`
	Origin         *FixtureOrigin      `json:"origin,omitempty"`
	DurationMS     float64             `json:"duration_ms"`
	CommandMS      *float64            `json:"command_ms,omitempty"`
	ProfileSamples *int                `json:"profile_samples,omitempty"`
	Profile        *FixtureProfile     `json:"profile,omitempty"`
	GoProfiles     []GoProfileArtifact `json:"go_profiles,omitempty"`
	SQLProfile     *SQLProfile         `json:"sql_profile,omitempty"`
	Violations     []string            `json:"violations,omitempty"`
}

type FixtureProfile struct {
	Scope             string         `json:"scope"`
	PID               int            `json:"pid,omitempty"`
	SampleCount       int            `json:"sample_count"`
	PeakCPUPercent    float64        `json:"peak_cpu_percent"`
	PeakMemoryPercent float64        `json:"peak_memory_percent"`
	PeakRSSBytes      uint64         `json:"peak_rss_bytes"`
	DiskIO            *ProfileDiskIO `json:"disk_io,omitempty"`
}

type ProfileDiskIO struct {
	DiskReadBytes    uint64 `json:"disk_read_bytes"`
	DiskWriteBytes   uint64 `json:"disk_write_bytes"`
	SampleCount      int    `json:"sample_count"`
	MissingProcesses int    `json:"missing_processes,omitempty"`
}

type BenchmarkDelta struct {
	Key                 string   `json:"key"`
	Name                string   `json:"name"`
	Kind                string   `json:"kind"`
	BaselineMS          float64  `json:"baseline_ms"`
	CurrentMS           float64  `json:"current_ms"`
	DeltaMS             float64  `json:"delta_ms"`
	DeltaPct            *float64 `json:"delta_pct,omitempty"`
	DiskReadBytesDelta  *int64   `json:"disk_read_bytes_delta,omitempty"`
	DiskWriteBytesDelta *int64   `json:"disk_write_bytes_delta,omitempty"`
}

type BenchmarkComparison struct {
	Baseline            string           `json:"baseline"`
	Deltas              []BenchmarkDelta `json:"deltas"`
	Added               []string         `json:"added,omitempty"`
	Missing             []string         `json:"missing,omitempty"`
	Unmeasured          []string         `json:"unmeasured,omitempty"`
	PeakCPUPercentDelta *float64         `json:"peak_cpu_percent_delta,omitempty"`
	PeakRSSBytesDelta   *int64           `json:"peak_rss_bytes_delta,omitempty"`
}

type BenchmarkReport struct {
	Version    int                  `json:"version"`
	Mode       string               `json:"mode"`
	StartedAt  time.Time            `json:"started_at"`
	FinishedAt time.Time            `json:"finished_at"`
	DurationMS float64              `json:"duration_ms"`
	Status     ExecutionState       `json:"status"`
	Files      []string             `json:"files"`
	Phases     []BenchmarkEntry     `json:"phases"`
	Fixtures   []BenchmarkEntry     `json:"fixtures"`
	Profile    *ProfileReport       `json:"profile,omitempty"`
	Comparison *BenchmarkComparison `json:"comparison,omitempty"`
	Limits     *BenchmarkLimits     `json:"limits,omitempty"`
	Violations []BenchmarkViolation `json:"violations,omitempty"`
	Path       string               `json:"-"`
}

func LoadBenchmarkReport(path string) (*BenchmarkReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var envelope struct {
		Status json.RawMessage `json:"status"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if len(envelope.Status) > 0 && envelope.Status[0] == '{' {
		return loadSnapshotBenchmarkReport(data, path)
	}
	var report BenchmarkReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if report.Version != benchmarkReportVersion {
		return nil, fmt.Errorf("unsupported benchmark report version %d in %s", report.Version, path)
	}
	if report.Mode != "benchmark" {
		return nil, fmt.Errorf("%s is a %q report, expected a benchmark report", path, report.Mode)
	}
	return &report, nil
}

func (r *Runner) saveBenchmarkReport(started, finished time.Time, profile *ProfileReport, runErr error) error {
	snapshot := r.progress.Snapshot()
	report := &BenchmarkReport{
		Version: benchmarkReportVersion, StartedAt: started.UTC(), FinishedAt: finished.UTC(),
		DurationMS: milliseconds(finished.Sub(started)), Status: snapshot.State,
		Files: []string{}, Phases: []BenchmarkEntry{}, Fixtures: []BenchmarkEntry{}, Profile: profile,
	}
	if r.options.Benchmark {
		report.Mode = "benchmark"
	} else {
		report.Mode = "profile"
	}
	if r.options.Limits.Enabled() {
		limits := r.options.Limits
		report.Limits = &limits
	}
	if runErr != nil {
		report.Status = ExecutionErrored
	}
	r.fillBenchmarkEntries(report, snapshot.Root)
	if r.baseline != nil {
		report.Comparison = compareBenchmarkReports(r.options.Baseline, r.baseline, report)
	}
	limitErr := r.applyBenchmarkLimits(report)
	path, err := writeBenchmarkReport(r.options.WorkDir, report)
	if err != nil {
		return err
	}
	report.Path = path
	r.benchmark = report
	logger.Infof("Fixture %s report: %s", report.Mode, path)
	return limitErr
}

func (r *Runner) fillBenchmarkEntries(report *BenchmarkReport, root *ExecutionNode) {
	files := make(map[string]bool)
	commandDurations := make(map[string]float64)
	results := make(map[string]*FixtureResult)
	for source, node := range r.progress.fixture {
		results[node.Key] = source.Results
		if source.Origin != nil {
			files[source.Origin.File] = true
		}
		if source.Results != nil && source.Results.CommandDuration > 0 {
			value := milliseconds(source.Results.CommandDuration)
			commandDurations[node.Key] = value
		}
	}
	for file := range files {
		report.Files = append(report.Files, file)
	}
	sort.Strings(report.Files)
	walkExecution(root, func(node *ExecutionNode) {
		if !isWorkNode(node) || node.State == ExecutionQueued {
			return
		}
		entry := BenchmarkEntry{Key: node.Key, Name: node.Name, Kind: string(node.Kind), Status: node.State,
			Origin: node.Origin, DurationMS: milliseconds(node.Duration)}
		if result := results[node.Key]; result != nil {
			entry.GoProfiles = result.GoProfiles
			entry.SQLProfile = result.SQLProfile
		}
		if report.Profile != nil {
			pid := 0
			if result := results[node.Key]; result != nil {
				if command, ok := result.Actual.(*clickyexec.ExecResult); ok {
					pid = command.PID
				}
			}
			entry.Profile = profileFixtureNode(report.Profile.Samples, node, pid)
			entry.ProfileSamples = &entry.Profile.SampleCount
			if result := results[node.Key]; result != nil {
				result.Profile = entry.Profile
			}
		}
		if value, ok := commandDurations[node.Key]; ok {
			entry.CommandMS = &value
		}
		switch node.Kind {
		case ExecutionKindSetup, ExecutionKindBuild, ExecutionKindDaemon, ExecutionKindDaemonStop, ExecutionKindCleanup:
			report.Phases = append(report.Phases, entry)
		default:
			report.Fixtures = append(report.Fixtures, entry)
		}
	})
}

func profileFixtureNode(samples []ProfileSample, node *ExecutionNode, pid int) *FixtureProfile {
	profile := &FixtureProfile{Scope: "run_tree", PID: pid}
	if pid > 0 {
		profile.Scope = "process_tree"
	}
	if node.StartedAt == nil || node.FinishedAt == nil {
		return profile
	}
	for _, sample := range samples {
		if sample.SampledAt.Before(*node.StartedAt) || sample.SampledAt.After(*node.FinishedAt) {
			continue
		}
		cpu, memory, rss := sample.CPUPercent, sample.MemoryPercent, sample.RSSBytes
		if pid > 0 {
			usage := process.NewSnapshot(sample.Processes).AggregateSubtree(pid)
			if len(usage.Processes) == 0 {
				continue
			}
			cpu, memory, rss = usage.CPUPercent, usage.MemoryPercent, usage.RSSBytes
		}
		profile.SampleCount++
		profile.PeakCPUPercent = max(profile.PeakCPUPercent, cpu)
		profile.PeakMemoryPercent = max(profile.PeakMemoryPercent, memory)
		profile.PeakRSSBytes = max(profile.PeakRSSBytes, rss)
	}
	profile.DiskIO = profileDiskIO(samples, node, pid, runtime.GOOS)
	return profile
}

func profileDiskIO(samples []ProfileSample, node *ExecutionNode, pid int, goos string) *ProfileDiskIO {
	if node.StartedAt == nil || node.FinishedAt == nil {
		return nil
	}
	result := &ProfileDiskIO{}
	previous := make(map[string]process.ProcessIO)
	var baselineRead, baselineWrite uint64
	for _, sample := range samples {
		if sample.SampledAt.Before(*node.StartedAt) || sample.SampledAt.After(*node.FinishedAt) {
			continue
		}
		members := sample.Processes
		if pid > 0 {
			members = process.NewSnapshot(members).Subtree(pid)
		}
		var read, write uint64
		measured := 0
		for _, member := range members {
			if member.IO == nil {
				result.MissingProcesses++
				continue
			}
			measured++
			read += member.IO.DiskReadBytes
			write += member.IO.DiskWriteBytes
			if goos != "linux" {
				identity := fmt.Sprintf("%d", member.PID)
				if member.StartedAt != nil {
					identity = fmt.Sprintf("%d/%d", member.PID, member.StartedAt.UnixNano())
				}
				if prior, found := previous[identity]; found {
					result.DiskReadBytes += member.IO.DiskReadBytes - min(member.IO.DiskReadBytes, prior.DiskReadBytes)
					result.DiskWriteBytes += member.IO.DiskWriteBytes - min(member.IO.DiskWriteBytes, prior.DiskWriteBytes)
				} else if pid > 0 || member.StartedAt != nil && !member.StartedAt.Before(*node.StartedAt) {
					result.DiskReadBytes += member.IO.DiskReadBytes
					result.DiskWriteBytes += member.IO.DiskWriteBytes
				}
				previous[identity] = *member.IO
			}
		}
		if measured == 0 {
			continue
		}
		result.SampleCount++
		if goos == "linux" {
			if result.SampleCount == 1 && pid == 0 {
				baselineRead, baselineWrite = read, write
			}
			if read >= baselineRead {
				result.DiskReadBytes = max(result.DiskReadBytes, read-baselineRead)
			}
			if write >= baselineWrite {
				result.DiskWriteBytes = max(result.DiskWriteBytes, write-baselineWrite)
			}
		}
	}
	if result.SampleCount == 0 {
		return nil
	}
	return result
}

func milliseconds(duration time.Duration) float64 {
	return float64(duration) / float64(time.Millisecond)
}

func writeBenchmarkReport(workDir string, report *BenchmarkReport) (string, error) {
	root, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("resolve fixture benchmark working directory: %w", err)
	}
	dir := filepath.Join(root, ".gavel", "benchmarks", "fixtures")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create fixture benchmark directory: %w", err)
	}
	path := filepath.Join(dir, "run-"+report.StartedAt.Format("2006-01-02T15-04-05.000000000Z")+".json")
	data, err := json.MarshalIndent(benchmarkSnapshot{
		Metadata:    benchmarkSnapshotMetadata{Version: "1", Started: report.StartedAt, Ended: report.FinishedAt, Kind: "fixtures"},
		Status:      benchmarkSnapshotStatus{Running: false},
		Tests:       BenchmarkReportToTests(report),
		Performance: BenchmarkPerformanceFromReport(report, path),
	}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode fixture benchmark report: %w", err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", fmt.Errorf("create fixture benchmark report %s: %w", path, err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("write fixture benchmark report %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close fixture benchmark report %s: %w", path, err)
	}
	return path, nil
}

func compareBenchmarkReports(path string, baseline, current *BenchmarkReport) *BenchmarkComparison {
	comparison := &BenchmarkComparison{Baseline: path, Deltas: []BenchmarkDelta{}}
	previous := make(map[string]BenchmarkEntry)
	for _, entry := range append(append([]BenchmarkEntry{}, baseline.Phases...), baseline.Fixtures...) {
		previous[entry.Key] = entry
	}
	for _, entry := range append(append([]BenchmarkEntry{}, current.Phases...), current.Fixtures...) {
		prior, ok := previous[entry.Key]
		if entry.Status == ExecutionCancelled && entry.DurationMS == 0 {
			comparison.Unmeasured = append(comparison.Unmeasured, entry.Key)
			delete(previous, entry.Key)
			continue
		}
		if !ok {
			comparison.Added = append(comparison.Added, entry.Key)
			continue
		}
		delete(previous, entry.Key)
		if prior.Status == ExecutionCancelled && prior.DurationMS == 0 {
			comparison.Unmeasured = append(comparison.Unmeasured, entry.Key)
			continue
		}
		delta := entry.DurationMS - prior.DurationMS
		change := BenchmarkDelta{Key: entry.Key, Name: entry.Name, Kind: entry.Kind,
			BaselineMS: prior.DurationMS, CurrentMS: entry.DurationMS, DeltaMS: delta}
		if prior.DurationMS > 0 {
			percent := delta / prior.DurationMS * 100
			change.DeltaPct = &percent
		}
		if prior.Profile != nil && entry.Profile != nil && prior.Profile.DiskIO != nil && entry.Profile.DiskIO != nil {
			read := int64(entry.Profile.DiskIO.DiskReadBytes) - int64(prior.Profile.DiskIO.DiskReadBytes)
			write := int64(entry.Profile.DiskIO.DiskWriteBytes) - int64(prior.Profile.DiskIO.DiskWriteBytes)
			change.DiskReadBytesDelta, change.DiskWriteBytesDelta = &read, &write
		}
		comparison.Deltas = append(comparison.Deltas, change)
	}
	for key := range previous {
		comparison.Missing = append(comparison.Missing, key)
	}
	sort.Strings(comparison.Added)
	sort.Strings(comparison.Missing)
	sort.Strings(comparison.Unmeasured)
	sort.Slice(comparison.Deltas, func(i, j int) bool { return comparison.Deltas[i].Key < comparison.Deltas[j].Key })
	if baseline.Profile != nil && current.Profile != nil && len(baseline.Profile.Samples) > 0 && len(current.Profile.Samples) > 0 {
		cpu := current.Profile.PeakCPUPercent - baseline.Profile.PeakCPUPercent
		rss := int64(current.Profile.PeakRSSBytes) - int64(baseline.Profile.PeakRSSBytes)
		comparison.PeakCPUPercentDelta = &cpu
		comparison.PeakRSSBytesDelta = &rss
	}
	return comparison
}

func (r BenchmarkReport) Pretty() api.Text {
	text := clicky.Text("Fixture "+r.Mode, "font-bold").Space().Append(fmt.Sprintf("%.2fms", r.DurationMS), "text-muted")
	for _, entry := range append(append([]BenchmarkEntry{}, r.Phases...), r.Fixtures...) {
		text = text.NewLine().Append("  " + entry.Name)
		if entry.Status == ExecutionCancelled && entry.DurationMS == 0 {
			text = text.Space().Append("(not run)", "text-muted")
		} else {
			text = text.Space().Append(fmt.Sprintf("%.2fms", entry.DurationMS), "text-muted")
		}
		if entry.CommandMS != nil {
			text = text.Space().Append(fmt.Sprintf("(command %.2fms)", *entry.CommandMS), "text-muted")
		}
		if entry.ProfileSamples != nil && *entry.ProfileSamples == 0 {
			text = text.Space().Append("(resource usage unobserved)", "text-muted")
		}
		if entry.Profile != nil && entry.Profile.DiskIO != nil {
			text = text.Space().Append(fmt.Sprintf("(observed disk read %s, write %s)", api.HumanizeBytes(int64(entry.Profile.DiskIO.DiskReadBytes)), api.HumanizeBytes(int64(entry.Profile.DiskIO.DiskWriteBytes))), "text-muted")
		}
		if entry.SQLProfile != nil {
			text = text.Space().Append(fmt.Sprintf("(SQL %d queries, %.2fms total, %.2fms max)", entry.SQLProfile.QueryCount, entry.SQLProfile.TotalDurationMS, entry.SQLProfile.MaxQueryMS), "text-muted")
		}
		for _, violation := range entry.Violations {
			text = text.NewLine().Append("    limit: "+violation, "text-red-600")
		}
		for _, artifact := range entry.GoProfiles {
			text = text.NewLine().Append(fmt.Sprintf("    Go profile %s: %s", artifact.Name, artifact.Status), "text-muted")
			if artifact.Status == "captured" {
				text = text.Space().Append(artifact.Path, "text-muted")
			}
		}
	}
	if r.Profile != nil {
		text = text.NewLine().Append(fmt.Sprintf("Peak CPU %.1f%%, memory %.1f%%, RSS ", r.Profile.PeakCPUPercent, r.Profile.PeakMemoryPercent), "text-muted").Add(api.HumanizeBytes(int64(r.Profile.PeakRSSBytes)))
	}
	if r.Comparison != nil {
		text = text.NewLine().Append("Baseline comparison", "font-bold")
		for _, delta := range r.Comparison.Deltas {
			text = text.NewLine().Append(fmt.Sprintf("  %s: %+.2fms", delta.Name, delta.DeltaMS))
			if delta.DeltaPct != nil {
				text = text.Space().Append(fmt.Sprintf("(%+.1f%%)", *delta.DeltaPct))
			}
			if delta.DiskReadBytesDelta != nil {
				text = text.Space().Append(fmt.Sprintf("(disk read %+d B, write %+d B)", *delta.DiskReadBytesDelta, *delta.DiskWriteBytesDelta), "text-muted")
			}
		}
		for _, key := range r.Comparison.Added {
			text = text.NewLine().Append("  added: " + key)
		}
		for _, key := range r.Comparison.Missing {
			text = text.NewLine().Append("  missing: " + key)
		}
		for _, key := range r.Comparison.Unmeasured {
			text = text.NewLine().Append("  unmeasured: " + key)
		}
		if r.Comparison.PeakCPUPercentDelta != nil {
			text = text.NewLine().Append(fmt.Sprintf("  peak CPU: %+.1f points, RSS: %+d bytes", *r.Comparison.PeakCPUPercentDelta, *r.Comparison.PeakRSSBytesDelta))
		}
	}
	return text.NewLine().Append("Report: "+r.Path, "text-muted")
}
