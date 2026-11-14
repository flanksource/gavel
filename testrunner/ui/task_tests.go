package testui

import (
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// TaskTests projects scoped tasks into virtual test groups with zero authored
// test counts. Embedded runners can place these inside their owning plan.
func TaskTests(snapshots []task.TaskSnapshot) []parsers.Test {
	return scopedTaskTests(snapshots)
}
