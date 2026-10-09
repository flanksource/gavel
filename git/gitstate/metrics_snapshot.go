package gitstate

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// LatencyStat summarises one labelled series of a gavel_git_* histogram.
type LatencyStat struct {
	Labels  map[string]string `json:"labels"`
	Count   uint64            `json:"count"`
	TotalMs float64           `json:"totalMs"`
	AvgMs   float64           `json:"avgMs"`
	P50Ms   float64           `json:"p50Ms"`
	P95Ms   float64           `json:"p95Ms"`
}

// CountStat is one labelled series of a gavel_git_* counter.
type CountStat struct {
	Labels map[string]string `json:"labels"`
	Count  uint64            `json:"count"`
}

// MetricsSnapshot is the process's git tracking metrics as the activity page
// shows them: the cache settings scans run with now, the tracker's size and
// load, what requested its scans, and the latency of each kind of scan, git
// command and read.
type MetricsSnapshot struct {
	FSMonitor      bool `json:"fsmonitor"`
	UntrackedCache bool `json:"untrackedCache"`
	TrackedRepos   int  `json:"trackedRepos"`
	HotWorktrees   int  `json:"hotWorktrees"`
	IdleWorktrees  int  `json:"idleWorktrees"`
	// BackoffWorktrees are idle worktrees polled less often after scans that
	// found nothing changed.
	BackoffWorktrees int           `json:"backoffWorktrees"`
	ScansInFlight    int           `json:"scansInFlight"`
	RangesComputed   int           `json:"rangesComputed"`
	ScanTriggers     []CountStat   `json:"scanTriggers"`
	RefsScans        []LatencyStat `json:"refsScans"`
	StatusScans      []LatencyStat `json:"statusScans"`
	Commands         []LatencyStat `json:"commands"`
	QueueWait        []LatencyStat `json:"queueWait"`
	Reads            []LatencyStat `json:"reads"`
}

// Metrics reads the gavel_git_* metrics registered with gatherer.
func Metrics(gatherer prometheus.Gatherer) (MetricsSnapshot, error) {
	families, err := gatherer.Gather()
	if err != nil {
		return MetricsSnapshot{}, fmt.Errorf("gather git metrics: %w", err)
	}
	config := currentPollConfig()
	snapshot := MetricsSnapshot{FSMonitor: config.FSMonitor, UntrackedCache: config.UntrackedCache}
	for _, family := range families {
		metrics := family.GetMetric()
		switch family.GetName() {
		case "gavel_git_tracked_repos":
			snapshot.TrackedRepos = int(gaugeSum(metrics, nil))
		case "gavel_git_tracked_worktrees":
			snapshot.HotWorktrees = int(gaugeSum(metrics, map[string]string{"cadence": cadenceHot}))
			snapshot.IdleWorktrees = int(gaugeSum(metrics, map[string]string{"cadence": cadenceIdle}))
			snapshot.BackoffWorktrees = int(gaugeSum(metrics, map[string]string{"cadence": cadenceBackoff}))
		case "gavel_git_scans_in_flight":
			snapshot.ScansInFlight = int(gaugeSum(metrics, nil))
		case "gavel_git_ranges_computed_total":
			for _, metric := range metrics {
				snapshot.RangesComputed += int(metric.GetCounter().GetValue())
			}
		case "gavel_git_scans_enqueued_total":
			snapshot.ScanTriggers = countStats(metrics)
		case "gavel_git_refs_scan_duration_seconds":
			snapshot.RefsScans = latencyStats(metrics)
		case "gavel_git_status_scan_duration_seconds":
			snapshot.StatusScans = latencyStats(metrics)
		case "gavel_git_command_duration_seconds":
			snapshot.Commands = latencyStats(metrics)
		case "gavel_git_scan_queue_wait_seconds":
			snapshot.QueueWait = latencyStats(metrics)
		case "gavel_git_read_duration_seconds":
			snapshot.Reads = latencyStats(metrics)
		}
	}
	return snapshot, nil
}

// gaugeSum adds the gauges whose labels include every pair of match.
func gaugeSum(metrics []*dto.Metric, match map[string]string) float64 {
	var sum float64
	for _, metric := range metrics {
		labels := labelMap(metric)
		matched := true
		for name, value := range match {
			matched = matched && labels[name] == value
		}
		if matched {
			sum += metric.GetGauge().GetValue()
		}
	}
	return sum
}

func labelMap(metric *dto.Metric) map[string]string {
	labels := make(map[string]string, len(metric.GetLabel()))
	for _, pair := range metric.GetLabel() {
		labels[pair.GetName()] = pair.GetValue()
	}
	return labels
}

// latencyStats summarises each series that observed a value, slowest total
// first.
func latencyStats(metrics []*dto.Metric) []LatencyStat {
	stats := make([]LatencyStat, 0, len(metrics))
	for _, metric := range metrics {
		histogram := metric.GetHistogram()
		count := histogram.GetSampleCount()
		if count == 0 {
			continue
		}
		total := histogram.GetSampleSum() * 1000
		stats = append(stats, LatencyStat{
			Labels:  labelMap(metric),
			Count:   count,
			TotalMs: total,
			AvgMs:   total / float64(count),
			P50Ms:   histogramQuantile(histogram, 0.5) * 1000,
			P95Ms:   histogramQuantile(histogram, 0.95) * 1000,
		})
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].TotalMs != stats[j].TotalMs {
			return stats[i].TotalMs > stats[j].TotalMs
		}
		return labelKey(stats[i].Labels) < labelKey(stats[j].Labels)
	})
	return stats
}

// countStats lists each counter series, highest count first.
func countStats(metrics []*dto.Metric) []CountStat {
	stats := make([]CountStat, 0, len(metrics))
	for _, metric := range metrics {
		stats = append(stats, CountStat{Labels: labelMap(metric), Count: uint64(metric.GetCounter().GetValue())})
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count != stats[j].Count {
			return stats[i].Count > stats[j].Count
		}
		return labelKey(stats[i].Labels) < labelKey(stats[j].Labels)
	})
	return stats
}

func labelKey(labels map[string]string) string {
	pairs := make([]string, 0, len(labels))
	for name, value := range labels {
		pairs = append(pairs, name+"="+value)
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}

// histogramQuantile estimates the q quantile of a histogram, in seconds, by
// interpolating linearly inside the bucket the rank falls in, as Prometheus's
// histogram_quantile does. A rank beyond the last finite bound reports that
// bound.
func histogramQuantile(histogram *dto.Histogram, q float64) float64 {
	count := histogram.GetSampleCount()
	if count == 0 {
		return 0
	}
	rank := q * float64(count)
	lower, below := 0.0, uint64(0)
	for _, bucket := range histogram.GetBucket() {
		upper, cumulative := bucket.GetUpperBound(), bucket.GetCumulativeCount()
		if math.IsInf(upper, 1) {
			break
		}
		if float64(cumulative) >= rank {
			inBucket := cumulative - below
			if inBucket == 0 {
				return upper
			}
			return lower + (upper-lower)*(rank-float64(below))/float64(inBucket)
		}
		lower, below = upper, cumulative
	}
	return lower
}
