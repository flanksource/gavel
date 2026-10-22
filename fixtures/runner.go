package fixtures

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"

	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/processruntime"
	"github.com/flanksource/gavel/traces"
)

// RunnerOptions configures the fixture runner
type RunnerOptions struct {
	ProcessRuntime processruntime.Options
	Args           []string
	Context        context.Context `json:"-" yaml:"-"`
	worker         bool
	Paths          []string // Fixture file paths/patterns
	Format         string   // Output format: tree, table, json, yaml, csv
	Filter         string   // Filter tests by name pattern (glob)
	NoColor        bool     // Disable colored output
	WorkDir        string   // Working directory
	// MaxWorkers caps how many fixture rows run at once. Zero means 1: rows
	// shell out to real binaries against shared databases, files and ports, so
	// parallelism is opt-in rather than the default.
	MaxWorkers     int
	Logger         logger.Logger                                         `json:"-" yaml:"-"`
	ExecutablePath string                                                // Path to the current executable (for fixtures to use)
	ProgressSink   ProgressSink                                          `json:"-" yaml:"-"` // Receives immutable execution-tree snapshots
	UpdateGolden   bool                                                  // When true, mismatched @file expectations are rewritten with actual output instead of failing
	Display        *DisplayOptions                                       // Optional result visibility controls for CLI rendering
	Spec           *api.Spec                                             // Runtime options for embedded AI prompts
	ResolveSpec    func(context.Context) (api.ResolveSpecOptions, error) `json:"-" yaml:"-"` // Lazy authored layers and saved snapshot; Spec takes precedence
	// DefaultTraces is the run-wide `--traces` default, armed only for fixtures
	// that declared no `traces:` of their own. An explicit `traces: []` is a
	// non-nil empty declaration precisely so it outranks this.
	DefaultTraces *TraceSpecs
	// TraceHost is the run's trace store, shared with the run that owns it
	// (gavel test); nil opens one under WorkDir once a fixture arms a trace.
	TraceHost *traces.Host `json:"-" yaml:"-"`
	// ProcessTrace is the run's process capture, which --profile reads; nil
	// starts one on the trace host for --profile.
	ProcessTrace *traces.ProcessTrace `json:"-" yaml:"-"`
	// TraceArchive receives a copy of the trace store a runner opened itself, under
	// traces/<runID>/, once the run ends. nil — no storage configured — leaves it only in
	// .gavel/traces.
	TraceArchive blob.Store `json:"-" yaml:"-"`
	Benchmark    bool
	Baseline     string
	Profile      bool
	SQLProfile   bool
	Limits       BenchmarkLimits
}

// Runner manages fixture test execution using typed tasks
type Runner struct {
	options    RunnerOptions
	fixtures   []FixtureTest
	evaluator  *CELEvaluator
	tree       *FixtureNode // Hierarchical tree structure
	daemonCmd  *exec.Cmd
	daemonPort int
	// setups holds one prepared `setup:` per markdown file that declared one,
	// keyed on FixtureNode.Origin.File. Keyed on the file rather than the source
	// directory because two fixture files in one directory are independent.
	setups map[string]*PreparedSetup
	// traceHost is the store the run's traces commit to, nil while no fixture
	// arms one; ownsTraceHost when the runner opened it. sharedTraces are the
	// running file- and run-scope captures, by sharedTraceKey.
	traceHost     *traces.Host
	ownsTraceHost bool
	sharedTraces  map[string]*traces.Capture
	// profileTrace is the process capture --profile reads from profileMark
	// on, ownsProfileTrace when the runner started it; profile is what it
	// read once the run ended.
	profileTrace     *traces.ProcessTrace
	ownsProfileTrace bool
	profileMark      traces.ProcessMark
	profile          *ProfileReport
	progress         *executionTracker
	resultMu         sync.Mutex
	benchmark        *BenchmarkReport
	workerSnapshot   *ExecutionSnapshot
	baseline         *BenchmarkReport
	// backstopGrace is how long a fixture task's deadline outlasts a declared
	// timeout (declaredTimeoutGrace outside tests).
	backstopGrace time.Duration
	// cutOff holds the fixtures whose task was cut off by its backstop deadline
	// or failed outright: their verdict is final, and a run function that
	// returns late must not overwrite it.
	cutOff map[*FixtureNode]bool
}

// NewRunner creates a new fixture runner
func NewRunner(opts RunnerOptions) (*Runner, error) {
	if err := opts.ProcessRuntime.Validate(); err != nil {
		return nil, err
	}
	if opts.Benchmark && opts.Display != nil {
		display := *opts.Display
		display.ShowPassed = true
		opts.Display = &display
	}
	if opts.Baseline != "" && !opts.Benchmark {
		return nil, fmt.Errorf("--baseline requires --benchmark")
	}
	if opts.Limits.Enabled() && !opts.Benchmark {
		return nil, fmt.Errorf("benchmark limits require --benchmark")
	}
	if opts.Limits.MaxDurationMS < 0 || opts.Limits.MaxSQLDurationMS < 0 || opts.Limits.MaxSQLQueryMS < 0 {
		return nil, fmt.Errorf("benchmark time limits must be non-negative")
	}
	if opts.Limits.MaxSQLQueries != nil && *opts.Limits.MaxSQLQueries < 0 || opts.Limits.MaxSlowSQL != nil && *opts.Limits.MaxSlowSQL < 0 {
		return nil, fmt.Errorf("SQL query limits must be non-negative")
	}
	if opts.Limits.MaxDeviationPct != nil {
		if opts.Baseline == "" {
			return nil, fmt.Errorf("--max-deviation-pct requires --baseline")
		}
		if *opts.Limits.MaxDeviationPct < 0 || math.IsNaN(*opts.Limits.MaxDeviationPct) || math.IsInf(*opts.Limits.MaxDeviationPct, 0) {
			return nil, fmt.Errorf("--max-deviation-pct must be a finite non-negative number")
		}
	}
	if opts.Limits.NeedsProcessProfile() {
		opts.Profile = true
	}
	if opts.Limits.MaxSQLDurationMS > 0 || opts.Limits.MaxSQLQueryMS > 0 || opts.Limits.MaxSQLQueries != nil || opts.Limits.MaxSlowSQL != nil || opts.Profile {
		opts.SQLProfile = true
	}
	var baseline *BenchmarkReport
	if opts.Baseline != "" {
		if !filepath.IsAbs(opts.Baseline) && opts.WorkDir != "" {
			opts.Baseline = filepath.Join(opts.WorkDir, opts.Baseline)
		}
		var err error
		baseline, err = LoadBenchmarkReport(opts.Baseline)
		if err != nil {
			return nil, fmt.Errorf("--baseline: %w", err)
		}
		if opts.Limits.MaxDeviationPct != nil {
			for _, entry := range baseline.Fixtures {
				if entry.Profile != nil {
					opts.Profile = true
				}
				if entry.SQLProfile != nil {
					opts.SQLProfile = true
				}
			}
		}
	}
	// Create CEL evaluator
	evaluator, err := NewCELEvaluator()
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL evaluator: %w", err)
	}

	return &Runner{
		options:       opts,
		baseline:      baseline,
		fixtures:      []FixtureTest{},
		evaluator:     evaluator,
		backstopGrace: declaredTimeoutGrace,
		tree: &FixtureNode{
			Name: "Fixtures",
			Type: SectionNode,
		},
	}, nil
}

// Run executes the fixture tests and returns the result tree.
// The caller is responsible for formatting/printing the output.
func (r *Runner) Run() (*FixtureNode, error) {
	if r.options.ProcessRuntime.Isolated() && !r.options.worker {
		return r.runIsolated()
	}
	r.benchmark = nil
	started := time.Now()
	if _, err := r.prepareFixtureTree(); err != nil {
		return nil, err
	}
	r.progress = newExecutionTracker(r.tree, r.options.WorkDir, r.executionSteps(), r.options.ProgressSink)
	if err := r.progress.Publish(context.Background()); err != nil {
		return nil, fmt.Errorf("publish queued fixture progress: %w", err)
	}

	r.profile = nil
	if err := r.startProfile(); err != nil {
		return nil, errors.Join(err, r.closeTraces())
	}
	results, err := r.executeFixtures()
	// executeFixtures closes the traces it prepared; one that returned before
	// preparing them leaves the profile's for here.
	err = errors.Join(err, r.closeTraces())

	if err == nil {
		clicky.WaitForGlobalCompletion()
	}
	if r.options.Benchmark || r.options.Profile {
		if reportErr := r.saveBenchmarkReport(started, time.Now(), r.profile, err); reportErr != nil {
			err = errors.Join(err, reportErr)
		}
	}
	if err != nil {
		return r.tree, fmt.Errorf("failed to execute fixtures: %w", err)
	}

	if results.Summary.HasFailures() {
		return r.tree, fmt.Errorf("fixture tests failed")
	}

	return r.tree, nil
}

func (r *Runner) BenchmarkReport() *BenchmarkReport { return r.benchmark }

func (r *Runner) ExecutionSnapshot() ExecutionSnapshot {
	r.resultMu.Lock()
	workerSnapshot := r.workerSnapshot
	r.resultMu.Unlock()
	if workerSnapshot != nil {
		return *workerSnapshot
	}
	if r.progress == nil {
		panic("fixture execution snapshot requested before Run")
	}
	return r.progress.Snapshot()
}

// Parse builds the fixture tree without executing any fixture work.
func (r *Runner) Parse() (*FixtureNode, error) {
	return r.prepareFixtureTree()
}

func (r *Runner) prepareFixtureTree() (*FixtureNode, error) {
	if err := r.parseFixtureFiles(); err != nil {
		return nil, fmt.Errorf("failed to parse fixture files: %w", err)
	}

	if r.options.Filter != "" {
		r.filterTests()
	}

	if len(r.fixtures) == 0 {
		return nil, fmt.Errorf("no fixtures found")
	}

	return r.tree, nil
}

// parseFixtureFiles parses all fixture files from the provided paths and builds tree structure
func (r *Runner) parseFixtureFiles() error {
	var allFixtures []FixtureTest
	r.tree = &FixtureNode{
		Name: "Fixtures",
		Type: SectionNode,
	}

	for _, pattern := range r.options.Paths {
		// Callers that already resolved concrete files (outline discovery) would
		// otherwise pay a second glob per path.
		matches := []string{pattern}
		if info, err := os.Stat(pattern); err != nil || !info.Mode().IsRegular() {
			matches, err = doublestar.FilepathGlob(pattern)
			if err != nil {
				return fmt.Errorf("invalid glob pattern '%s': %w", pattern, err)
			}
		}

		if len(matches) == 0 {
			logger.Warnf("No files matched pattern: %s", pattern)
			continue
		}

		for _, filepath := range matches {
			// Parse with tree structure
			fileTree, err := ParseFixtureFile(filepath)
			if err != nil {
				return fmt.Errorf("failed to parse fixture file '%s': %w", filepath, err)
			}

			// Merge file tree into main tree
			if fileTree != nil {
				r.tree.AddChild(fileTree)
			}

			fileFixtureCount := 0
			fileTree.Walk(func(node *FixtureNode) {
				if node.Test != nil {
					allFixtures = append(allFixtures, *node.Test)
					fileFixtureCount++
				}
			})
			logger.Debugf("Parsed %d fixtures from %s", fileFixtureCount, filepath)
		}
	}

	r.fixtures = allFixtures

	// Log the loaded fixtures
	fileCount := len(r.tree.Children)
	logger.Infof("Loaded %d total fixtures in %d files", len(allFixtures), fileCount)
	return nil
}

// filterTests applies name filtering to loaded tests
func (r *Runner) filterTests() {
	var filtered []FixtureTest

	for _, fixture := range r.fixtures {
		match, err := doublestar.Match(r.options.Filter, fixture.Name)
		if err != nil {
			logger.Warnf("Invalid filter pattern '%s': %v", r.options.Filter, err)
			continue
		}
		if match {
			filtered = append(filtered, fixture)
		}
	}

	logger.Infof("Filtered to %d fixtures matching '%s'", len(filtered), r.options.Filter)
	r.fixtures = filtered
	if r.tree != nil {
		filterFixtureTree(r.tree, r.options.Filter)
	}
}

func filterFixtureTree(node *FixtureNode, pattern string) bool {
	if node == nil {
		return false
	}

	if node.Test != nil {
		match, err := doublestar.Match(pattern, node.Test.Name)
		return err == nil && match
	}

	children := make([]*FixtureNode, 0, len(node.Children))
	for _, child := range node.Children {
		if filterFixtureTree(child, pattern) {
			children = append(children, child)
		}
	}
	node.Children = children

	return node.Type == FileNode || len(node.Children) > 0
}

// executeFixtures runs all fixtures using typed task groups
func (r *Runner) executeFixtures() (results *FixtureGroup, runErr error) {
	results = &FixtureGroup{
		Tests:   make([]FixtureNode, 0, len(r.fixtures)),
		Summary: Stats{},
	}

	parent := r.options.Context
	if parent == nil {
		parent = context.Background()
	}
	ctx := flanksourceContext.NewContext(parent)
	if r.hasSetup() {
		defer func() {
			startErr := r.startProgressStep(ctx, ExecutionKindCleanup)
			cleanupErr := r.cleanupSetups()
			state := ExecutionPassed
			if cleanupErr != nil {
				state = ExecutionErrored
			}
			completeErr := r.progress.CompleteStep(ctx, "cleanup", state, cleanupErr)
			runErr = errors.Join(runErr, cleanupErr, startErr, completeErr)
		}()
	}

	// Setup runs before the build: it can relocate a file's tests into a
	// worktree, and a build that ran in the original repo would build tree A
	// while the tests exercise tree B — and pass.
	if err := r.startProgressStep(ctx, ExecutionKindSetup); err != nil {
		return nil, err
	}
	if err := r.prepareSetups(ctx); err != nil {
		if progressErr := r.failPrerequisite(ctx, ExecutionKindSetup, err); progressErr != nil {
			return nil, errors.Join(err, progressErr)
		}
		return nil, err
	}
	if err := r.completeProgressStep(ctx, ExecutionKindSetup); err != nil {
		return nil, err
	}
	// Run build command synchronously before any fixtures
	buildCmd, buildSetup := r.getBuildCommand()
	if buildCmd != "" {
		if err := r.startProgressStep(ctx, ExecutionKindBuild); err != nil {
			return nil, err
		}
		logger.V(2).Infof("Running build command: %s", buildCmd)
		if err := r.executeBuildCommand(ctx, buildCmd, buildSetup); err != nil {
			if progressErr := r.failPrerequisite(ctx, ExecutionKindBuild, err); progressErr != nil {
				return nil, errors.Join(err, progressErr)
			}
			return nil, fmt.Errorf("build failed, skipping all fixtures: %w", err)
		}
		if err := r.completeProgressStep(ctx, ExecutionKindBuild); err != nil {
			return nil, err
		}
		logger.V(2).Infof("Build completed successfully")
	}

	// Start daemon if configured
	daemonCmd, daemonSetup := r.getDaemonCommand()
	if daemonCmd != "" {
		if err := r.startProgressStep(ctx, ExecutionKindDaemon); err != nil {
			return nil, err
		}
		if err := r.startDaemon(ctx, daemonCmd, daemonSetup); err != nil {
			if progressErr := r.failPrerequisite(ctx, ExecutionKindDaemon, err); progressErr != nil {
				return nil, errors.Join(err, progressErr)
			}
			return nil, fmt.Errorf("daemon failed to start: %w", err)
		}
		defer func() {
			startErr := r.startProgressStep(ctx, ExecutionKindDaemonStop)
			stopErr := r.stopDaemon()
			state := ExecutionPassed
			if stopErr != nil {
				state = ExecutionErrored
			}
			completeErr := r.progress.CompleteStep(ctx, "daemon-stop", state, stopErr)
			runErr = errors.Join(runErr, stopErr, startErr, completeErr)
		}()
		if err := r.completeProgressStep(ctx, ExecutionKindDaemon); err != nil {
			return nil, err
		}
	}

	// Traces start after the daemon, which runs untraced, and stop once every
	// file's scope is torn down, so a service a step started is traced until
	// it stops.
	if err := r.prepareTraces(); err != nil {
		return nil, errors.Join(err, r.closeTraces())
	}
	defer func() { runErr = errors.Join(runErr, r.closeTraces()) }()

	// Create typed task group for fixture execution. The group's own semaphore
	// is the only thing that bounds a row's concurrency — the manager's worker
	// pool is process-wide and already sized — so MaxWorkers has to arrive here
	// as WithConcurrency or it bounds nothing at all.
	fixtureGroup := task.StartGroup[FixtureResult]("Fixture Tests", task.WithConcurrency(r.maxWorkers()))
	r.cutOff = map[*FixtureNode]bool{}
	scopes, queued := r.addFixtureTasks(fixtureGroup, parent)

	// Wait on each task rather than the group: a task's own wait is what cuts it
	// off at its deadline when its run function ignores the context.
	tests, collectErr := r.collectFixtureResults(ctx, queued)
	results.Tests = tests
	r.attachTeardownResults(ctx, scopes)

	r.tree.UpdateStatsRecursive()
	results.Summary = *r.tree.Stats

	if r.options.Display != nil {
		r.tree.ApplyDisplayOptions(*r.options.Display)
	}

	// Prune empty sections from the tree
	r.tree.PruneEmptySections()

	return results, collectErr
}

// fixtureTimeout is a fixture's task deadline. A declared timeout is enforced by
// runKind, which reports it as a timeout; the task deadline only backstops it,
// so it fires after the declared one.
func fixtureTimeout(fixture FixtureTest, grace time.Duration) time.Duration {
	if timeout := fixture.declaredTimeout(); timeout != nil {
		return *timeout + grace
	}
	if fixture.Step != nil {
		return stepTaskTimeout(*fixture.Step)
	}
	return defaultFixtureTimeout
}

func (r *Runner) executionSteps() []ExecutionStep {
	var steps []ExecutionStep
	if r.hasSetup() {
		steps = append(steps, ExecutionStep{Key: "setup", Name: "Setup", Kind: ExecutionKindSetup})
	}
	if command, _ := r.getBuildCommand(); command != "" {
		steps = append(steps, ExecutionStep{Key: "build", Name: "Build", Kind: ExecutionKindBuild})
	}
	if command, _ := r.getDaemonCommand(); command != "" {
		steps = append(steps, ExecutionStep{Key: "daemon", Name: "Daemon", Kind: ExecutionKindDaemon})
		steps = append(steps, ExecutionStep{Key: "daemon-stop", Name: "Daemon stop", Kind: ExecutionKindDaemonStop})
	}
	if r.hasSetup() {
		steps = append(steps, ExecutionStep{Key: "cleanup", Name: "Cleanup", Kind: ExecutionKindCleanup})
	}
	return steps
}

func (r *Runner) hasSetup() bool {
	for _, file := range r.tree.Children {
		if setup, _, _ := fileSetup(file); setup != nil {
			return true
		}
	}
	return false
}

func (r *Runner) startProgressStep(ctx context.Context, kind ExecutionKind) error {
	key := progressStepKey(kind)
	if key == "" || r.progress.byKey[key] == nil {
		return nil
	}
	return r.progress.StartStep(ctx, key)
}

func (r *Runner) completeProgressStep(ctx context.Context, kind ExecutionKind) error {
	key := progressStepKey(kind)
	if key == "" || r.progress.byKey[key] == nil {
		return nil
	}
	return r.progress.CompleteStep(ctx, key, ExecutionPassed, nil)
}

func (r *Runner) failPrerequisite(ctx context.Context, kind ExecutionKind, cause error) error {
	key := progressStepKey(kind)
	if key == "" || r.progress.byKey[key] == nil {
		return r.progress.CancelQueued(ctx, cause)
	}
	return errors.Join(
		r.progress.CompleteStep(ctx, key, ExecutionErrored, cause),
		r.progress.CancelQueued(ctx, cause),
	)
}

func progressStepKey(kind ExecutionKind) string {
	switch kind {
	case ExecutionKindSetup:
		return "setup"
	case ExecutionKindBuild:
		return "build"
	case ExecutionKindDaemon:
		return "daemon"
	case ExecutionKindDaemonStop:
		return "daemon-stop"
	case ExecutionKindCleanup:
		return "cleanup"
	default:
		return ""
	}
}

// attachResult puts a result on its node; the caller holds resultMu.
func (r *Runner) attachResult(node *FixtureNode, result FixtureResult) {
	children := result.Children
	result.Children = nil
	stampSource(node, &result)
	node.Results = &result
	for _, child := range children {
		node.AddChild(child)
	}
}

// executeFixture runs a single fixture test in the environment prepared for the
// file that declared it (zero-valued when that file declared no `setup:`),
// under the traces it arms. The dispatch itself is RunNode's: this only
// assembles the run's environment around it, lands its traces and stamps the
// runner's own presentation.
func (r *Runner) executeFixture(ctx flanksourceContext.Context, fixture FixtureTest, env fixtureEnv) (FixtureResult, error) {
	if r.options.WorkDir == "" {
		r.options.WorkDir, _ = os.Getwd()
	}
	ctx.Logger.V(5).Infof("Using CWD: %s", r.options.WorkDir)

	opts := RunOptions{
		ProcessRuntime: r.options.ProcessRuntime,
		Args:           r.options.Args,
		Context:        ctx,
		WorkDir:        r.options.WorkDir,
		Verbose:        ctx.Logger.IsLevelEnabled(logger.Debug),
		Spec:           r.options.Spec,
		ResolveSpec:    r.options.ResolveSpec,
		Evaluator:      r.evaluator,
		ExecutablePath: r.options.ExecutablePath,
		UpdateGolden:   r.options.UpdateGolden,
		Profile:        r.options.Profile,
		SQLProfile:     r.options.SQLProfile,
		Setup:          env.setup,
		Scope:          env.scope,
		DaemonPort:     r.daemonPort,
		Progress:       fixtureProgressFromContext(ctx),
		ExtraArgs: map[string]interface{}{
			"flanksource_context": ctx,
		},
	}

	start := time.Now()
	armed, err := r.armTraces(ctx, fixture, env)
	if err != nil {
		return nodeError(fixture, fmt.Sprintf("arm traces: %v", err)), nil
	}
	opts.Traces = armed
	result := applyTraces(RunNode(ctx, fixture, opts), armed.Collect(ctx))
	result.Duration = time.Since(start)
	if r.options.Display != nil {
		result.Display = r.options.Display
	}
	return result, nil
}
