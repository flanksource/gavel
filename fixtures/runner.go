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

	"github.com/flanksource/gavel/fixtures/record"
)

// RunnerOptions configures the fixture runner
type RunnerOptions struct {
	Paths   []string // Fixture file paths/patterns
	Format  string   // Output format: tree, table, json, yaml, csv
	Filter  string   // Filter tests by name pattern (glob)
	NoColor bool     // Disable colored output
	WorkDir string   // Working directory
	// MaxWorkers caps how many fixture rows run at once. Zero means 1: rows
	// shell out to real binaries against shared databases, files and ports, so
	// parallelism is opt-in rather than the default.
	MaxWorkers     int
	Logger         logger.Logger
	ExecutablePath string                                                // Path to the current executable (for fixtures to use)
	ProgressSink   ProgressSink                                          // Receives immutable execution-tree snapshots
	UpdateGolden   bool                                                  // When true, mismatched @file expectations are rewritten with actual output instead of failing
	Display        *DisplayOptions                                       // Optional result visibility controls for CLI rendering
	Spec           *api.Spec                                             // Runtime options for embedded AI prompts
	ResolveSpec    func(context.Context) (api.ResolveSpecOptions, error) `json:"-" yaml:"-"` // Lazy authored layers and saved snapshot; Spec takes precedence
	// Record is the run-wide `--record` default, applied only to fixtures that
	// declared no `record:` of their own. An explicit `record: none` parses to an
	// empty (non-nil) Spec precisely so it outranks this.
	Record     *record.Spec
	Benchmark  bool
	Baseline   string
	Profile    bool
	SQLProfile bool
	Limits     BenchmarkLimits
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
	// recorders holds the diagnostics recorders per markdown file, keyed the
	// same way. nil when nothing in the run declared `record:` — no goroutine,
	// no listener, no file.
	recorders map[string]*recorderSet
	store     *record.Store
	progress  *executionTracker
	resultMu  sync.Mutex
	benchmark *BenchmarkReport
	baseline  *BenchmarkReport
}

// NewRunner creates a new fixture runner
func NewRunner(opts RunnerOptions) (*Runner, error) {
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
		options:   opts,
		baseline:  baseline,
		fixtures:  []FixtureTest{},
		evaluator: evaluator,
		tree: &FixtureNode{
			Name: "Fixtures",
			Type: SectionNode,
		},
	}, nil
}

// Run executes the fixture tests and returns the result tree.
// The caller is responsible for formatting/printing the output.
func (r *Runner) Run() (*FixtureNode, error) {
	r.benchmark = nil
	started := time.Now()
	if _, err := r.prepareFixtureTree(); err != nil {
		return nil, err
	}
	r.progress = newExecutionTracker(r.tree, r.options.WorkDir, r.executionSteps(), r.options.ProgressSink)
	if err := r.progress.Publish(context.Background()); err != nil {
		return nil, fmt.Errorf("publish queued fixture progress: %w", err)
	}

	var sampler *profileSampler
	if r.options.Profile {
		sampler = startProfileSampler()
	}
	results, err := r.executeFixtures()

	if err == nil {
		clicky.WaitForGlobalCompletion()
	}
	if r.options.Benchmark || r.options.Profile {
		var profile *ProfileReport
		if sampler != nil {
			var profileErr error
			profile, profileErr = sampler.Stop()
			err = errors.Join(err, profileErr)
		}
		if reportErr := r.saveBenchmarkReport(started, time.Now(), profile, err); reportErr != nil {
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
			fileTree, err := ParseMarkdownFixturesWithTree(filepath)
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

	ctx := flanksourceContext.NewContext(context.Background())
	if r.hasSetup() {
		defer func() {
			startErr := r.startProgressStep(ctx, ExecutionKindCleanup)
			cleanupErr := r.cleanupSetups()
			state := ExecutionPassed
			if cleanupErr != nil {
				state = ExecutionWarned
			}
			completeErr := r.progress.CompleteStep(ctx, "cleanup", state, cleanupErr)
			runErr = errors.Join(runErr, startErr, completeErr)
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
		if err := r.completeProgressStep(ctx, ExecutionKindDaemon); err != nil {
			return nil, err
		}
		defer func() {
			startErr := r.startProgressStep(ctx, ExecutionKindDaemonStop)
			stopErr := r.stopDaemon()
			state := ExecutionPassed
			if stopErr != nil {
				state = ExecutionWarned
			}
			completeErr := r.progress.CompleteStep(ctx, "daemon-stop", state, stopErr)
			runErr = errors.Join(runErr, startErr, completeErr)
		}()
	}

	// Recorders start after the daemon so its port can be excluded from the
	// proxy, and their defer is registered after the daemon's so LIFO closes
	// them first — a daemon shutting down can still emit requests worth
	// recording.
	if err := r.prepareRecorders(); err != nil {
		return nil, err
	}
	defer r.closeRecorders()

	// Create typed task group for fixture execution. The group's own semaphore
	// is the only thing that bounds a row's concurrency — the manager's worker
	// pool is process-wide and already sized — so MaxWorkers has to arrive here
	// as WithConcurrency or it bounds nothing at all.
	fixtureGroup := task.StartGroup[FixtureResult]("Fixture Tests", task.WithConcurrency(r.maxWorkers()))

	r.tree.Walk(func(node *FixtureNode) {
		if node.Test != nil {
			env := r.envForNode(node)
			fixtureGroup.Add(node.Test.String(), func(ctx flanksourceContext.Context, t *task.Task) (FixtureResult, error) {
				if err := r.progress.Start(ctx, node); err != nil {
					return FixtureResult{Name: node.Name, Status: task.StatusERR, Error: err.Error()}, err
				}
				runContext := flanksourceContext.NewContext(withFixtureProgress(ctx, func(done, total int) error {
					return r.progress.Update(ctx, node, done, total)
				}))
				result, err := r.executeFixture(runContext, *node.Test, env)
				r.attachResult(node, result)
				if progressErr := r.progress.Complete(ctx, node, result); progressErr != nil {
					return result, progressErr
				}
				return result, err
			}, clicky.WithTaskTimeout(fixtureTimeout(*node.Test)))
		}
	})

	// Wait for all fixture tasks to complete and collect results
	groupResult := fixtureGroup.WaitFor()
	if groupResult.Error != nil {
		logger.Warnf("Some fixture tests failed: %v", groupResult.Error)
	}

	// Process results
	fixtureResults, err := fixtureGroup.GetResults()
	if err != nil {
		return nil, fmt.Errorf("failed to get fixture results: %w", err)
	}

	for _, result := range fixtureResults {
		// Create a FixtureNode for the result
		resultNode := FixtureNode{
			Name:    result.Name,
			Type:    TestNode,
			Results: &result,
		}
		results.Tests = append(results.Tests, resultNode)

	}

	r.tree.UpdateStatsRecursive()
	results.Summary = *r.tree.Stats

	if r.options.Display != nil {
		r.tree.ApplyDisplayOptions(*r.options.Display)
	}

	// Prune empty sections from the tree
	r.tree.PruneEmptySections()

	return results, nil
}

func fixtureTimeout(fixture FixtureTest) time.Duration {
	if fixture.Expected.Timeout != nil {
		return *fixture.Expected.Timeout
	}
	if fixture.Timeout != nil {
		return *fixture.Timeout
	}
	return 2 * time.Minute
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

func (r *Runner) attachResult(node *FixtureNode, result FixtureResult) {
	r.resultMu.Lock()
	defer r.resultMu.Unlock()
	if len(result.Children) == 0 {
		node.Results = &result
		return
	}
	for _, child := range result.Children {
		node.AddChild(child)
	}
}

// executeFixture runs a single fixture test in the environment prepared for the
// file that declared it (zero-valued when that file declared neither `setup:`
// nor `record:`). The dispatch itself is RunNode's: this only assembles the
// run's environment around it and stamps the runner's own presentation.
func (r *Runner) executeFixture(ctx flanksourceContext.Context, fixture FixtureTest, env fixtureEnv) (FixtureResult, error) {
	if r.options.WorkDir == "" {
		r.options.WorkDir, _ = os.Getwd()
	}
	ctx.Logger.V(5).Infof("Using CWD: %s", r.options.WorkDir)

	opts := RunOptions{
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
		Recorder:       r.recorderContext(env.file),
		Record:         r.effectiveRecord(fixture),
		DaemonPort:     r.daemonPort,
		Progress:       fixtureProgressFromContext(ctx),
		ExtraArgs: map[string]interface{}{
			"flanksource_context": ctx,
		},
	}

	start := time.Now()
	result := RunNode(ctx, fixture, opts)
	result.Duration = time.Since(start)
	if r.options.Display != nil {
		result.Display = r.options.Display
	}
	return result, nil
}
