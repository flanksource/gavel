package main

import (
	"context"

	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/linters"
)

func init() {
	commitpkg.SetLintRunner(runCommitLint)
}

// runCommitLint is the LintRunner adapter the commit package uses to invoke
// the same executeLinters pipeline as `gavel lint`.
//
// linterNames semantics match commit.applyLintGate:
//   - nil          → run every detected linter
//   - {"betterleaks"} → run only betterleaks
//   - {"!betterleaks"} → run every detected linter except betterleaks
//
// Anything else is passed through to executeLinters as an explicit list.
func runCommitLint(ctx context.Context, workDir string, linterNames, files []string) ([]*linters.LinterResult, error) {
	requested, exclude := splitCommitLintRequest(linterNames)
	opts := LintOptions{
		WorkDir: workDir,
		Linters: requested,
		Files:   files,
		Timeout: "5m",
		Context: ctx,
	}
	results, err := executeLinters(opts)
	if err != nil {
		return nil, err
	}
	if len(exclude) > 0 {
		results = filterOutLinters(results, exclude)
	}
	return results, nil
}

// splitCommitLintRequest splits a request list into (positive, negative)
// linter names. A leading "!" marks an exclusion. We use this to express
// "everything except betterleaks" without changing the registry contract.
func splitCommitLintRequest(in []string) (positive, negative []string) {
	for _, name := range in {
		if len(name) > 1 && name[0] == '!' {
			negative = append(negative, name[1:])
			continue
		}
		positive = append(positive, name)
	}
	return positive, negative
}

func filterOutLinters(results []*linters.LinterResult, names []string) []*linters.LinterResult {
	excluded := make(map[string]struct{}, len(names))
	for _, n := range names {
		excluded[n] = struct{}{}
	}
	out := make([]*linters.LinterResult, 0, len(results))
	for _, r := range results {
		if r == nil {
			continue
		}
		if _, drop := excluded[r.Linter]; drop {
			continue
		}
		out = append(out, r)
	}
	return out
}
