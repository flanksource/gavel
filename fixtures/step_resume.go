package fixtures

import (
	"fmt"
	"sort"

	"github.com/flanksource/clicky/task"
)

// Resume restarts a fixture file at FromStep: every step before it, in file
// order, is not run but reuses its result from an earlier run (Prior).
type Resume struct {
	FromStep string               `json:"from_step"`
	Prior    map[string]PriorStep `json:"prior"`
}

// PriorStep is what a reused step left behind in the run it is reused from.
type PriorStep struct {
	Outputs map[string]any          `json:"outputs,omitempty"`
	Exports map[string]ExportRecord `json:"exports,omitempty"`
	Skipped bool                    `json:"skipped,omitempty"`
}

// CleanupMetadataKey is Metadata[key] = true on a step that registered
// teardown work, such as a service it started: a resumed run cannot reuse it.
const CleanupMetadataKey = "cleanup"

// ReusedStep is Metadata["reused"] on a step a resumed run did not run.
type ReusedStep struct {
	Step string `json:"step"`
	From string `json:"from"`
}

// reuseStep stands in for running a step before Resume.FromStep: it seeds the
// step's prior outputs and exports into the scope and passes it (or skips it,
// when it was skipped before), marked reused.
func reuseStep(step *StepContext, resume Resume) FixtureResult {
	id := step.Spec.ID
	prior, ok := resume.Prior[id]
	if !ok {
		step.Scope.MarkFailed(id)
		return stepError(step, fmt.Errorf("resume: step %q runs before %q but the resumed run has no result for it", id, resume.FromStep))
	}
	if prior.Skipped {
		result := normalizeResult(FixtureResult{Type: step.Spec.Kind, Status: task.StatusSKIP, Error: "skipped in the resumed run"}, step)
		stampReused(&result, step, resume.FromStep, PriorStep{})
		return recordSkipped(result, step, "skipped in the resumed run")
	}
	if err := step.Scope.SetOutputs(id, prior.Outputs); err != nil {
		return stepError(step, fmt.Errorf("resume: step %q: %w", id, err))
	}
	if err := reapplyExports(step, prior.Exports); err != nil {
		step.Scope.MarkFailed(id)
		return stepError(step, fmt.Errorf("resume: step %q: %w", id, err))
	}
	result := normalizeResult(FixtureResult{Status: task.StatusPASS}, step)
	stampReused(&result, step, resume.FromStep, prior)
	return result
}

func reapplyExports(step *StepContext, exports map[string]ExportRecord) error {
	names := make([]string, 0, len(exports))
	for name := range exports {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		record := exports[name]
		if err := step.Scope.Env.Export(record.Step, name, record.Value, step.Spec.Override); err != nil {
			return err
		}
	}
	return nil
}

func stampReused(result *FixtureResult, step *StepContext, from string, prior PriorStep) {
	exports := prior.Exports
	if exports == nil {
		exports = map[string]ExportRecord{}
	}
	stampStepMetadata(result, step.Spec, prior.Outputs, exports, nil, nil)
	result.Metadata["reused"] = ReusedStep{Step: step.Spec.ID, From: from}
}
