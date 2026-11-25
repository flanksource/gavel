package slots

import (
	"slices"
	"sync"
	"time"

	"github.com/flanksource/gavel/sdk/status"
)

// tracked are this process's held slots and the pools it is queued on, by
// the order they were taken, which it publishes in its status.
var tracked = struct {
	sync.Mutex
	next   int
	leases map[int]status.Lease
}{leases: map[int]status.Lease{}}

// Leases are the slots this process holds and the pools it waits for,
// oldest first.
func Leases() []status.Lease {
	tracked.Lock()
	defer tracked.Unlock()
	ids := make([]int, 0, len(tracked.leases))
	for id := range tracked.leases {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	leases := make([]status.Lease, 0, len(ids))
	for _, id := range ids {
		leases = append(leases, tracked.leases[id])
	}
	return leases
}

func track(lease status.Lease) int {
	tracked.Lock()
	defer tracked.Unlock()
	tracked.next++
	tracked.leases[tracked.next] = lease
	return tracked.next
}

func untrack(ids ...int) {
	tracked.Lock()
	defer tracked.Unlock()
	for _, id := range ids {
		delete(tracked.leases, id)
	}
}

func (p pool) lease(state string, index int, command string) status.Lease {
	return status.Lease{Pool: p.name, Project: p.project, Limit: p.limit, Index: index, State: state, Since: time.Now().UTC(), Command: command}
}

// queue is an Acquire's place in the pool it currently waits for.
type queue struct {
	command string
	pool    string
	id      int
}

// wait records waiting for p, keeping the original queue time while it is
// the same pool.
func (q *queue) wait(p pool) {
	if q.id != 0 && q.pool == p.name {
		return
	}
	q.leave()
	q.pool, q.id = p.name, track(p.lease(status.LeaseWaiting, -1, q.command))
}

func (q *queue) leave() {
	if q.id != 0 {
		untrack(q.id)
		q.id = 0
	}
}
