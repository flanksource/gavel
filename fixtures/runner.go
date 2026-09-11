package fixtures

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	Paths          []string // Fixture file paths/patterns
	Format         string   // Output format: tree, table, json, yaml, csv
	Filter         string   // Filter tests by name pattern (glob)
	NoColor        bool     // Disable colored output
	WorkDir        string   // Working directory
	MaxWorkers     int      // Maximum number of parallel workers
	Logger         logger.Logger
	ExecutablePath string                                                // Path to the current executable (for fixtures to use)
	OnResult       func(FixtureResult)                                   // Called once after each logical fixture row is finalized
	ProgressSink   ProgressSink                                          // Receives immutable execution-tree snapshots
	UpdateGolden   bool                                                  // When true, mismatched @file expectations are rewritten with actual output instead of failing
	Display        *DisplayOptions                                       // Optional result visibility controls for CLI rendering
	Spec           *api.Spec                                             // Runtime options for embedded AI prompts
	ResolveSpec    func(context.Context) (api.ResolveSpecOptions, error) `json:"-" yaml:"-"` // Lazy authored layers and saved snapshot; Spec takes precedence
	// Record is the run-wide `--record` default, applied only to fixtures that
	// declared no `record:` of their own. An explicit `record: none` parses to an
	// empty (non-nil) Spec precisely so it outranks this.
	Record *record.Spec
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
}

// NewRunner creates a new fixture runner
func NewRunner(opts RunnerOptions) (*Runner, error) {
	// Create CEL evaluator
	evaluator, err := NewCELEvaluator()
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL evaluator: %w", err)
	}

	return &Runner{
		options:   opts,
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
	if _, err := r.prepareFixtureTree(); err != nil {
		return nil, err
	}
	r.progress = newExecutionTracker(r.tree, r.options.WorkDir, r.executionSteps(), r.options.ProgressSink)
	if err := r.progress.Publish(context.Background()); err != nil {
		return nil, fmt.Errorf("publish queued fixture progress: %w", err)
	}

	results, err := r.executeFixtures()
	if err != nil {
		return nil, fmt.Errorf("failed to execute fixtures: %w", err)
	}

	clicky.WaitForGlobalCompletion()

	if results.Summary.HasFailures() {
		return r.tree, fmt.Errorf("fixture tests failed")
	}

	return r.tree, nil
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
	if err := validateFixtureConfiguration(r.fixtures); err != nil {
		return nil, fmt.Errorf("invalid fixture configuration: %w", err)
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
func (r *Runner) executeFixtures() (*FixtureGroup, error) {
	results := &FixtureGroup{
		Tests:   make([]FixtureNode, 0, len(r.fixtures)),
		Summary: Stats{},
	}

	ctx := flanksourceContext.NewContext(context.Background())

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
	// Registered before the daemon's defer so LIFO stops the daemon first:
	// removing a worktree a live process is sitting in leaves a stale git
	// worktree registration behind.
	defer r.cleanupSetups()

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
		defer r.stopDaemon()
	}

	// Recorders start after the daemon so its port can be excluded from the
	// proxy, and their defer is registered after the daemon's so LIFO closes
	// them first — a daemon shutting down can still emit requests worth
	// recording.
	if err := r.prepareRecorders(); err != nil {
		return nil, err
	}
	defer r.closeRecorders()

	// Create typed task group for fixture execution
	fixtureGroup := task.StartGroup[FixtureResult]("Fixture Tests")

	taskToNodeMap := make(map[task.TypedTask[FixtureResult]]*FixtureNode)
	r.tree.Walk(func(node *FixtureNode) {
		if node.Test != nil {
			env := r.envForNode(node)
			typedTask := fixtureGroup.Add(node.Test.String(), func(ctx flanksourceContext.Context, t *task.Task) (FixtureResult, error) {
				if err := r.progress.Start(ctx, node); err != nil {
					return FixtureResult{Name: node.Name, Status: task.StatusERR, Error: err.Error()}, err
				}
				runContext := flanksourceContext.NewContext(withFixtureProgress(ctx, func(done, total int) error {
					return r.progress.Update(ctx, node, done, total)
				}))
				return r.executeLogicalFixture(runContext, *node.Test, env), nil
			}, clicky.WithTaskTimeout(fixtureTimeout(*node.Test)))
			taskToNodeMap[typedTask] = node
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

	finalized := make(map[task.TypedTask[FixtureResult]]*FixtureResult, len(fixtureResults))
	allResults := make([]*FixtureResult, 0, len(fixtureResults))
	for typedTask, result := range fixtureResults {
		result := result
		finalized[typedTask] = &result
		allResults = append(allResults, &result)
	}
	finalizeMetricComparisons(allResults)

	for typedTask, result := range finalized {
		// The task initially completes before cross-row comparisons are possible.
		// Replace its stored value so subsequent Clicky renders use the finalized
		// baseline-aware verdict rather than the pre-comparison result.
		typedTask.SetResult(*result)
		// Create a FixtureNode for the result
		resultNode := FixtureNode{
			Name:    result.Name,
			Type:    TestNode,
			Results: result,
		}
		results.Tests = append(results.Tests, resultNode)

		if testNode, exists := taskToNodeMap[typedTask]; exists {
			r.attachResult(testNode, *result)
			if err := r.progress.Complete(ctx, testNode, *result); err != nil {
				return nil, err
			}
		} else {
			logger.Warnf("No tree node found for task: %s", typedTask.Name())
		}
		if r.options.OnResult != nil {
			r.options.OnResult(*result)
		}
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

// executeLogicalFixture preserves one task/result per Markdown row while
// collecting serial samples within the row's shared task deadline.
func (r *Runner) executeLogicalFixture(ctx flanksourceContext.Context, fixture FixtureTest, env fixtureEnv) FixtureResult {
	if !fixture.hasSampleConfiguration() {
		result, _ := r.executeFixture(ctx, fixture, env)
		return result
	}
	if reason := fixture.ShouldSkip(); reason != "" {
		return FixtureResult{
			Name:   fixture.Name,
			Status: task.StatusSKIP,
			Test:   fixture,
			Error:  reason,
		}
	}

	started := time.Now()
	result := FixtureResult{
		Name:    fixture.Name,
		Test:    fixture,
		Samples: make([]FixtureSample, 0, fixture.repeatCount()),
		Outcomes: &FixtureOutcomes{
			Command: FixtureOutcome{Status: OutcomePASS},
		},
	}
	if fixture.Expected.CEL != "" {
		result.Outcomes.Assertions = &FixtureOutcome{Status: OutcomeNotEvaluated}
	}

	var rowError string
	for index := 1; index <= fixture.repeatCount(); index++ {
		if err := ctx.Err(); err != nil {
			rowError = fmt.Sprintf("fixture timeout exhausted before sample %d: %v", index, err)
			break
		}

		sampleFixture := fixture
		sampleFixture.Expected.CEL = ""
		sampleResult, _ := r.executeFixture(ctx, sampleFixture, env)
		if sampleResult.Status == task.StatusSKIP {
			return sampleResult
		}
		sample := FixtureSample{
			Index:    index,
			Duration: sampleResult.Duration,
			Command:  sampleResult.Command,
			CWD:      sampleResult.CWD,
			ExitCode: sampleResult.ExitCode,
			Stdout:   sampleResult.Stdout,
			Stderr:   sampleResult.Stderr,
			Outcome:  outcomeFromTaskStatus(sampleResult.Status, sampleResult.Error),
		}
		if sample.Outcome.Status != OutcomePASS {
			sample.Error = sampleResult.Error
		}

		result.Type = sampleResult.Type
		result.Command = sampleResult.Command
		result.CWD = sampleResult.CWD
		result.Stdout = sampleResult.Stdout
		result.Stderr = sampleResult.Stderr
		result.ExitCode = sampleResult.ExitCode
		result.Actual = sampleResult.Actual
		result.Recordings = append(result.Recordings, sampleResult.Recordings...)

		metrics := fixture.metricSpecs()
		if len(metrics) > 0 {
			sample.Metrics = make(map[string]MetricSample, len(metrics))
		}
		if sample.Outcome.Status == OutcomePASS {
			variables := EvaluationContext(&sampleResult, EvaluateOptions{CELVars: sampleResult.EvaluationVars})
			if fixture.Expected.CEL != "" || len(metrics) > 0 {
				result.Metadata = sampleResult.Metadata
			}
			if fixture.Expected.CEL != "" {
				assertionResult := EvaluateCEL(sampleResult, fixture.Expected.CEL, variables)
				celOutcome := AssertionOutcome{FixtureOutcome: outcomeFromTaskStatus(assertionResult.Status, assertionResult.Error)}
				celOutcome.Expression = fixture.Expected.CEL
				celOutcome.Result = assertionResult.celResult
				sample.CEL = &celOutcome
				if assertionResult.Status != task.StatusPASS && result.CELExpression == "" {
					result.CELExpression = fixture.Expected.CEL
					result.CELTrace = assertionResult.CELTrace
					result.CELVars = variables
				}
			}
			for _, metric := range metrics {
				value, err := extractMetric(metric, variables)
				if err != nil {
					sample.Metrics[metric.Name] = MetricSample{Status: OutcomeERR, Error: err.Error()}
				} else {
					sample.Metrics[metric.Name] = MetricSample{Status: OutcomePASS, Value: value}
				}
			}
		} else {
			if fixture.Expected.CEL != "" {
				sample.CEL = &AssertionOutcome{
					FixtureOutcome: FixtureOutcome{Status: OutcomeNotEvaluated},
					Expression:     fixture.Expected.CEL,
				}
			}
			for _, metric := range metrics {
				sample.Metrics[metric.Name] = MetricSample{Status: OutcomeNotEvaluated}
			}
		}

		result.Samples = append(result.Samples, sample)
		if sampleHasError(sample) {
			break
		}
	}

	result.Duration = time.Since(started)
	result.Outcomes.Command = aggregateCommandOutcome(result.Samples)
	if rowError != "" {
		result.Outcomes.Command.Status = OutcomeERR
		result.Outcomes.Command.Error = joinErrors(result.Outcomes.Command.Error, rowError)
	}
	if fixture.Expected.CEL != "" {
		result.Outcomes.Assertions = aggregateAssertionOutcome(result.Samples)
	}
	summarizeMetrics(&result)
	finalizeLogicalResult(&result)
	return result
}

func outcomeFromTaskStatus(status task.Status, err string) FixtureOutcome {
	outcome := FixtureOutcome{Error: err}
	switch status {
	case task.StatusPASS, task.StatusSuccess:
		outcome.Status = OutcomePASS
	case task.StatusFAIL, task.StatusFailed:
		outcome.Status = OutcomeFAIL
	default:
		outcome.Status = OutcomeERR
	}
	return outcome
}

func sampleHasError(sample FixtureSample) bool {
	if sample.Outcome.Status == OutcomeERR || (sample.CEL != nil && sample.CEL.Status == OutcomeERR) {
		return true
	}
	for _, metric := range sample.Metrics {
		if metric.Status == OutcomeERR {
			return true
		}
	}
	return false
}

func aggregateCommandOutcome(samples []FixtureSample) FixtureOutcome {
	outcome := FixtureOutcome{Status: OutcomePASS}
	for _, sample := range samples {
		if sample.Outcome.Status == OutcomeERR {
			outcome.Status = OutcomeERR
		} else if sample.Outcome.Status == OutcomeFAIL && outcome.Status != OutcomeERR {
			outcome.Status = OutcomeFAIL
		}
		if sample.Outcome.Error != "" {
			outcome.Error = joinErrors(outcome.Error, fmt.Sprintf("sample %d: %s", sample.Index, sample.Outcome.Error))
		}
	}
	return outcome
}

func aggregateAssertionOutcome(samples []FixtureSample) *FixtureOutcome {
	outcome := &FixtureOutcome{Status: OutcomeNotEvaluated}
	evaluated := false
	for _, sample := range samples {
		if sample.CEL == nil || sample.CEL.Status == OutcomeNotEvaluated {
			continue
		}
		evaluated = true
		if sample.CEL.Status == OutcomeERR {
			outcome.Status = OutcomeERR
		} else if sample.CEL.Status == OutcomeFAIL && outcome.Status != OutcomeERR {
			outcome.Status = OutcomeFAIL
		} else if outcome.Status == OutcomeNotEvaluated {
			outcome.Status = OutcomePASS
		}
		if sample.CEL.Error != "" {
			outcome.Error = joinErrors(outcome.Error, fmt.Sprintf("sample %d: %s", sample.Index, sample.CEL.Error))
		}
	}
	if !evaluated {
		outcome.Status = OutcomeNotEvaluated
	}
	return outcome
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
