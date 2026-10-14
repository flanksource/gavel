package fixtures

import (
	"context"
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/flanksource/clicky/task"
)

// scopeSeed is the initial fixture Env: the process environment, then the
// file's prepared setup environment, then its front-matter `env:`.
func scopeSeed(setup *PreparedSetup, frontMatter FrontMatter) []string {
	seed := os.Environ()
	if setup != nil {
		for _, k := range slices.Sorted(maps.Keys(setup.Env)) {
			seed = append(seed, k+"="+setup.Env[k])
		}
	}
	for _, k := range slices.Sorted(maps.Keys(frontMatter.Env)) {
		seed = append(seed, fmt.Sprintf("%s=%v", k, frontMatter.Env[k]))
	}
	return seed
}

// firstFrontMatter is the front matter the nodes' document declared: every
// test of one document carries the same copy.
func firstFrontMatter(nodes []*FixtureNode) FrontMatter {
	var found *FrontMatter
	for _, node := range nodes {
		if node == nil || found != nil {
			continue
		}
		node.Walk(func(n *FixtureNode) {
			if found == nil && n.Test != nil {
				found = &n.Test.FrontMatter
			}
		})
	}
	if found == nil {
		return FrontMatter{}
	}
	return *found
}

// teardownResults tears scope down and reports what that produced: one result
// per background step's verdict, and an errored `teardown` result when a
// cleanup failed — never a swallowed error. Each is marked
// Metadata["teardown"] so a caller mapping results back onto tests can tell
// them apart. Teardown runs even when ctx is already cancelled.
func teardownResults(ctx context.Context, scope *StepScope) []FixtureResult {
	results, err := scope.Teardown(context.WithoutCancel(ctx))
	if err != nil {
		results = append(results, FixtureResult{Name: "teardown", Type: "teardown", Status: task.StatusERR, Error: err.Error(),
			Metadata: map[string]interface{}{}})
	}
	for i := range results {
		results[i].Metadata["teardown"] = true
	}
	return results
}

// isTeardownResult reports whether a result came from teardownResults.
func isTeardownResult(result FixtureResult) bool {
	marked, _ := result.Metadata["teardown"].(bool)
	return marked
}

// foldTeardown attaches a standalone node's teardown results to it: background
// verdicts become children, and a failing one fails the node.
func foldTeardown(result FixtureResult, teardown []FixtureResult) FixtureResult {
	for i := range teardown {
		verdict := teardown[i]
		result.Children = append(result.Children, &FixtureNode{Name: verdict.Name, Type: TestNode, Results: &verdict})
		if isFailureStatus(verdict.Status) && !isFailureStatus(result.Status) {
			result.Status = verdict.Status
			result.Error = fmt.Sprintf("%s: %s", verdict.Name, verdict.Error)
		}
	}
	return result
}
