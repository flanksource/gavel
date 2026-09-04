package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/google/uuid"

	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
)

// phaseRuns returns every issue's latest run per phase for the workspace,
// loading the whole set once per provider.
//
// Same bargain as labelResolver: the API opens a provider per request, so this
// memo is exactly one phase query per request, and a CLI command loads it once.
// The alternative — reading a todo's runs as each row is rendered — is the N+1
// that made /api/projects take 46 seconds, and it would be worse here because
// the per-row read joins Captain's session overview.
func (p *Provider) phaseRuns(ctx context.Context, workspaceID uuid.UUID) (map[uuid.UUID]types.PhaseRuns, error) {
	// Only the mapping unit tests construct a Provider without a repository.
	// An empty index is the correct answer for a source with no run history:
	// every phase renders as never run.
	if p.repository == nil {
		return nil, nil
	}

	p.phasesMu.RLock()
	cached, ok := p.phasesCache[workspaceID]
	p.phasesMu.RUnlock()
	if ok {
		return cached, nil
	}

	rows, err := p.repository.ListIssuePhaseRuns(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	index := make(map[uuid.UUID]types.PhaseRuns, len(rows))
	for _, row := range rows {
		runs, ok := index[row.IssueID]
		if !ok {
			runs = types.PhaseRuns{}
			index[row.IssueID] = runs
		}
		run, err := phaseRunFromNative(row)
		if err != nil {
			return nil, err
		}
		runs[types.Phase(row.Phase)] = run
	}

	p.phasesMu.Lock()
	if p.phasesCache == nil {
		p.phasesCache = map[uuid.UUID]map[uuid.UUID]types.PhaseRuns{}
	}
	p.phasesCache[workspaceID] = index
	p.phasesMu.Unlock()
	return index, nil
}

// dropPhaseRuns invalidates the memo so a read after a dispatch through this
// provider sees the run it just started.
func (p *Provider) dropPhaseRuns(workspaceID uuid.UUID) {
	p.phasesMu.Lock()
	delete(p.phasesCache, workspaceID)
	p.phasesMu.Unlock()
}

func phaseRunFromNative(row native.IssuePhaseRun) (types.PhaseRun, error) {
	progress, err := phaseProgress(row)
	if err != nil {
		return types.PhaseRun{}, fmt.Errorf("phase %s of issue %s: %w", row.Phase, row.IssueID, err)
	}
	run := types.PhaseRun{
		Phase:      types.Phase(row.Phase),
		State:      row.State,
		StartedAt:  row.StartedAt,
		FinishedAt: row.FinishedAt,
		CostUSD:    row.CostUSD,
		Active:     row.Active,
		Progress:   progress,
	}
	if row.DurationSeconds != nil {
		run.DurationMS = int64(*row.DurationSeconds * float64(time.Second/time.Millisecond))
	}
	return run, nil
}

// phaseProgress reads the unit the phase actually counts in. Plan, run and
// triage count agent iterations; verification counts the checks in its fixture.
//
// The verify counts come from captain's own record —
// captain_prompt_run_iterations.verification_result, surfaced by
// captain_prompt_run_overview.latest_verification_result — and from nowhere
// else. They used to be read back out of a gavel-shaped copy under the run's
// result_json.definitionOfDone.progress, which meant a run that produced a
// verdict but never wrote the copy rendered as "no progress" while captain held
// the counts all along. There is one record of a verification now.
func phaseProgress(row native.IssuePhaseRun) (types.PhaseProgress, error) {
	if row.Phase == native.StepVerify {
		return verificationProgress(row.VerificationResult)
	}
	return types.PhaseProgress{
		Done:   row.Succeeded,
		Failed: row.Failed,
		Total:  row.Iterations,
	}, nil
}

// verificationProgress is the check counts of captain's stored report. An empty
// column is a phase that has not produced a verdict — no progress to show — but
// a column that will not decode is corrupt, and reporting that as an empty pass
// is exactly the silence this column exists to prevent.
func verificationProgress(verificationResult string) (types.PhaseProgress, error) {
	if strings.TrimSpace(verificationResult) == "" {
		return types.PhaseProgress{}, nil
	}
	var report api.VerifyReport
	if err := json.Unmarshal([]byte(verificationResult), &report); err != nil {
		return types.PhaseProgress{}, fmt.Errorf("decode Captain verification result: %w", err)
	}
	return types.PhaseProgress{
		Done:   report.Summary.Passed,
		Failed: report.Summary.Failed,
		Total:  report.Summary.Total,
	}, nil
}
