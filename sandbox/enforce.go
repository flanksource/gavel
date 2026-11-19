// Stops an organization's running runs once one of its cumulative budgets is used up, saying why:
// every run for CPU and memory, the agent runs for tokens and model cost, which only they spend.
package sandbox

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/budget"
	"github.com/google/uuid"
)

// activeRun is a running run the dispatcher can stop.
type activeRun struct {
	org        uuid.UUID
	modelCalls bool
	stop       func(reason error)
	// stopped is why the run was stopped; nil while it runs.
	stopped error
}

// track keeps run id of org, which makes model calls when modelCalls is set, until it finishes,
// stopping it with stop.
func (d *Dispatcher) track(id, org uuid.UUID, modelCalls bool, stop func(reason error)) {
	d.activeMu.Lock()
	defer d.activeMu.Unlock()
	if d.active == nil {
		d.active = map[uuid.UUID]*activeRun{}
	}
	d.active[id] = &activeRun{org: org, modelCalls: modelCalls, stop: stop}
}

// untrack forgets run id as it finishes, returning why it was stopped: nil when it wasn't.
func (d *Dispatcher) untrack(id uuid.UUID) error {
	d.activeMu.Lock()
	defer d.activeMu.Unlock()
	run, ok := d.active[id]
	delete(d.active, id)
	if !ok {
		return nil
	}
	return run.stopped
}

// Enforce stops org's running runs that one of its used-up cumulative budgets covers, recording
// why on each.
func (d *Dispatcher) Enforce(ctx context.Context, org uuid.UUID) error {
	if d.Admission == nil {
		return nil
	}
	statuses, err := d.Admission.Budgets.Status(ctx, org, d.RunLifetime())
	if err != nil {
		return fmt.Errorf("enforce the budgets of organization %s: %w", org, err)
	}
	every, modelCalls := usedUpBudgets(statuses)
	if len(every) == 0 && len(modelCalls) == 0 {
		return nil
	}
	var stops []func()
	d.activeMu.Lock()
	for _, run := range d.active {
		reasons := every
		if run.modelCalls {
			reasons = append(append([]string{}, every...), modelCalls...)
		}
		if run.org != org || run.stopped != nil || len(reasons) == 0 {
			continue
		}
		reason := fmt.Errorf("%w: %s; an organization admin can raise it", ErrBudgetUsedUp, strings.Join(reasons, "; "))
		run.stopped = reason
		stops = append(stops, func() { run.stop(reason) })
	}
	d.activeMu.Unlock()
	for _, stop := range stops {
		stop()
	}
	return nil
}

// usedUpBudgets say which of statuses are used-up cumulative budgets: those that stop every run,
// and those that stop the runs making model calls.
func usedUpBudgets(statuses []budget.Status) (every, modelCalls []string) {
	for _, status := range statuses {
		switch {
		case status.Kind != budget.Cumulative || !status.Reached():
		case isModelBudget(status.Resource):
			modelCalls = append(modelCalls, usedUp(status))
		default:
			every = append(every, usedUp(status))
		}
	}
	return every, modelCalls
}

// UsedUpModelBudgets says which of an organization's token and model cost budgets are used up,
// and how far; none when its model calls may go on.
func UsedUpModelBudgets(statuses []budget.Status) []string {
	_, modelCalls := usedUpBudgets(statuses)
	return modelCalls
}

// WatchBudgets enforces the budgets of every organization with a running run, every interval
// until ctx ends.
func (d *Dispatcher) WatchBudgets(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		d.activeMu.Lock()
		orgs := map[uuid.UUID]bool{}
		for _, run := range d.active {
			orgs[run.org] = true
		}
		d.activeMu.Unlock()
		for org := range orgs {
			if err := d.Enforce(ctx, org); err != nil && ctx.Err() == nil {
				logger.Warnf("%v", err)
			}
		}
	}
}

// RemainingModelCost is how much of org's model cost budget is left, in USD; capped is false
// when it has none.
func (d *Dispatcher) RemainingModelCost(ctx context.Context, org uuid.UUID) (remaining float64, capped bool, err error) {
	if d.Admission == nil {
		return 0, false, nil
	}
	statuses, err := d.Admission.Budgets.Status(ctx, org, d.RunLifetime())
	if err != nil {
		return 0, false, fmt.Errorf("read the model cost budget of organization %s: %w", org, err)
	}
	for _, status := range statuses {
		if status.Resource == budget.USD && status.Kind == budget.Cumulative {
			return max(0, status.Amount-status.Used), true, nil
		}
	}
	return 0, false, nil
}
