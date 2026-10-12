package gitstate

import (
	"math/rand/v2"
	"time"
)

// pollSchedule is when one scan key falls due: its hot interval counts from
// last, its idle interval ends at next, which backs off while consecutive
// scans find nothing changed.
type pollSchedule struct {
	last, next time.Time
	streak     int
}

// record notes a scan that finished at now. A change resets the idle interval;
// an unchanged, skipped or failed scan doubles it, up to maxIdle.
func (s *pollSchedule) record(now time.Time, changed bool, idle, maxIdle time.Duration) {
	if changed {
		s.streak = 0
	} else {
		s.streak++
	}
	s.last = now
	s.next = now.Add(jitter(backoff(idle, maxIdle, s.streak)))
}

// backoff is the idle interval after streak consecutive unchanged scans.
func backoff(idle, maxIdle time.Duration, streak int) time.Duration {
	interval := idle
	for range streak {
		if interval >= maxIdle {
			break
		}
		interval *= 2
	}
	return min(interval, maxIdle)
}

// jitter spreads interval by up to a tenth either way, so keys first scanned
// together do not stay due in the same second.
func jitter(interval time.Duration) time.Duration {
	spread := int64(interval / 5)
	if spread == 0 {
		return interval
	}
	return interval - interval/10 + time.Duration(rand.Int64N(spread+1))
}
