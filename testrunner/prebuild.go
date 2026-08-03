package testrunner

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// goPackagesToWarm returns the deduplicated, sorted set of Go packages that
// benefit from cache warming: the union of the GoTest and Ginkgo frameworks
// (both compile via the Go toolchain). JS/TS frameworks are excluded — `go
// test -count=0` does nothing for them.
func goPackagesToWarm(packagesByFramework map[parsers.Framework][]string) []string {
	seen := map[string]struct{}{}
	for _, fw := range []parsers.Framework{parsers.GoTest, parsers.Ginkgo} {
		for _, pkg := range packagesByFramework[fw] {
			seen[pkg] = struct{}{}
		}
	}
	pkgs := make([]string, 0, len(seen))
	for pkg := range seen {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)
	return pkgs
}

// goPreBuildArgs builds the `go test -count=0 <pkgs...>` argument vector that
// compiles every test binary without executing a single test.
func goPreBuildArgs(pkgs, tags []string) []string {
	args := []string{"test", "-count=0"}
	if len(tags) > 0 {
		args = append(args, "-tags="+strings.Join(tags, ","))
	}
	return append(args, pkgs...)
}

// preBuildGoPackages compiles all Go test binaries for pkgs in one `go test
// -count=0` invocation, warming the build cache before the timed per-package
// run. It is rendered as a single phase task. A compile failure aborts the run
// (a broken build would fail every package anyway), so the compiler output is
// surfaced in the returned error.
func (o *TestOrchestrator) preBuildGoPackages(pkgs []string) error {
	if len(pkgs) == 0 {
		return nil
	}

	name := fmt.Sprintf("Pre-build (compiling %d Go test %s)", len(pkgs), plural(len(pkgs), "package", "packages"))
	process := exec.NewExec("go", goPreBuildArgs(pkgs)...).WithCwd(o.WorkDir).WithProcessGroup()
	if o.OutputTee != nil {
		process = process.Stream(o.OutputTee, o.OutputTee)
	}

	ctx := o.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if o.Timeout > 0 {
		if _, hasDeadline := ctx.Deadline(); !hasDeadline {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, o.Timeout)
			defer cancel()
		}
	}
	result, err := runPreBuildProcess(ctx, process, name)
	if ctx.Err() != nil {
		return fmt.Errorf("pre-build: %w", ctx.Err())
	}
	if result == nil {
		return fmt.Errorf("pre-build: go test -count=0 produced no result: %w", err)
	}
	if !result.IsOk() {
		detail := strings.TrimSpace(stripExitStatus(result.Stderr))
		if detail == "" {
			detail = strings.TrimSpace(result.Stdout)
		}
		return fmt.Errorf("pre-build: compiling Go test binaries failed (exit %d): %s", result.ExitCode, detail)
	}
	if err != nil {
		return fmt.Errorf("pre-build: %w", err)
	}
	return nil
}

func runPreBuildProcess(ctx context.Context, process *exec.Process, name string) (*exec.ExecResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	run := process.RunAsTask(name)
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			run.Cancel()
		case <-done:
		}
	}()
	_, err := run.GetResult()
	close(done)
	if ctx.Err() != nil {
		return process.Result(), ctx.Err()
	}
	return process.Result(), err
}

func plural(n int, singular, pluralForm string) string {
	if n == 1 {
		return singular
	}
	return pluralForm
}
