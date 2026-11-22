package testrunner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	osexec "os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	commonsCtx "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/duration"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/baseline"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/git/changeset"
	"github.com/flanksource/gavel/processruntime"
	"github.com/flanksource/gavel/slots"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
	"github.com/flanksource/gavel/traces"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/gavel/workspace"
	"github.com/samber/lo"
)

var (
	exitStatusRe = regexp.MustCompile(`(?m)^exit status \d+\s*$`)
	// ErrNoTestsExecuted reports a completed non-dry run with no test results.
	ErrNoTestsExecuted = errors.New("no tests were executed")
)

func stripExitStatus(s string) string {
	return strings.TrimSpace(exitStatusRe.ReplaceAllString(s, ""))
}

// dedupeGinkgoBootstraps removes GoTest wrapper tests (the TestXxx function that
// calls ginkgo.RunSpecs) from the final results when the same package was also
// run through the Ginkgo runner and produced at least one real spec. Packages
// that have no Ginkgo coverage keep the wrapper so the UI still shows a
// pass/fail signal for the whole suite.
func dedupeGinkgoBootstraps(results parsers.TestSuiteResults) {
	ginkgoPackages := make(map[string]bool)
	for _, tr := range results {
		if tr.Framework != parsers.Ginkgo {
			continue
		}
		for _, t := range tr.Tests {
			if t.PackagePath != "" {
				ginkgoPackages[t.PackagePath] = true
			}
		}
	}
	if len(ginkgoPackages) == 0 {
		return
	}
	for i := range results {
		if results[i].Framework != parsers.GoTest {
			continue
		}
		kept := results[i].Tests[:0]
		for _, t := range results[i].Tests {
			if t.IsGinkgoBootstrap && ginkgoPackages[t.PackagePath] {
				continue
			}
			kept = append(kept, t)
		}
		results[i].Tests = kept
	}
}

// OutputMode controls when stdout/stderr are displayed in test output.
type OutputMode string

const (
	OutputNever     OutputMode = "Never"
	OutputOnFailure OutputMode = "OnFailure"
	OutputAlways    OutputMode = "Always"
)

// ShouldShow returns true if output should be shown based on the mode and test failure status.
func (o OutputMode) ShouldShow(failed bool) bool {
	switch o {
	case OutputAlways, "true":
		return true
	case OutputNever, "false":
		return false
	default: // OnFailure
		return failed
	}
}

// TestOrchestrator orchestrates test execution.
type TestOrchestrator struct {
	RunOptions
	configurationScoped bool
	registry            *Registry
	streamer            *TestStreamer
	selector            *selectorContext
}

// RunOptions configures test execution behavior.
//
// The yaml tags mirror the CLI flag names so a `yaml test` fixture code block
// unmarshals directly onto this struct (see fixtures/types/runstep.go).
type RunOptions struct {
	runtimeConfigDir         string
	Runtime                  processruntime.Options           `json:"runtime,omitempty" yaml:"runtime,omitempty" flag:"-"`
	StartingPaths            []string                         `json:"starting_paths,omitempty" yaml:"paths,omitempty" args:"true"`                                                                                                                                            // Package paths to test (e.g., ["./pkg/testrunner"]). If empty, all packages are discovered.
	ExtraArgs                []string                         `json:"extra_args,omitempty" yaml:"extra-args,omitempty" flag:"extra-args" help:"Additional arguments passed to every selected test runner"`                                                                    // Additional arguments broadcast to every selected test runner.
	Tags                     []string                         `json:"tags,omitempty" yaml:"tags,omitempty" flag:"tags" help:"Go build tags for Go and Ginkgo tests (comma-separated or repeated)"`                                                                            // Go build tags mapped to each supported framework.
	ShowPassed               bool                             `json:"show_passed,omitempty" yaml:"show-passed,omitempty" flag:"show-passed" help:"Include passing tests in the report"`                                                                                       // Whether to show passed tests in output
	Ignore                   []string                         `json:"ignore,omitempty" yaml:"ignore,omitempty" flag:"ignore" help:"Glob patterns for packages and paths to exclude from discovery"`                                                                           // Glob patterns for test packages/paths to exclude from discovery.
	ShowStdout               OutputMode                       `json:"show_stdout,omitempty" yaml:"show-stdout,omitempty" flag:"show-stdout" default:"OnFailure" help:"When to show test stdout (Never, OnFailure, Always)"`                                                   // When to show stdout: false|Never, OnFailure (default), true|Always
	ShowStderr               OutputMode                       `json:"show_stderr,omitempty" yaml:"show-stderr,omitempty" flag:"show-stderr" default:"OnFailure" help:"When to show test stderr (Never, OnFailure, Always)"`                                                   // When to show stderr: false|Never, OnFailure (default), true|Always
	WorkDir                  string                           `json:"work_dir,omitempty" yaml:"work-dir,omitempty" flag:"work-dir" help:"Run tests from this directory instead of the invocation directory"`                                                                  // Working directory to run tests in
	DryRun                   bool                             `json:"dry_run,omitempty" yaml:"dry-run,omitempty" flag:"dry-run" help:"Show the planned test commands without running tests"`                                                                                  // Show what tests would be executed without running them
	Recursive                bool                             `json:"recursive,omitempty" yaml:"recursive,omitempty" flag:"recursive" default:"true" help:"Discover test packages below each starting path; --recursive=false tests only the named directories"`              // Recursively discover test packages in subdirectories
	Nodes                    int                              `json:"nodes,omitempty" yaml:"nodes,omitempty" flag:"nodes" short:"p" help:"Node count accepted by the unified Go test runner (0 or 1)"`                                                                        // Go and Ginkgo share the same runner.
	Concurrency              int                              `json:"concurrency,omitempty" yaml:"concurrency,omitempty" flag:"concurrency" help:"Maximum concurrent test package subprocesses (0 = automatic)"`                                                              // Max test package subprocesses to run at once. 0 = auto-bounded default.
	UI                       bool                             `json:"ui,omitempty" yaml:"-" flag:"ui" help:"Launch a live browser dashboard; also records the run's process tree, samples, stacks and facets into .gavel/traces/<run>"`                                       // Launch browser with real-time task progress dashboard
	Addr                     string                           `json:"addr,omitempty" yaml:"-" flag:"addr" default:"0.0.0.0" help:"Interface for the --ui server (use localhost for local-only access)"`                                                                       // Interface to bind --ui HTTP server. Defaults to 0.0.0.0 (all interfaces); set localhost to restrict to this machine.
	Dev                      bool                             `json:"dev,omitempty" yaml:"-" flag:"dev" help:"With --ui, serve the dashboard from the testrunner/ui Vite dev server, with clicky-ui from the sibling checkout's source (requires a source checkout)"`         // Serve the --ui page through the testrunner/ui Vite dev server instead of the embedded bundle.
	Diagnostics              bool                             `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty" flag:"diagnostics" help:"Record the run's process tree, CPU/RSS samples and stack dumps into .gavel/traces/<run> without --ui"`                      // Capture the run's process tree, samples and stacks into its trace store (.gavel/traces/<run>) without --ui; the periodic, timeout, package-timeout and final goroutine dumps are stacks checkpoints of it.
	ProcessFacets            bool                             `json:"process_facets,omitempty" yaml:"process-facets,omitempty" flag:"process-facets" help:"Also record each process's open files, sockets, ports, connections and environment (needs --ui or --diagnostics)"` // Read every process's open files, sockets, ports and connections each sample, and its environment, into the run's process capture, as --ui does. Needs --ui or --diagnostics.
	SkipHooks                bool                             `json:"skip_hooks,omitempty" yaml:"skip-hooks,omitempty" flag:"skip-hooks" help:"Skip .gavel.yaml pre/post hooks (skipped by default outside CI)"`                                                              // When true, .gavel.yaml pre/post hooks do not run. Default is computed in runTests from $CI: skip when unset, run when set.
	AutoStop                 time.Duration                    `json:"auto_stop,omitempty" yaml:"-"`                                                                                                                                                                           // Hard wall-clock deadline for the detached UI child. Passed through when --detach is set. 0 = use default (30m). Flag wired imperatively from cmd/gavel/test.go because clicky doesn't bind time.Duration.
	IdleTimeout              time.Duration                    `json:"idle_timeout,omitempty" yaml:"-"`                                                                                                                                                                        // Idle deadline for the detached UI child; resets on every HTTP request. 0 = use default (10m). Only meaningful with --detach.
	Timeout                  time.Duration                    `json:"timeout,omitempty" yaml:"timeout,omitempty"`                                                                                                                                                             // Global wall-clock deadline for the entire test+lint run. 0 = default 20m. Cancels every in-flight subprocess when it fires.
	LintTimeout              time.Duration                    `json:"lint_timeout,omitempty" yaml:"lint-timeout,omitempty"`                                                                                                                                                   // Per-linter subprocess deadline. 0 = the linter default (5m). Forwarded to executeLinters when --lint is set.
	TestTimeout              time.Duration                    `json:"test_timeout,omitempty" yaml:"test-timeout,omitempty"`                                                                                                                                                   // Per-test-package subprocess deadline. 0 = none (test.testTimeout still reaches each package through its runtime).
	Dirs                     workspace.Dirs                   `json:"-" yaml:"-"`                                                                                                                                                                                             // The run's resolved directories; .gavel state lives under Dirs.StateRoot.
	Context                  context.Context                  `json:"-" yaml:"-"`                                                                                                                                                                                             // Parent context for the run. Nil = context.Background(). Global --timeout is applied on top.
	Lint                     bool                             `json:"lint,omitempty" yaml:"lint,omitempty" flag:"lint" help:"Run linters alongside tests"`                                                                                                                    // Run linters in parallel with tests
	Cache                    bool                             `json:"cache,omitempty" yaml:"cache,omitempty" flag:"cache" help:"Skip unchanged packages that passed the previous run"`                                                                                        // Skip packages whose content fingerprint matches the last passing run
	Changed                  bool                             `json:"changed,omitempty" yaml:"changed,omitempty" flag:"changed" help:"Run tests affected by local edits, or by the runner changeset in CI"`
	Since                    string                           `json:"since,omitempty" yaml:"since,omitempty" flag:"since" help:"Select changes since a ref merge-base plus local edits; HEAD = local edits, staged = staged files"`
	ChangedFiles             []string                         `json:"changed_files,omitempty" yaml:"changed-files,omitempty" flag:"changed-files" help:"Select tests affected by these files, combined with --changed or --since when supplied"` // Only run packages affected by these workdir-relative files (no git query); unions with --changed/--since when either is set
	Bench                    string                           `json:"bench,omitempty" yaml:"bench,omitempty" flag:"bench" help:"Run Go benchmarks matching this regex (. runs all)"`                                                             // Run Go benchmarks matching this regex ("." or "true" runs all). Auto-enabled for packages containing only Benchmark* funcs.
	Fixtures                 bool                             `json:"fixtures,omitempty" yaml:"fixtures,omitempty" flag:"fixtures" help:"Discover and run fixtures alongside tests (can also be enabled in .gavel.yaml)"`                        // Discover and run fixture files. Off by default; can also be enabled via .gavel.yaml fixtures.enabled
	FixtureFiles             []string                         `json:"fixture_files,omitempty" yaml:"fixture-files,omitempty" flag:"fixture-files" help:"Glob patterns for fixture discovery, overriding .gavel.yaml fixtures.files"`             // Globs for fixture discovery. Overrides .gavel.yaml fixtures.files. Default: **/*.fixture.{md,yaml,yml,json}
	FixturesOnly             bool                             `json:"-" yaml:"-"`                                                                                                                                                                // Run only the supplied fixture runner without discovering ordinary test frameworks.
	FixtureRunner            *fixtureOptions                  `json:"-" yaml:"-"`                                                                                                                                                                // Direct fixture command options; nil uses normal fixture discovery.
	TraceHost                *traces.Host                     `json:"-" yaml:"-"`                                                                                                                                                                // The run's trace store, shared by the fixtures it runs; nil lets a fixture runner open its own when a fixture arms a trace.
	ProcessTrace             *traces.ProcessTrace             `json:"-" yaml:"-"`                                                                                                                                                                // The run's process capture: attempts read their process's peaks from its samples, and a package that times out gets its Go test binary's goroutine dump checkpointed into its stacks. Nil records the dumps on the attempts only.
	SigquitDumper            *SigquitDumper                   `json:"-" yaml:"-"`                                                                                                                                                                // The process capture's sigquit stack transport, which watches the go test batches for the Go test binaries it can dump. Nil watches none.
	FixtureBenchmarkProgress func(fixtures.ExecutionSnapshot) `json:"-" yaml:"-"`
	FixtureBenchmarkResult   func(*fixtures.BenchmarkReport)  `json:"-" yaml:"-"`
	FixturePrompter          fixtures.Prompter                `json:"-" yaml:"-"`                                                                                                                                   // Answers the fixtures' form steps (the --ui server's prompts); overrides the fixture runner's own.
	FixtureRunID             string                           `json:"-" yaml:"-"`                                                                                                                                   // The prompt owner FixturePrompter files the fixtures' forms under.
	FixtureParams            map[string]any                   `json:"-" yaml:"-"`                                                                                                                                   // Pre-supplied values for the fixtures' params: forms (a UI run request's params); overrides the fixture runner's own.
	FixtureResume            *fixtures.Resume                 `json:"-" yaml:"-"`                                                                                                                                   // Restarts the fixture run at a step, reusing the earlier steps' results (--from-step, a UI "retry from here" rerun).
	FixtureCommand           *fixtures.RerunBase              `json:"-" yaml:"-"`                                                                                                                                   // The command a copied rerun of the run's fixtures starts with.
	Frameworks               []string                         `json:"frameworks,omitempty" yaml:"framework,omitempty" flag:"framework" help:"Select one or more frameworks (go, ginkgo, jest, vitest, playwright)"` // Restrict execution to these frameworks (e.g. jest, vitest, playwright, go, ginkgo). Empty = run every detected framework. Unknown names hard-fail.
	Baseline                 string                           `json:"baseline,omitempty" yaml:"baseline,omitempty" flag:"baseline" help:"Compare against a saved results snapshot and report new failures"`         // Path to previous results JSON; only report NEW failures not in baseline
	Failed                   string                           `json:"failed,omitempty" yaml:"failed,omitempty" flag:"failed"`                                                                                       // Path to previous results JSON; re-run only failed tests
	PR                       string                           `json:"-" yaml:"-" flag:"pr"`                                                                                                                         // PR reference (12, #12, owner/repo#12, or a PR URL); re-run only the tests that failed in that PR's gavel CI artifacts. CLI-only: the command resolves it into Failed before the run starts.
	Updates                  chan<- []parsers.Test            `json:"-" yaml:"-"`                                                                                                                                   // Channel for streaming test result updates to UI
	OutputTee                io.Writer                        `json:"-" yaml:"-"`                                                                                                                                   // Optional writer that receives a copy of raw process stdout/stderr
	RunKind                  string                           `json:"run_kind,omitempty" yaml:"run-kind,omitempty"`                                                                                                 // "initial" (default) or "rerun" — tagged onto each TestAttempt produced
	SummaryOut               *parsers.TestSummary             `json:"-" yaml:"-"`                                                                                                                                   // If non-nil, the runner writes the aggregate pass/fail/skip/total/duration counts here before returning, so CLI callers can print an end-of-run summary even when the run errors mid-way and returns a partial tree.
	ResultsOut               *[]parsers.Test                  `json:"-" yaml:"-"`                                                                                                                                   // If non-nil, the runner appends every test it parsed here, flat and before the display tree is built, so exporters see each test's own File, Package and Attempts. Filled even when the run errors mid-way.
	CoverDir                 string                           `json:"-" yaml:"-"`                                                                                                                                   // If set, every go test batch writes its -coverprofile into this directory, one batch-<n>.out per batch.
	UIRExport                string                           `json:"uir_export,omitempty" yaml:"-" flag:"uir-export" help:"Write the run's results, coverage and inputs digest as a uir tests import (v1 normally; v2 with --uir-skip)"`
	UIRTestCoverage          bool                             `json:"uir_test_coverage,omitempty" yaml:"-" flag:"uir-test-coverage" help:"With --uir-export, also rerun each Go test and Ginkgo spec alone to record its own coverage (slow)"`
	UIRSkip                  bool                             `json:"uir_skip,omitempty" yaml:"-" flag:"uir-skip" help:"Opt in to UIR-proved ordinary Go test skipping; requires --framework go, --uir-export and --uir-input-context"`
	UIRBinary                string                           `json:"uir_binary,omitempty" yaml:"-" flag:"uir-binary" default:"uir" help:"UIR executable used by --uir-skip"`
	UIRInputContext          string                           `json:"uir_input_context,omitempty" yaml:"-" flag:"uir-input-context" help:"Identity of declared service, environment and other runtime inputs for --uir-skip"`
	UIRSkippable             map[string][]string              `json:"-" yaml:"-"`
	UIRRunAll                bool                             `json:"-" yaml:"-"`
	UIRAllowEmpty            bool                             `json:"-" yaml:"-"`

	PassThroughArgs []string `json:"pass_through_args,omitempty" yaml:"-"`                                                                                                                                                           // CLI-only arguments supplied after --, parsed by every runner into the shared test options; arguments no runner recognises require one framework family.
	Focus           string   `json:"focus,omitempty" yaml:"focus,omitempty" flag:"focus" help:"Run only tests whose name matches this regexp, in every framework (go test -run, Ginkgo --focus, vitest/jest -t, playwright --grep)"` // Common focus regexp mapped to each framework's name filter.

	Selection    *changeset.Selection `flag:"-" json:"-" yaml:"-"`
	OptionFields map[string]bool      `json:"-" yaml:"-" flag:"-"`
}

type fixtureOptions = fixtures.RunnerOptions

// withFixtureForms hands runner the run's form prompter, prompt owner and
// pre-supplied params, where the run has them.
func (opts RunOptions) withFixtureForms(runner fixtures.RunnerOptions) fixtures.RunnerOptions {
	if opts.FixturePrompter != nil {
		runner.Prompter, runner.RunID = opts.FixturePrompter, opts.FixtureRunID
	}
	if opts.FixtureParams != nil {
		runner.Params = opts.FixtureParams
	}
	if opts.FixtureResume != nil {
		runner.Resume = opts.FixtureResume
	}
	return runner
}

func (opts RunOptions) Pretty() api.Text {
	text := clicky.Text("")
	if opts.WorkDir != "" {
		text = text.Space().Append("WorkDir: ", "text-muted").Append(opts.WorkDir, "text-blue-500")
	}
	if len(opts.StartingPaths) > 0 {
		text = text.Space().Append("StartingPaths: ", "text-muted").Append(clicky.CompactList(opts.StartingPaths), "text-blue-500")
	}
	if len(opts.ExtraArgs) > 0 {
		text = text.Space().Append("ExtraArgs: ", "text-muted").Append(clicky.CompactList(opts.ExtraArgs), "text-blue-500")
	}
	if len(opts.Tags) > 0 {
		text = text.Space().Append("Tags: ", "text-muted").Append(clicky.CompactList(opts.Tags), "text-blue-500")
	}
	if len(opts.PassThroughArgs) > 0 {
		text = text.Space().Append("PassThroughArgs: ", "text-muted").Append(clicky.CompactList(opts.PassThroughArgs), "text-blue-500")
	}
	if len(opts.Ignore) > 0 {
		text = text.Space().Append("Ignore: ", "text-muted").Append(clicky.CompactList(opts.Ignore), "text-blue-500")
	}
	if opts.ShowPassed {
		text = text.Space().Append("ShowPassed: ", "text-muted").Append(icons.Check, "text-green-500")
	}
	if opts.ShowStdout != "" && opts.ShowStdout != OutputOnFailure {
		text = text.Space().Append("ShowStdout: ", "text-muted").Append(string(opts.ShowStdout), "text-blue-500")
	}
	if opts.ShowStderr != "" && opts.ShowStderr != OutputOnFailure {
		text = text.Space().Append("ShowStderr: ", "text-muted").Append(string(opts.ShowStderr), "text-blue-500")
	}
	if opts.DryRun {
		text = text.NewLine().Append("DryRun: ", "text-muted").Append(icons.Check, "text-green-500")
	}
	if opts.Cache {
		text = text.Space().Append("Cache: ", "text-muted").Append(icons.Check, "text-green-500")
	}
	if opts.Changed {
		text = text.Space().Append("Changed: ", "text-muted").Append(icons.Check, "text-green-500")
	}
	if opts.Since != "" {
		text = text.Space().Append("Since: ", "text-muted").Append(opts.Since, "text-blue-500")
	}
	if opts.Fixtures {
		text = text.Space().Append("Fixtures: ", "text-muted").Append(icons.Check, "text-green-500")
	}
	if opts.Diagnostics {
		text = text.Space().Append("Diagnostics: ", "text-muted").Append(icons.Check, "text-green-500")
	}
	if opts.ProcessFacets {
		text = text.Space().Append("ProcessFacets: ", "text-muted").Append(icons.Check, "text-green-500")
	}
	if len(opts.FixtureFiles) > 0 {
		text = text.Space().Append("FixtureFiles: ", "text-muted").Append(clicky.CompactList(opts.FixtureFiles), "text-blue-500")
	}
	if opts.Timeout > 0 {
		text = text.Space().Append("Timeout: ", "text-muted").Append(opts.Timeout.String(), "text-blue-500")
	}
	if opts.LintTimeout > 0 {
		text = text.Space().Append("LintTimeout: ", "text-muted").Append(opts.LintTimeout.String(), "text-blue-500")
	}
	if opts.TestTimeout > 0 {
		text = text.Space().Append("TestTimeout: ", "text-muted").Append(opts.TestTimeout.String(), "text-blue-500")
	}
	if opts.Concurrency > 0 {
		text = text.Space().Append("Concurrency: ", "text-muted").Append(fmt.Sprintf("%d", opts.Concurrency), "text-blue-500")
	}
	return text
}

// hasChangeSelector reports whether any flag in opts narrows execution to a
// subset of packages via the change graph.
func (opts RunOptions) hasChangeSelector() bool {
	return opts.ChangeOptions().Active() || opts.Selection.Active()
}

func (r RunOptions) Help() api.Textable {
	return clicky.Text(`Discover and run Go, Ginkgo, Jest, Vitest and Playwright tests.

Choose a framework subcommand or use --framework to limit discovery. Enable
fixture discovery with --fixtures. Use --lint to run linters alongside tests.

Positional args restrict the run to those package paths (recursive by default,
--recursive=false for only the named directories). A test file runs just that
file: a Go test file its Test functions, a Ginkgo spec file its specs, a
vitest/jest/playwright file through its nearest package.json. Each path runs
from its project root (git root or nearest go.mod).

--focus PATTERN runs only the tests whose name matches, in every framework.
After "--", every runner parses the arguments: the focus spellings of go test
(-run), Ginkgo (--focus, --ginkgo.focus) and vitest/jest (-t,
--testNamePattern) are accepted anywhere and mean the same, as do the skip
(-skip, --ginkgo.skip) and fail-fast (-failfast, --fail-fast, --bail) flags.
go test and Ginkgo count as one framework, so -race or -count reach both.
Arguments no runner recognises need a single framework family.

Examples:
  gavel test                                       # all frameworks
  gavel test ./testrunner ./cmd                    # only these paths
  gavel test ./testrunner/focus_test.go            # only that file's tests
  gavel test --lint                                # tests + linters in parallel
  gavel test --changed                             # only changed packages
  gavel test --failed                              # re-run last run's failures (.gavel/last.json)
  gavel test --failed=1d                           # re-run tests that failed in the last day
  gavel test --pr 12                               # re-run the tests that failed on PR #12 in CI
  gavel test --framework go,ginkgo                 # restrict frameworks
  gavel test go --tags integration ./cmd/gavel     # enable Go build tags
  gavel test --focus "TestName" ./testrunner         # focus every framework
  gavel test ./testrunner -- -race --ginkgo.focus "a|b" # Go flags + a Ginkgo focus
  gavel test ./web -- -t "renders"                   # vitest-style focus
  gavel test ginkgo . -- --label-filter "integration" # Ginkgo label filter
  gavel test --ui                                  # live browser UI (HTTP + SSE)`)
}

// packageResult contains results from running tests in a single package
type packageResult struct {
	packagePath string
	framework   Framework
	testResults parsers.TestSuiteResults
	err         error
	timedOut    bool
}

// testGroup holds a set of starting paths that share the same execution root.
type testGroup struct {
	workDir string
	paths   []string
}

// groupPathsByGitRoot partitions starting paths by execution root. Git roots
// still split runs, and nested Go modules get their own WorkDir rooted at the
// nearest containing go.mod. When a group discovers from its root (paths
// empty), nested descendant modules become additional groups.
func groupPathsByGitRoot(workDir string, startingPaths []string) ([]testGroup, error) {
	workDir, _ = filepath.Abs(workDir)

	if len(startingPaths) == 0 {
		return expandNestedModuleGroups([]testGroup{{workDir: workDir}})
	}

	workDirRoot := utils.FindGitRoot(workDir)
	if workDirRoot == "" {
		workDirRoot = workDir
	}

	groups := make(map[string][]string)
	var order []string

	for _, p := range startingPaths {
		abs := p
		if !filepath.IsAbs(p) {
			abs = filepath.Join(workDir, p)
		}
		abs, _ = filepath.Abs(abs)

		groupRoot := executionRootForPath(workDirRoot, abs)
		if _, ok := groups[groupRoot]; !ok {
			order = append(order, groupRoot)
		}

		rel, err := filepath.Rel(groupRoot, abs)
		if err != nil {
			groups[groupRoot] = append(groups[groupRoot], p)
			continue
		}
		if rel == "." {
			// Path IS the execution root — leave paths empty so the runner
			// discovers from the root without a path filter.
			continue
		}
		groups[groupRoot] = append(groups[groupRoot], "./"+filepath.ToSlash(rel))
	}

	result := make([]testGroup, 0, len(order))
	for _, root := range order {
		result = append(result, testGroup{workDir: root, paths: groups[root]})
	}
	return expandNestedModuleGroups(result)
}

func executionRootForPath(workDirRoot, abs string) string {
	groupRoot := workDirRoot
	if gitRoot := utils.FindGitRoot(abs); gitRoot != "" {
		groupRoot = gitRoot
	}
	if moduleRoot := findNearestGoModRoot(abs); moduleRoot != "" && utils.IsWithin(moduleRoot, groupRoot) {
		groupRoot = moduleRoot
	}
	return groupRoot
}

func expandNestedModuleGroups(groups []testGroup) ([]testGroup, error) {
	result := append([]testGroup(nil), groups...)
	seen := make(map[string]struct{}, len(result))
	for _, g := range result {
		seen[g.workDir] = struct{}{}
	}

	for i := 0; i < len(result); i++ {
		if len(result[i].paths) > 0 {
			continue
		}
		modules, err := findNestedGoModuleRoots(result[i].workDir)
		if err != nil {
			return nil, err
		}
		for _, moduleRoot := range modules {
			if _, ok := seen[moduleRoot]; ok {
				continue
			}
			seen[moduleRoot] = struct{}{}
			result = append(result, testGroup{workDir: moduleRoot})
		}
	}

	return result, nil
}

func findNestedGoModuleRoots(root string) ([]string, error) {
	root, _ = filepath.Abs(root)
	var modules []string

	err := utils.WalkGitIgnored(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() || path == root {
			return nil
		}

		if info, err := os.Stat(filepath.Join(path, ".git")); err == nil && info.IsDir() {
			return fs.SkipDir
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}

		if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
			modules = append(modules, path)
			return nil
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(modules)
	return lo.Uniq(modules), nil
}

func findNearestGoModRoot(dir string) string {
	dir, _ = filepath.Abs(dir)
	for {
		if info, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func Run(opts RunOptions) (any, error) {
	if err := opts.Configure(); err != nil {
		return nil, err
	}
	if opts.WorkDir == "" {
		opts.WorkDir, _ = os.Getwd()
	}
	var configDirErr error
	opts.runtimeConfigDir, configDirErr = filepath.Abs(opts.WorkDir)
	if configDirErr != nil {
		return nil, configDirErr
	}
	opts.Selection = opts.ChangeSelection()
	if opts.FixturesOnly {
		opts.WorkDir, _ = filepath.Abs(opts.WorkDir)
		result, err := runSingleRoot(opts)
		return requireExecutedTests(result, err, opts.DryRun || opts.hasChangeSelector())
	}

	// Split starting paths by execution root so each group runs with the
	// correct WorkDir. Nested Go modules get their own groups.
	groups, err := groupPathsByGitRoot(opts.WorkDir, opts.StartingPaths)
	if err != nil {
		return nil, err
	}
	var result any
	if len(groups) > 1 {
		result, err = runMultiRoot(opts, groups)
	} else if len(groups) == 1 {
		// Single-root fast path: update WorkDir/StartingPaths from the
		// (possibly rebased) group so the rest of the function sees
		// normalised values.
		opts.WorkDir = groups[0].workDir
		opts.StartingPaths = groups[0].paths
		result, err = runSingleRoot(opts)
	}
	return requireExecutedTests(result, err, opts.DryRun || opts.hasChangeSelector() || opts.UIRAllowEmpty)
}

func requireExecutedTests(result any, err error, allowEmpty bool) (any, error) {
	if err != nil || allowEmpty {
		return result, err
	}
	tests, ok := result.([]parsers.Test)
	if ok && parsers.Tests(tests).Sum().Total == 0 {
		return result, ErrNoTestsExecuted
	}
	return result, nil
}

// runMultiRoot handles the case where starting paths span multiple git roots
// by running a separate orchestration per root and merging the results.
func runMultiRoot(base RunOptions, groups []testGroup) (any, error) {
	if base.Updates != nil {
		defer close(base.Updates)
	}

	baseWorkDir, err := filepath.Abs(base.WorkDir)
	if err != nil {
		return nil, err
	}
	var allTree []parsers.Test
	var streamedTree []parsers.Test
	for _, g := range groups {
		opts := base
		opts.WorkDir = g.workDir
		opts.StartingPaths = g.paths
		if opts.UIRSkippable != nil && filepath.Clean(g.workDir) != filepath.Clean(baseWorkDir) {
			// UIRSkippable keys are relative to the planned module. A child module can have the same
			// package path, so it must run normally rather than inherit another module's evidence.
			opts.UIRSkippable = nil
			opts.UIRAllowEmpty = false
		}

		var result any
		var err error
		var streamedRoot *parsers.Test
		if base.Updates != nil {
			groupUpdates := make(chan []parsers.Test, 16)
			prefix := append([]parsers.Test(nil), streamedTree...)
			opts.Updates = groupUpdates

			done := make(chan *parsers.Test, 1)
			go func(workDir string, completed []parsers.Test) {
				var last *parsers.Test
				for batch := range groupUpdates {
					wrapped := wrapExecutionRootTree(baseWorkDir, workDir, annotateTestsWorkDir(batch, workDir))
					wrappedCopy := wrapped
					last = &wrappedCopy
					merged := append(append([]parsers.Test(nil), completed...), wrapped)
					sendTestUpdate(base.Updates, merged)
				}
				done <- last
			}(g.workDir, prefix)

			result, err = runSingleRootWithUpdateOwnership(opts, true)
			streamedRoot = <-done
		} else {
			result, err = runSingleRootWithUpdateOwnership(opts, false)
		}
		if err != nil {
			return nil, err
		}
		if streamedRoot != nil {
			streamedTree = append(streamedTree, *streamedRoot)
			sendTestUpdate(base.Updates, append([]parsers.Test(nil), streamedTree...))
		}
		if tests, ok := result.([]parsers.Test); ok && len(tests) > 0 {
			allTree = append(allTree, wrapExecutionRootTree(baseWorkDir, g.workDir, tests))
		}
	}
	return allTree, nil
}

func wrapExecutionRootTree(baseWorkDir, rootWorkDir string, tests []parsers.Test) parsers.Test {
	root := parsers.Test{
		Name:     executionRootLabel(baseWorkDir, rootWorkDir),
		WorkDir:  rootWorkDir,
		Children: tests,
	}

	summary := root.Sum()
	switch {
	case summary.Pending > 0:
		root.Pending = true
	case summary.Failed > 0:
		root.Failed = true
	case summary.Passed > 0:
		root.Passed = true
	case summary.Skipped > 0:
		root.Skipped = true
	}

	return root
}

func executionRootLabel(baseWorkDir, rootWorkDir string) string {
	if label, ok := executionRootProjectLabel(rootWorkDir); ok {
		return label
	}
	if baseWorkDir == "" {
		baseWorkDir = rootWorkDir
	}

	baseWorkDir, _ = filepath.Abs(baseWorkDir)
	rootWorkDir, _ = filepath.Abs(rootWorkDir)
	rel, err := filepath.Rel(baseWorkDir, rootWorkDir)
	if err != nil || rel == "" {
		rel = rootWorkDir
	}
	rel = filepath.ToSlash(rel)

	switch {
	case rel == ".":
		return "./"
	case strings.HasPrefix(rel, "."):
		return rel + "/"
	default:
		return "./" + rel + "/"
	}
}

func executionRootProjectLabel(rootWorkDir string) (string, bool) {
	if modulePath, ok := parsers.GoModuleName(rootWorkDir); ok {
		return modulePath, true
	}
	if packageName, ok := packageJSONName(rootWorkDir); ok {
		return packageName, true
	}
	return "", false
}

func packageJSONName(rootWorkDir string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(rootWorkDir, "package.json"))
	if err != nil {
		return "", false
	}
	var pkg struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", false
	}
	pkg.Name = strings.TrimSpace(pkg.Name)
	return pkg.Name, pkg.Name != ""
}

func sendTestUpdate(updates chan<- []parsers.Test, tree []parsers.Test) {
	if updates == nil {
		return
	}
	select {
	case updates <- tree:
	default:
	}
}

func runSingleRoot(opts RunOptions) (any, error) {
	return runSingleRootWithUpdateOwnership(opts, true)
}

func runSingleRootWithUpdateOwnership(opts RunOptions, closeUpdates bool) (any, error) {
	opts.Selection = opts.ChangeSelection()
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	selected, selectionErr := opts.Selection.Resolve(ctx, opts.WorkDir)
	if selectionErr != nil {
		return nil, selectionErr
	}
	if selected != nil && len(selected.Changes) == 0 {
		logger.Infof("%s; no applicable test work", selected.Summary())
		return []parsers.Test{}, nil
	}
	if _, err := verify.LoadGavelConfig(opts.runtimeWorkDir()); err != nil {
		return nil, err
	}
	logger.Infof("Running tests  %s", renderText(opts.Pretty()))

	var streamer *TestStreamer
	if opts.Updates != nil {
		if closeUpdates {
			streamer = NewTestStreamer(opts.Updates)
		} else {
			streamer = NewSharedTestStreamer(opts.Updates)
		}
		defer streamer.Done()
	}

	t := &TestOrchestrator{
		RunOptions: opts,
		registry:   DefaultRegistry(opts.WorkDir),
		streamer:   streamer,
	}
	var results parsers.TestSuiteResults
	var err error
	if !opts.FixturesOnly {
		results, err = t.Run()
	}
	// Populate SummaryOut from whatever completed before the error, so the
	// CLI can still print an end-of-run summary when a subprocess failed.
	if opts.SummaryOut != nil {
		*opts.SummaryOut = opts.SummaryOut.Add(results.All().Sum())
	}
	if opts.ResultsOut != nil {
		for _, result := range results {
			for _, test := range result.Tests {
				if test.WorkDir == "" {
					test.WorkDir = opts.WorkDir
				}
				*opts.ResultsOut = append(*opts.ResultsOut, test)
			}
		}
	}
	if err != nil {
		return nil, err
	}
	// Return the complete, unmutated tree. --show-passed / --show-stdout /
	// --show-stderr are presentation concerns applied at pretty-print time
	// (cmd/gavel/test.go), never here — filtering or clearing stdout/stderr at
	// the runner level would make the JSON/UI snapshot lossy.
	tree := attachFailedPackageStderr(results.Tree(), results)

	// Discover and run fixture tests, gated on the --fixtures flag or
	// fixtures.enabled in .gavel.yaml. Keep the partial tree and return direct
	// runner errors so invalid baselines and report failures reach the command.
	var fixtureErr error
	if opts.FixtureRunner != nil {
		opts.FixtureRunner.Context = opts.Context
		opts.FixtureRunner.TraceHost = opts.TraceHost
		opts.FixtureRunner.ProcessTrace = opts.ProcessTrace
		opts.FixtureRunner.ProcessRuntime = opts.FixtureRunner.ProcessRuntime.Merge(opts.fixtureRuntime())
		if opts.FixtureRunner.ExecutablePath == "" {
			// A runtime cap (test.testTimeout, --cpu, --mem) isolates the run in a
			// worker spawned from this binary, exactly as fixture discovery does.
			execPath, err := os.Executable()
			if err != nil {
				return nil, fmt.Errorf("resolve fixture worker executable: %w", err)
			}
			opts.FixtureRunner.ExecutablePath = execPath
		}
		fixtureTests, err := t.runFixtureRunner(opts.withFixtureForms(*opts.FixtureRunner), opts.runtimeWorkDir())
		if err != nil {
			fixtureErr = err
		}
		tree = append(tree, fixtureTests...)
	} else if globs := resolveFixtureGlobs(opts); globs != nil {
		fixtureTests, err := t.runDiscoveredFixtures(globs)
		if err != nil {
			fixtureErr = err
		}
		tree = append(tree, fixtureTests...)
	}

	tree = annotateTestsWorkDir(tree, opts.WorkDir)
	if opts.FixturesOnly && opts.SummaryOut != nil {
		*opts.SummaryOut = opts.SummaryOut.Add(parsers.Tests(tree).Sum())
	}

	return tree, fixtureErr
}

func annotateTestsWorkDir(tests []parsers.Test, workDir string) []parsers.Test {
	for i := range tests {
		tests[i].WorkDir = workDir
		if len(tests[i].Children) > 0 {
			tests[i].Children = annotateTestsWorkDir(tests[i].Children, workDir)
		}
	}
	return tests
}

// Run executes tests.
// Returns []parsers.TestResults with all package test results (passed, failed, skipped), or an error.
func (o *TestOrchestrator) Run() (parsers.TestSuiteResults, error) {
	cwd, err := filepath.Abs(o.WorkDir)
	if err != nil {
		return nil, err
	}
	detected, err := o.registry.DetectAll(cwd, o.testArgv())
	if err != nil {
		return nil, fmt.Errorf("failed to detect frameworks: %w", err)
	}

	if len(o.Frameworks) > 0 {
		detected, err = filterFrameworks(detected, o.Frameworks)
		if err != nil {
			return nil, err
		}
	}

	results, err := o.detectAndRun(detected)
	// Safe as Info again: clicky's task renderer installs a serializing
	// logger writer while active, so this line waits on the same mutex
	// as tick renders and can't interleave mid-frame.
	if results.All().Sum().Total == 0 {
		logger.Infof("No framework tests executed.")
	} else {
		logger.Infof("Test run completed. %d total failures.", len(results.Failed()))
	}

	if err != nil {
		return results, err
	}

	return results, nil
}

func packageIgnored(pkg string, ignore []string) bool {
	pkg = filepath.ToSlash(strings.TrimPrefix(pkg, "./"))
	for _, pattern := range ignore {
		pattern = strings.TrimSpace(filepath.ToSlash(pattern))
		if pattern == "" {
			continue
		}
		pattern = strings.TrimPrefix(pattern, "./")
		pattern = strings.TrimSuffix(pattern, "/")
		candidates := []string{pattern, pattern + "/**"}
		for _, candidate := range candidates {
			if ok, _ := doublestar.Match(candidate, pkg); ok {
				return true
			}
		}
	}
	return false
}

func (o *TestOrchestrator) packageConcurrency() int {
	if o.Concurrency > 0 {
		return o.Concurrency
	}
	if n := runtime.NumCPU(); n < 4 {
		return n
	}
	return 4
}

// filterFrameworks rejects unknown or undetected frameworks before execution.
func filterFrameworks(detected map[Framework]*runners.TestOptions, requested []string) (map[Framework]*runners.TestOptions, error) {
	kept := map[Framework]*runners.TestOptions{}
	for _, name := range requested {
		fw, err := parsers.ParseFramework(name)
		if err != nil {
			return nil, err
		}
		if detected[fw] == nil {
			return nil, fmt.Errorf("framework %q not detected in workdir", fw)
		}
		kept[fw] = detected[fw]
	}
	return kept, nil
}

// loadFailedTests reads the tests --failed names. A duration (1d, 6h) unions
// every run recorded under stateRoot that started within it; any other value
// is a results JSON path. The second result describes the source for errors.
func loadFailedTests(stateRoot, failed string) ([]parsers.Test, string, error) {
	window, err := duration.ParseDuration(failed)
	if err != nil {
		snapshot, err := baseline.LoadSnapshot(failed)
		if err != nil {
			return nil, "", fmt.Errorf("--failed: %w", err)
		}
		return snapshot.Tests, failed, nil
	}
	if time.Duration(window) <= 0 {
		return nil, "", fmt.Errorf("--failed=%s: want a positive window such as 1d or 6h", failed)
	}
	source := "runs started within " + failed
	tests, err := snapshots.TestsOfFailedRunsSince(stateRoot, time.Now().Add(-time.Duration(window)))
	if err != nil {
		return nil, "", fmt.Errorf("--failed: %s: %w", source, err)
	}
	return tests, source, nil
}

// applyFailedFilter loads the previous failures --failed names and narrows
// packagesByFramework to only packages that had failures, returning the
// de-duplicated failed test names per framework for the orchestrator to focus on.
func applyFailedFilter(packagesByFramework map[Framework][]string, stateRoot, failed string) (map[Framework][]string, map[Framework][]string, error) {
	tests, source, err := loadFailedTests(stateRoot, failed)
	if err != nil {
		return nil, nil, err
	}
	failedPkgs := baseline.ExtractFailedTestPackages(tests)
	if len(failedPkgs) == 0 {
		return nil, nil, fmt.Errorf("--failed: no failed tests found in %s", source)
	}

	filtered := make(map[Framework][]string)
	namesByFramework := make(map[Framework][]string)
	failedPackageCount := 0
	for _, packages := range failedPkgs {
		failedPackageCount += len(packages)
	}

	for fw, pkgs := range packagesByFramework {
		failedForFW, ok := failedPkgs[parsers.Framework(fw)]
		if !ok {
			continue
		}
		for _, pkg := range pkgs {
			if names, exists := failedForFW[pkg]; exists {
				filtered[fw] = append(filtered[fw], pkg)
				namesByFramework[fw] = append(namesByFramework[fw], names...)
			}
		}
	}
	if len(filtered) == 0 {
		return nil, nil, fmt.Errorf("--failed: %d failed test packages in %s did not match any detected packages", failedPackageCount, source)
	}
	for fw, names := range namesByFramework {
		sort.Strings(names)
		namesByFramework[fw] = slices.Compact(names)
	}

	logger.Infof("--failed: narrowed to %d frameworks from %s", len(filtered), source)
	return filtered, namesByFramework, nil
}

func (o *TestOrchestrator) detectAndRun(detected map[Framework]*runners.TestOptions) (parsers.TestSuiteResults, error) {
	if len(detected) == 0 {
		logger.Infof("No test frameworks detected")
		return nil, nil
	}
	if o.Nodes != 0 && o.Nodes != 1 {
		return nil, fmt.Errorf("--nodes=%d is unsupported by the unified go test runner", o.Nodes)
	}
	base, err := o.baseTestOptions()
	if err != nil {
		return nil, err
	}

	// Discover each detected framework's packages under its starting paths
	// and test files.
	packagesByFramework := make(map[Framework][]string)
	for fw, opts := range detected {
		merged, err := base.Merge(*opts)
		if err != nil {
			return nil, err
		}
		detected[fw] = &merged
		runner, ok := o.registry.Get(fw)
		if !ok {
			return nil, fmt.Errorf("no runner registered for framework %s, registry = %s", fw, renderText(o.registry.Pretty()))
		}
		packages, err := runner.DiscoverPackages(merged)
		if err != nil {
			return nil, fmt.Errorf("failed to discover packages for %s: %w", fw, err)
		}
		packagesByFramework[fw], err = o.configuredPackages(fw, packages)
		if err != nil {
			return nil, err
		}
	}

	for framework, items := range packagesByFramework {
		logger.V(4).Infof("[%s] discovered %s", framework, clicky.CompactList(items).ANSI())
	}

	optsByFramework, err := o.resolveTestOptions(detected, packagesByFramework)
	if err != nil {
		return nil, err
	}
	var frameworks []Framework
	var allPackages []string
	for fw := range optsByFramework {
		frameworks = append(frameworks, fw)
		allPackages = append(allPackages, packagesByFramework[fw]...)
	}
	sort.Slice(frameworks, func(i, j int) bool { return frameworks[i] < frameworks[j] })
	for fw := range packagesByFramework {
		if _, ok := optsByFramework[fw]; !ok {
			delete(packagesByFramework, fw)
		}
	}

	if len(allPackages) == 0 {
		if len(o.StartingPaths) > 0 {
			logger.Infof("No test packages found in starting paths %v", o.StartingPaths)
		} else {
			logger.Infof("No test packages found")
		}
		return nil, nil
	}

	if err := o.selectChanges(packagesByFramework); err != nil {
		return nil, err
	}

	cachedByFramework := map[Framework][]cacheHit{}
	if o.Cache && o.selector != nil {
		for fw, pkgs := range packagesByFramework {
			if fw == parsers.GoTest || fw == parsers.Ginkgo {
				continue
			}
			need, hits, err := o.selector.filterByRunCache(fw, pkgs)
			if err != nil {
				return nil, fmt.Errorf("run-cache filter for %s: %w", fw, err)
			}
			if len(hits) > 0 {
				logger.Infof("[%s] run-cache: %d hits, %d to run", fw, len(hits), len(need))
			}
			packagesByFramework[fw] = need
			cachedByFramework[fw] = hits
		}
	}

	// Narrow to previously-failed packages/tests when --failed is set.
	if o.Failed != "" {
		var failedNames map[Framework][]string
		packagesByFramework, failedNames, err = applyFailedFilter(packagesByFramework, o.Dirs.StateRoot, o.Failed)
		if err != nil {
			return nil, err
		}
		failedFocus := failedTestOptions(failedNames)
		for fw, names := range failedNames {
			focus, ok := failedFocus[fw]
			if !ok {
				return nil, fmt.Errorf("framework %s cannot focus previously failed tests %v", fw, names)
			}
			opts := optsByFramework[fw]
			opts.Focus = opts.Focus.Merge(focus)
			optsByFramework[fw] = opts
		}
	}

	if o.UIRSkippable != nil {
		seen := map[string]bool{}
		for _, pkg := range packagesByFramework[parsers.GoTest] {
			seen[pkg] = true
		}
		for pkg := range o.UIRSkippable {
			if !seen[pkg] {
				packagesByFramework[parsers.GoTest] = append(packagesByFramework[parsers.GoTest], pkg)
			}
		}
		sort.Strings(packagesByFramework[parsers.GoTest])
	}

	// If dry-run mode, display what would be executed and return early
	if o.DryRun {
		return o.displayDryRun(packagesByFramework, optsByFramework, o.unfocusedGoRun())
	}

	// Send pending package outline to UI before execution
	if o.streamer != nil {
		var outline []parsers.Test
		for fw, packages := range packagesByFramework {
			for _, pkg := range packages {
				outline = append(outline, parsers.Test{
					Name:        pkg,
					PackagePath: pkg,
					Framework:   fw,
					Pending:     true,
				})
			}
		}
		o.streamer.SetPackageOutline(outline)
	}

	// Create task group to orchestrate parallel package test execution.
	// Keep package fanout bounded; CI runners can OOM when every package starts
	// a Go/Ginkgo subprocess at once.
	group := task.StartGroup[packageResult]("Running tests across packages", task.WithConcurrency(o.packageConcurrency()))
	goBatches, err := o.planGoBatches(packagesByFramework, optsByFramework, o.unfocusedGoRun())
	if err != nil {
		return nil, err
	}
	for _, batch := range goBatches {
		batch := batch
		group.Add("go test", func(ctx commonsCtx.Context, t *task.Task) (packageResult, error) {
			return o.runGoBatch(ctx, t, batch)
		})
	}

	// Launch a task for each package per framework
	for _, fw := range frameworks {
		if fw == parsers.GoTest || fw == parsers.Ginkgo {
			continue
		}
		packages := packagesByFramework[fw]
		runner, ok := o.registry.Get(fw)
		if !ok {
			continue
		}

		for _, pkgPath := range packages {
			// Capture variables for closure
			pkgPath := pkgPath
			fw := fw
			runner := runner
			opts := optsByFramework[fw]

			taskName := fmt.Sprintf("%s %s", fw, pkgPath)
			group.Add(taskName, func(ctx commonsCtx.Context, t *task.Task) (packageResult, error) {
				result, err := o.runPackageTest(ctx, pkgPath, fw, runner, t, opts)
				if o.streamer != nil {
					o.streamer.CompletePackage(pkgPath, fw, result.testResults)
				}
				return result, err
			})
		}
	}

	// Wait for all tasks to complete and collect results
	resultMap, err := group.GetResults()
	if err != nil {
		return nil, fmt.Errorf("test execution failed: %w", err)
	}

	// Collect all test results and failures
	var allResults parsers.TestSuiteResults

	for _, result := range resultMap {
		allResults = append(allResults, result.testResults...)
	}

	// Fold cached hits into the final results and stream them to the UI so
	// the user sees them alongside freshly-run packages.
	for fw, hits := range cachedByFramework {
		for _, hit := range hits {
			suite := cachedSuiteResults(hit)
			allResults = append(allResults, suite...)
			if o.streamer != nil {
				o.streamer.CompletePackage(hit.PkgPath, fw, suite)
			}
		}
	}

	dedupeGinkgoBootstraps(allResults)
	allResults.Sort()

	return allResults, nil
}

// subprocessTimeoutFor computes the deadline to pass into the test
// subprocess (go test -timeout / ginkgo --timeout) so the subprocess fires
// its own timeout handler — and emits a goroutine dump / progress report —
// before gavel's supervisor SIGKILLs the process group. The margin is
// max(20s, testTimeout/10): short budgets keep ~20s of head-room, long
// budgets reserve 10%. Returns 0 when testTimeout is too small (or zero)
// to meaningfully carve out a margin; callers then skip the flag.
func subprocessTimeoutFor(testTimeout time.Duration) time.Duration {
	if testTimeout <= 20*time.Second {
		return 0
	}
	margin := testTimeout / 10
	if margin < 20*time.Second {
		margin = 20 * time.Second
	}
	return testTimeout - margin
}

func (o *TestOrchestrator) runPackageTest(
	ctx commonsCtx.Context,
	pkgPath string,
	framework Framework,
	runner runners.Runner, // runners.Runner interface
	t *task.Task,
	opts runners.TestOptions,
) (packageResult, error) {
	testRun, err := o.buildCommand(runner, pkgPath, opts)
	if err != nil {
		// Fold the build-command error into the compact task label so
		// the single CI line carries the reason. Avoid t.Errorf which
		// renders as indented child lines under the task.
		buildFail := parsers.Test{
			Failed:  true,
			Name:    fmt.Sprintf("%s Build", framework),
			Message: err.Error(),
		}
		t.SetName(formatPackageLabel(framework, pkgPath, parsers.TestSummary{Failed: 1, Total: 1}, &buildFail))
		t.SetDescription("")
		t.Failed()
		return packageResult{
			packagePath: pkgPath,
			framework:   framework,
			err:         fmt.Errorf("failed to build command: %w", err),
		}, nil
	}

	testRun.Process.SucceedOnNonZero = true
	if o.OutputTee != nil {
		testRun.Process.Stream(o.OutputTee, o.OutputTee)
	}
	// Tap stdout into a streaming parser so the UI reflects per-test progress
	// while the subprocess is still running. The final authoritative result
	// still comes from the framework's report once the run finishes.
	progressStream := o.newProgressStream(framework, pkgPath)
	if progressStream != nil {
		testRun.Process.Stream(progressStream, io.Discard)
	}
	// Execute the test process in the orchestrator
	process := testRun.Process.WithTask(t)
	// Keep the RUNNING-state label short; it may be rendered if the
	// task group dumps dirty tasks before the package finishes.
	t.SetName(formatRunningCommand(testRun.Process.Cmd, testRun.Process.Args))

	lease, err := o.acquireSlots(ctx, slots.Request{Limits: testRun.Slots, WorkDir: process.Cwd, Command: formatRunningCommand(process.Cmd, process.Args)})
	if err != nil {
		return packageResult{}, err
	}
	defer func() {
		if err := lease.Close(); err != nil {
			logger.Errorf("release test slots: %v", err)
		}
	}()
	process.WithEnv(lease.Env())
	pkgCtx, cancelPkg, supervision := o.supervisePackage(ctx, process, pkgPath)
	defer cancelPkg()
	process.WithContext(pkgCtx)

	span, err := o.beginProcessSpan()
	if err != nil {
		return packageResult{}, err
	}
	runStart := time.Now()
	result := clearOrphanedPipeError(process.WithContext(pkgCtx).Run().Result(), pkgPath)
	if err := lease.Close(); err != nil {
		return packageResult{}, err
	}
	runDuration := time.Since(runStart)
	cancelPkg()
	timedOut := supervision.TimedOut()
	if progressStream != nil {
		progressStream.Flush()
	}
	usage, err := span.usage(process.Pid())
	if err != nil {
		return packageResult{}, err
	}
	stamp := attemptStamp{RunKind: o.RunKind, Started: runStart, Ended: runStart.Add(runDuration), PID: process.Pid(), ExitCode: result.ExitCode, Usage: usage}

	// Always attempt to parse test results first, even if there was an execution error
	// This handles cases like `go test -json` which outputs valid JSON even when tests fail
	testResults, parseErr := o.parseTestResults(testRun, result, pkgPath)
	if result.Error != nil && len(testResults.All()) > 0 {
		testResults = append(testResults, parsers.TestResults{Framework: framework, Tests: parsers.Tests{{Name: fmt.Sprintf("%s Runtime", framework), Failed: true, TimedOut: timedOut, Message: result.Error.Error()}}})
	}

	// If parsing failed or returned no tests, create a fallback execution test
	if parseErr != nil || len(testResults.All()) == 0 {
		var message string
		if parseErr != nil {
			message = fmt.Sprintf("%s: %s", errors.Join(result.Error, parseErr), result.Output())
		} else if result.Error != nil {
			message = result.Error.Error()
		} else {
			message = strings.TrimSpace(result.Output())
		}

		fallback := parsers.Test{
			Failed:   true,
			TimedOut: timedOut,
			Name:     fmt.Sprintf("%s Execution", framework),
			Message:  message,
			Stderr:   result.Stderr,
			Stdout:   result.Stdout,
		}
		if timedOut {
			fallback.Name = fmt.Sprintf("%s Timeout", framework)
			fallback.Message = ""
		}
		testResults = parsers.TestSuiteResults{{
			Command:   process.Cmd,
			Framework: testRun.Framework,
			Stderr:    result.Stderr,
			Stdout:    result.Stdout,
			ExitCode:  result.ExitCode,
			Tests:     parsers.Tests{fallback},
		}}
		stampAttempts(testResults, stamp)

		// Fold the parse/execution failure reason into the compact task
		// label so the single line in the CI step log is self-explanatory.
		// Don't use t.Errorf here — it buffers indented child lines that
		// clobber the one-line-per-task budget the user asked for.
		failSum := parsers.TestSummary{Failed: 1, Total: 1, Duration: runDuration}
		t.SetName(formatPackageLabel(framework, pkgPath, failSum, &fallback))
		t.SetDescription("")
		t.Failed()

		return packageResult{
			packagePath: pkgPath,
			framework:   framework,
			testResults: testResults,
			timedOut:    timedOut,
		}, nil
	}

	// If the package was killed by our timeout supervisor but the parser
	// still yielded tests (happens when JSON events were flushed before the
	// SIGKILL), propagate TimedOut onto any test the parser saw start but
	// never observed finishing. Tests that already reported pass/fail/skip
	// before the kill keep their actual outcome.
	if timedOut {
		markTimedOutTests(testResults, runDuration, result.Stdout, result.Stderr, framework)
	}

	stampAttempts(testResults, stamp)

	// Update task status based on results. Build a compact, no-ANSI label
	// that fits in the one-line-per-task CI budget. On failure the label
	// carries the first failing leaf's stderr/stdout/message snippet so
	// the line is self-explanatory in the GitHub step log.
	sum := testResults.All().Sum()
	var firstFail *parsers.Test
	if sum.Failed > 0 {
		firstFail = firstFailingLeaf(testResults)
	}
	t.SetName(formatPackageLabel(framework, pkgPath, sum, firstFail))
	t.SetDescription("")

	if sum.Failed > 0 {
		t.Failed()
	} else if sum.Skipped > 0 {
		t.Warning()
	} else {
		t.Success()
		// Persist the passing outcome to the run cache so the next
		// invocation with --cache can skip it. No-op when --cache is off
		// (selector is nil).
		if o.selector != nil {
			o.selector.recordSuccess(framework, pkgPath, testResults, runDuration)
		}
	}

	return packageResult{
		packagePath: pkgPath,
		framework:   framework,
		testResults: testResults,
		timedOut:    timedOut,
	}, nil
}

// clearOrphanedPipeError keeps a test package that exited 0 from failing only
// because something it spawned (a Go toolchain helper, a grandchild) still held
// its stdout when clicky's WaitDelay force-closed the pipes. The output that
// matters was already read; any other error, or a non-zero exit, is kept.
func clearOrphanedPipeError(result *exec.ExecResult, pkgPath string) *exec.ExecResult {
	if result.ExitCode == 0 && errors.Is(result.Error, osexec.ErrWaitDelay) {
		logger.Warnf("test package %s: exited 0 but a descendant kept its output pipe open; pipes were closed after WaitDelay", pkgPath)
		result.Error = nil
	}
	return result
}

// markTimedOutTests walks the parsed results from a package whose subprocess
// was killed by the timeout supervisor and annotates every Test that did not
// report a terminal outcome (passed/failed/skipped) with TimedOut=true +
// Failed=true. When every parsed test is already terminal but the package
// itself ran past its deadline, append a synthetic timeout record to the
// first suite so the UI still surfaces one timeout entry with captured
// stdout/stderr. Also marks the enclosing TestSuiteResult so aggregators can
// show the package-level timeout icon.
func markTimedOutTests(results parsers.TestSuiteResults, runDuration time.Duration, stdout, stderr string, framework Framework) {
	marker := fmt.Sprintf("killed by gavel test-timeout supervisor after %s", runDuration.Round(time.Millisecond))
	touched := false
	for ri := range results {
		for ti := range results[ri].Tests {
			if markSubtreeTimedOut(&results[ri].Tests[ti], marker) {
				touched = true
			}
		}
	}
	if touched {
		return
	}
	if len(results) == 0 {
		return
	}
	results[0].Tests = append(results[0].Tests, parsers.Test{
		Failed:   true,
		TimedOut: true,
		Name:     fmt.Sprintf("%s Timeout", framework),
		Message:  marker,
		Stdout:   stdout,
		Stderr:   stderr,
	})
}

// markSubtreeTimedOut recursively flags any non-terminal leaf under t as
// TimedOut + Failed. Returns true when at least one leaf was marked.
func markSubtreeTimedOut(t *parsers.Test, marker string) bool {
	if len(t.Children) > 0 {
		marked := false
		for i := range t.Children {
			if markSubtreeTimedOut(&t.Children[i], marker) {
				marked = true
			}
		}
		if marked {
			t.TimedOut = true
		}
		return marked
	}
	if t.Passed || t.Failed || t.Skipped {
		return false
	}
	t.TimedOut = true
	t.Failed = true
	if t.Message == "" {
		t.Message = marker
	}
	return true
}

// parseTestResults parses test results from either stdout (go test) or report file (Ginkgo).
func (o *TestOrchestrator) parseTestResults(testRun *runners.TestRun, result *exec.ExecResult, pkgPath string) (parsers.TestSuiteResults, error) {
	var tests parsers.Tests
	var err error

	if testRun.ReportPath != "" {
		tests, err = o.parseReportFile(testRun)
		if err != nil {
			return nil, err
		}
	} else {
		// Stdout-based parsers (go test).
		tests, err = testRun.Parser.Parse(strings.NewReader(result.Stdout))
		if err != nil {
			return nil, fmt.Errorf("failed to parse test results: %w", err)
		}
	}

	// The package stderr stays once on the suite result; only a failed test
	// without stderr of its own gets a copy, as it may explain the failure.
	processStderr := stripExitStatus(strings.TrimSpace(result.Stderr))
	for i := range tests {
		tests[i].PackagePath = pkgPath
		tests[i].Command = testRun.Process.Cmd + " " + strings.Join(testRun.Process.Args, " ")
		if tests[i].Failed && tests[i].Stderr == "" {
			tests[i].Stderr = processStderr
		}
	}

	// Parsers own file-path normalization — see parsers.GinkgoJSON.relToWorkDir,
	// parsers.JestJSON.relPath, parsers.PlaywrightJSON.relPath, and the
	// location map in parsers.GoTestJSON. Orchestrator just stamps package
	// context on top; no second-pass normalization here.

	return parsers.TestSuiteResults{{
		Command:   testRun.Process.Cmd,
		Framework: testRun.Framework,
		Tests:     tests,
		Stdout:    result.Stdout,
		Stderr:    result.Stderr,
		ExitCode:  result.ExitCode,
	}}, nil
}

// parseReportFile reads and parses a runner's JSON report file using the
// TestRun's own Parser. Used by Ginkgo, Jest, Vitest, and Playwright runners
// that write results to a file rather than stdout.
func (o *TestOrchestrator) parseReportFile(testRun *runners.TestRun) (parsers.Tests, error) {
	reportPath := testRun.ReportPath
	if o.WorkDir != "" && !filepath.IsAbs(reportPath) {
		reportPath = filepath.Join(o.WorkDir, reportPath)
	}
	execName := fmt.Sprintf("%s Execution", testRun.Framework)
	if _, err := os.Stat(reportPath); err != nil {
		absPath, _ := filepath.Abs(reportPath)
		return parsers.Tests{{
			Name:    execName,
			Failed:  true,
			Message: fmt.Sprintf("'%s' not found", absPath),
		}}, nil
	}

	reportFile, err := os.Open(reportPath)
	if err != nil {
		return parsers.Tests{{
			Name:    execName,
			Failed:  true,
			Message: fmt.Sprintf("Failed to open %s report: %v", testRun.Framework, err),
		}}, nil
	}
	defer reportFile.Close()

	tests, err := testRun.Parser.Parse(reportFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s report: %w", testRun.Framework, err)
	}

	return tests, nil
}

// displayDryRun shows what tests would be executed without running them.
func (o *TestOrchestrator) displayDryRun(packagesByFramework map[Framework][]string, optsByFramework map[Framework]runners.TestOptions, unfocused bool) (parsers.TestSuiteResults, error) {
	logger.Infof("🔍 Dry-run mode: showing what would be executed\n")

	var totalPackages int
	for _, packages := range packagesByFramework {
		totalPackages += len(packages)
	}

	logger.Infof("Detected Frameworks: %s", clicky.CompactList(lo.Keys(packagesByFramework)).ANSI())
	logger.Infof("Total Packages: %d\n", totalPackages)
	goBatches, err := o.planGoBatches(packagesByFramework, optsByFramework, unfocused)
	if err != nil {
		return nil, err
	}
	for _, batch := range goBatches {
		process, err := o.buildGoBatchCommand(batch, 0)
		if err != nil {
			return nil, err
		}
		PrintDryRunCommand("test:go", fmt.Sprintf("%d packages", len(batch.packages)), process.Cmd, process.Args, process.Cwd)
	}

	for framework, packages := range packagesByFramework {
		if framework == parsers.GoTest || framework == parsers.Ginkgo {
			continue
		}
		runner, ok := o.registry.Get(framework)
		if !ok {
			continue
		}

		for _, pkg := range packages {
			testRun, err := o.buildCommand(runner, pkg, optsByFramework[framework])
			if err != nil {
				return nil, fmt.Errorf("[test:%s] %s: build failed: %w", framework, pkg, err)
			}
			PrintDryRunCommand(
				fmt.Sprintf("test:%s", framework),
				pkg,
				testRun.Process.Cmd,
				testRun.Process.Args,
				testRun.Process.Cwd,
			)
			runtime, _, runtimeErr := o.runtimeFor(framework, "")
			if runtimeErr != nil {
				return nil, runtimeErr
			}
			if summary := runtime.Summary().String(); summary != "" {
				logger.Infof("  runtime: %s", summary)
			}
		}
	}

	return nil, nil
}

// PrintDryRunCommand renders a single dry-run command line in the shared
// format used by gavel test and gavel lint. It is exported so the lint
// command (in package main) can render test and lint previews identically.
//
//	[label] title
//	  $ cmd arg1 arg2
//	  cwd: /working/dir
func PrintDryRunCommand(label, title, cmd string, args []string, cwd string) {
	logger.Infof("[%s] %s", label, title)
	logger.Infof("  $ %s", shellJoin(append([]string{cmd}, args...)))
	if cwd != "" {
		logger.Infof("  cwd: %s", cwd)
	}
	logger.Infof("")
}

// PrintDryRunSkipped renders a "would be skipped" line in dry-run output.
func PrintDryRunSkipped(label, title, reason string) {
	logger.Infof("[%s] %s  (skipped: %s)", label, title, reason)
}

func shellJoin(tokens []string) string {
	parts := make([]string, 0, len(tokens))
	for _, t := range tokens {
		parts = append(parts, shellQuote(t))
	}
	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " \t\n\"'\\$`*?[]{}()&;|<>#~!") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// Runner is an alias for TestOrchestrator for backwards compatibility during migration.
// Deprecated: use TestOrchestrator directly.
type Runner = TestOrchestrator

// resolveFixtureGlobs decides whether fixtures should run and with which globs.
// Returns nil when fixtures are disabled (neither the --fixtures flag nor
// .gavel.yaml fixtures.enabled is set). Flag-supplied globs take precedence
// over config; config takes precedence over the default glob.
func resolveFixtureGlobs(opts RunOptions) []string {
	if opts.OptionFields["fixtures"] && !opts.Fixtures {
		return nil
	}
	cfg, err := verify.LoadGavelConfig(opts.runtimeWorkDir())
	if err != nil {
		logger.Warnf("failed to load .gavel.yaml for fixture config: %v", err)
	}
	if !opts.Fixtures && !cfg.Fixtures.Enabled {
		return nil
	}
	if configured := cfg.TestOptionsFor("").Fixtures; configured != nil && !*configured && !opts.OptionFields["fixtures"] {
		return nil
	}
	if len(opts.FixtureFiles) > 0 {
		return opts.FixtureFiles
	}
	return cfg.Fixtures.ResolvedFiles()
}

// discoverFixtures resolves globs against workDir, then limits matches to the
// supplied starting paths. Discovery is gitignore-aware and stops at nested
// checkouts, so a scratch worktree holding a full copy of the repo doesn't
// double-run every fixture.
func discoverFixtures(workDir string, startingPaths []string, globs []string) ([]string, error) {
	if len(globs) == 0 {
		return nil, nil
	}
	matches, err := utils.GlobFilesBounded(workDir, globs)
	if err != nil || len(startingPaths) == 0 {
		return matches, err
	}

	roots := make([]string, 0, len(startingPaths))
	for _, startingPath := range startingPaths {
		root := startingPath
		if !filepath.IsAbs(root) {
			root = filepath.Join(workDir, root)
		}
		root, err := filepath.Abs(root)
		if err != nil {
			return nil, fmt.Errorf("resolve fixture starting path %q: %w", startingPath, err)
		}
		roots = append(roots, root)
	}

	filtered := make([]string, 0, len(matches))
	for _, match := range matches {
		for _, root := range roots {
			if utils.IsWithin(match, root) {
				filtered = append(filtered, match)
				break
			}
		}
	}
	return filtered, nil
}

func (t *TestOrchestrator) runDiscoveredFixtures(globs []string) ([]parsers.Test, error) {
	opts := t.RunOptions
	fixtureFiles, err := discoverFixtures(opts.WorkDir, opts.StartingPaths, globs)
	if err != nil {
		return nil, err
	}
	if len(fixtureFiles) == 0 {
		return nil, nil
	}

	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	// A run sharing its caller's trace host leaves archiving the store to that caller.
	var archive blob.Store
	if opts.TraceHost == nil {
		ctx := opts.Context
		if ctx == nil {
			ctx = context.Background()
		}
		if archive, err = verify.OpenWorkspaceStorage(ctx, opts.WorkDir); err != nil {
			return nil, err
		}
	}
	var tests []parsers.Test
	for _, file := range fixtureFiles {
		fixtureTests, err := t.runFixtureRunner(opts.withFixtureForms(fixtures.RunnerOptions{Context: opts.Context, ProcessRuntime: opts.fixtureRuntime(), Paths: []string{file}, WorkDir: opts.WorkDir, Logger: logger.StandardLogger(), ExecutablePath: execPath, TraceHost: opts.TraceHost, ProcessTrace: opts.ProcessTrace, TraceArchive: archive}), filepath.Dir(file))
		// A broad discovery glob also matches markdown that only supports
		// fixtures (shared templates, notes); such a file is not a fixture file.
		if errors.Is(err, fixtures.ErrNoFixtures) {
			logger.Debugf("fixture discovery: %s declares no fixtures, skipping", file)
			continue
		}
		tests = append(tests, fixtureTests...)
		if err != nil {
			return tests, err
		}
	}
	return tests, nil

}

func (t *TestOrchestrator) runFixtureRunner(runnerOpts fixtures.RunnerOptions, configDir string) ([]parsers.Test, error) {
	streamer := t.streamer
	onProgress, onReport := t.FixtureBenchmarkProgress, t.FixtureBenchmarkResult
	config, err := verify.LoadGavelConfig(configDir)
	if err != nil {
		return nil, err
	}
	if framework := config.Test.Frameworks["fixture"]; framework.Enabled != nil && !*framework.Enabled {
		return nil, nil
	}
	var runtimeArgs []string
	configured := *t
	configured.Runtime = runnerOpts.ProcessRuntime
	runnerOpts.ProcessRuntime, runtimeArgs, err = configured.configuredRuntime(config, parsers.Fixture)
	if err != nil {
		return nil, err
	}
	runnerOpts.Args = append(runtimeArgs, runnerOpts.Args...)
	// The runner's display options prune passing results from its tree, but
	// every result is merged into the tests; RunOptions' Show* filter them.
	runnerOpts.Display = nil
	if streamer != nil || onProgress != nil {
		upstream := runnerOpts.ProgressSink
		runnerOpts.ProgressSink = func(ctx context.Context, snapshot fixtures.ExecutionSnapshot) error {
			if upstream != nil {
				if err := upstream(ctx, snapshot); err != nil {
					return err
				}
			}
			if streamer != nil {
				tests, err := executionSnapshotToTests(snapshot)
				if err != nil {
					return err
				}
				streamer.UpdateFixtures(tests)
			}
			if onProgress != nil {
				onProgress(snapshot)
			}
			return nil
		}
	}

	runner, err := fixtures.NewRunner(runnerOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create fixture runner: %w", err)
	}
	if t.DryRun {
		tree, err := runner.Parse()
		if err != nil {
			return nil, err
		}
		printFixtureDryRun(tree, runnerOpts)
		return nil, nil
	}

	result, err := runner.Run()
	if onReport != nil {
		onReport(runner.BenchmarkReport())
	}

	if result == nil {
		return nil, err
	}
	tests, projectErr := FixtureRunTests(runner, result, runnerOpts.WorkDir)
	if projectErr != nil {
		return tests, errors.Join(err, projectErr)
	}
	if streamer != nil {
		streamer.UpdateFixtures(tests)
	}
	return tests, err
}

func fixtureNodeToTests(node *fixtures.FixtureNode) []parsers.Test {
	return fixtureNodeToTestsWithFile(node, "")
}

func fixtureNodeToTestsWithFile(node *fixtures.FixtureNode, inheritedFile string) []parsers.Test {
	fixtureFile := fixtureNodeFile(node, inheritedFile)
	if node.Results != nil {
		r := node.Results
		t := parsers.Test{
			Name:      node.Name,
			Framework: parsers.Fixture,
			File:      fixtureFile,
			Line:      fixtureNodeLine(node),
			EndLine:   fixtureNodeEndLine(node),
			Duration:  r.Duration,
			Stdout:    r.Stdout,
			Stderr:    r.Stderr,
			Failed:    r.Status == task.StatusFAIL || r.Status == task.StatusFailed || r.Status == task.StatusERR,
			Passed:    r.Status == task.StatusPASS || r.Status == task.StatusSuccess,
			Warned:    r.Status == task.StatusWarning,
			Message:   r.Error,
			Context: parsers.FixtureContext{
				Command:       r.Command,
				ExitCode:      r.ExitCode,
				CWD:           r.CWD,
				CELExpression: r.CELExpression,
				CELTrace:      r.CELTrace,
				CELVars:       r.CELVars,
				Expected:      r.Expected,
				Actual:        r.Actual,
				Traces:        r.Traces,
			},
		}
		if detail := fixtures.ResultDetail(r); detail != nil {
			t.Detail = fixtureDetail(detail)
		}
		// A step's own result children — a query step's findings — are part of
		// what it reports, so they become the test's children. The Runner moves
		// them onto the node (attachResult); RunNodes leaves them on the result.
		for _, child := range append(append([]*fixtures.FixtureNode{}, node.Children...), r.Children...) {
			t.Children = append(t.Children, fixtureNodeToTestsWithFile(child, fixtureFile)...)
		}
		if r.Profile != nil {
			t.FixtureProfile = &parsers.FixtureProfile{
				Scope: r.Profile.Scope, PID: r.Profile.PID, SampleCount: r.Profile.SampleCount,
				PeakCPUPercent: r.Profile.PeakCPUPercent, PeakMemoryPercent: r.Profile.PeakMemoryPercent,
				PeakRSSBytes: r.Profile.PeakRSSBytes,
			}
			if r.Profile.DiskIO != nil {
				t.FixtureProfile.DiskIO = &parsers.ProfileDiskIO{
					DiskReadBytes: r.Profile.DiskIO.DiskReadBytes, DiskWriteBytes: r.Profile.DiskIO.DiskWriteBytes,
					SampleCount: r.Profile.DiskIO.SampleCount, MissingProcesses: r.Profile.DiskIO.MissingProcesses,
				}
			}
		}
		if r.SQLProfile != nil {
			t.SQLProfile = &parsers.SQLProfile{
				Path: r.SQLProfile.Path, QueryCount: r.SQLProfile.QueryCount,
				SlowQueryCount:  r.SQLProfile.SlowQueryCount,
				TotalDurationMS: r.SQLProfile.TotalDurationMS, MaxQueryMS: r.SQLProfile.MaxQueryMS,
			}
			for _, statement := range r.SQLProfile.Statements {
				t.SQLProfile.Statements = append(t.SQLProfile.Statements, parsers.SQLProfileStatement{
					SQL: statement.SQL, Params: statement.Params, DurationMS: statement.DurationMS,
					Rows: statement.Rows, Slow: statement.Slow, Error: statement.Error,
				})
			}
		}
		return []parsers.Test{t}
	}

	var children parsers.Tests
	for _, child := range node.Children {
		children = append(children, fixtureNodeToTestsWithFile(child, fixtureFile)...)
	}

	if node.Type == fixtures.FileNode || node.Type == fixtures.SectionNode {
		failed := false
		for _, child := range children {
			if child.Failed || child.Sum().Failed > 0 {
				failed = true
				break
			}
		}
		return []parsers.Test{{
			Name:      node.Name,
			Framework: parsers.Fixture,
			File:      fixtureFile,
			Line:      fixtureNodeLine(node),
			Children:  children,
			Failed:    failed,
		}}
	}
	return children
}

func fixtureNodeFile(node *fixtures.FixtureNode, inherited string) string {
	if node == nil || node.Origin == nil || node.Origin.File == "" {
		return inherited
	}
	if inherited != "" && !filepath.IsAbs(node.Origin.File) {
		return inherited
	}
	return node.Origin.File
}

func fixtureNodeLine(node *fixtures.FixtureNode) int {
	if node == nil || node.Origin == nil {
		return 0
	}
	return node.Origin.Line
}

func fixtureNodeEndLine(node *fixtures.FixtureNode) int {
	if node == nil || node.Origin == nil {
		return 0
	}
	return node.Origin.EndLine
}
