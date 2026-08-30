package runtime

import (
	"context"
	"encoding/json"
	"time"

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
		runs[types.Phase(row.Phase)] = phaseRunFromNative(row)
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

func phaseRunFromNative(row native.IssuePhaseRun) types.PhaseRun {
	run := types.PhaseRun{
		Phase:      types.Phase(row.Phase),
		State:      row.State,
		StartedAt:  row.StartedAt,
		FinishedAt: row.FinishedAt,
		CostUSD:    row.CostUSD,
		Active:     row.Active,
		Progress:   phaseProgress(row),
	}
	if row.DurationSeconds != nil {
		run.DurationMS = int64(*row.DurationSeconds * float64(time.Second/time.Millisecond))
	}
	return run
}

// phaseProgress reads the unit the phase actually counts in. Plan, run and
// triage count agent iterations; verification counts the checks in its fixture,
// which the lifecycle records as a definition-of-done snapshot on the run.
func phaseProgress(row native.IssuePhaseRun) types.PhaseProgress {
	if row.Phase == native.StepVerify {
		if progress, ok := verificationProgress(row.ResultJSON); ok {
			return progress
		}
	}
	return types.PhaseProgress{
		Done:   row.Succeeded,
		Failed: row.Failed,
		Total:  row.Iterations,
	}
}

// verificationSnapshot is the shape lifecycle.go writes under
// definitionOfDone.progress — the fixtures.ExecutionSnapshot summary. Only the
// counts are read here; the node tree belongs to the detail view.
type verificationSnapshot struct {
	DefinitionOfDone struct {
		Progress struct {
			Summary struct {
				Total   int `json:"total"`
				Passed  int `json:"passed"`
				Failed  int `json:"failed"`
				Running int `json:"running"`
			} `json:"summary"`
		} `json:"progress"`
	} `json:"definitionOfDone"`
}

func verificationProgress(resultJSON []byte) (types.PhaseProgress, bool) {
	if len(resultJSON) == 0 {
		return types.PhaseProgress{}, false
	}
	var snapshot verificationSnapshot
	if err := json.Unmarshal(resultJSON, &snapshot); err != nil {
		return types.PhaseProgress{}, false
	}
	summary := snapshot.DefinitionOfDone.Progress.Summary
	if summary.Total == 0 && summary.Passed == 0 && summary.Failed == 0 {
		return types.PhaseProgress{}, false
	}
	return types.PhaseProgress{Done: summary.Passed, Failed: summary.Failed, Total: summary.Total}, true
}
