package fixtures

import (
	"fmt"
	"sort"
	"strings"

	"github.com/flanksource/clicky/task"
)

// Resume restarts a fixture file at FromStep: every step before it, in file
// order, is not run but reuses whatever it left in an earlier run (Prior),
// passed or not. A step that needs an output the earlier run does not have
// fails on reading it. The `params:` step is not reused: it resolves again from
// RunOptions.Params, so a setup that reads it is prepared again.
type Resume struct {
	FromStep string               `json:"from_step"`
	Prior    map[string]PriorStep `json:"prior"`
}

// PriorStep is what a reused step left behind in the run it is reused from.
type PriorStep struct {
	Outputs map[string]any          `json:"outputs,omitempty"`
	Exports map[string]ExportRecord `json:"exports,omitempty"`
	Skipped bool                    `json:"skipped,omitempty"`
	Failed  bool                    `json:"failed,omitempty"`
}

// RerunBase is the command that runs a run's fixtures again, up to their files:
// Argv run in Cwd, which the run's relative fixture files are relative to.
type RerunBase struct {
	Cwd  string   `json:"cwd"`
	Argv []string `json:"argv"`
}

// ReusedStep is Metadata["reused"] on a step a resumed run did not run.
type ReusedStep struct {
	Step string `json:"step"`
	From string `json:"from"`
}

// validate refuses a FromStep the fixtures do not have: no step would run.
func (r *Resume) validate(tests []FixtureTest) error {
	if r == nil {
		return nil
	}
	var ids []string
	for _, test := range tests {
		if test.Step == nil || test.Step.ID == "" {
			continue
		}
		if test.Step.ID == r.FromStep {
			return nil
		}
		ids = append(ids, test.Step.ID)
	}
	return fmt.Errorf("no step %q to resume from; steps: %s", r.FromStep, strings.Join(ids, ", "))
}

// reuseStep stands in for running a step before Resume.FromStep: it seeds the
// step's prior outputs and exports into the scope and reports how it ended in
// the resumed run, marked reused. It never fails the scope over that outcome.
func reuseStep(step *StepContext, resume Resume) FixtureResult {
	id := step.Spec.ID
	prior, ok := resume.Prior[id]
	if !ok {
		result := normalizeResult(FixtureResult{Status: task.StatusSKIP, Error: "has no result in the resumed run"}, step)
		stampReused(&result, step, resume.FromStep, prior)
		return result
	}
	if prior.Skipped {
		result := normalizeResult(FixtureResult{Status: task.StatusSKIP, Error: "skipped in the resumed run"}, step)
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
	result := FixtureResult{Status: task.StatusPASS}
	if prior.Failed {
		result = FixtureResult{Status: task.StatusWarning, Error: "failed in the resumed run"}
	}
	result = normalizeResult(result, step)
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
