package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/kinds/form"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// fixtureRerun is what rerunning a fixture from a step resumes: the fixture
// run's Resume, and the earlier params form's answered values.
type fixtureRerun struct {
	Resume *fixtures.Resume
	Params map[string]any
}

// fixtureStepDetail is the part of a fixture step test's Detail
// (fixtures.ResultDetail) a resumed rerun reuses.
type fixtureStepDetail struct {
	Step    fixtures.StepRef                 `json:"step"`
	Outputs map[string]any                   `json:"outputs"`
	Exports map[string]fixtures.ExportRecord `json:"exports"`
	Reused  *fixtures.ReusedStep             `json:"reused"`
	Form    *struct {
		State  string         `json:"state"`
		Values map[string]any `json:"values"`
	} `json:"form"`
}

// fixtureResumeRequest names the fixture file and the step to resume it at.
// WorkDir is what a relative File, here or in the resumed run, is relative to.
type fixtureResumeRequest struct {
	File     string
	FromStep string
	WorkDir  string
}

// resumeFixture builds the resume of req from tests, an earlier run: whatever
// each step of that fixture left behind, passed or not.
func resumeFixture(tests []parsers.Test, req fixtureResumeRequest) (fixtureRerun, error) {
	rerun := fixtureRerun{Resume: &fixtures.Resume{FromStep: req.FromStep, Prior: map[string]fixtures.PriorStep{}}}
	file := absoluteFixtureFile(req.File, req.WorkDir)
	var walk func([]parsers.Test) error
	walk = func(nodes []parsers.Test) error {
		for _, test := range nodes {
			if test.Framework == parsers.Fixture && test.Detail != nil && absoluteFixtureFile(test.File, req.WorkDir) == file {
				if err := rerun.reuse(test); err != nil {
					return err
				}
			}
			if err := walk(test.Children); err != nil {
				return err
			}
		}
		return nil
	}
	return rerun, walk(tests)
}

// reuse records what test, a step of the resumed fixture, left behind.
func (f *fixtureRerun) reuse(test parsers.Test) error {
	var detail fixtureStepDetail
	raw, err := json.Marshal(test.Detail)
	if err == nil {
		err = json.Unmarshal(raw, &detail)
	}
	if err != nil {
		return fmt.Errorf("read fixture step %q of %s: %w", test.Name, test.File, err)
	}
	switch detail.Step.ID {
	case "":
	case fixtures.ParamsStepID:
		if detail.Form != nil && (detail.Form.State == form.FormAnswered || detail.Form.State == form.FormSupplied) {
			f.Params = detail.Form.Values
		}
	default:
		f.Resume.Prior[detail.Step.ID] = fixtures.PriorStep{
			Outputs: detail.Outputs, Exports: detail.Exports, Skipped: test.Skipped,
			// A reused step warns only when it failed in the run it came from.
			Failed: !test.Passed && !test.Skipped && (!test.Warned || detail.Reused != nil),
		}
	}
	return nil
}

func absoluteFixtureFile(file, workDir string) string {
	if !filepath.IsAbs(file) {
		file = filepath.Join(workDir, file)
	}
	return filepath.Clean(file)
}

// resumeFixtureRerun builds the fixture resume a UI rerun request asks for from
// tests, the current run; a request that resumes nothing gives the zero
// fixtureRerun. A step the file does not have is refused
// (testui.BadRerunRequest) before anything runs.
func resumeFixtureRerun(tests []parsers.Test, req testui.RerunRequest, workDir string) (fixtureRerun, error) {
	if req.FromStep == "" {
		return fixtureRerun{}, nil
	}
	rerun, err := resumeFixture(tests, fixtureResumeRequest{File: req.Fixture, FromStep: req.FromStep, WorkDir: workDir})
	if err != nil {
		return fixtureRerun{}, err
	}
	runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{absoluteFixtureFile(req.Fixture, workDir)}, WorkDir: workDir, Resume: rerun.Resume})
	if err != nil {
		return fixtureRerun{}, err
	}
	if _, err := runner.Parse(); err != nil {
		return fixtureRerun{}, testui.BadRerunRequest{Reason: fmt.Sprintf("cannot retry from step %q: %v", req.FromStep, err)}
	}
	return rerun, nil
}

// scope limits a rerun to the one fixture file it resumes, over a copy of the
// run's fixture runner options so the run's own are left as they were.
func (f fixtureRerun) scope(opts testrunner.RunOptions, file string) testrunner.RunOptions {
	opts.FixtureResume = f.Resume
	if f.Params != nil {
		opts.FixtureParams = f.Params
	}
	opts.Fixtures, opts.FixturesOnly = true, true
	opts.FixtureFiles = []string{file}
	if opts.FixtureRunner != nil {
		runner := *opts.FixtureRunner
		runner.Paths = []string{file}
		opts.FixtureRunner = &runner
	}
	return opts
}
