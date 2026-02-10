package testrunner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	commonsCtx "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
	"github.com/samber/lo"
)

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

// TestOrchestrator orchestrates test execution and TODO syncing.
type TestOrchestrator struct {
	RunOptions
	registry *Registry
}

// RunOptions configures test execution behavior.
type RunOptions struct {
	SyncTodos     bool       `json:"sync_todos,omitempty" flag:"sync-todos"`                       // Whether to sync test failures to TODOs
	StartingPaths []string   `json:"starting_paths,omitempty" args:"true"`                         // Package paths to test (e.g., ["./pkg/testrunner"]). If empty, all packages are discovered.
	ExtraArgs     []string   `json:"extra_args,omitempty" flag:"extra-args"`                       // Additional arguments to pass to test runners (e.g., ["--focus", "TestName"])
	ShowPassed    bool       `json:"show_passed,omitempty" flag:"show-passed"`                     // Whether to show passed tests in output
	ShowStdout    OutputMode `json:"show_stdout,omitempty" flag:"show-stdout" default:"OnFailure"` // When to show stdout: false|Never, OnFailure (default), true|Always
	ShowStderr    OutputMode `json:"show_stderr,omitempty" flag:"show-stderr" default:"OnFailure"` // When to show stderr: false|Never, OnFailure (default), true|Always
	TodosDir      string     `json:"todos_dir,omitempty" flag:"todos-dir" default:".todos"`        // Directory to store TODO files (default: .todos/)
	TodoTemplate  string     `json:"todo_template,omitempty" flag:"todo-template"`                 // Path to TODO template file
	WorkDir       string     `json:"work_dir,omitempty" flag:"work-dir"`                           // Working directory to run tests in
	DryRun        bool       `json:"dry_run,omitempty" flag:"dry-run"`                             // Show what tests would be executed without running them
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
	if opts.SyncTodos {
		text = text.Space().Append("SyncTodos: ", "text-muted").Append(icons.Check, "text-green-500")
		text = text.Space().Append("Dir: ", "text-muted").Append(opts.TodosDir, "text-blue-500")
		if opts.TodoTemplate != "" {
			text = text.Space().Append("TodoTemplate: ", "text-muted").Append(opts.TodoTemplate, "text-blue-500")
		}
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
	return text
}

func (r RunOptions) Help() string {
	return `Run all test frameworks in the project (go test and Ginkgo).

Automatically detects which frameworks are present and runs them.
Captures output in JSON format for reliable parsing.
Optionally syncs failures to TODO files in .todos/ directory.

Optionally accepts package paths as arguments to run only tests in those packages.
Additional arguments after "--" are passed to the test runners (e.g., ginkgo flags).

Examples:
  arch-unit test
  arch-unit test ./pkg/testrunner
  arch-unit test ./pkg/testrunner ./cmd
  arch-unit test ./pkg/testrunner -- --focus "TestName"
  arch-unit test . -- --focus "Integration" --skip "Slow"`
}

// packageResult contains results from running tests in a single package
type packageResult struct {
	packagePath string
	framework   Framework
	testResults parsers.TestSuiteResults
	err         error
}

func Run(opts RunOptions) (any, error) {
	if opts.WorkDir == "" {
		opts.WorkDir, _ = os.Getwd()
	}
	logger.Infof("Running tests  %s", opts.Pretty().ANSI())

	t := &TestOrchestrator{
		RunOptions: opts,
		registry:   DefaultRegistry(opts.WorkDir),
	}
	results, err := t.Run()
	if err != nil {
		return nil, err
	}
	tests := []parsers.Test{}
	for _, result := range results {
		for _, test := range result.Tests {
			if test.Failed || opts.ShowPassed {
				if !opts.ShowStdout.ShouldShow(test.Failed) {
					test.Stdout = ""
				}
				if !opts.ShowStderr.ShouldShow(test.Failed) {
					test.Stderr = ""
				}
				tests = append(tests, test)
			}
		}
	}

	// Build hierarchical tree from flat test results
	tree := parsers.BuildTestTree(tests)
	return tree, nil
}

// Run executes tests and optionally syncs failures to TODOs.
// Returns []parsers.TestResults with all package test results (passed, failed, skipped), or an error.
func (o *TestOrchestrator) Run() (parsers.TestSuiteResults, error) {
	frameworks, err := o.registry.DetectAll()
	if err != nil {
		return nil, fmt.Errorf("failed to detect frameworks: %w", err)
	}

	results, err := o.detectAndRun(frameworks, o.StartingPaths, o.ExtraArgs)
	failed := results.Failed()
	logger.Infof("Test run completed. %d total failures.", len(failed))

	if o.SyncTodos && len(failed) > 0 {
		if err := o.syncTodos(failed); err != nil {
			return nil, fmt.Errorf("failed to sync todos: %w", err)
		}
	}
	if err != nil {
		return results, err
	}

	return results, nil
}

func (o *TestOrchestrator) detectAndRun(frameworks []Framework, startingPaths []string, extraArgs []string) (parsers.TestSuiteResults, error) {
	if len(frameworks) == 0 {
		return nil, fmt.Errorf("no frameworks provided for test execution")
	}
	// Validate that starting paths exist
	if len(startingPaths) > 0 {
		for _, path := range startingPaths {
			fullPath := path
			if !filepath.IsAbs(path) {
				fullPath = filepath.Join(o.WorkDir, path)
			}
			info, err := os.Stat(fullPath)
			if err != nil {
				return nil, fmt.Errorf("starting path %s does not exist: %w", path, err)
			}
			if !info.IsDir() {
				return nil, fmt.Errorf("starting path %s is not a directory", path)
			}
		}
	}

	// Discover all unique packages across all frameworks
	packagesByFramework := make(map[Framework][]string)

	for _, fw := range frameworks {
		runner, ok := o.registry.Get(fw)
		if !ok {
			return nil, fmt.Errorf("no runner registered for framework %s, registry = %s", fw, o.registry.Pretty().ANSI())
		}

		var packages []string
		var err error

		// Discover packages starting from specified paths or entire workDir
		if len(startingPaths) > 0 {
			packages, err = o.discoverPackagesInPaths(runner, startingPaths)
		} else {
			packages, err = runner.DiscoverPackages(o.WorkDir)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to discover packages for %s: %w", fw, err)
		}

		packagesByFramework[fw] = packages
	}

	for framework, items := range packagesByFramework {
		logger.V(4).Infof("[%s] discovered %s", framework, clicky.CompactList(items).ANSI())
	}

	// Collect all packages for error checking
	var allPackages []string
	for _, packages := range packagesByFramework {
		allPackages = append(allPackages, packages...)
	}

	if len(allPackages) == 0 {
		if len(startingPaths) > 0 {
			return nil, fmt.Errorf("no test packages found in starting paths %v", startingPaths)
		}
		return nil, fmt.Errorf("no test packages found")
	}

	// If dry-run mode, display what would be executed and return early
	if o.DryRun {
		return o.displayDryRun(packagesByFramework, extraArgs), nil
	}

	// Create task group to orchestrate parallel package test execution
	group := task.StartGroup[packageResult]("Running tests across packages")

	// Launch a task for each package per framework
	for _, fw := range frameworks {
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

			taskName := fmt.Sprintf("%s %s", fw, pkgPath)
			group.Add(taskName, func(ctx commonsCtx.Context, t *task.Task) (packageResult, error) {
				return o.runPackageTest(ctx, pkgPath, fw, runner, t, extraArgs)
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

	allResults.Sort()

	return allResults, nil
}

func (o *TestOrchestrator) runPackageTest(
	ctx commonsCtx.Context,
	pkgPath string,
	framework Framework,
	runner runners.Runner, // runners.Runner interface
	t *task.Task,
	extraArgs []string,
) (packageResult, error) {
	// Build the test command (without executing)
	testRun, err := runner.BuildCommand(pkgPath, extraArgs...)
	if err != nil {
		t.Errorf("Failed to build command: %v", err)
		t.Failed()
		return packageResult{
			packagePath: pkgPath,
			framework:   framework,
			err:         fmt.Errorf("failed to build command: %w", err),
		}, nil
	}

	testRun.Process.SucceedOnNonZero = true
	// Execute the test process in the orchestrator
	process := testRun.Process.WithTask(t)
	t.SetName(fmt.Sprintf("%s %s", framework, pkgPath))
	result := process.Run().Result()

	// Always attempt to parse test results first, even if there was an execution error
	// This handles cases like `go test -json` which outputs valid JSON even when tests fail
	testResults, parseErr := o.parseTestResults(testRun, result, pkgPath)

	// If parsing failed or returned no tests, create a fallback execution test
	if parseErr != nil || len(testResults.All()) == 0 {
		var message string
		if parseErr != nil {
			message = fmt.Sprintf("%s: %s", parseErr.Error(), result.Output())
		} else {
			message = strings.TrimSpace(result.Output())
		}

		testResults = parsers.TestSuiteResults{{
			Command:   process.Cmd,
			Framework: testRun.Framework,
			Stderr:    result.Stderr,
			Stdout:    result.Stdout,
			ExitCode:  result.ExitCode,
			Tests: parsers.Tests{{
				Failed:  true,
				Name:    fmt.Sprintf("%s Execution", framework),
				Message: message,
			}},
		}}

		if parseErr != nil {
			t.Errorf("Failed to parse results: %v", parseErr)
		}
		if result.Error != nil {
			t.Errorf("Execution error: %v", result.Error)
		}
		t.Failed()

		return packageResult{
			packagePath: pkgPath,
			framework:   framework,
			testResults: testResults,
		}, nil
	}

	// Update task status based on results
	sum := testResults.All().Sum()
	t.SetName(clicky.Text("").Append(framework, "text-muted font-medium").Space().Append(pkgPath, "text-blue-500").Space().Append(sum).ANSI())

	if sum.Failed > 0 || sum.Skipped > 0 {
		t.Failed()
	} else {
		t.Success()
	}

	return packageResult{
		packagePath: pkgPath,
		framework:   framework,
		testResults: testResults,
	}, nil
}

// discoverPackagesInPaths discovers packages only within the specified starting paths.
func (o *TestOrchestrator) discoverPackagesInPaths(runner runners.Runner, startingPaths []string) ([]string, error) {
	var allPackages []string

	for _, startPath := range startingPaths {
		fullPath := startPath
		if !filepath.IsAbs(startPath) {
			fullPath = filepath.Join(o.WorkDir, startPath)
		}
		logger.Debugf("Discovering packages in path: %s", fullPath)

		packages, err := runner.DiscoverPackages(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to discover packages in %s: %w", startPath, err)
		}

		// Filter packages to only include those under the starting path
		relStartPath, err := filepath.Rel(o.WorkDir, fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get relative path for %s: %w", fullPath, err)
		}
		relStartPath = "./" + filepath.ToSlash(relStartPath)

		for _, pkg := range packages {
			// Check if package is under the starting path
			if strings.HasPrefix(pkg, relStartPath) {
				allPackages = append(allPackages, pkg)
			}
		}

	}

	return lo.Uniq(allPackages), nil
}

// parseTestResults parses test results from either stdout (go test) or report file (Ginkgo).
func (o *TestOrchestrator) parseTestResults(testRun *runners.TestRun, result *exec.ExecResult, pkgPath string) (parsers.TestSuiteResults, error) {
	var tests parsers.Tests
	var err error

	// For Ginkgo, parse from report file
	if testRun.Framework == parsers.Ginkgo && testRun.ReportPath != "" {
		tests, err = o.parseGinkgoReportFile(testRun.ReportPath)
		if err != nil {
			return nil, err
		}
	} else {
		// For go test, parse from stdout
		tests, err = testRun.Parser.Parse(strings.NewReader(result.Stdout))
		if err != nil {
			return nil, fmt.Errorf("failed to parse test results: %w", err)
		}
	}

	// Set package path and stderr on each test
	for i := range tests {
		tests[i].PackagePath = pkgPath
		if tests[i].Stderr == "" {
			tests[i].Stderr = strings.TrimSpace(result.Stderr)
		}
	}

	// Normalize file paths using runner's method
	runner, ok := o.registry.Get(testRun.Framework)
	if ok {
		switch r := runner.(type) {
		case interface{ NormalizeFilePath(string) string }:
			for i := range tests {
				if tests[i].File != "" {
					tests[i].File = r.NormalizeFilePath(tests[i].File)
				}
			}
		}
	}

	return parsers.TestSuiteResults{{
		Command:   testRun.Process.Cmd,
		Framework: testRun.Framework,
		Tests:     tests,
		Stdout:    result.Stdout,
		Stderr:    result.Stderr,
		ExitCode:  result.ExitCode,
	}}, nil
}

// parseGinkgoReportFile reads and parses a Ginkgo JSON report file.
func (o *TestOrchestrator) parseGinkgoReportFile(reportPath string) (parsers.Tests, error) {
	if o.WorkDir != "" && !filepath.IsAbs(reportPath) {
		reportPath = filepath.Join(o.WorkDir, reportPath)
	}
	if _, err := os.Stat(reportPath); err != nil {
		absPath, _ := filepath.Abs(reportPath)
		return parsers.Tests{{
			Name:    "Ginkgo Execution",
			Failed:  true,
			Message: fmt.Sprintf("'%s' not found", absPath),
		}}, nil
	}

	reportFile, err := os.Open(reportPath)
	if err != nil {
		return parsers.Tests{{
			Name:    "Ginkgo Execution",
			Failed:  true,
			Message: fmt.Sprintf("Failed to open ginkgo report: %v", err),
		}}, nil
	}
	defer reportFile.Close()

	parser := parsers.NewGinkgoJSON()
	tests, err := parser.Parse(reportFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ginkgo report: %w", err)
	}

	return tests, nil
}

// displayDryRun shows what tests would be executed without running them.
func (o *TestOrchestrator) displayDryRun(packagesByFramework map[Framework][]string, extraArgs []string) parsers.TestSuiteResults {
	logger.Infof("🔍 Dry-run mode: showing what would be executed\n")

	var totalPackages int
	for _, packages := range packagesByFramework {
		totalPackages += len(packages)
	}

	logger.Infof("Detected Frameworks: %s", clicky.CompactList(lo.Keys(packagesByFramework)).ANSI())
	logger.Infof("Total Packages: %d\n", totalPackages)

	for framework, packages := range packagesByFramework {
		runner, ok := o.registry.Get(framework)
		if !ok {
			continue
		}

		logger.Infof("[%s] Packages (%d):", framework, len(packages))
		for _, pkg := range packages {
			testRun, err := runner.BuildCommand(pkg, extraArgs...)
			if err != nil {
				logger.Errorf("  ❌ %s: %v", pkg, err)
				continue
			}
			logger.Infof("  %s", testRun.Pretty().ANSI())
		}
		logger.Infof("")
	}

	return nil
}

// Runner is an alias for TestOrchestrator for backwards compatibility during migration.
// Deprecated: use TestOrchestrator directly.
type Runner = TestOrchestrator

func (o *TestOrchestrator) syncTodos(failures []TestFailure) error {
	sync := NewTodoSync(o.TodosDir, o.TodoTemplate)

	for _, failure := range failures {
		failure := failure // Capture for goroutine
		taskName := fmt.Sprintf("Creating TODO for %s", failure.Name)

		taskFunc := func(ctx commonsCtx.Context, t *task.Task) (string, error) {
			return sync.SyncFailure(failure)
		}

		clickyTask := clicky.StartTask[string](taskName, taskFunc)
		todoPath, err := clickyTask.GetResult()

		if err != nil {
			clickyTask.Errorf("Failed to sync TODO: %v", err)
			clickyTask.Failed()
			return err
		}

		relPath, _ := filepath.Rel(o.WorkDir, todoPath)
		clickyTask.SetName(fmt.Sprintf("✓ %s", relPath))
		clickyTask.Success()
	}

	return nil
}
