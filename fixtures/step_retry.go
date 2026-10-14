package fixtures

import (
	"context"
	"fmt"

	"github.com/flanksource/clicky/task"
)

// AttemptReport is one Execute attempt, recorded under Metadata["attempts"].
type AttemptReport struct {
	Attempt    int    `json:"attempt"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"durationMs"`
}

// retrier re-runs a failed Execute while the policy's `when` holds, up to its
// attempts cap, sleeping delay·factor^n ± jitter between attempts. clicky's
// task retry matches error substrings only; a CEL predicate over the failed
// result with a per-attempt record is what a fixture step needs, so the loop
// lives here.
type retrier struct {
	policy RetryPolicy
	check  celCheck
	vars   func(outcome StepOutcome, err error) map[string]any
	clock  stepClock
	rand   func() float64
}

func (r retrier) run(ctx context.Context, execute func(context.Context) (StepOutcome, error)) (StepOutcome, []AttemptReport) {
	attempts := max(r.policy.Attempts, 1)
	spread, jitterErr := parseJitter(r.policy.Jitter)
	var reports []AttemptReport
	for n := 1; ; n++ {
		started := r.clock.Now()
		outcome, err := execute(ctx)
		report := AttemptReport{Attempt: n, Status: attemptStatus(outcome.Result.Status),
			DurationMs: r.clock.Now().Sub(started).Milliseconds(), Error: outcome.Result.Error}
		if err != nil {
			report.Status, report.Error = attemptStatus(task.StatusERR), err.Error()
		}
		reports = append(reports, report)
		if err == nil && isPassingStatus(outcome.Result.Status) {
			return outcome, reports
		}
		outcome = withExecuteError(outcome, err)
		if n >= attempts || r.policy.When == "" {
			return outcome, reports
		}
		if jitterErr != nil {
			return failOutcome(outcome, jitterErr), reports
		}
		retry, _, checkErr := r.check(r.policy.When, r.vars(outcome, err))
		if checkErr != nil {
			return failOutcome(outcome, fmt.Errorf("retry when %s: %w", r.policy.When, checkErr)), reports
		}
		if !retry {
			return outcome, reports
		}
		var base Duration
		if r.policy.Delay != nil {
			base = *r.policy.Delay
		}
		delay := backoffDelay(base.Std(), r.policy.Backoff, spread, n-1, r.rand)
		if err := r.clock.Sleep(ctx, delay); err != nil {
			return failOutcome(outcome, fmt.Errorf("retry interrupted: %w", err)), reports
		}
	}
}

// Attempt statuses recorded under Metadata["attempts"][].status.
const (
	AttemptPassed    = "passed"
	AttemptFailed    = "failed"
	AttemptErrored   = "errored"
	AttemptSkipped   = "skipped"
	AttemptWarned    = "warned"
	AttemptCancelled = "cancelled"
)

// attemptStatus maps a result status onto the lowercase verdict vocabulary the
// step detail renderer reads. A status no verdict covers is a kind bug.
func attemptStatus(status task.Status) string {
	switch status {
	case task.StatusPASS, task.StatusSuccess:
		return AttemptPassed
	case task.StatusFAIL, task.StatusFailed:
		return AttemptFailed
	case task.StatusSKIP:
		return AttemptSkipped
	case task.StatusWarning:
		return AttemptWarned
	case task.StatusCancelled:
		return AttemptCancelled
	default:
		return AttemptErrored
	}
}

// retryVarsFor is the CEL environment `retry.when` sees: the failed attempt's
// outputs as top-level roots (a kind may return outputs alongside an Execute
// error, e.g. todo_prompt's `launch`), then the reserved roots error, exitCode,
// outputs and status, which win over an output of the same name.
func retryVarsFor(outcome StepOutcome, err error) map[string]any {
	message := outcome.Result.Error
	if err != nil {
		message = err.Error()
	}
	outputs := outcome.Outputs
	if outputs == nil {
		outputs = map[string]any{}
	}
	vars := make(map[string]any, len(outputs)+4)
	for k, v := range outputs {
		vars[k] = v
	}
	vars["error"] = message
	vars["exitCode"] = outcome.Result.ExitCode
	vars["outputs"] = outputs
	vars["status"] = string(outcome.Result.Status)
	return vars
}

// withExecuteError turns an Execute error into the errored result it stands for.
func withExecuteError(outcome StepOutcome, err error) StepOutcome {
	if err == nil {
		return outcome
	}
	outcome.Result.Status = task.StatusERR
	outcome.Result.Error = err.Error()
	return outcome
}

func failOutcome(outcome StepOutcome, err error) StepOutcome {
	outcome.Result.Status = task.StatusERR
	if outcome.Result.Error != "" {
		outcome.Result.Error += "; " + err.Error()
	} else {
		outcome.Result.Error = err.Error()
	}
	return outcome
}
