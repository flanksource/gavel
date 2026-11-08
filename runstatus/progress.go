package runstatus

import (
	"fmt"
	"maps"
	"slices"
	"time"

	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/sdk/status"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// maxRunning bounds the running tasks a status lists.
const maxRunning = 20

// groupKinds are the groups the live UI shows, and fixture runs, by name.
var groupKinds = map[string]string{
	testui.TestTaskGroupName:       status.GroupPackages,
	testui.LintTaskGroupName:       status.GroupLint,
	testui.PromptLintTaskGroupName: status.GroupPrompt,
	fixtures.TaskGroupName:         status.GroupFixtures,
}

func liveTasks() []clickytask.TaskSnapshot {
	return clickytask.SnapshotAll(slices.Collect(maps.Keys(groupKinds))...)
}

// progress projects task snapshots the way the live UI does (testui.TaskTests),
// so a group's counts here match the UI's.
func progress(snapshots []clickytask.TaskSnapshot) ([]status.Group, []status.Item) {
	finished := map[string]string{}
	for _, snapshot := range snapshots {
		if snapshot.Type == "group" {
			finished[snapshot.Name] = snapshot.FinishedAt
		}
	}
	var groups []status.Group
	var running []status.Item
	for _, node := range testui.TaskTests(snapshots) {
		// A group nested under one of these has no kind of its own.
		group := status.Group{Name: node.Name, Kind: groupKinds[node.Name], Summary: summaryOf(parsers.Tests(node.Children).Sum()), Finished: parseTime(finished[node.Name])}
		if node.Started != nil {
			group.Started = *node.Started
		}
		groups = append(groups, group)
		for _, child := range node.Children {
			if child.Running && len(running) < maxRunning {
				running = append(running, item(node.Name, child))
			}
		}
	}
	return groups, running
}

func item(group string, test parsers.Test) status.Item {
	item := status.Item{Name: test.Name, Group: group, Description: test.Message}
	if test.Started != nil {
		item.Started = *test.Started
	}
	if test.Progress != nil {
		item.Done, item.Total = test.Progress.Done, test.Progress.Total
	}
	if len(test.Attempts) > 0 {
		item.PID = test.Attempts[len(test.Attempts)-1].PID
	}
	return item
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		panic(fmt.Sprintf("runstatus: task finishedAt %q: %v", value, err))
	}
	return parsed
}

func summaryOf(s parsers.TestSummary) status.Summary {
	return status.Summary{
		Total: s.Total, Passed: s.Passed, Failed: s.Failed, Warned: s.Warned, Aborted: s.Aborted, Skipped: s.Skipped,
		Pending: s.Pending, Running: s.Running, TimedOut: s.TimedOut,
	}
}
