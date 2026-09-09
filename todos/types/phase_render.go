package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

// phaseTone maps a phase run's outcome onto the palette the rest of the TODO
// renderers use: red for failure, amber while waiting on a person, blue while
// working, green when done.
func phaseTone(run PhaseRun) string {
	switch {
	case run.Failed():
		return "text-red-600"
	case run.State == "waiting":
		return "text-amber-600"
	case run.Running():
		return "text-blue-600"
	case run.State == "succeeded":
		return "text-green-600"
	default:
		return "text-muted"
	}
}

// phaseGlyph is a single character standing in for the run's state, so a phase
// column stays one cell wide in a terminal.
func phaseGlyph(run PhaseRun) string {
	switch {
	case run.Failed():
		return "✗"
	case run.State == "waiting":
		return "?"
	case run.Running():
		return "◐"
	case run.State == "succeeded":
		return "✓"
	default:
		return "·"
	}
}

// phaseDuration renders elapsed time. A running phase is measured against now
// rather than the duration recorded when the row was read, so a terminal that
// re-renders shows the timer advancing.
func phaseDuration(run PhaseRun) string {
	elapsed := time.Duration(run.DurationMS) * time.Millisecond
	if run.Running() && run.StartedAt != nil {
		elapsed = time.Since(*run.StartedAt)
	}
	if elapsed <= 0 {
		return ""
	}
	return summaryAge(elapsed)
}

// Pretty renders one phase as "glyph progress duration" — the status, how far
// it got, and how long it took. Progress is omitted when the phase counts
// nothing, so a single-iteration run does not read as "1/1".
func (r PhaseRun) Pretty() api.Text {
	parts := []string{phaseGlyph(r)}
	if !r.Progress.Empty() && r.Progress.Total > 1 {
		parts = append(parts, fmt.Sprintf("%d/%d", r.Progress.Done, r.Progress.Total))
	} else if r.Progress.Failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", r.Progress.Failed))
	}
	if elapsed := phaseDuration(r); elapsed != "" {
		parts = append(parts, elapsed)
	}
	return clicky.Text(strings.Join(parts, " "), phaseTone(r))
}

// phaseCell renders one phase column. A phase that has never run gets an
// em-dash rather than an empty cell, so "not started" is visibly different from
// a column that failed to render.
func (t TODO) phaseCell(phase Phase) api.Text {
	run, ok := t.PhaseRuns[phase]
	if !ok {
		return clicky.Text("—", "text-muted")
	}
	return run.Pretty()
}
