package testrunner

import (
	"fmt"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	commonsCtx "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
)

type goBatchPlan struct {
	packages  []string
	ginkgo    []string
	args      []string
	targets   []string
	skipEmpty []string
	framework Framework
}

func (o *TestOrchestrator) unfocusedGoRun(startingPaths []string) bool {
	return len(startingPaths) == 0 && len(o.Frameworks) == 0 && len(o.Ignore) == 0 &&
		!o.hasChangeSelector() && o.Failed == "" && len(o.PassThroughArgs) == 0
}

func (o *TestOrchestrator) planGoBatches(packagesByFramework map[Framework][]string, argsByFramework map[Framework][]string, unfocused bool) ([]goBatchPlan, error) {
	all, ginkgo := sortedGoPackages(packagesByFramework)
	if len(all) == 0 {
		return nil, nil
	}
	goArgs := argsByFramework[parsers.GoTest]
	ginkgoArgs := append([]string(nil), argsByFramework[parsers.Ginkgo]...)
	if len(ginkgo) > 0 {
		var err error
		ginkgoArgs, err = runners.GinkgoGoTestArgs(ginkgoArgs)
		if err != nil {
			return nil, err
		}
	}
	runtimeEqual := true
	if len(packagesByFramework[parsers.GoTest]) > 0 && len(ginkgo) > 0 {
		goRuntime, goRuntimeArgs, err := o.runtimeFor(parsers.GoTest)
		if err != nil {
			return nil, err
		}
		ginkgoRuntime, ginkgoRuntimeArgs, err := o.runtimeFor(parsers.Ginkgo)
		if err != nil {
			return nil, err
		}
		ginkgoRuntimeArgs, err = runners.GinkgoGoTestArgs(ginkgoRuntimeArgs)
		if err != nil {
			return nil, err
		}
		runtimeEqual = reflect.DeepEqual(goRuntime, ginkgoRuntime) && slices.Equal(goRuntimeArgs, ginkgoRuntimeArgs)
	}
	var bases []goBatchPlan
	if len(packagesByFramework[parsers.GoTest]) == 0 {
		bases = append(bases, goBatchPlan{packages: all, ginkgo: ginkgo, args: ginkgoArgs, framework: parsers.Ginkgo})
	} else if len(ginkgo) == 0 || slices.Equal(goArgs, ginkgoArgs) && runtimeEqual {
		bases = append(bases, goBatchPlan{packages: all, ginkgo: ginkgo, args: goArgs, framework: parsers.GoTest})
	} else {
		bases = append(bases, goBatchPlan{packages: packagesByFramework[parsers.GoTest], args: goArgs, framework: parsers.GoTest})
		bases = append(bases, goBatchPlan{packages: ginkgo, ginkgo: ginkgo, args: ginkgoArgs, framework: parsers.Ginkgo})
	}
	if unfocused && len(bases) == 1 {
		return o.planUnfocusedGoBatch(bases[0]), nil
	}
	runner := runners.NewGoTest(o.WorkDir)
	var plans []goBatchPlan
	for _, base := range bases {
		for _, pkg := range base.packages {
			args := o.augmentBenchArgs(runner, pkg, base.args)
			index := -1
			for i := range plans {
				if plans[i].framework == base.framework && slices.Equal(plans[i].args, args) {
					index = i
					break
				}
			}
			if index < 0 {
				plans = append(plans, goBatchPlan{args: args, framework: base.framework})
				index = len(plans) - 1
			}
			plans[index].packages = append(plans[index].packages, pkg)
			plans[index].targets = append(plans[index].targets, pkg)
			if slices.Contains(base.ginkgo, pkg) {
				plans[index].ginkgo = append(plans[index].ginkgo, pkg)
			}
		}
	}
	return plans, nil
}

func (o *TestOrchestrator) planUnfocusedGoBatch(base goBatchPlan) []goBatchPlan {
	selector := "./..."
	if !o.Recursive {
		selector = "."
	}
	base.targets = []string{selector}
	plans := []goBatchPlan{base}
	runner := runners.NewGoTest(o.WorkDir)
	for _, pkg := range base.packages {
		args := o.augmentBenchArgs(runner, pkg, base.args)
		if slices.Equal(args, base.args) {
			continue
		}
		plans[0].skipEmpty = append(plans[0].skipEmpty, pkg)
		index := -1
		for i := 1; i < len(plans); i++ {
			if slices.Equal(plans[i].args, args) {
				index = i
				break
			}
		}
		if index < 0 {
			plans = append(plans, goBatchPlan{args: args, framework: base.framework})
			index = len(plans) - 1
		}
		plans[index].packages = append(plans[index].packages, pkg)
		plans[index].targets = append(plans[index].targets, pkg)
		if slices.Contains(base.ginkgo, pkg) {
			plans[index].ginkgo = append(plans[index].ginkgo, pkg)
		}
	}
	return plans
}

func (o *TestOrchestrator) runGoBatch(ctx commonsCtx.Context, t *task.Task, plan goBatchPlan) (packageResult, error) {
	process, err := o.buildGoBatchCommand(plan)
	if err != nil {
		return packageResult{}, err
	}
	process.WithTask(t)
	if o.OutputTee != nil {
		process.Stream(o.OutputTee, o.OutputTee)
	}
	t.SetName(formatRunningCommand(process.Cmd, process.Args))
	started := time.Now()
	runContext, cancel, supervision := o.supervisePackage(ctx, process, strings.Join(plan.targets, " "))
	defer cancel()
	if o.SigquitDumper != nil {
		defer o.SigquitDumper.watch(process)()
	}
	span, err := o.beginProcessSpan()
	if err != nil {
		return packageResult{}, err
	}
	result := process.WithContext(runContext).Run().Result()
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
				return packageResult{}, fmt.Errorf("package %s passed without Ginkgo metadata", pkgPath)
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
		if supervision.TimedOut() {
			markTimedOutTests(packageSuites, time.Since(started), result.Stdout, result.Stderr, parsers.GoTest)
		}
		dedupeGinkgoBootstraps(packageSuites)
		for i := range packageSuites {
			stampAttempts(parsers.TestSuiteResults{packageSuites[i]}, attemptStamp{
				RunKind: o.RunKind, Started: started, Ended: time.Now(), PID: process.Pid(),
				ExitCode: packageSuites[i].ExitCode, Usage: usage, SharedProcess: len(paths) > 1,
			})
		}
		if supervision.TimedOut() {
			o.checkpointTimeoutDump(pkgPath, packageSuites, supervision.Dumped())
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
			if rel == "." && importPath == module || rel != "." && (importPath == module+"/"+rel || strings.HasSuffix(importPath, "/"+rel)) {
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
	if len(remainingImports) == 1 && len(remainingPackages) == 1 {
		paths[remainingImports[0]] = remainingPackages[0]
		return paths, nil
	}
	if allowExtras && len(remainingImports) > 0 {
		prefix, err := goImportPrefix(workDir)
		if err != nil {
			return nil, err
		}
		for _, importPath := range remainingImports {
			rel, ok := strings.CutPrefix(importPath, prefix+"/")
			if !ok || rel == "" || filepath.Clean(rel) != rel || strings.HasPrefix(rel, "..") {
				return nil, fmt.Errorf("go test package %s is outside %s", importPath, prefix)
			}
			paths[importPath] = "./" + rel
		}
		remainingImports = nil
	}
	if len(remainingImports) > 0 {
		return nil, fmt.Errorf("cannot map go test packages %v to requested paths %v", remainingImports, remainingPackages)
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

func (o *TestOrchestrator) buildGoBatchCommand(plan goBatchPlan) (*exec.Process, error) {
	overlay, err := runners.BuildGinkgoOverlay(o.WorkDir, plan.ginkgo, o.Tags)
	if err != nil {
		return nil, err
	}
	runtime, runtimeArgs, err := o.runtimeFor(plan.framework)
	if err != nil {
		return nil, err
	}
	timeout := 10 * time.Minute
	if runtime.Timeout != nil {
		if *runtime.Timeout == "" {
			timeout = 0
		} else {
			timeout, err = time.ParseDuration(*runtime.Timeout)
			if err != nil {
				return nil, fmt.Errorf("invalid go test runtime timeout: %w", err)
			}
		}
	}
	args := []string{"test", "-json", "-v"}
	if overlay != "" {
		args = append(args, "-overlay="+overlay)
	}
	if subTimeout := subprocessTimeoutFor(timeout); subTimeout > 0 {
		args = append(args, subprocessTimeoutArgs(parsers.GoTest, subTimeout)...)
	}
	if plan.framework == parsers.Ginkgo {
		runtimeArgs, err = runners.GinkgoGoTestArgs(runtimeArgs)
		if err != nil {
			return nil, err
		}
	}
	var ginkgoArgs []string
	for _, arg := range append(runtimeArgs, plan.args...) {
		if strings.HasPrefix(arg, "-ginkgo.") {
			ginkgoArgs = append(ginkgoArgs, arg)
		} else {
			args = append(args, arg)
		}
	}
	args = append(args, plan.targets...)
	if len(ginkgoArgs) > 0 {
		args = append(args, "-args")
		args = append(args, ginkgoArgs...)
	}
	process := exec.NewExec("go", args...).WithCwd(o.WorkDir).WithProcessGroup()
	process.SucceedOnNonZero = true
	if err := runtime.Apply(process); err != nil {
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
