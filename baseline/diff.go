package baseline

import (
	"path/filepath"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// FilterNewTestFailures walks the test tree and flips failures that exist
// in the baseline to passed, so only genuinely new failures remain.
// Returns the modified tree (mutated in place).
func FilterNewTestFailures(tests []parsers.Test, baselineKeys map[TestKey]bool) []parsers.Test {
	for i := range tests {
		filterTestNode(&tests[i], baselineKeys)
	}
	return tests
}

func filterTestNode(t *parsers.Test, baselineKeys map[TestKey]bool) {
	if t.Failed {
		key := TestKey{
			Framework:   string(t.Framework),
			PackagePath: t.PackagePath,
			FullName:    t.FullName(),
		}
		if baselineKeys[key] {
			t.Failed = false
			t.Passed = true
		}
	}
	for i := range t.Children {
		filterTestNode(&t.Children[i], baselineKeys)
	}
	// Clear cached summary so it gets recomputed from the mutated children.
	t.Summary = nil
}

// FilterNewViolations removes violations that exist in the baseline from each
// LinterResult. Mutates results in place and returns them.
func FilterNewViolations(results []*linters.LinterResult, baselineKeys map[ViolationKey]bool) []*linters.LinterResult {
	for _, r := range results {
		if r == nil || r.Skipped || len(r.Violations) == 0 {
			continue
		}
		kept := r.Violations[:0]
		for _, v := range r.Violations {
			file := v.File
			if r.WorkDir != "" && filepath.IsAbs(file) {
				if rel, err := filepath.Rel(r.WorkDir, file); err == nil {
					file = rel
				}
			}
			rule := ""
			if v.Rule != nil {
				rule = v.Rule.Method
			}
			key := ViolationKey{Linter: r.Linter, File: file, Rule: rule}
			if !baselineKeys[key] {
				kept = append(kept, v)
			}
		}
		r.Violations = kept
	}
	return results
}
