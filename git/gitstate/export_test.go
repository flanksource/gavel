package gitstate

import (
	"context"
	"path/filepath"
	"time"
)

// Rescan runs the scans the background would, synchronously: the refs of the
// repository containing dir and then the status of each of its worktrees.
func (t *Tracker) Rescan(ctx context.Context, dir string) error {
	if _, err := t.Track(ctx, dir); err != nil {
		return err
	}
	t.mu.Lock()
	repo := t.byDir[filepath.Clean(dir)]
	t.mu.Unlock()
	return t.scanRepo(ctx, repo)
}

// HoldSlot takes one of the tracker's concurrency slots until release is
// called, keeping background scans queued.
func (t *Tracker) HoldSlot() (release func()) {
	t.sem <- struct{}{}
	return func() { <-t.sem }
}

// IdleIntervals records one scan per entry of changed on a fresh schedule and
// returns how long after each its idle interval ends.
func IdleIntervals(changed []bool, idle, maxIdle time.Duration) []time.Duration {
	var schedule pollSchedule
	now := time.Now()
	intervals := make([]time.Duration, 0, len(changed))
	for _, c := range changed {
		schedule.record(now, c, idle, maxIdle)
		intervals = append(intervals, schedule.next.Sub(now))
		now = schedule.next
	}
	return intervals
}

var (
	PorcelainPaths = porcelainPaths
	Backoff        = backoff
	Jitter         = jitter
)

func PollGitArgs() []string {
	return currentPollConfig().gitArgs()
}
