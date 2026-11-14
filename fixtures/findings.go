package fixtures

import (
	"github.com/flanksource/clicky/task"

	"github.com/flanksource/gavel/models"
)

// ViolationNode is a finding as a child of the step that reported it: an
// error fails it, a warning warns it and an info passes it. The row shows the
// rule and message; the violation itself rides in ViolationMetadataKey.
func ViolationNode(v models.Violation) *FixtureNode {
	name := v.Rule.Method + ": " + *v.Message
	status := map[models.ViolationSeverity]task.Status{
		models.SeverityError: task.StatusFAIL, models.SeverityWarning: task.StatusWarning, models.SeverityInfo: task.StatusPASS,
	}[v.Severity]
	result := &FixtureResult{Name: name, Type: "finding", Status: status, Test: FixtureTest{Name: name},
		Metadata: map[string]any{ViolationMetadataKey: v}}
	if status != task.StatusPASS {
		result.Error = *v.Message
	}
	return &FixtureNode{Name: name, Type: TestNode, Results: result}
}

// FindingsStatus is the status a step's findings give it: failed by any
// error, else warned by any warning.
func FindingsStatus(errors, warnings int) task.Status {
	switch {
	case errors > 0:
		return task.StatusFAIL
	case warnings > 0:
		return task.StatusWarning
	}
	return task.StatusPASS
}
