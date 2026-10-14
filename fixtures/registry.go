package fixtures

import (
	"context"

	"github.com/flanksource/captain/pkg/api"

	"github.com/flanksource/gavel/fixtures/record"
	"github.com/flanksource/gavel/processruntime"
)

// RunOptions provides configuration options for executing fixture tests.
// It controls execution behavior including working directory, verbosity, and caching.
type RunOptions struct {
	ProcessRuntime processruntime.Options
	Args           []string
	Context        context.Context `json:"-" yaml:"-"`
	WorkDir        string
	Verbose        bool
	NoCache        bool
	Spec           *api.Spec
	Runtime        *api.ResolveSpecOptions                               `json:"-" yaml:"-"`
	ResolveSpec    func(context.Context) (api.ResolveSpecOptions, error) `json:"-" yaml:"-"`
	Evaluator      *CELEvaluator                                         `json:"-" yaml:"-"`
	ExtraArgs      map[string]interface{}
	ExecutablePath string // Path to the current executable
	UpdateGolden   bool   // When true, mismatched @file expectations are rewritten with actual output instead of failing
	Profile        bool
	SQLProfile     bool
	Progress       func(done, total int) error `json:"-" yaml:"-"`

	// Setup is the environment prepared for the markdown file this fixture came
	// from, or nil when that file declared no `setup:`. It lives here rather than
	// on FixtureTest because RunOptions is the runtime-environment carrier and
	// FixtureTest is serialized into results — the resolved environment holds
	// credentials.
	Setup *PreparedSetup

	// Recorder is the runtime state a fixture type needs to run under the
	// recorders watching its file — nil when nothing is recording it. Here for
	// the same reason as Setup: a proxy address is runtime state, not part of the
	// fixture's declaration.
	Recorder *RecorderContext `json:"-" yaml:"-"`

	// Record is the recorder declaration in force for this fixture — its own
	// `record:` or the run-wide default — so RunNode can refuse a recorder that
	// has no implementation before the fixture runs. Nil records nothing.
	Record *record.Spec

	// DaemonPort is the port the file's `daemon:` was started on, exposed to
	// exec fixtures as the `{{.port}}` template variable. Zero means no daemon.
	DaemonPort int

	// Changed are the workdir-relative files the change under verification
	// touched, when the caller scoped verification to them (captain's
	// `workflow.verify.scope: changed`). Nil or empty verifies the whole tree.
	//
	// Every node kind honours it: a `yaml test` step narrows to the packages
	// those files affect, a `yaml lint` step lints only them, an `ai` step names
	// them to the grader, and an exec fixture's expectation can read them as the
	// CEL variable `changed_files`. RunNode records the list under
	// Metadata["changed_files"] on every result so a verdict shows the scope it
	// was reached under.
	Changed []string
	// PreviousResults are completed fixture steps available to a later AI check.
	PreviousResults []FixtureResult

	// Scope is the fixture-scoped environment, step outputs and teardown stack
	// this node runs in: one per fixture file (Runner) or per RunNodes call.
	// RunNode given none runs the node as a fixture of its own.
	Scope *StepScope `json:"-" yaml:"-"`
}

// OutputMode controls when captured command output is shown in rendered
// fixture results.
type OutputMode string

const (
	OutputNever     OutputMode = "Never"
	OutputOnFailure OutputMode = "OnFailure"
	OutputAlways    OutputMode = "Always"
)

func (o OutputMode) ShouldShow(failed bool) bool {
	switch o {
	case OutputAlways, "true":
		return true
	case OutputNever, "false":
		return false
	default:
		return failed
	}
}

func ParseOutputMode(value string) OutputMode {
	switch value {
	case "", "OnFailure", "onfailure", "on-failure":
		return OutputOnFailure
	case "Always", "always", "true":
		return OutputAlways
	case "Never", "never", "false":
		return OutputNever
	default:
		return OutputMode(value)
	}
}

// DisplayOptions controls which fixture results and captured fields are
// included in human-facing command output.
type DisplayOptions struct {
	ShowPassed  bool
	ShowCommand bool
	ShowCELVars bool
	ShowStdout  OutputMode
	ShowStderr  OutputMode
}

// DisplayOptionsForVerbosity layers fixture-specific display affordances on
// top of the global logging verbosity count.
func DisplayOptionsForVerbosity(level int, base DisplayOptions, stdoutExplicit, stderrExplicit bool) DisplayOptions {
	if base.ShowStdout == "" {
		base.ShowStdout = OutputOnFailure
	}
	if base.ShowStderr == "" {
		base.ShowStderr = OutputOnFailure
	}
	if level >= 1 {
		base.ShowPassed = true
	}
	if level >= 2 {
		base.ShowCommand = true
	}
	if level >= 3 {
		base.ShowCELVars = true
		if !stdoutExplicit {
			base.ShowStdout = OutputAlways
		}
		if !stderrExplicit {
			base.ShowStderr = OutputAlways
		}
	}
	return base
}

// FixtureGroup represents a logical grouping of fixture tests with summary statistics.
type FixtureGroup struct {
	Name     string
	Children []FixtureResult `json:"children" pretty:"type:tree"`
	Tests    []FixtureNode   `json:"tests"`
	Summary  Stats           `json:"summary"`
}
