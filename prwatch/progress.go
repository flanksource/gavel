package prwatch

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/gavel/github"
)

// followProgressLine is the heartbeat a follower that cannot redraw gets in
// place of the full status frame.
//
// `gavel pr status --follow` used to reprint the entire PR table every
// --interval whenever stderr was not a TTY, which under an agent's verify hook
// is 15-20 copies of the same 40-line block inside one CI cycle. A reader that
// cannot redraw needs to know something is still moving and what changed, not
// the frame again — the frame is printed once, at the end, by the caller, which
// is also what a verify hook's feedback tail keeps.
func followProgressLine(result *PRWatchResult, interval time.Duration) string {
	complete, failed, running := 0, 0, 0
	for _, check := range result.PR.StatusCheckRollup {
		switch {
		case check.Status != "COMPLETED":
			running++
		case github.IsFailureConclusion(check.Conclusion):
			complete, failed = complete+1, failed+1
		default:
			complete++
		}
	}

	parts := []string{fmt.Sprintf("#%d: %d/%d checks complete",
		result.PR.Number, complete, len(result.PR.StatusCheckRollup))}
	if running > 0 {
		parts = append(parts, fmt.Sprintf("%d running", running))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failing", failed))
	}
	if n := result.UnresolvedComments(); n > 0 {
		parts = append(parts, fmt.Sprintf("%d unresolved comment(s)", n))
	}
	if n := failedGavelResults(result); n > 0 {
		parts = append(parts, fmt.Sprintf("%d gavel artifact(s) failing", n))
	}
	if result.HasFailedRun() {
		parts = append(parts, "failed jobs present")
	}
	return strings.Join(parts, ", ") + fmt.Sprintf(" — next poll in %s", interval)
}

// failedGavelResults counts the harvested gavel artifacts reporting a failure,
// the signal a green rollup hides (see statusExitCode).
func failedGavelResults(result *PRWatchResult) int {
	n := 0
	for _, summary := range result.GavelResults {
		if summary != nil && summary.HasFailure() {
			n++
		}
	}
	return n
}
