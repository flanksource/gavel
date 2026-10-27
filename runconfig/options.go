package runconfig

// Test and Lint hold configured command defaults. Pointer and non-nil slice
// fields distinguish an explicit reset from an omitted option.
type Test struct {
	StartingPaths []string `json:"paths,omitempty" yaml:"paths,omitempty" scope:"run" help:"Package paths to discover." path:"true"`
	ExtraArgs     []string `json:"extra-args,omitempty" yaml:"extra-args,omitempty" help:"Native runner arguments."`
	Tags          []string `json:"tags,omitempty" yaml:"tags,omitempty" help:"Build tags."`
	Ignore        []string `json:"ignore,omitempty" yaml:"ignore,omitempty" help:"Paths excluded from discovery." path:"true"`
	ShowPassed    *bool    `json:"show-passed,omitempty" yaml:"show-passed,omitempty" scope:"run" help:"Include passing tests in terminal output."`
	ShowStdout    *string  `json:"show-stdout,omitempty" yaml:"show-stdout,omitempty" scope:"run" enum:"Never,OnFailure,Always,true,false" help:"When to display stdout."`
	ShowStderr    *string  `json:"show-stderr,omitempty" yaml:"show-stderr,omitempty" scope:"run" enum:"Never,OnFailure,Always,true,false" help:"When to display stderr."`
	WorkDir       *string  `json:"work-dir,omitempty" yaml:"work-dir,omitempty" scope:"run" path:"true" help:"Execution working directory."`
	DryRun        *bool    `json:"dry-run,omitempty" yaml:"dry-run,omitempty" scope:"run" help:"Preview commands without executing."`
	Recursive     *bool    `json:"recursive,omitempty" yaml:"recursive,omitempty" scope:"run" help:"Discover packages recursively."`
	Nodes         *int     `json:"nodes,omitempty" yaml:"nodes,omitempty" min:"-1" help:"Parallel Ginkgo nodes; -1 selects automatically."`
	Concurrency   *int     `json:"concurrency,omitempty" yaml:"concurrency,omitempty" min:"0" scope:"run" help:"Maximum package concurrency; zero selects automatically."`
	UI            *bool    `json:"ui,omitempty" yaml:"ui,omitempty" scope:"run" help:"Open the test dashboard."`
	Addr          *string  `json:"addr,omitempty" yaml:"addr,omitempty" scope:"run" help:"Dashboard listen address."`
	Dev           *bool    `json:"dev,omitempty" yaml:"dev,omitempty" scope:"run" help:"Use the dashboard development server."`
	Diagnostics   *bool    `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty" scope:"run" help:"Capture process diagnostics."`
	ProcessFacets *bool    `json:"process-facets,omitempty" yaml:"process-facets,omitempty" scope:"run" help:"Capture process files, connections, and environment."`
	SkipHooks     *bool    `json:"skip-hooks,omitempty" yaml:"skip-hooks,omitempty" scope:"run" help:"Skip pre/post hooks."`
	AutoStop      *string  `json:"auto-stop,omitempty" yaml:"auto-stop,omitempty" duration:"true" scope:"run" help:"Detached dashboard lifetime."`
	IdleTimeout   *string  `json:"idle-timeout,omitempty" yaml:"idle-timeout,omitempty" duration:"true" scope:"run" help:"Detached dashboard idle deadline."`
	Timeout       *string  `json:"timeout,omitempty" yaml:"timeout,omitempty" duration:"true" scope:"run" help:"Whole-run deadline."`
	LintTimeout   *string  `json:"lint-timeout,omitempty" yaml:"lint-timeout,omitempty" duration:"true" scope:"run" help:"Per-linter deadline."`
	TestTimeout   *string  `json:"test-timeout,omitempty" yaml:"test-timeout,omitempty" duration:"true" help:"Per-package deadline."`
	Lint          *bool    `json:"lint,omitempty" yaml:"lint,omitempty" scope:"run" help:"Run lint with tests."`
	Cache         *bool    `json:"cache,omitempty" yaml:"cache,omitempty" scope:"run" help:"Reuse passing package results."`
	Changed       *bool    `json:"changed,omitempty" yaml:"changed,omitempty" scope:"run" help:"Select tests affected by changes."`
	Since         *string  `json:"since,omitempty" yaml:"since,omitempty" scope:"run" help:"Select changes since a Git reference."`
	ChangedFiles  []string `json:"changed-files,omitempty" yaml:"changed-files,omitempty" scope:"run" path:"true" help:"Explicit changed paths."`
	Bench         *string  `json:"bench,omitempty" yaml:"bench,omitempty" help:"Benchmark name pattern."`
	Fixtures      *bool    `json:"fixtures,omitempty" yaml:"fixtures,omitempty" scope:"run" help:"Discover fixture files."`
	FixtureFiles  []string `json:"fixture-files,omitempty" yaml:"fixture-files,omitempty" scope:"run" path:"true" help:"Fixture discovery globs."`
	Frameworks    []string `json:"framework,omitempty" yaml:"framework,omitempty" scope:"run" help:"Selected test frameworks."`
	Baseline      *string  `json:"baseline,omitempty" yaml:"baseline,omitempty" scope:"run" path:"true" help:"Baseline snapshot."`
	Failed        *string  `json:"failed,omitempty" yaml:"failed,omitempty" scope:"run" path:"true" help:"Snapshot used to select failed tests."`
	PR            *string  `json:"pr,omitempty" yaml:"pr,omitempty" scope:"run" help:"Pull request whose failed tests are selected."`
}

type Lint struct {
	Linters       []string `json:"linters,omitempty" yaml:"linters,omitempty" scope:"run" help:"Selected linters."`
	Ignore        []string `json:"ignore,omitempty" yaml:"ignore,omitempty" path:"true" help:"Excluded file globs."`
	Generated     *bool    `json:"generated,omitempty" yaml:"generated,omitempty" help:"Include generated files."`
	Tags          []string `json:"tags,omitempty" yaml:"tags,omitempty" help:"Build tags."`
	Exclude       []string `json:"exclude,omitempty" yaml:"exclude,omitempty" help:"Excluded native rule identifiers."`
	Triage        *bool    `json:"triage,omitempty" yaml:"triage,omitempty" scope:"run" help:"Interactively triage findings."`
	Fix           *bool    `json:"fix,omitempty" yaml:"fix,omitempty" help:"Enable native fixes."`
	NoCache       *bool    `json:"no-lint-cache,omitempty" yaml:"no-lint-cache,omitempty" help:"Disable lint result caching."`
	Timeout       *string  `json:"timeout,omitempty" yaml:"timeout,omitempty" duration:"true" help:"Per-linter deadline."`
	WorkDir       *string  `json:"work-dir,omitempty" yaml:"work-dir,omitempty" scope:"run" path:"true" help:"Execution working directory."`
	Changed       *bool    `json:"changed,omitempty" yaml:"changed,omitempty" scope:"run" help:"Select changed files."`
	Since         *string  `json:"since,omitempty" yaml:"since,omitempty" scope:"run" help:"Select changes since a Git reference."`
	ChangedFiles  []string `json:"changed-files,omitempty" yaml:"changed-files,omitempty" scope:"run" path:"true" help:"Explicit changed paths."`
	UI            *bool    `json:"ui,omitempty" yaml:"ui,omitempty" scope:"run" help:"Open the lint dashboard."`
	Addr          *string  `json:"addr,omitempty" yaml:"addr,omitempty" scope:"run" help:"Dashboard listen address."`
	DryRun        *bool    `json:"dry-run,omitempty" yaml:"dry-run,omitempty" scope:"run" help:"Preview commands without executing."`
	Baseline      *string  `json:"baseline,omitempty" yaml:"baseline,omitempty" scope:"run" path:"true" help:"Baseline snapshot."`
	Failed        *string  `json:"failed,omitempty" yaml:"failed,omitempty" scope:"run" path:"true" help:"Snapshot used to select failed lint."`
	PR            *string  `json:"pr,omitempty" yaml:"pr,omitempty" scope:"run" help:"Pull request whose failed lint is selected."`
	Summary       *bool    `json:"summary,omitempty" yaml:"summary,omitempty" scope:"run" help:"Summarize findings by rule."`
	SummaryLimit  *int     `json:"summary-limit,omitempty" yaml:"summary-limit,omitempty" min:"0" scope:"run" help:"Maximum examples per rule."`
	Files         []string `json:"files,omitempty" yaml:"files,omitempty" scope:"run" path:"true" help:"Selected file paths."`
	AIFix         *bool    `json:"ai-fix,omitempty" yaml:"ai-fix,omitempty" scope:"run" help:"Repair findings with lint.fix."`
	AIFixMaxIters *int     `json:"ai-fix-max-iterations,omitempty" yaml:"ai-fix-max-iterations,omitempty" min:"1" scope:"run" help:"Maximum repair iterations."`
	Yes           *bool    `json:"yes,omitempty" yaml:"yes,omitempty" scope:"run" help:"Accept AI repair without prompting."`
}
