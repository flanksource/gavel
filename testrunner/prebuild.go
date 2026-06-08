package testrunner

import (
	"fmt"
	"sort"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	commonsCtx "github.com/flanksource/commons/context"
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
func goPreBuildArgs(pkgs []string) []string {
	args := []string{"test", "-count=0"}
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
	t := clicky.StartTask[string](name, func(_ commonsCtx.Context, t *task.Task) (string, error) {
		process := exec.NewExec("go", goPreBuildArgs(pkgs)...).WithCwd(o.WorkDir).WithProcessGroup().WithTask(t)
		// Bound a hung build by the global --timeout; the per-package
		// --test-timeout deliberately does not apply — compilation is the
		// whole point of this phase.
		if o.Timeout > 0 {
			process = process.WithTimeout(o.Timeout)
		}
		if o.OutputTee != nil {
			process = process.Stream(o.OutputTee, o.OutputTee)
		}

		result := process.Run().Result()
		if result == nil {
			return "", fmt.Errorf("pre-build: go test -count=0 produced no result")
		}
		if !result.IsOk() {
			detail := strings.TrimSpace(stripExitStatus(result.Stderr))
			if detail == "" {
				detail = strings.TrimSpace(result.Stdout)
			}
			return "", fmt.Errorf("pre-build: compiling Go test binaries failed (exit %d): %s", result.ExitCode, detail)
		}
		return "", nil
	})

	if _, err := t.GetResult(); err != nil {
		t.Errorf("%v", err)
		t.Failed()
		return err
	}
	t.Success()
	return nil
}

func plural(n int, singular, pluralForm string) string {
	if n == 1 {
		return singular
	}
	return pluralForm
}
