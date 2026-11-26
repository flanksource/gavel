// Derives the status a run is shown with from its events: where in its life it is, or where it
// failed, more finely than the status recorded on it.
package sandbox

import "time"

// RunEvent is one row of sandbox_run_events.
type RunEvent struct {
	At      time.Time `json:"at"`
	Source  string    `json:"source"`
	Kind    string    `json:"kind"`
	Message string    `json:"message"`
}

// RunPhase is the status a run of backend, recorded as status, is shown with, from its events:
//
//   - rejected: it failed before it was submitted, as when its budgets refused it;
//   - waiting for worker, then starting once substrate placed it on a worker;
//   - failed to assign: it failed before substrate placed it on a worker;
//   - failed to start: it failed on its worker before its job ran;
//
// and status otherwise. Only substrate reports placing a run, so only its runs wait for a worker.
// A run recorded before runs had events, whose history lacks the queued event that creates every
// run now, is shown as status.
func RunPhase(status, backend string, events []RunEvent) string {
	if len(events) == 0 || events[0].Source != "gavel" || events[0].Kind != "queued" {
		return status
	}
	var submitted, assigned, ran bool
	for _, event := range events {
		switch {
		case event.Source == "substrate" && event.Kind == "assigned":
			assigned = true
		case event.Source == "gavel" && event.Kind == "provisioning":
			submitted = true
		case event.Source == "gavel" && event.Kind == "running":
			submitted, ran = true, true
		}
	}
	placesOnWorkers := backend == "substrate"
	switch {
	case status == "provisioning" && placesOnWorkers && assigned:
		return "starting"
	case status == "provisioning" && placesOnWorkers:
		return "waiting for worker"
	case status != "failed" || ran:
		return status
	case !submitted:
		return "rejected"
	case placesOnWorkers && !assigned:
		return "failed to assign"
	case placesOnWorkers:
		return "failed to start"
	}
	return status
}
