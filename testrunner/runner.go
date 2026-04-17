package testrunner

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
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
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/baseline"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
	"github.com/samber/lo"
)

var exitStatusRe = regexp.MustCompile(`(?m)^exit status \d+\s*$`)

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

// TestOrchestrator orchestrates test execution and TODO syncing.
type TestOrchestrator struct {
	RunOptions
	registry *Registry
	streamer *TestStreamer
	selector *selectorContext
}

// RunOptions configures test execution behavior.
type RunOptions struct {
	SyncTodos     bool                  `json:"sync_todos,omitempty" flag:"sync-todos"`                       // Whether to sync test failures to TODOs
	StartingPaths []string              `json:"starting_paths,omitempty" args:"true"`                         // Package paths to test (e.g., ["./pkg/testrunner"]). If empty, all packages are discovered.
	ExtraArgs     []string              `json:"extra_args,omitempty" flag:"extra-args"`                       // Additional arguments to pass to test runners (e.g., ["--focus", "TestName"])
	ShowPassed    bool                  `json:"show_passed,omitempty" flag:"show-passed"`                     // Whether to show passed tests in output
	ShowStdout    OutputMode            `json:"show_stdout,omitempty" flag:"show-stdout" default:"OnFailure"` // When to show stdout: false|Never, OnFailure (default), true|Always
	ShowStderr    OutputMode            `json:"show_stderr,omitempty" flag:"show-stderr" default:"OnFailure"` // When to show stderr: false|Never, OnFailure (default), true|Always
	TodosDir      string                `json:"todos_dir,omitempty" flag:"todos-dir" default:".todos"`        // Directory to store TODO files (default: .todos/)
	TodoTemplate  string                `json:"todo_template,omitempty" flag:"todo-template"`                 // Path to TODO template file
	WorkDir       string                `json:"work_dir,omitempty" flag:"work-dir"`                           // Working directory to run tests in
	DryRun        bool                  `json:"dry_run,omitempty" flag:"dry-run"`                             // Show what tests would be executed without running them
	Recursive     bool                  `json:"recursive,omitempty" flag:"recursive" default:"true"`          // Recursively discover test packages in subdirectories
	Nodes         int                   `json:"nodes,omitempty" flag:"nodes" short:"p"`                       // Number of parallel ginkgo nodes (0 = default, -1 = auto)
	UI            bool                  `json:"ui,omitempty" flag:"ui"`                                       // Launch browser with real-time task progress dashboard
	Addr          string                `json:"addr,omitempty" flag:"addr" default:"localhost"`               // Interface to bind --ui HTTP server. Use 0.0.0.0 to expose on the LAN.
	Diagnostics   bool                  `json:"diagnostics,omitempty" flag:"diagnostics"`                     // Capture a final diagnostics snapshot and embed it in JSON results / detached UI handoff artifacts.
	SkipHooks     bool                  `json:"skip_hooks,omitempty" flag:"skip-hooks"`                       // When true, .gavel.yaml pre/post hooks do not run. Default is computed in runTests from $CI: skip when unset, run when set.
	AutoStop      time.Duration         `json:"auto_stop,omitempty"`                                          // Hard wall-clock deadline for the detached UI child. Passed through when --detach is set. 0 = use default (30m). Flag wired imperatively from cmd/gavel/test.go because clicky doesn't bind time.Duration.
	IdleTimeout   time.Duration         `json:"idle_timeout,omitempty"`                                       // Idle deadline for the detached UI child; resets on every HTTP request. 0 = use default (5m). Only meaningful with --detach.
	Lint          bool                  `json:"lint,omitempty" flag:"lint"`                                   // Run linters in parallel with tests
	Cache         bool                  `json:"cache,omitempty" flag:"cache"`                                 // Skip packages whose content fingerprint matches the last passing run
	Changed       bool                  `json:"changed,omitempty" flag:"changed"`                             // Only run packages affected by staged/unstaged/untracked changes and the diff against origin/main
	Since         string                `json:"since,omitempty" flag:"since"`                                 // Only run packages affected by the diff since <ref> (merge-base(HEAD,ref)..HEAD) plus the working tree
	Bench         string                `json:"bench,omitempty" flag:"bench"`                                 // Run Go benchmarks matching this regex ("." or "true" runs all). Auto-enabled for packages containing only Benchmark* funcs.
	Fixtures      bool                  `json:"fixtures,omitempty" flag:"fixtures"`                           // Discover and run fixture files. Off by default; can also be enabled via .gavel.yaml fixtures.enabled
	FixtureFiles  []string              `json:"fixture_files,omitempty" flag:"fixture-files"`                 // Globs for fixture discovery. Overrides .gavel.yaml fixtures.files. Default: **/*.fixture.md
	Frameworks    []string              `json:"frameworks,omitempty" flag:"framework"`                        // Restrict execution to these frameworks (e.g. jest, vitest, playwright, go, ginkgo). Empty = run every detected framework. Unknown names hard-fail.
	Baseline      string                `json:"baseline,omitempty" flag:"baseline"`                           // Path to previous results JSON; only report NEW failures not in baseline
	Failed        string                `json:"failed,omitempty" flag:"failed"`                               // Path to previous results JSON; re-run only failed tests
	Updates       chan<- []parsers.Test `json:"-"`                                                            // Channel for streaming test result updates to UI
	OutputTee     io.Writer             `json:"-"`                                                            // Optional writer that receives a copy of raw process stdout/stderr
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
	if len(opts.FixtureFiles) > 0 {
		text = text.Space().Append("FixtureFiles: ", "text-muted").Append(clicky.CompactList(opts.FixtureFiles), "text-blue-500")
	}
	return text
}

// hasChangeSelector reports whether any flag in opts narrows execution to a
// subset of packages via the change graph.
func (opts RunOptions) hasChangeSelector() bool {
	return opts.Changed || opts.Since != ""
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
	if moduleRoot := findNearestGoModRoot(abs); moduleRoot != "" && isWithin(moduleRoot, groupRoot) {
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

func isWithin(path, root string) bool {
	if path == root {
		return true
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func Run(opts RunOptions) (any, error) {
	if opts.WorkDir == "" {
		opts.WorkDir, _ = os.Getwd()
	}

	// Split starting paths by execution root so each group runs with the
	// correct WorkDir. Nested Go modules get their own groups.
	groups, err := groupPathsByGitRoot(opts.WorkDir, opts.StartingPaths)
	if err != nil {
		return nil, err
	}
	if len(groups) > 1 {
		return runMultiRoot(opts, groups)
	}

	// Single-root fast path: update WorkDir/StartingPaths from the (possibly
	// rebased) group so the rest of the function sees normalised values.
	if len(groups) == 1 {
		opts.WorkDir = groups[0].workDir
		opts.StartingPaths = groups[0].paths
	}

	return runSingleRoot(opts)
}

// runMultiRoot handles the case where starting paths span multiple git roots
// by running a separate orchestration per root and merging the results.
func runMultiRoot(base RunOptions, groups []testGroup) (any, error) {
	if base.Updates != nil {
		defer close(base.Updates)
	}

	var allTree []parsers.Test
	for _, g := range groups {
		opts := base
		opts.WorkDir = g.workDir
		opts.StartingPaths = g.paths
		result, err := runSingleRootWithUpdateOwnership(opts, false)
		if err != nil {
			return nil, err
		}
		if tests, ok := result.([]parsers.Test); ok {
			allTree = append(allTree, tests...)
		}
	}
	return allTree, nil
}

func runSingleRoot(opts RunOptions) (any, error) {
	return runSingleRootWithUpdateOwnership(opts, true)
}

func runSingleRootWithUpdateOwnership(opts RunOptions, closeUpdates bool) (any, error) {
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
	results, err := t.Run()
	if err != nil {
		return nil, err
	}
	tests := []parsers.Test{}
	showAll := opts.UI || opts.ShowPassed
	for _, result := range results {
		for _, test := range result.Tests {
			if test.Failed || showAll {
				if !opts.UI {
					if !opts.ShowStdout.ShouldShow(test.Failed) {
						test.Stdout = ""
					}
					if !opts.ShowStderr.ShouldShow(test.Failed) {
						test.Stderr = ""
					}
				}
				tests = append(tests, test)
			}
		}
	}

	// Build hierarchical tree from flat test results
	tree := parsers.BuildTestTree(tests)

	// Discover and run fixture tests, gated on the --fixtures flag or
	// fixtures.enabled in .gavel.yaml. Fixture failures are captured in the
	// tree, not returned as errors, so AddNamedCommand still prints results.
	if globs := resolveFixtureGlobs(opts); globs != nil {
		fixtureTree, err := runDiscoveredFixtures(opts.WorkDir, opts.StartingPaths, globs, streamer)
		if err != nil {
			logger.Warnf("fixture discovery failed: %v", err)
		}
		if fixtureTree != nil {
			for _, child := range fixtureTree.Children {
				tree = append(tree, fixtureNodeToTests(child)...)
			}
		}
	}

	tree = annotateTestsWorkDir(tree, opts.WorkDir)

	return tree, nil
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

// Run executes tests and optionally syncs failures to TODOs.
// Returns []parsers.TestResults with all package test results (passed, failed, skipped), or an error.
func (o *TestOrchestrator) Run() (parsers.TestSuiteResults, error) {
	frameworks, err := o.registry.DetectAll()
	if err != nil {
		return nil, fmt.Errorf("failed to detect frameworks: %w", err)
	}

	if len(o.Frameworks) > 0 {
		frameworks, err = filterFrameworks(frameworks, o.Frameworks)
		if err != nil {
			return nil, err
		}
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

// filterFrameworks narrows detected frameworks to the set the user asked for
// via --framework. Each requested name must resolve to a known framework
// (parsers.ParseFramework handles aliases and errors on typos). A requested
// framework that exists but wasn't detected in this workdir is also an error,
// so `gavel test --framework jest` in a repo without jest fails loudly instead
// of silently running nothing.
func filterFrameworks(detected []Framework, requested []string) ([]Framework, error) {
	detectedSet := make(map[Framework]bool, len(detected))
	for _, fw := range detected {
		detectedSet[fw] = true
	}
	var kept []Framework
	seen := make(map[Framework]bool, len(requested))
	for _, name := range requested {
		fw, err := parsers.ParseFramework(name)
		if err != nil {
			return nil, err
		}
		if seen[fw] {
			continue
		}
		seen[fw] = true
		if !detectedSet[fw] {
			return nil, fmt.Errorf("framework %q not detected in workdir", fw)
		}
		kept = append(kept, fw)
	}
	return kept, nil
}

// applyFailedFilter loads a previous Snapshot and narrows packagesByFramework
// to only packages that had failures. It also injects framework-specific
// --run/--focus args to target specific failed tests.
func applyFailedFilter(packagesByFramework map[Framework][]string, extraArgs []string, failedPath string) (map[Framework][]string, []string, error) {
	snapshot, err := baseline.LoadSnapshot(failedPath)
	if err != nil {
		return nil, nil, fmt.Errorf("--failed: %w", err)
	}
	failedPkgs := baseline.ExtractFailedTestPackages(snapshot.Tests)
	if len(failedPkgs) == 0 {
		return nil, nil, fmt.Errorf("--failed: no failed tests found in %s", failedPath)
	}

	filtered := make(map[Framework][]string)
	var goTestNames, ginkgoTestNames []string

	for fw, pkgs := range packagesByFramework {
		failedForFW, ok := failedPkgs[parsers.Framework(fw)]
		if !ok {
			continue
		}
		pkgSet := make(map[string]bool, len(failedForFW))
		for pkgPath := range failedForFW {
			pkgSet[pkgPath] = true
		}
		for _, pkg := range pkgs {
			if pkgSet[pkg] {
				filtered[fw] = append(filtered[fw], pkg)
			}
		}
		for _, names := range failedForFW {
			switch parsers.Framework(fw) {
			case parsers.GoTest:
				goTestNames = append(goTestNames, names...)
			case parsers.Ginkgo:
				ginkgoTestNames = append(ginkgoTestNames, names...)
			}
		}
	}

	if len(goTestNames) > 0 {
		pattern := "^(" + strings.Join(escapeRegexNames(goTestNames), "|") + ")$"
		extraArgs = append(extraArgs, "-run", pattern)
	}
	if len(ginkgoTestNames) > 0 {
		pattern := strings.Join(ginkgoTestNames, "|")
		extraArgs = append(extraArgs, "--focus", pattern)
	}

	logger.Infof("--failed: narrowed to %d frameworks from %s", len(filtered), failedPath)
	return filtered, extraArgs, nil
}

func escapeRegexNames(names []string) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = regexp.QuoteMeta(n)
	}
	return out
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
			return nil, fmt.Errorf("no runner registered for framework %s, registry = %s", fw, renderText(o.registry.Pretty()))
		}

		var packages []string
		var err error

		// Discover packages starting from specified paths or entire workDir
		if len(startingPaths) > 0 {
			packages, err = o.discoverPackagesInPaths(runner, startingPaths)
		} else {
			packages, err = runner.DiscoverPackages(o.WorkDir, o.Recursive)
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

	// Narrow via change graph and/or run cache. Both selectors share a
	// single lazily-loaded graph/hasher instance via selectorContext.
	if o.hasChangeSelector() || o.Cache {
		var err error
		o.selector, err = newSelectorContext(o.WorkDir)
		if err != nil {
			return nil, fmt.Errorf("initialize selector: %w", err)
		}
	}

	if o.hasChangeSelector() && o.selector != nil {
		diff := diffOptionsFromRunOptions(o.RunOptions)
		for fw, pkgs := range packagesByFramework {
			kept, err := o.selector.filterByChangeGraph(pkgs, diff)
			if err != nil {
				return nil, fmt.Errorf("change-graph filter for %s: %w", fw, err)
			}
			if len(kept) != len(pkgs) {
				logger.Infof("[%s] change-graph: %d → %d packages", fw, len(pkgs), len(kept))
			}
			packagesByFramework[fw] = kept
		}
	}

	cachedByFramework := map[Framework][]cacheHit{}
	if o.Cache && o.selector != nil {
		for fw, pkgs := range packagesByFramework {
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
		var err error
		packagesByFramework, extraArgs, err = applyFailedFilter(packagesByFramework, extraArgs, o.Failed)
		if err != nil {
			return nil, err
		}
	}

	// If dry-run mode, display what would be executed and return early
	if o.DryRun {
		return o.displayDryRun(packagesByFramework, extraArgs), nil
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

	// Create task group to orchestrate parallel package test execution
	group := task.StartGroup[packageResult]("Running tests across packages")

	// Launch a task for each package per framework
	for _, fw := range frameworks {
		packages := packagesByFramework[fw]
		runner, ok := o.registry.Get(fw)
		if !ok {
			continue
		}

		fwExtraArgs := extraArgs
		if fw == parsers.Ginkgo && o.Nodes != 0 {
			fwExtraArgs = append([]string{fmt.Sprintf("--nodes=%d", o.Nodes)}, extraArgs...)
		}

		for _, pkgPath := range packages {
			// Capture variables for closure
			pkgPath := pkgPath
			fw := fw
			runner := runner
			pkgExtraArgs := fwExtraArgs

			if fw == parsers.GoTest {
				pkgExtraArgs = o.augmentBenchArgs(runner, pkgPath, pkgExtraArgs)
			}

			taskName := fmt.Sprintf("%s %s", fw, pkgPath)
			group.Add(taskName, func(ctx commonsCtx.Context, t *task.Task) (packageResult, error) {
				result, err := o.runPackageTest(ctx, pkgPath, fw, runner, t, pkgExtraArgs)
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
	// Execute the test process in the orchestrator
	process := testRun.Process.WithTask(t)
	// Keep the RUNNING-state label short; it may be rendered if the
	// task group dumps dirty tasks before the package finishes.
	t.SetName(formatRunningCommand(testRun.Process.Cmd, testRun.Process.Args))
	runStart := time.Now()
	result := process.Run().Result()
	runDuration := time.Since(runStart)

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

		fallback := parsers.Test{
			Failed:  true,
			Name:    fmt.Sprintf("%s Execution", framework),
			Message: message,
			Stderr:  result.Stderr,
			Stdout:  result.Stdout,
		}
		testResults = parsers.TestSuiteResults{{
			Command:   process.Cmd,
			Framework: testRun.Framework,
			Stderr:    result.Stderr,
			Stdout:    result.Stdout,
			ExitCode:  result.ExitCode,
			Tests:     parsers.Tests{fallback},
		}}

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
		}, nil
	}

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
	}, nil
}

// augmentBenchArgs decides whether benchmarks should run for a specific Go test package
// and returns the adjusted extraArgs with `-bench=<pattern>` (and `-run=^$` when the
// package has no regular Test* funcs) appended. It is a no-op when:
//   - any of the incoming extraArgs already contain `-bench` (user override)
//   - bench is neither requested via RunOptions.Bench nor auto-enabled for this package
func (o *TestOrchestrator) augmentBenchArgs(r runners.Runner, pkgPath string, extraArgs []string) []string {
	for _, a := range extraArgs {
		if a == "-bench" || strings.HasPrefix(a, "-bench=") {
			return extraArgs
		}
	}

	gt, ok := r.(*runners.GoTest)
	if !ok {
		return extraArgs
	}

	pattern := strings.TrimSpace(o.Bench)
	switch pattern {
	case "", "false":
		// Auto-enable only when the package has benchmarks but no runnable tests.
		if gt.PackageHasBenchmarks(pkgPath) && !gt.PackageHasGoTests(pkgPath) {
			pattern = "."
		} else {
			return extraArgs
		}
	case "true":
		pattern = "."
	}

	if !gt.PackageHasBenchmarks(pkgPath) {
		return extraArgs
	}

	args := append([]string{}, extraArgs...)
	args = append(args, "-bench="+pattern)
	if !gt.PackageHasGoTests(pkgPath) {
		hasRun := false
		for _, a := range extraArgs {
			if a == "-run" || strings.HasPrefix(a, "-run=") {
				hasRun = true
				break
			}
		}
		if !hasRun {
			args = append(args, "-run=^$")
		}
	}
	return args
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

		packages, err := runner.DiscoverPackages(fullPath, o.Recursive)
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

	// Set package path and stderr on each test
	processStderr := stripExitStatus(strings.TrimSpace(result.Stderr))
	for i := range tests {
		tests[i].PackagePath = pkgPath
		tests[i].Command = testRun.Process.Cmd + " " + strings.Join(testRun.Process.Args, " ")
		if tests[i].Stderr == "" {
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

		for _, pkg := range packages {
			pkgExtraArgs := extraArgs
			if framework == parsers.GoTest {
				pkgExtraArgs = o.augmentBenchArgs(runner, pkg, pkgExtraArgs)
			}
			testRun, err := runner.BuildCommand(pkg, pkgExtraArgs...)
			if err != nil {
				logger.Errorf("[test:%s] %s: build failed: %v", framework, pkg, err)
				continue
			}
			PrintDryRunCommand(
				fmt.Sprintf("test:%s", framework),
				pkg,
				testRun.Process.Cmd,
				testRun.Process.Args,
				testRun.Process.Cwd,
			)
		}
	}

	return nil
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
	cfg, err := verify.LoadGavelConfig(opts.WorkDir)
	if err != nil {
		logger.Warnf("failed to load .gavel.yaml for fixture config: %v", err)
	}
	if !opts.Fixtures && !cfg.Fixtures.Enabled {
		return nil
	}
	if len(opts.FixtureFiles) > 0 {
		return opts.FixtureFiles
	}
	return cfg.Fixtures.ResolvedFiles()
}

// discoverFixtures resolves each glob against every starting path (or workDir
// when no paths are given) and returns the union of matching files.
// Absolute globs are used as-is; relative globs are joined with the path root.
func discoverFixtures(workDir string, startingPaths []string, globs []string) []string {
	if len(globs) == 0 {
		return nil
	}
	roots := startingPaths
	if len(roots) == 0 {
		roots = []string{workDir}
	}

	var patterns []string
	for _, root := range roots {
		if !filepath.IsAbs(root) {
			root = filepath.Join(workDir, root)
		}
		for _, g := range globs {
			if filepath.IsAbs(g) {
				patterns = append(patterns, g)
			} else {
				patterns = append(patterns, filepath.Join(root, g))
			}
		}
	}
	return globFixtures(patterns)
}

func globFixtures(patterns []string) []string {
	var found []string
	for _, pattern := range patterns {
		matches, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			continue
		}
		found = append(found, matches...)
	}
	return lo.Uniq(found)
}

func runDiscoveredFixtures(workDir string, startingPaths []string, globs []string, streamer *TestStreamer) (*fixtures.FixtureNode, error) {
	fixtureFiles := discoverFixtures(workDir, startingPaths, globs)
	if len(fixtureFiles) == 0 {
		return nil, nil
	}

	execPath, _ := os.Executable()
	runnerOpts := fixtures.RunnerOptions{
		Paths:          fixtureFiles,
		WorkDir:        workDir,
		Logger:         logger.StandardLogger(),
		ExecutablePath: execPath,
	}

	if streamer != nil {
		runnerOpts.OnParsed = func(tree *fixtures.FixtureNode) {
			streamer.SetFixtureOutline(fixtureTreeToPending(tree))
		}
	}

	runner, err := fixtures.NewRunner(runnerOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create fixture runner: %w", err)
	}

	if streamer != nil {
		runner.SetOnResult(func(_ fixtures.FixtureResult) {
			tree := runner.Tree()
			var fixtureTests []parsers.Test
			for _, child := range tree.Children {
				fixtureTests = append(fixtureTests, fixtureNodeToTests(child)...)
			}
			streamer.UpdateFixtures(fixtureTests)
		})
	}

	result, err := runner.Run()

	if streamer != nil && result != nil {
		var fixtureTests []parsers.Test
		for _, child := range result.Children {
			fixtureTests = append(fixtureTests, fixtureNodeToTests(child)...)
		}
		streamer.UpdateFixtures(fixtureTests)
	}

	return result, err
}

func fixtureNodeToTests(node *fixtures.FixtureNode) []parsers.Test {
	if node.Results != nil {
		r := node.Results
		t := parsers.Test{
			Name:      node.Name,
			Framework: "fixture",
			Duration:  r.Duration,
			Stdout:    r.Stdout,
			Stderr:    r.Stderr,
			Failed:    r.Status == task.StatusFAIL || r.Status == task.StatusFailed || r.Status == task.StatusERR,
			Passed:    r.Status == task.StatusPASS || r.Status == task.StatusSuccess,
			Message:   r.Error,
			Context: parsers.FixtureContext{
				Command:       r.Command,
				ExitCode:      r.ExitCode,
				CWD:           r.CWD,
				CELExpression: r.CELExpression,
				CELVars:       r.CELVars,
				Expected:      r.Expected,
				Actual:        r.Actual,
			},
		}
		return []parsers.Test{t}
	}

	var children parsers.Tests
	for _, child := range node.Children {
		children = append(children, fixtureNodeToTests(child)...)
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
			Framework: "fixture",
			Children:  children,
			Failed:    failed,
		}}
	}
	return children
}

func fixtureTreeToPending(node *fixtures.FixtureNode) []parsers.Test {
	if node == nil {
		return nil
	}
	var result []parsers.Test
	for _, child := range node.Children {
		result = append(result, fixtureNodeToPending(child)...)
	}
	return result
}

func fixtureNodeToPending(node *fixtures.FixtureNode) []parsers.Test {
	if node.Test != nil {
		return []parsers.Test{{
			Name:      node.Name,
			Framework: "fixture",
			Pending:   true,
		}}
	}
	var children parsers.Tests
	for _, child := range node.Children {
		children = append(children, fixtureNodeToPending(child)...)
	}
	if node.Type == fixtures.FileNode || node.Type == fixtures.SectionNode {
		return []parsers.Test{{
			Name:      node.Name,
			Framework: "fixture",
			Children:  children,
			Pending:   true,
		}}
	}
	return children
}

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
