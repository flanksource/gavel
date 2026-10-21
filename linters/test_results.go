package linters

import (
	"fmt"

	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// Test renders the linter run as a test group: framework = linter name, one
// leaf per violation with the severity mapped onto the test state flags
// (error→failed, warning→warned, info→skipped; no severity is an error). A run
// without violations is itself the leaf: passed, skipped, timed out or failed.
// A run that crashed or timed out after reporting violations adds that as a
// failed leaf of its own, so it is not lost behind the findings.
func (lr *LinterResult) Test() parsers.Test {
	group := lr.runTest()
	if len(lr.Violations) == 0 {
		return group
	}
	for _, v := range lr.Violations {
		group.Children = append(group.Children, violationTest(lr.Linter, v))
	}
	if group.Failed {
		group.Children = append(group.Children, parsers.Test{Name: lr.Linter, Framework: group.Framework,
			Failed: true, TimedOut: group.TimedOut, Message: group.Message})
	}
	return group
}

// runTest is the linter run's own verdict.
func (lr *LinterResult) runTest() parsers.Test {
	group := parsers.Test{Name: lr.Linter, Framework: parsers.Framework(lr.Linter), Duration: lr.Duration}
	switch {
	case lr.Skipped:
		group.Skipped, group.Message = true, lr.Error
	case lr.TimedOut:
		group.Failed, group.TimedOut, group.Message = true, true, lr.Error
	case !lr.Success:
		group.Failed, group.Message = true, lr.Error
	case len(lr.Violations) == 0:
		group.Passed = true
	}
	return group
}

func violationTest(linter string, v models.Violation) parsers.Test {
	leaf := parsers.Test{Name: v.File, Framework: parsers.Framework(linter), File: v.File, Line: v.Line}
	if v.Line > 0 {
		leaf.Name = fmt.Sprintf("%s:%d", v.File, v.Line)
	}
	if v.Rule != nil && v.Rule.Method != "" {
		leaf.Name += " " + v.Rule.Method
	}
	if v.Message != nil {
		leaf.Message = *v.Message
	}
	switch v.Severity {
	case models.SeverityError, "":
		leaf.Failed = true
	case models.SeverityWarning:
		leaf.Warned = true
	case models.SeverityInfo:
		leaf.Skipped = true
	default:
		leaf.Failed = true
		leaf.Message = fmt.Sprintf("unknown lint severity %q: %s", v.Severity, leaf.Message)
	}
	return leaf
}
