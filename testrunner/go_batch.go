package testrunner

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	goruntime "runtime"
	"slices"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	commonsCtx "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/processruntime"
	"github.com/flanksource/gavel/slots"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
	"github.com/flanksource/gavel/utils"
)

type goBatchPlan struct {
	runtime   processruntime.Options
	opts      runners.TestOptions
	key       string
	packages  []string
	ginkgo    []string
	targets   []string
	skipEmpty []string
	framework Framework
	// coverProfile, when set, is the -coverprofile the batch writes.
	coverProfile string
}

// goBatchCount numbers the batches' coverage profiles under CoverDir.
var goBatchCount atomic.Int64

func (o *TestOrchestrator) unfocusedGoRun() bool {
	return !o.configurationScoped && len(o.StartingPaths) == 0 && (len(o.Frameworks) == 0 || o.UIRRunAll) && len(o.Ignore) == 0 &&
		!o.hasChangeSelector() && o.Failed == "" && len(o.PassThroughArgs) == 0 && o.Focus == "" && o.UIRSkippable == nil
}

// goPackageOptions is a Go-family package's options: its framework's, then
// the package's .gavel.yaml, then the UIR-proved skips.
func (o *TestOrchestrator) goPackageOptions(framework Framework, pkg string, opts runners.TestOptions) (processruntime.Options, runners.TestOptions, error) {
	runtime, configured, err := o.runtimeFor(framework, pkg)
	if err != nil {
		return runtime, opts, err
	}
	opts, err = configured.Merge(opts)
	if err != nil {
		return runtime, opts, fmt.Errorf("%s %s: %w", framework, pkg, err)
	}
	if skippable := o.UIRSkippable[pkg]; framework == parsers.GoTest && len(skippable) > 0 {
		opts.Skip = opts.Skip.Merge(runners.FileSelection{Exact: skippable})
	}
	return runtime, opts, nil
}

// goBatchKey groups packages that can share one go test command: the same
// options and, when the options filter tests, the same framework (go test
// and Ginkgo spell filters differently) and the same selected files.
func goBatchKey(framework Framework, pkg string, opts runners.TestOptions) (string, error) {
	opts.Packages = nil
	files, err := opts.PackageFiles(pkg, false)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(opts)
	if err != nil {
		return "", err
	}
	key := string(data)
	ginkgoExtra := slices.ContainsFunc(opts.ExtraArgs, func(arg string) bool { return strings.HasPrefix(strings.TrimLeft(arg, "-"), "ginkgo.") })
	if !opts.Focus.Empty() || !opts.Skip.Empty() || !opts.Files.Empty() || opts.LabelFilter != "" || ginkgoExtra {
		key += "|" + string(framework)
	}
	if len(files) > 0 {
		key += "|" + pkg + ":" + strings.Join(files, ",")
	}
	return key, nil
}

func addGoBatch(plans []goBatchPlan, plan goBatchPlan, framework Framework, pkg string) []goBatchPlan {
	index := slices.IndexFunc(plans, func(p goBatchPlan) bool {
		return p.key == plan.key && reflect.DeepEqual(p.runtime, plan.runtime)
	})
	if index < 0 {
		plan.framework = framework
		plans = append(plans, plan)
		index = len(plans) - 1
	}
	if !slices.Contains(plans[index].packages, pkg) {
		plans[index].packages = append(plans[index].packages, pkg)
		plans[index].targets = append(plans[index].targets, pkg)
	}
	if framework == parsers.Ginkgo && !slices.Contains(plans[index].ginkgo, pkg) {
		plans[index].ginkgo = append(plans[index].ginkgo, pkg)
	}
	return plans
}

func (o *TestOrchestrator) planGoBatches(packagesByFramework map[Framework][]string, optsByFramework map[Framework]runners.TestOptions, unfocused bool) ([]goBatchPlan, error) {
	all, _ := sortedGoPackages(packagesByFramework)
	if len(all) == 0 {
		return nil, nil
	}
	var bases []goBatchPlan
	for _, framework := range []Framework{parsers.GoTest, parsers.Ginkgo} {
		for _, pkg := range packagesByFramework[framework] {
			runtime, opts, err := o.goPackageOptions(framework, pkg, optsByFramework[framework])
			if err != nil {
				return nil, err
			}
			key, err := goBatchKey(framework, pkg, opts)
			if err != nil {
				return nil, err
			}
			bases = addGoBatch(bases, goBatchPlan{runtime: runtime, opts: opts, key: key}, framework, pkg)
		}
	}
	if unfocused && len(bases) == 1 {
		return o.planUnfocusedGoBatch(bases[0])
	}
	var plans []goBatchPlan
	for _, base := range bases {
		for _, pkg := range base.packages {
			framework, opts := parsers.GoTest, base.opts
			if slices.Contains(base.ginkgo, pkg) {
				framework = parsers.Ginkgo
			} else {
				opts = o.benchOptions(pkg, opts)
			}
			key, err := goBatchKey(framework, pkg, opts)
			if err != nil {
				return nil, err
			}
			plans = addGoBatch(plans, goBatchPlan{runtime: base.runtime, opts: opts, key: key}, framework, pkg)
		}
	}
	return plans, nil
}

func (o *TestOrchestrator) planUnfocusedGoBatch(base goBatchPlan) ([]goBatchPlan, error) {
	selector := "./..."
	if !o.Recursive {
		selector = "."
	}
	plans := []goBatchPlan{base}
	plans[0].targets = []string{selector}
	for _, pkg := range base.packages {
		if slices.Contains(base.ginkgo, pkg) {
			continue
		}
		opts := o.benchOptions(pkg, base.opts)
		if opts.Bench == base.opts.Bench {
			continue
		}
		plans[0].skipEmpty = append(plans[0].skipEmpty, pkg)
		key, err := goBatchKey(parsers.GoTest, pkg, opts)
		if err != nil {
			return nil, err
		}
		plans = addGoBatch(plans, goBatchPlan{runtime: base.runtime, opts: opts, key: key}, parsers.GoTest, pkg)
	}
	return plans, nil
}

// acquireGoBatchSlots takes up to one slot per test binary the batch would run
// at once (-p) and returns the -p it may use: 0 with slots disabled, else the granted
// count (or the wanted one when running on an inherited lease).
func (o *TestOrchestrator) acquireGoBatchSlots(ctx commonsCtx.Context, plan goBatchPlan) (*slots.Lease, int, error) {
	wanted, err := goBatchParallelism(plan)
	if err != nil {
		return nil, 0, err
	}
	lease, err := o.acquireSlots(ctx, slots.Request{Limits: plan.runtime.Slots, WorkDir: o.WorkDir, Command: "go test " + strings.Join(plan.targets, " "), Count: wanted})
	if err != nil || !plan.runtime.Slots.Enabled() {
		return lease, 0, err
	}
	if granted := lease.Count(); granted > 0 && granted < wanted {
		logger.Infof("Downgraded go test parallelism from %d to %d to fit available slots", wanted, granted)
		return lease, granted, nil
	}
	return lease, wanted, nil
}

func (o *TestOrchestrator) runGoBatch(ctx commonsCtx.Context, t *task.Task, plan goBatchPlan) (packageResult, error) {
	lease, parallel, err := o.acquireGoBatchSlots(ctx, plan)
	if err != nil {
		return packageResult{}, err
	}
	defer func() {
		if err := lease.Close(); err != nil {
			logger.Errorf("release test slots: %v", err)
		}
	}()
	if o.CoverDir != "" {
		plan.coverProfile = filepath.Join(o.CoverDir, fmt.Sprintf("batch-%d.out", goBatchCount.Add(1)))
	}
	process, err := o.buildGoBatchCommand(plan, parallel)
	if err != nil {
		return packageResult{}, err
	}
	process.WithTask(t)
	if o.OutputTee != nil {
		utils.AnnounceCommand(o.OutputTee, process.Cwd, append([]string{process.Cmd}, process.Args...))
		process.Stream(o.OutputTee, o.OutputTee)
	}
	t.SetName(formatRunningCommand(process.Cmd, process.Args))
	process.WithEnv(lease.Env())
	started := time.Now()
	// --test-timeout bounds each package, not the batch: the process itself
	// is held only to the run's --timeout.
	budget := o.packageTimeout(process)
	runContext, cancel, supervision := o.superviseProcess(ctx, process, strings.Join(plan.targets, " "), 0)
	defer cancel()
	var deadlines *goPackageDeadlines
	if budget > 0 {
		deadlines = newGoPackageDeadlines(budget, process, func() { _ = process.KillTree() })
		process.Stream(deadlines, io.Discard)
	}
	if o.SigquitDumper != nil {
		defer o.SigquitDumper.watch(process)()
	}
	span, err := o.beginProcessSpan()
	if err != nil {
		return packageResult{}, err
	}
	result := process.WithContext(runContext).Run().Result()
	if err := lease.Close(); err != nil {
		return packageResult{}, err
	}
	var dumped []goTestBinary
	if deadlines != nil {
		deadlines.stop()
		dumped = deadlines.Dumped()
	}
	if supervision.TimedOut() {
		dumped = append(dumped, supervision.Dumped()...)
	}
	if result == nil {
		return packageResult{}, fmt.Errorf("go test produced no process result")
	}
	usage, err := span.usage(process.Pid())
	if err != nil {
		return packageResult{}, err
	}
	parsed, err := parsers.ParseGoTestBatch(strings.NewReader(result.Stdout), o.WorkDir, nil)
	if err != nil {
		return packageResult{}, err
	}
	importPaths := make([]string, 0, len(parsed))
	for importPath := range parsed {
		importPaths = append(importPaths, importPath)
	}
	paths, err := resolveGoBatchPaths(o.WorkDir, plan.packages, importPaths, len(plan.targets) == 1 && (plan.targets[0] == "./..." || plan.targets[0] == "."))
	if err != nil {
		return packageResult{}, err
	}
	var suites parsers.TestSuiteResults
	for importPath, pkgPath := range paths {
		packageSuites, exists := parsed[importPath]
		if !exists {
			packageSuites = parsers.TestSuiteResults{{Framework: parsers.GoTest, ExitCode: result.ExitCode, Stderr: result.Stderr, Tests: parsers.Tests{{Name: "Go test execution", Package: importPath, Failed: true, Message: "go test emitted no results for this package: " + result.Output()}}}}
		}
		if slices.Contains(plan.ginkgo, pkgPath) {
			hasReport := false
			for _, suite := range packageSuites {
				hasReport = hasReport || suite.Framework == parsers.Ginkgo
			}
			if !hasReport && packageSuites.All().Sum().Failed == 0 {
				packageSuites = append(packageSuites, parsers.TestResults{
					Framework: parsers.Ginkgo, ExitCode: 1,
					Tests: parsers.Tests{{Name: "Ginkgo metadata", Package: importPath, Framework: parsers.Ginkgo, Failed: true,
						Message: fmt.Sprintf("package %s passed without Ginkgo metadata", pkgPath)}},
				})
			}
		}
		if (len(plan.targets) == 1 && plan.targets[0] == "./..." && !slices.Contains(plan.packages, pkgPath) || slices.Contains(plan.skipEmpty, pkgPath)) && emptyGoPackage(packageSuites) {
			continue
		}
		for i := range packageSuites {
			packageSuites[i].Command = process.Cmd + " " + strings.Join(process.Args, " ")
			for j := range packageSuites[i].Tests {
				packageSuites[i].Tests[j].PackagePath = pkgPath
				packageSuites[i].Tests[j].Command = packageSuites[i].Command
			}
		}
		packageTimedOut := deadlines != nil && deadlines.TimedOut(importPath)
		if packageTimedOut {
			markTimedOutTests(packageSuites, budget, result.Stdout, result.Stderr, parsers.GoTest)
		} else if supervision.TimedOut() {
			markTimedOutTests(packageSuites, time.Since(started), result.Stdout, result.Stderr, parsers.GoTest)
		}
		dedupeGinkgoBootstraps(packageSuites)
		for i := range packageSuites {
			stampAttempts(parsers.TestSuiteResults{packageSuites[i]}, attemptStamp{
				RunKind: o.RunKind, Started: started, Ended: time.Now(), PID: process.Pid(),
				ExitCode: packageSuites[i].ExitCode, Usage: usage, SharedProcess: len(paths) > 1,
			})
		}
		if packageTimedOut || supervision.TimedOut() {
			o.checkpointTimeoutDump(pkgPath, packageSuites, dumped)
		}
		suites = append(suites, packageSuites...)
		if o.streamer != nil {
			for _, suite := range packageSuites {
				o.streamer.CompletePackage(pkgPath, suite.Framework, parsers.TestSuiteResults{suite})
			}
		}
	}
	if result.ExitCode != 0 && suites.All().Sum().Failed == 0 {
		suites = append(suites, parsers.TestResults{Framework: parsers.GoTest, Stderr: result.Stderr, Tests: parsers.Tests{{Name: "Go test execution", Failed: true, Message: result.Output()}}})
	}
	var firstFail *parsers.Test
	if suites.All().Sum().Failed > 0 {
		firstFail = firstFailingLeaf(suites)
	}
	t.SetName(formatPackageLabel(parsers.GoTest, fmt.Sprintf("%d packages", len(plan.packages)), suites.All().Sum(), firstFail))
	t.SetDescription("")
	if firstFail != nil {
		t.Failed()
	} else {
		t.Success()
	}
	return packageResult{framework: parsers.GoTest, testResults: suites}, nil
}

func emptyGoPackage(suites parsers.TestSuiteResults) bool {
	if len(suites) != 1 || suites[0].Framework != parsers.GoTest || len(suites[0].Tests) != 1 {
		return false
	}
	test := suites[0].Tests[0]
	return test.Skipped && (test.Name == "No test files" || test.Name == "No tests to run")
}

func resolveGoBatchPaths(workDir string, packages, importPaths []string, allowExtras bool) (map[string]string, error) {
	module, _ := parsers.GoModuleName(workDir)
	rels := make(map[string]string, len(packages))
	for _, pkg := range packages {
		rel := filepath.Clean(pkg)
		if filepath.IsAbs(rel) {
			var err error
			rel, err = filepath.Rel(workDir, rel)
			if err != nil {
				return nil, fmt.Errorf("resolve package %s: %w", pkg, err)
			}
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("package %s is outside workdir %s", pkg, workDir)
		}
		rels[pkg] = filepath.ToSlash(rel)
	}
	paths := make(map[string]string, len(packages))
	used := map[string]bool{}
	for _, importPath := range importPaths {
		var matches []string
		longest := 0
		for _, pkg := range packages {
			rel := rels[pkg]
			if used[pkg] {
				continue
			}
			if rel == "." && importPath == module || rel != "." && (importPath == module+"/"+rel || module == "" && strings.HasSuffix(importPath, "/"+rel)) {
				if len(rel) > longest {
					matches = nil
					longest = len(rel)
				}
				if len(rel) == longest {
					matches = append(matches, pkg)
				}
			}
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("ambiguous go test package %s matches %v", importPath, matches)
		}
		if len(matches) == 1 {
			paths[importPath] = matches[0]
			used[matches[0]] = true
		}
	}
	var remainingImports, remainingPackages []string
	for _, importPath := range importPaths {
		if _, ok := paths[importPath]; !ok {
			remainingImports = append(remainingImports, importPath)
		}
	}
	for _, pkg := range packages {
		if !used[pkg] {
			remainingPackages = append(remainingPackages, pkg)
		}
	}
	if module == "" && len(remainingImports) == 1 && len(remainingPackages) == 1 {
		paths[remainingImports[0]] = remainingPackages[0]
		return paths, nil
	}
	if len(remainingImports) > 0 {
		prefix, err := goImportPrefix(workDir)
		if err != nil {
			return nil, err
		}
		var unmapped []string
		for _, importPath := range remainingImports {
			rel, ok := strings.CutPrefix(importPath, prefix+"/")
			if !ok || rel == "" || filepath.Clean(rel) != rel || strings.HasPrefix(rel, "..") {
				// A dependency outside the module only reaches here when its
				// build failed without go test naming the package it broke.
				paths[importPath] = "build-failure:" + importPath
			} else if allowExtras {
				paths[importPath] = "./" + rel
			} else {
				unmapped = append(unmapped, importPath)
			}
		}
		if len(unmapped) > 0 {
			return nil, fmt.Errorf("cannot map go test packages %v to requested paths %v", unmapped, remainingPackages)
		}
	}
	for _, pkg := range remainingPackages {
		paths["gavel-unreported:"+pkg] = pkg
	}
	return paths, nil
}

func goImportPrefix(workDir string) (string, error) {
	root, err := filepath.Abs(workDir)
	if err != nil {
		return "", err
	}
	for dir := root; ; dir = filepath.Dir(dir) {
		if module, ok := parsers.GoModuleName(dir); ok {
			rel, err := filepath.Rel(dir, root)
			if err != nil {
				return "", err
			}
			if rel == "." {
				return module, nil
			}
			return module + "/" + filepath.ToSlash(rel), nil
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("cannot map go test packages: no go.mod above %s", workDir)
		}
	}
}

// goBatchParallelism is how many test binaries a batch wants to run at once:
// its own -p, or go test's default of GOMAXPROCS.
func goBatchParallelism(plan goBatchPlan) (int, error) {
	if plan.opts.Parallel > 0 {
		return plan.opts.Parallel, nil
	}
	return goruntime.GOMAXPROCS(0), nil
}

// buildGoBatchCommand runs the batch with -p=parallel when parallel is
// positive, replacing the batch's own -p.
func (o *TestOrchestrator) buildGoBatchCommand(plan goBatchPlan, parallel int) (*exec.Process, error) {
	timeout, err := runtimeTimeout(plan.runtime)
	if err != nil {
		return nil, err
	}
	opts := plan.opts.ForPackages(plan.targets...)
	opts.Timeout = subprocessTimeoutFor(timeout)
	opts.CoverProfile = plan.coverProfile
	if parallel > 0 {
		opts.Parallel = parallel
	}
	process, err := runners.GoTestCommand(opts, plan.ginkgo)
	if err != nil {
		return nil, err
	}
	if err := plan.runtime.Apply(process); err != nil {
		return nil, err
	}
	process.WithTimeout(timeout)
	return process, nil
}

func sortedGoPackages(packages map[Framework][]string) (all, ginkgo []string) {
	seen := map[string]bool{}
	for _, fw := range []Framework{parsers.GoTest, parsers.Ginkgo} {
		for _, pkg := range packages[fw] {
			if !seen[pkg] {
				seen[pkg] = true
				all = append(all, pkg)
			}
			if fw == parsers.Ginkgo {
				ginkgo = append(ginkgo, pkg)
			}
		}
	}
	sort.Strings(all)
	sort.Strings(ginkgo)
	return all, ginkgo
}
