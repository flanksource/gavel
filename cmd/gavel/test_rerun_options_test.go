package main

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func TestPrepareRerunOptionsDisablesRecursiveDiscovery(t *testing.T) {
	base := testrunner.RunOptions{
		Lint:            true,
		Recursive:       true,
		PassThroughArgs: []string{"--label-filter", "smoke"},
	}

	got := prepareRerunOptions(base, testui.RerunRequest{
		PackagePaths: []string{"./git"},
		WorkDir:      "/tmp/submodule",
		Framework:    "ginkgo",
	}, fixtureRerun{}, nil)

	if got.Lint {
		t.Fatalf("Lint should be disabled for reruns")
	}
	if got.Recursive {
		t.Fatalf("Recursive discovery should be disabled for explicit rerun packages")
	}
	if len(got.StartingPaths) != 1 || got.StartingPaths[0] != "./git" {
		t.Fatalf("unexpected rerun package paths: %#v", got.StartingPaths)
	}
	if got.WorkDir != "/tmp/submodule" {
		t.Fatalf("unexpected rerun workdir: %q", got.WorkDir)
	}
	if len(got.PassThroughArgs) != 0 {
		t.Fatalf("rerun should clear original pass-through args: %#v", got.PassThroughArgs)
	}
}

const resumeFixtureFile = "checks/purge.fixture.yaml"

// fixtureStepTest is a fixture step test as the run reports it: its Detail is
// the fixtures.ResultDetail record.
func fixtureStepTest(t *testing.T, file string, detail map[string]any, status func(*parsers.Test)) parsers.Test {
	t.Helper()
	raw, err := json.Marshal(detail)
	if err != nil {
		t.Fatal(err)
	}
	step := detail["step"].(map[string]any)
	test := parsers.Test{Name: step["id"].(string), Framework: parsers.Fixture, File: file, Detail: json.RawMessage(raw)}
	status(&test)
	return test
}

func markStepPassed(test *parsers.Test) { test.Passed = true }
func markStepFailed(test *parsers.Test) { test.Failed = true }

func stepDetail(id, kind string, extra map[string]any) map[string]any {
	detail := map[string]any{"step": map[string]any{"id": id, "kind": kind}, "outputs": map[string]any{}, "exports": map[string]any{}, "background": false}
	for key, value := range extra {
		detail[key] = value
	}
	return detail
}

// resumeRun is a last run of purge.fixture.yaml: params, a lookup that
// exported CLUSTER_URL, then stage, with another file's step alongside.
func resumeRun(t *testing.T, lookup func(*parsers.Test), extra map[string]any) []parsers.Test {
	params := stepDetail(fixtures.ParamsStepID, fixtures.FormKind, map[string]any{
		"outputs": map[string]any{"db": "lab", "values": map[string]any{"db": "lab"}},
		"form":    map[string]any{"state": "answered", "values": map[string]any{"db": "lab"}},
	})
	lookupDetail := stepDetail("lookup", "query", map[string]any{
		"outputs": map[string]any{"cluster": "east"},
		"exports": map[string]any{"CLUSTER_URL": map[string]any{"value": "https://east.example", "step": "lookup"}},
	})
	for key, value := range extra {
		lookupDetail[key] = value
	}
	return []parsers.Test{{Name: "purge", Framework: parsers.Fixture, File: resumeFixtureFile, Children: []parsers.Test{
		fixtureStepTest(t, resumeFixtureFile, params, markStepPassed),
		fixtureStepTest(t, resumeFixtureFile, lookupDetail, lookup),
		fixtureStepTest(t, resumeFixtureFile, stepDetail("stage", "query", nil), markStepFailed),
	}}, fixtureStepTest(t, "other.fixture.yaml", stepDetail("lookup", "query", nil), markStepPassed)}
}

func TestResumeFixtureRerunBuildsResumeFromTheLastRun(t *testing.T) {
	req := testui.RerunRequest{Fixture: resumeFixtureFile, FromStep: "stage"}

	got, err := resumeFixtureRerun(resumeRun(t, markStepPassed, nil), req)
	if err != nil {
		t.Fatal(err)
	}

	want := fixtureRerun{
		Resume: &fixtures.Resume{FromStep: "stage", Prior: map[string]fixtures.PriorStep{
			fixtures.ParamsStepID: {Outputs: map[string]any{"db": "lab", "values": map[string]any{"db": "lab"}}, Exports: map[string]fixtures.ExportRecord{}},
			"lookup": {Outputs: map[string]any{"cluster": "east"}, Exports: map[string]fixtures.ExportRecord{
				"CLUSTER_URL": {Value: "https://east.example", Step: "lookup"},
			}},
		}},
		Params: map[string]any{"db": "lab"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resume = %#v\nwant %#v", got, want)
	}
}

func TestResumeFixtureRerunRefusesAStepItCannotResumeAt(t *testing.T) {
	for _, tc := range []struct {
		name   string
		from   string
		lookup func(*parsers.Test)
		extra  map[string]any
		reason string
	}{
		{"a step not in the last run", "deploy", markStepPassed, nil, `step "deploy" is not in the last run of ` + resumeFixtureFile},
		{"an earlier failed step", "stage", markStepFailed, nil, `step "lookup" before "stage" did not pass`},
		{"an earlier background step", "stage", markStepPassed, map[string]any{"background": true}, `step "lookup" before "stage" ran in the background`},
		{"an earlier step that started a service", "stage", markStepPassed, map[string]any{fixtures.CleanupMetadataKey: true}, `step "lookup" before "stage" started a service`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := resumeFixtureRerun(resumeRun(t, tc.lookup, tc.extra), testui.RerunRequest{Fixture: resumeFixtureFile, FromStep: tc.from})

			var refusal testui.BadRerunRequest
			if !errors.As(err, &refusal) {
				t.Fatalf("err = %v, want a BadRerunRequest", err)
			}
			if !strings.Contains(refusal.Reason, tc.reason) {
				t.Fatalf("reason = %q, want it to contain %q", refusal.Reason, tc.reason)
			}
		})
	}
}

func TestPrepareRerunOptionsScopesAResumeToItsFixtureWithoutTouchingTheRun(t *testing.T) {
	runner := &fixtures.RunnerOptions{Paths: []string{"a.fixture.yaml", resumeFixtureFile}}
	base := testrunner.RunOptions{FixtureRunner: runner, FixtureFiles: []string{"**/*.fixture.yaml"}}
	rerun := fixtureRerun{Resume: &fixtures.Resume{FromStep: "stage"}, Params: map[string]any{"db": "lab"}}

	got := prepareRerunOptions(base, testui.RerunRequest{Fixture: resumeFixtureFile, FromStep: "stage"}, rerun, nil)

	if got.FixtureResume != rerun.Resume || !reflect.DeepEqual(got.FixtureParams, rerun.Params) {
		t.Fatalf("resume = %#v params = %#v", got.FixtureResume, got.FixtureParams)
	}
	if !reflect.DeepEqual(got.FixtureRunner.Paths, []string{resumeFixtureFile}) || !reflect.DeepEqual(got.FixtureFiles, []string{resumeFixtureFile}) {
		t.Fatalf("rerun paths = %v files = %v, want only %s", got.FixtureRunner.Paths, got.FixtureFiles, resumeFixtureFile)
	}
	if !got.FixturesOnly {
		t.Fatalf("a resumed rerun runs only its fixture")
	}
	if !reflect.DeepEqual(runner.Paths, []string{"a.fixture.yaml", resumeFixtureFile}) {
		t.Fatalf("the run's own fixture runner was mutated: %v", runner.Paths)
	}
}
