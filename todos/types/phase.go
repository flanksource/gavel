package types

import "time"

// Phase is one step of a TODO's lifecycle, as it is recorded rather than as it
// behaves: triage runs as a plan-class run but records itself separately, so a
// backlog can tell a triage pass from a planning pass.
type Phase string

// Named <noun>Phase rather than Phase<noun> so the run phase does not collide
// with PhaseRun, the record of what happened during a phase.
const (
	PlanPhase   Phase = "plan"
	TriagePhase Phase = "triage"
	RunPhase    Phase = "run"
	VerifyPhase Phase = "verify"
)

// Phases is the order the pipeline reads in: what you plan, you run; what you
// run, you verify. Triage sits with plan because it is the other read-only pass.
var Phases = []Phase{PlanPhase, TriagePhase, RunPhase, VerifyPhase}

// PhaseProgress is how far through its own work a phase got. The unit differs
// by phase and deliberately is not normalised: plan, run and triage count
// agent iterations, while verification counts the checks in its fixture.
type PhaseProgress struct {
	Done   int `json:"done"`
	Failed int `json:"failed,omitempty"`
	Total  int `json:"total"`
}

// Empty reports whether there is nothing to show, so a renderer can omit the
// progress rather than draw "0/0".
func (p PhaseProgress) Empty() bool { return p.Total == 0 && p.Done == 0 && p.Failed == 0 }

// PhaseRun is the latest run a TODO has for one phase: enough to render that
// phase's status, progress and elapsed time in a list without opening the todo.
type PhaseRun struct {
	Phase Phase `json:"phase"`
	// State is the run's own outcome: pending, running, waiting, succeeded,
	// failed or cancelled. It is NOT the TODO's status, which folds every phase
	// into one value.
	State string `json:"state"`
	// Progress carries no omitempty because encoding/json does not honour it for
	// structs: a phase that counts nothing serialises as zeroes, which readers
	// must treat as "no progress to show" rather than "made no progress".
	Progress   PhaseProgress `json:"progress"`
	StartedAt  *time.Time    `json:"started_at,omitempty"`
	FinishedAt *time.Time    `json:"finished_at,omitempty"`
	// DurationMS is the elapsed time at the moment this was read. A running
	// phase keeps accruing, so a live renderer should tick from StartedAt
	// instead of re-reading this.
	DurationMS int64   `json:"duration_ms,omitempty"`
	CostUSD    float64 `json:"cost_usd,omitempty"`
	// Active marks the phase the TODO is executing right now, as opposed to the
	// one it ran most recently.
	Active bool `json:"active,omitempty"`
}

// Running reports whether this phase is the one currently executing.
func (r PhaseRun) Running() bool { return r.State == "running" || r.State == "pending" }

// Failed reports whether the phase ended badly, which for verification includes
// producing failing checks.
func (r PhaseRun) Failed() bool {
	return r.State == "failed" || r.State == "cancelled" || r.Progress.Failed > 0
}

// PhaseRuns indexes a TODO's latest run per phase. A phase that has never run
// is absent rather than zero-valued, so "not started" and "started and produced
// nothing" stay distinguishable.
type PhaseRuns map[Phase]PhaseRun

// Ordered returns the runs that exist, in pipeline order.
//
// Each run is stamped with the phase it is keyed under: the key is what a
// caller looked the run up by, so a value that disagrees with it would render
// one phase under another's heading.
func (p PhaseRuns) Ordered() []PhaseRun {
	runs := make([]PhaseRun, 0, len(p))
	for _, phase := range Phases {
		if run, ok := p[phase]; ok {
			run.Phase = phase
			runs = append(runs, run)
		}
	}
	return runs
}
