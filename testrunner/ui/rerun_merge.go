package testui

import (
	"github.com/flanksource/gavel/testrunner/parsers"
)

// mergeRerunTests overlays rerun results on top of the existing tree.
// Identity is (framework, package_path, suite path, name). Matching tests
// have their top-level flags refreshed to reflect the new attempt, and the
// incoming Attempts slice is appended onto the preserved tree's Attempts.
// Tests that only exist in the incoming snapshot are appended; tests only
// in the preserved tree keep their prior state.
func mergeRerunTests(existing, incoming []parsers.Test) []parsers.Test {
	if len(existing) == 0 {
		return incoming
	}
	if len(incoming) == 0 {
		return existing
	}

	byKey := make(map[string]int, len(existing))
	for i, t := range existing {
		byKey[testMergeKey(t)] = i
	}
	merged := make([]parsers.Test, len(existing))
	copy(merged, existing)
	for _, in := range incoming {
		key := testMergeKey(in)
		if idx, ok := byKey[key]; ok {
			merged[idx] = mergeTestNode(merged[idx], in)
		} else {
			merged = append(merged, in)
			byKey[key] = len(merged) - 1
		}
	}
	return merged
}

// mergeTestNode merges a single incoming test into an existing one.
// Top-level state (pass/fail/skip/pending/timedout/duration/stdout/stderr)
// adopts the most recent attempt's values so existing filters keep working.
// Attempts from the incoming test are appended to the existing Attempts so
// history survives.
func mergeTestNode(existing, incoming parsers.Test) parsers.Test {
	merged := existing
	merged.Attempts = append(append([]parsers.TestAttempt(nil), existing.Attempts...), incoming.Attempts...)
	// Renumber so Sequence is globally monotonic across the merged history.
	// The runner stamps each attempt with len(existing.Attempts)+1 locally,
	// but a rerun ships a fresh Test whose local count starts from 0, so
	// without this every rerun would re-land as "Attempt 1".
	for i := range merged.Attempts {
		merged.Attempts[i].Sequence = i + 1
	}

	// Reset terminal flags before copying so a passing rerun of a previously
	// failing test is not reported as both passed+failed.
	merged.Passed = false
	merged.Failed = false
	merged.Skipped = false
	merged.Pending = false
	merged.TimedOut = false

	merged.Passed = incoming.Passed
	merged.Failed = incoming.Failed
	merged.Skipped = incoming.Skipped
	merged.Pending = incoming.Pending
	merged.TimedOut = incoming.TimedOut
	merged.Duration = incoming.Duration
	merged.Message = incoming.Message
	merged.Stdout = incoming.Stdout
	merged.Stderr = incoming.Stderr
	merged.Command = incoming.Command
	merged.Summary = incoming.Summary

	// Children: recurse with the same key-based merge.
	merged.Children = mergeRerunTests(existing.Children, incoming.Children)
	return merged
}

func testMergeKey(t parsers.Test) string {
	// Framework + package path + name uniquely identifies a leaf in practice.
	// Children get their own merge via recursive mergeRerunTests, so suite
	// path is not needed here.
	return string(t.Framework) + "\x00" + t.PackagePath + "\x00" + t.Name
}
