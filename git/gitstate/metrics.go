package gitstate

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Scan results, the result label of the scan histograms.
const (
	resultChanged   = "changed"
	resultUnchanged = "unchanged"
	// resultSkipped is a background scan another process held the lock for.
	resultSkipped = "skipped"
	resultError   = "error"
)

// scanBuckets span a fingerprint-only status scan of a small worktree (a few
// ms) to a cold ref scan comparing many branches (tens of seconds).
var scanBuckets = prometheus.ExponentialBuckets(0.002, 2, 15)

var (
	refsScanDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gavel_git_refs_scan_duration_seconds",
		Help:    "Duration of a repository ref scan, by result (changed, unchanged, skipped, error).",
		Buckets: scanBuckets,
	}, []string{"result"})
	statusScanDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gavel_git_status_scan_duration_seconds",
		Help:    "Duration of a worktree status scan, by result and the git.fsmonitor / git.untrackedCache properties it ran with.",
		Buckets: scanBuckets,
	}, []string{"result", "fsmonitor", "untracked_cache"})
	gitCommandDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gavel_git_command_duration_seconds",
		Help:    "Duration of a git command a scan ran, by subcommand and whether core.fsmonitor was on.",
		Buckets: scanBuckets,
	}, []string{"command", "fsmonitor"})
	scanQueueWait = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gavel_git_scan_queue_wait_seconds",
		Help:    "Time a background scan waited for one of the tracker's concurrency slots, by kind (refs, status).",
		Buckets: scanBuckets,
	}, []string{"kind"})
	scansInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gavel_git_scans_in_flight",
		Help: "Background scans holding a concurrency slot.",
	})
	rangesComputed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gavel_git_ranges_computed_total",
		Help: "Base/head comparisons computed because no cached git_range_stats row had them.",
	})
	trackedRepos = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gavel_git_tracked_repos",
		Help: "Repositories the tracker keeps current.",
	})
	trackedWorktrees = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gavel_git_tracked_worktrees",
		Help: "Live worktrees of the tracked repositories, by cadence (hot, idle).",
	}, []string{"cadence"})
	readDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gavel_git_read_duration_seconds",
		Help:    "Duration of a read of the stored git state, by operation (states, files).",
		Buckets: prometheus.ExponentialBuckets(0.0005, 2, 15),
	}, []string{"op"})
)

func scanResult(ran, changed bool, err error) string {
	switch {
	case err != nil:
		return resultError
	case !ran:
		return resultSkipped
	case changed:
		return resultChanged
	default:
		return resultUnchanged
	}
}

func jobKind(key jobKey) string {
	if key.worktree == "" {
		return "refs"
	}
	return "status"
}

func observeSince(observer prometheus.Observer, start time.Time) {
	observer.Observe(time.Since(start).Seconds())
}

// observeGitCommand records how long git took to run args, labelled by the
// subcommand: the first argument after the global options.
func observeGitCommand(args []string, start time.Time) {
	command := ""
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "-c" || args[i] == "-C":
			i++
		case strings.HasPrefix(args[i], "-"):
		default:
			command = args[i]
		}
		if command != "" {
			break
		}
	}
	fsmonitor := strconv.FormatBool(slices.Contains(args, "core.fsmonitor=true"))
	observeSince(gitCommandDuration.WithLabelValues(command, fsmonitor), start)
}
