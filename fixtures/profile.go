package fixtures

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/flanksource/clicky/process"
	tracedprocess "github.com/flanksource/commons-db/tracing/process"

	"github.com/flanksource/gavel/traces"
)

// profileInterval is how often a run the fixture runner traces for --profile
// samples its tree.
const profileInterval = 500 * time.Millisecond

type ProfileSample struct {
	SampledAt     time.Time         `json:"sampled_at"`
	CPUPercent    float64           `json:"cpu_percent"`
	MemoryPercent float64           `json:"memory_percent"`
	RSSBytes      uint64            `json:"rss_bytes"`
	Processes     []process.Process `json:"processes"`
}

type ProfileReport struct {
	IntervalMS        int64           `json:"interval_ms"`
	Samples           []ProfileSample `json:"samples"`
	PeakCPUPercent    float64         `json:"peak_cpu_percent"`
	PeakMemoryPercent float64         `json:"peak_memory_percent"`
	PeakRSSBytes      uint64          `json:"peak_rss_bytes"`
	DiskIO            *ProfileDiskIO  `json:"disk_io,omitempty"`
}

// startProfile has --profile read the run's process_tree and process_samples:
// the run's own process capture when the caller handed one in, else one the
// runner starts on its trace host. Its readings from here on are the report.
func (r *Runner) startProfile() error {
	if !r.options.Profile {
		return nil
	}
	r.profileTrace = r.options.ProcessTrace
	if r.profileTrace == nil {
		if err := r.openTraceHost(); err != nil {
			return err
		}
		trace, err := r.traceHost.StartProcessTrace(traces.ProcessTraceOptions{RootPID: os.Getpid(), Interval: profileInterval})
		if err != nil {
			return fmt.Errorf("--profile: %w", err)
		}
		r.profileTrace, r.ownsProfileTrace = trace, true
	}
	var err error
	if r.profileMark, err = r.profileTrace.Mark(context.Background()); err != nil {
		return fmt.Errorf("--profile: %w", err)
	}
	return nil
}

// finishProfile reads the profile's samples into r.profile and stops a
// capture the runner started; a later call does nothing.
func (r *Runner) finishProfile(ctx context.Context) error {
	if r.profileTrace == nil {
		return nil
	}
	trace := r.profileTrace
	r.profileTrace = nil
	readings, err := trace.Readings(ctx, r.profileMark)
	if err == nil {
		r.profile = profileReport(readings, trace.Interval())
	}
	if r.ownsProfileTrace {
		err = errors.Join(err, trace.Stop(ctx))
		r.ownsProfileTrace = false
	}
	if err != nil {
		return fmt.Errorf("--profile: %w", err)
	}
	return nil
}

// profileReport is the run's samples by tick: each tick one ProfileSample of
// every process it sampled, with the parent and command its tree node holds.
func profileReport(readings traces.ProcessReadings, interval time.Duration) *ProfileReport {
	nodes := make(map[string]tracedprocess.Node, len(readings.Nodes))
	for _, node := range readings.Nodes {
		nodes[node.Key] = node
	}
	report := &ProfileReport{IntervalMS: interval.Milliseconds(), Samples: []ProfileSample{}}
	byTick := map[int64]*ProfileSample{}
	for _, sample := range readings.Samples {
		tick, ok := byTick[sample.Tick]
		if !ok {
			tick = &ProfileSample{SampledAt: sample.SampledAt}
			byTick[sample.Tick] = tick
		}
		member := sampledProcess(sample, nodes[sample.Key])
		tick.Processes = append(tick.Processes, member)
		tick.CPUPercent += member.CPUPercent
		tick.MemoryPercent += member.MemoryPercent
		tick.RSSBytes += member.RSSBytes
	}
	ticks := make([]int64, 0, len(byTick))
	for tick := range byTick {
		ticks = append(ticks, tick)
	}
	sort.Slice(ticks, func(i, j int) bool { return ticks[i] < ticks[j] })
	for _, tick := range ticks {
		sample := *byTick[tick]
		report.Samples = append(report.Samples, sample)
		report.PeakCPUPercent = max(report.PeakCPUPercent, sample.CPUPercent)
		report.PeakMemoryPercent = max(report.PeakMemoryPercent, sample.MemoryPercent)
		report.PeakRSSBytes = max(report.PeakRSSBytes, sample.RSSBytes)
	}
	if len(report.Samples) > 0 {
		started, finished := report.Samples[0].SampledAt, report.Samples[len(report.Samples)-1].SampledAt
		report.DiskIO = profileDiskIO(report.Samples, &ExecutionNode{StartedAt: &started, FinishedAt: &finished}, 0, runtime.GOOS)
	}
	return report
}

// sampledProcess is one process at one tick, as a sample and its node hold
// it. A CPU% not yet known counts as idle.
func sampledProcess(sample tracedprocess.Sample, node tracedprocess.Node) process.Process {
	member := process.Process{
		PID: sample.PID, PPID: node.PPID, Command: node.Command, Status: node.Status,
		MemoryPercent: sample.MemoryPercent, RSSBytes: sample.RSSBytes,
	}
	if sample.CPUPercent != nil {
		member.CPUPercent = *sample.CPUPercent
	}
	if !node.StartedAt.IsZero() {
		started := node.StartedAt
		member.StartedAt = &started
	}
	if sample.DiskReadBytes != nil && sample.DiskWriteBytes != nil {
		member.IO = &process.ProcessIO{DiskReadBytes: *sample.DiskReadBytes, DiskWriteBytes: *sample.DiskWriteBytes}
	}
	return member
}
