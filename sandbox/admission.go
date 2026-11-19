// Admits sandbox runs against their organization's budgets: a run is refused when a cumulative
// budget is used up or it can never fit a concurrent one, and otherwise waits its turn, first in
// first out, while the organization's runs hold all its concurrent budget allows.
package sandbox

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/gavel/budget"
	"github.com/google/uuid"
)

var (
	// ErrBudgetUsedUp refuses a run its organization's budgets don't leave room for.
	ErrBudgetUsedUp = errors.New("budget used up")
	// ErrQueueTimeout fails a run that waited longer than Admission.Wait for room.
	ErrQueueTimeout = errors.New("no room for the run")
)

// Admission admits an organization's runs within its budgets. What its runs hold at once is
// counted here, from a run's admission to its release, so a serve has one Admission for every
// run it starts: gavel serve runs as one replica.
type Admission struct {
	Budgets budget.Store
	// RunLifetime is the longest a run can last.
	RunLifetime time.Duration
	// Wait is the longest a run waits for room; then it fails.
	Wait time.Duration
	// Recheck is how often a waiting run reads its organization's budgets again, so it starts
	// when an admin raises one.
	Recheck time.Duration

	mu      sync.Mutex
	held    map[uuid.UUID]substrate.Size
	lines   map[uuid.UUID][]*Ticket
	changed map[uuid.UUID]chan struct{}
}

// Demand is what a run asks of its organization's budgets: its size, and whether it makes model
// calls, which only agent runs do and only token and model cost budgets count.
type Demand struct {
	Size       substrate.Size
	ModelCalls bool
}

// Ticket is a run's place in its organization's budgets: it holds room once admitted, which
// Release gives back. A nil Ticket is a run admitted without budgets, holding nothing.
type Ticket struct {
	admission *Admission
	org       uuid.UUID
	demand    Demand
	admitted  bool
	reason    string
}

// limits are an organization's concurrent budgets; nil is no cap.
type limits struct {
	cores, gib *float64
}

// Enter refuses a run demanding what org's budgets leave no room for, and otherwise gives it a
// ticket: admitted at once when the organization holds little enough and no run is waiting, or
// in line behind the runs waiting.
func (a *Admission) Enter(ctx context.Context, org uuid.UUID, demand Demand) (*Ticket, error) {
	caps, err := a.limits(ctx, org, demand)
	if err != nil {
		return nil, err
	}
	size := demand.Size
	a.mu.Lock()
	defer a.mu.Unlock()
	a.init()
	ticket := &Ticket{admission: a, org: org, demand: demand}
	line := a.lines[org]
	if len(line) == 0 && caps.fit(a.held[org], size) {
		a.hold(ticket)
		return ticket, nil
	}
	ticket.reason = "queued: " + a.waitingFor(org, size, caps, len(line))
	a.lines[org] = append(line, ticket)
	return ticket, nil
}

// Queued reports whether the ticket's run waits for room, and what for.
func (t *Ticket) Queued() (bool, string) {
	if t == nil {
		return false, ""
	}
	t.admission.mu.Lock()
	defer t.admission.mu.Unlock()
	return !t.admitted, t.reason
}

// Wait waits until the run has room: when it is first in line and fits beside what the
// organization holds. It reports what it waits for, and that it started. It fails after the
// Admission's Wait, when a cumulative budget is used up meanwhile, or when ctx ends, leaving the
// line.
func (t *Ticket) Wait(ctx context.Context, report func(string)) error {
	queued, reason := t.Queued()
	if !queued {
		return nil
	}
	a := t.admission
	report(reason)
	deadline := time.NewTimer(a.Wait)
	defer deadline.Stop()
	recheck := time.NewTicker(a.Recheck)
	defer recheck.Stop()
	for {
		caps, err := a.limits(ctx, t.org, t.demand)
		if err != nil {
			a.leave(t)
			return err
		}
		a.mu.Lock()
		line := a.lines[t.org]
		if len(line) > 0 && line[0] == t && caps.fit(a.held[t.org], t.demand.Size) {
			a.lines[t.org] = line[1:]
			a.hold(t)
			a.mu.Unlock()
			report("started after waiting for room")
			return nil
		}
		reason := a.waitingFor(t.org, t.demand.Size, caps, slices.Index(line, t))
		changed := a.changed[t.org]
		a.mu.Unlock()
		select {
		case <-changed:
		case <-recheck.C:
		case <-deadline.C:
			a.leave(t)
			return fmt.Errorf("%w: waited %s for room and gave up: %s", ErrQueueTimeout, a.Wait, reason)
		case <-ctx.Done():
			a.leave(t)
			return ctx.Err()
		}
	}
}

// Release gives back the room the ticket's run held, so the runs waiting can start.
func (t *Ticket) Release() {
	if t == nil {
		return
	}
	a := t.admission
	a.mu.Lock()
	defer a.mu.Unlock()
	if !t.admitted {
		return
	}
	t.admitted = false
	held, size := a.held[t.org], t.demand.Size
	a.held[t.org] = substrate.Size{CPUMillicores: held.CPUMillicores - size.CPUMillicores, MemoryBytes: held.MemoryBytes - size.MemoryBytes}
	a.changeLocked(t.org)
}

func (a *Admission) init() {
	if a.held == nil {
		a.held, a.lines, a.changed = map[uuid.UUID]substrate.Size{}, map[uuid.UUID][]*Ticket{}, map[uuid.UUID]chan struct{}{}
	}
}

// hold admits ticket, adding its size to what its organization holds. a.mu is held.
func (a *Admission) hold(ticket *Ticket) {
	ticket.admitted = true
	held, size := a.held[ticket.org], ticket.demand.Size
	a.held[ticket.org] = substrate.Size{CPUMillicores: held.CPUMillicores + size.CPUMillicores, MemoryBytes: held.MemoryBytes + size.MemoryBytes}
	a.changeLocked(ticket.org)
}

// leave takes ticket out of its organization's line.
func (a *Admission) leave(ticket *Ticket) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lines[ticket.org] = slices.DeleteFunc(a.lines[ticket.org], func(t *Ticket) bool { return t == ticket })
	a.changeLocked(ticket.org)
}

// changeLocked wakes the runs waiting in org's line. a.mu is held.
func (a *Admission) changeLocked(org uuid.UUID) {
	if changed, ok := a.changed[org]; ok {
		close(changed)
	}
	a.changed[org] = make(chan struct{})
}

// limits are org's concurrent budgets, refusing a run demanding what its cumulative budgets have
// no room for, or bigger than a concurrent budget on its own. Token and model cost budgets hold
// back only runs that make model calls.
func (a *Admission) limits(ctx context.Context, org uuid.UUID, demand Demand) (limits, error) {
	size := demand.Size
	statuses, err := a.Budgets.Status(ctx, org, a.RunLifetime)
	if err != nil {
		return limits{}, fmt.Errorf("check the budgets of organization %s: %w", org, err)
	}
	var caps limits
	var refusals []string
	for _, status := range statuses {
		modelBudget := status.Resource == budget.Tokens || status.Resource == budget.USD
		switch {
		case modelBudget && !demand.ModelCalls:
		case status.Kind == budget.Cumulative && status.Reached():
			refusals = append(refusals, fmt.Sprintf("this organization has used %s of its %s %s budget since %s",
				amount(status.Budget, status.Used), amount(status.Budget, status.Amount), resourceName(status.Resource), status.Since.UTC().Format(time.RFC3339)))
		case status.Kind == budget.Concurrent:
			want := float64(size.CPUMillicores) / 1000
			if status.Resource == budget.Memory {
				want = float64(size.MemoryBytes) / (1 << 30)
				caps.gib = &status.Amount
			} else {
				caps.cores = &status.Amount
			}
			if want > status.Amount {
				refusals = append(refusals, fmt.Sprintf("a run of %s can never start under this organization's concurrent %s budget of %s",
					amount(status.Budget, want), status.Resource, amount(status.Budget, status.Amount)))
			}
		}
	}
	if len(refusals) > 0 {
		return limits{}, fmt.Errorf("%w: %s; an organization admin can raise it", ErrBudgetUsedUp, strings.Join(refusals, "; "))
	}
	return caps, nil
}

// fit reports whether a run of size fits beside held.
func (l limits) fit(held, size substrate.Size) bool {
	cores := float64(held.CPUMillicores+size.CPUMillicores) / 1000
	gib := float64(held.MemoryBytes+size.MemoryBytes) / (1 << 30)
	return (l.cores == nil || cores <= *l.cores) && (l.gib == nil || gib <= *l.gib)
}

// waitingFor says what a run of size waits for with ahead runs in line before it. a.mu is held.
func (a *Admission) waitingFor(org uuid.UUID, size substrate.Size, caps limits, ahead int) string {
	held := a.held[org]
	var full []string
	if caps.cores != nil && float64(held.CPUMillicores+size.CPUMillicores)/1000 > *caps.cores {
		full = append(full, fmt.Sprintf("%s of its %s concurrent %s", figure(float64(held.CPUMillicores)/1000), figure(*caps.cores), plural(*caps.cores, "core")))
	}
	if caps.gib != nil && float64(held.MemoryBytes+size.MemoryBytes)/(1<<30) > *caps.gib {
		full = append(full, fmt.Sprintf("%s of its %s concurrent GiB", figure(float64(held.MemoryBytes)/(1<<30)), figure(*caps.gib)))
	}
	line := "no run is ahead of this one"
	if ahead == 1 {
		line = "1 run is ahead of this one"
	} else if ahead > 1 {
		line = fmt.Sprintf("%d runs are ahead of this one", ahead)
	}
	if len(full) == 0 {
		return line
	}
	return "this organization's runs hold " + strings.Join(full, " and ") + ", and " + line
}

// amount is v in b's unit.
func amount(b budget.Budget, v float64) string {
	switch {
	case b.Resource == budget.USD:
		return fmt.Sprintf("$%.2f", v)
	case b.Resource == budget.Tokens:
		return strconv.FormatFloat(v, 'f', 0, 64) + " tokens"
	case b.Resource == budget.CPU && b.Kind == budget.Concurrent:
		return figure(v) + " " + plural(v, "core")
	case b.Resource == budget.CPU:
		return figure(v) + " core-hours"
	case b.Kind == budget.Concurrent:
		return figure(v) + " GiB"
	default:
		return figure(v) + " GiB-hours"
	}
}

func resourceName(resource budget.Resource) string {
	switch resource {
	case budget.USD:
		return "model cost"
	case budget.Tokens:
		return "token"
	}
	return string(resource)
}

// figure is n rounded to two decimals, without trailing zeros.
func figure(n float64) string {
	return strconv.FormatFloat(math.Round(n*100)/100, 'f', -1, 64)
}

func plural(n float64, unit string) string {
	if n == 1 {
		return unit
	}
	return unit + "s"
}
