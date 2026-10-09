package fixtures

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/clicky/task"
)

type BenchmarkViolation struct {
	Key     string   `json:"key"`
	Name    string   `json:"name"`
	Reasons []string `json:"reasons"`
}

type BenchmarkLimits struct {
	MaxDeviationPct   *float64 `json:"max_deviation_pct,omitempty"`
	MaxDurationMS     float64  `json:"max_duration_ms,omitempty"`
	MaxRSSBytes       uint64   `json:"max_rss_bytes,omitempty"`
	MaxDiskReadBytes  uint64   `json:"max_disk_read_bytes,omitempty"`
	MaxDiskWriteBytes uint64   `json:"max_disk_write_bytes,omitempty"`
	MaxSQLDurationMS  float64  `json:"max_sql_duration_ms,omitempty"`
	MaxSQLQueryMS     float64  `json:"max_sql_query_ms,omitempty"`
	MaxSQLQueries     *int     `json:"max_sql_queries,omitempty"`
	MaxSlowSQL        *int     `json:"max_slow_sql,omitempty"`
}

func (l BenchmarkLimits) Enabled() bool {
	return l.MaxDeviationPct != nil || l.MaxDurationMS > 0 || l.MaxRSSBytes > 0 ||
		l.MaxDiskReadBytes > 0 || l.MaxDiskWriteBytes > 0 || l.MaxSQLDurationMS > 0 || l.MaxSQLQueryMS > 0 ||
		l.MaxSQLQueries != nil || l.MaxSlowSQL != nil
}

func (l BenchmarkLimits) NeedsProcessProfile() bool {
	return l.MaxRSSBytes > 0 || l.MaxDiskReadBytes > 0 || l.MaxDiskWriteBytes > 0
}

type benchmarkMetric struct {
	name    string
	value   *float64
	max     float64
	enabled bool
	unit    string
}

func benchmarkMetrics(entry BenchmarkEntry, limits BenchmarkLimits) []benchmarkMetric {
	metrics := []benchmarkMetric{{name: "time", value: &entry.DurationMS, max: limits.MaxDurationMS, enabled: limits.MaxDurationMS > 0, unit: "ms"}}
	if entry.Profile != nil && entry.Profile.Scope == "process_tree" && entry.Profile.SampleCount > 0 {
		value := float64(entry.Profile.PeakRSSBytes)
		metrics = append(metrics, benchmarkMetric{name: "memory", value: &value, max: float64(limits.MaxRSSBytes), enabled: limits.MaxRSSBytes > 0, unit: " B"})
	} else {
		metrics = append(metrics, benchmarkMetric{name: "memory", max: float64(limits.MaxRSSBytes), enabled: limits.MaxRSSBytes > 0, unit: " B"})
	}
	for _, item := range []struct {
		name string
		max  uint64
		read bool
	}{{"disk read", limits.MaxDiskReadBytes, true}, {"disk write", limits.MaxDiskWriteBytes, false}} {
		metric := benchmarkMetric{name: item.name, max: float64(item.max), enabled: item.max > 0, unit: " B"}
		if entry.Profile != nil && entry.Profile.Scope == "process_tree" && entry.Profile.DiskIO != nil {
			value := float64(entry.Profile.DiskIO.DiskWriteBytes)
			if item.read {
				value = float64(entry.Profile.DiskIO.DiskReadBytes)
			}
			metric.value = &value
		}
		metrics = append(metrics, metric)
	}
	if entry.SQLProfile != nil {
		metrics = append(metrics, benchmarkMetric{name: "SQL total", value: &entry.SQLProfile.TotalDurationMS, max: limits.MaxSQLDurationMS, enabled: limits.MaxSQLDurationMS > 0, unit: "ms"},
			benchmarkMetric{name: "SQL query", value: &entry.SQLProfile.MaxQueryMS, max: limits.MaxSQLQueryMS, enabled: limits.MaxSQLQueryMS > 0, unit: "ms"})
	} else {
		metrics = append(metrics, benchmarkMetric{name: "SQL total", max: limits.MaxSQLDurationMS, enabled: limits.MaxSQLDurationMS > 0, unit: "ms"}, benchmarkMetric{name: "SQL query", max: limits.MaxSQLQueryMS, enabled: limits.MaxSQLQueryMS > 0, unit: "ms"})
	}
	queries := benchmarkMetric{name: "SQL queries", enabled: limits.MaxSQLQueries != nil}
	slow := benchmarkMetric{name: "slow SQL queries", enabled: limits.MaxSlowSQL != nil}
	if limits.MaxSQLQueries != nil {
		queries.max = float64(*limits.MaxSQLQueries)
	}
	if limits.MaxSlowSQL != nil {
		slow.max = float64(*limits.MaxSlowSQL)
	}
	if entry.SQLProfile != nil {
		count, slowCount := float64(entry.SQLProfile.QueryCount), float64(entry.SQLProfile.SlowQueryCount)
		queries.value, slow.value = &count, &slowCount
	}
	return append(metrics, queries, slow)
}

func checkBenchmarkLimits(entry BenchmarkEntry, baseline *BenchmarkEntry, limits BenchmarkLimits) []string {
	if entry.Status != ExecutionPassed {
		return nil
	}
	current := benchmarkMetrics(entry, limits)
	var prior []benchmarkMetric
	if baseline != nil {
		prior = benchmarkMetrics(*baseline, limits)
	}
	var violations []string
	for i, metric := range current {
		if metric.enabled {
			if metric.value == nil {
				violations = append(violations, metric.name+" unmeasured")
			} else if *metric.value > metric.max {
				if metric.unit == "" {
					violations = append(violations, fmt.Sprintf("%s %.0f exceeds %.0f", metric.name, *metric.value, metric.max))
				} else {
					violations = append(violations, fmt.Sprintf("%s %.2f%s exceeds %.2f%s", metric.name, *metric.value, metric.unit, metric.max, metric.unit))
				}
			}
		}
		if limits.MaxDeviationPct == nil {
			continue
		}
		if baseline == nil {
			if i == 0 {
				violations = append(violations, "baseline missing")
			}
			continue
		}
		if prior[i].value == nil {
			continue
		}
		if metric.value == nil {
			violations = append(violations, metric.name+" baseline comparison unmeasured")
			continue
		}
		if *prior[i].value == 0 {
			if *metric.value > 0 {
				violations = append(violations, metric.name+" baseline is zero")
			}
			continue
		}
		pct := (*metric.value - *prior[i].value) / *prior[i].value * 100
		if pct > *limits.MaxDeviationPct {
			violations = append(violations, fmt.Sprintf("%s baseline deviation %.1f%% exceeds %.1f%%", metric.name, pct, *limits.MaxDeviationPct))
		}
	}
	return violations
}

func (r *Runner) applyBenchmarkLimits(report *BenchmarkReport) error {
	if !r.options.Limits.Enabled() {
		return nil
	}
	baseline := make(map[string]BenchmarkEntry)
	if r.baseline != nil {
		for _, entry := range r.baseline.Fixtures {
			baseline[entry.Key] = entry
		}
	}
	failures := make(map[string]string)
	for i := range report.Fixtures {
		entry := &report.Fixtures[i]
		prior, found := baseline[entry.Key]
		var previous *BenchmarkEntry
		if found {
			previous = &prior
		}
		entry.Violations = checkBenchmarkLimits(*entry, previous, r.options.Limits)
		if len(entry.Violations) == 0 {
			continue
		}
		entry.Status = ExecutionFailed
		report.Violations = append(report.Violations, BenchmarkViolation{Key: entry.Key, Name: entry.Name, Reasons: entry.Violations})
		failures[entry.Key] = strings.Join(entry.Violations, "; ")
	}
	if len(failures) == 0 {
		return nil
	}
	if report.Status != ExecutionErrored {
		report.Status = ExecutionFailed
	}
	for source, node := range r.progress.fixture {
		if reason, ok := failures[node.Key]; ok && source.Results != nil {
			source.Results.Status = task.StatusFAIL
			source.Results.Error = reason
		}
	}
	r.tree.UpdateStatsRecursive()
	if err := r.progress.mutateAndPublish(context.Background(), func(_ time.Time) error {
		for key, reason := range failures {
			node := r.progress.byKey[key]
			node.State = ExecutionFailed
			node.Error = reason
		}
		return nil
	}); err != nil {
		return fmt.Errorf("publish fixture benchmark limits: %w", err)
	}
	return fmt.Errorf("%d fixture benchmark limit(s) exceeded", len(failures))
}
