package testrunner

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
)

// testArgv is the argv every runner's Detect parses: the starting paths, then
// -- and the arguments given after it.
func (o *TestOrchestrator) testArgv() []string {
	argv := slices.Clone(o.StartingPaths)
	if len(o.PassThroughArgs) > 0 {
		argv = append(append(argv, "--"), o.PassThroughArgs...)
	}
	return argv
}

// baseTestOptions carries gavel's own flags into the shared options.
func (o *TestOrchestrator) baseTestOptions() (runners.TestOptions, error) {
	cwd, err := filepath.Abs(o.WorkDir)
	if err != nil {
		return runners.TestOptions{}, err
	}
	opts := runners.TestOptions{Cwd: cwd, Recursive: o.Recursive, Tags: slices.Clone(o.Tags)}
	if strings.TrimSpace(o.Focus) != "" {
		if _, err := regexp.Compile(o.Focus); err != nil {
			return opts, fmt.Errorf("--focus %q: %w", o.Focus, err)
		}
		opts.Focus.Regex = []string{o.Focus}
	}
	for _, pattern := range o.Ignore {
		if pattern = strings.TrimSpace(pattern); pattern != "" {
			opts.SkipFiles.Glob = append(opts.SkipFiles.Glob, pattern)
		}
	}
	return opts, nil
}

// goFamily is go test and Ginkgo: both run through one go test command, so
// they share runner arguments.
func goFamily(fw Framework) bool {
	return fw == parsers.GoTest || fw == parsers.Ginkgo
}

// resolveTestOptions finishes each detected framework's options once its
// packages are known: the Go family merges its go test and Ginkgo parses,
// a token any selected runner recognised is no longer an unknown argument,
// and unknown arguments are only allowed when one framework family runs.
func (o *TestOrchestrator) resolveTestOptions(detected map[Framework]*runners.TestOptions, packages map[Framework][]string) (map[Framework]runners.TestOptions, error) {
	var selected []Framework
	for fw := range detected {
		if len(packages[fw]) > 0 {
			selected = append(selected, fw)
		}
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i] < selected[j] })
	var union runners.TestOptions
	for _, fw := range selected {
		merged, err := union.Merge(*detected[fw])
		if err != nil {
			return nil, err
		}
		union = merged
	}
	families := map[string]bool{}
	var names []string
	for _, fw := range selected {
		family := string(fw)
		if goFamily(fw) {
			family = "go"
		}
		families[family] = true
		names = append(names, string(fw))
	}
	if len(union.ExtraArgs) > 0 && len(families) > 1 {
		return nil, fmt.Errorf("arguments after -- that no test runner recognises (%s) need one framework family; got %d (%s) for %s: use `gavel test <framework> -- ...` or a single --framework",
			strings.Join(union.ExtraArgs, " "), len(families), strings.Join(names, ", "), o.pathsLabel())
	}
	resolved := make(map[Framework]runners.TestOptions, len(selected))
	for _, fw := range selected {
		opts := *detected[fw]
		if goFamily(fw) {
			for _, other := range []Framework{parsers.GoTest, parsers.Ginkgo} {
				if other != fw && detected[other] != nil {
					merged, err := opts.Merge(*detected[other])
					if err != nil {
						return nil, err
					}
					opts = merged
				}
			}
		}
		if opts.LabelFilter != "" && fw != parsers.Ginkgo {
			logger.Infof("Label filter %q: skipping %s, which has no labels", opts.LabelFilter, fw)
			continue
		}
		opts.ExtraArgs = append(slices.Clone(union.ExtraArgs), o.ExtraArgs...)
		resolved[fw] = opts
	}
	if len(o.Tags) > 0 && resolved[parsers.GoTest].Cwd == "" && resolved[parsers.Ginkgo].Cwd == "" {
		return nil, fmt.Errorf("--tags is not supported by the selected frameworks")
	}
	return resolved, nil
}

func (o *TestOrchestrator) pathsLabel() string {
	if len(o.StartingPaths) == 0 {
		return o.WorkDir
	}
	return strings.Join(o.StartingPaths, " ")
}

// failedTestOptions focuses each framework on the tests a previous run failed.
// Node runners match their name filter against the full name (suite titles
// joined before the test title), so their leaf titles are end-anchored.
func failedTestOptions(names map[Framework][]string) map[Framework]runners.FileSelection {
	out := map[Framework]runners.FileSelection{}
	quote := func(list []string, suffix string) []string {
		var quoted []string
		for _, name := range list {
			quoted = append(quoted, regexp.QuoteMeta(name)+suffix)
		}
		return quoted
	}
	for fw, list := range names {
		switch fw {
		case parsers.GoTest:
			out[fw] = runners.FileSelection{Exact: list}
		case parsers.Ginkgo:
			out[fw] = runners.FileSelection{Regex: quote(list, "")}
		case parsers.Vitest, parsers.Jest, parsers.Playwright:
			out[fw] = runners.FileSelection{Regex: quote(list, "$")}
		}
	}
	return out
}

// benchOptions decides whether benchmarks run for a Go test package: with
// --bench, or automatically when it has benchmarks but no runnable tests (then
// -run=^$ keeps the empty test pass quiet). A benchmark pattern given after --
// wins.
func (o *TestOrchestrator) benchOptions(pkg string, opts runners.TestOptions) runners.TestOptions {
	if opts.Bench != "" {
		return opts
	}
	gt := runners.NewGoTest(o.WorkDir)
	pattern := strings.TrimSpace(o.Bench)
	switch pattern {
	case "false":
		return opts
	case "":
		if !gt.PackageHasBenchmarks(pkg) || gt.PackageHasRunnableTests(pkg) {
			return opts
		}
		pattern = "."
	case "true":
		pattern = "."
	}
	if !gt.PackageHasBenchmarks(pkg) {
		return opts
	}
	opts.Bench = pattern
	if !gt.PackageHasRunnableTests(pkg) && opts.Focus.Empty() {
		opts.Focus = runners.FileSelection{Regex: []string{"^$"}}
	}
	return opts
}
