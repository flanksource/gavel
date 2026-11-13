package main

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// fixtureRerun is what a "retry from here" rerun request resumes: the fixture
// run's Resume, and the earlier params form's answered values.
type fixtureRerun struct {
	Resume *fixtures.Resume
	Params map[string]any
}

// fixtureStepDetail is the part of a fixture step test's Detail
// (fixtures.ResultDetail) a resumed rerun reuses.
type fixtureStepDetail struct {
	Step       fixtures.StepRef                 `json:"step"`
	Outputs    map[string]any                   `json:"outputs"`
	Exports    map[string]fixtures.ExportRecord `json:"exports"`
	Background bool                             `json:"background"`
	Cleanup    bool                             `json:"cleanup"`
	Form       *struct {
		Values map[string]any `json:"values"`
	} `json:"form"`
}

type fixtureStepRun struct {
	test   parsers.Test
	detail fixtureStepDetail
}

// resumeFixtureRerun builds the fixture resume req asks for from tests, the
// current run: the steps of req.Fixture before req.FromStep, in file order.
// It refuses (testui.BadRerunRequest) a step that is not in that run, or one
// whose earlier steps did not pass or left a process behind. A request that
// resumes nothing gives the zero fixtureRerun.
func resumeFixtureRerun(tests []parsers.Test, req testui.RerunRequest) (fixtureRerun, error) {
	if req.FromStep == "" {
		return fixtureRerun{}, nil
	}
	steps, err := fixtureStepRuns(tests, req.Fixture)
	if err != nil {
		return fixtureRerun{}, err
	}
	from := slices.IndexFunc(steps, func(step fixtureStepRun) bool { return step.detail.Step.ID == req.FromStep })
	if from < 0 {
		return fixtureRerun{}, testui.BadRerunRequest{Reason: fmt.Sprintf("cannot retry from step %q: step %q is not in the last run of %s", req.FromStep, req.FromStep, req.Fixture)}
	}
	rerun := fixtureRerun{Resume: &fixtures.Resume{FromStep: req.FromStep, Prior: map[string]fixtures.PriorStep{}}}
	for _, step := range steps[:from] {
		if reason := unreusableStep(step); reason != "" {
			return fixtureRerun{}, testui.BadRerunRequest{Reason: fmt.Sprintf("cannot retry from step %q: step %q before %q %s", req.FromStep, step.detail.Step.ID, req.FromStep, reason)}
		}
		rerun.Resume.Prior[step.detail.Step.ID] = fixtures.PriorStep{Outputs: step.detail.Outputs, Exports: step.detail.Exports, Skipped: step.test.Skipped}
		if step.detail.Step.ID == fixtures.ParamsStepID && step.detail.Form != nil {
			rerun.Params = step.detail.Form.Values
		}
	}
	return rerun, nil
}

// unreusableStep is why a resumed rerun cannot reuse step, empty when it can.
func unreusableStep(step fixtureStepRun) string {
	switch {
	case step.detail.Background:
		return "ran in the background"
	case step.detail.Cleanup:
		return "started a service"
	case !step.test.Passed && !step.test.Skipped:
		return "did not pass"
	}
	return ""
}

// fixtureStepRuns are the typed step tests of file in tests, in tree (file) order.
func fixtureStepRuns(tests []parsers.Test, file string) ([]fixtureStepRun, error) {
	var steps []fixtureStepRun
	var walk func([]parsers.Test) error
	walk = func(nodes []parsers.Test) error {
		for _, test := range nodes {
			if test.Framework == parsers.Fixture && test.File == file && test.Detail != nil {
				detail, err := decodeFixtureStepDetail(test)
				if err != nil {
					return err
				}
				if detail.Step.ID != "" {
					steps = append(steps, fixtureStepRun{test: test, detail: detail})
				}
			}
			if err := walk(test.Children); err != nil {
				return err
			}
		}
		return nil
	}
	return steps, walk(tests)
}

func decodeFixtureStepDetail(test parsers.Test) (fixtureStepDetail, error) {
	var detail fixtureStepDetail
	raw, err := json.Marshal(test.Detail)
	if err == nil {
		err = json.Unmarshal(raw, &detail)
	}
	if err != nil {
		return detail, fmt.Errorf("read fixture step %q of %s: %w", test.Name, test.File, err)
	}
	return detail, nil
}

// scope limits a rerun to the one fixture file it resumes, over a copy of the
// run's fixture runner options so the run's own are left as they were.
func (f fixtureRerun) scope(opts testrunner.RunOptions, file string) testrunner.RunOptions {
	opts.FixtureResume, opts.FixtureParams = f.Resume, f.Params
	opts.Fixtures, opts.FixturesOnly = true, true
	opts.FixtureFiles = []string{file}
	if opts.FixtureRunner != nil {
		runner := *opts.FixtureRunner
		runner.Paths = []string{file}
		opts.FixtureRunner = &runner
	}
	return opts
}
