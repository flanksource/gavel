package testui

import (
	"strings"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// SetTaskSource scopes virtual tasks to this server's run. The source must return
// independent snapshots without I/O; it is frozen when the run finishes.
// A source returning no tasks explicitly disables the CLI's global task groups.
func (s *Server) SetTaskSource(source func() []task.TaskSnapshot) {
	if source == nil {
		panic("test UI: task source is required")
	}
	s.mu.Lock()
	s.taskSource = source
	s.frozenTasks, s.tasksFrozen = nil, false
	s.mu.Unlock()
	s.notify()
}

func taskSnapshotsRunning(snapshots []task.TaskSnapshot) bool {
	for _, snapshot := range snapshots {
		if snapshot.Status == "pending" || snapshot.Status == "running" {
			return true
		}
	}
	return false
}

func scopedTaskTests(snapshots []task.TaskSnapshot) []parsers.Test {
	byID := make(map[string]task.TaskSnapshot, len(snapshots))
	children := map[string][]parsers.Test{}
	for _, snapshot := range snapshots {
		byID[snapshot.ID] = snapshot
	}
	for _, snapshot := range snapshots {
		if snapshot.Type != "group" {
			children[snapshot.GroupID] = append(children[snapshot.GroupID], scopedTaskTest(snapshot, byID))
		}
	}
	var nodes []parsers.Test
	for _, snapshot := range snapshots {
		if snapshot.Type == "group" {
			node := scopedTaskTest(snapshot, byID)
			node.Children = children[snapshot.GroupID]
			node.Summary = &parsers.TestSummary{}
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func scopedTaskTest(snapshot task.TaskSnapshot, byID map[string]task.TaskSnapshot) parsers.Test {
	node := taskSnapshotToTest(snapshot)
	if snapshot.Status == "running" {
		node.Pending, node.Running = false, true
	}
	node.CanStop = false
	if snapshot.Error != "" {
		node.Message = snapshot.Error
	}
	if snapshot.StartedAt != "" {
		started, err := time.Parse(time.RFC3339Nano, snapshot.StartedAt)
		if err != nil {
			panic("test UI: invalid task start time: " + err.Error())
		}
		ended := time.Now()
		if snapshot.FinishedAt != "" {
			ended, err = time.Parse(time.RFC3339Nano, snapshot.FinishedAt)
			if err != nil {
				panic("test UI: invalid task finish time: " + err.Error())
			}
		}
		node.Duration = ended.Sub(started)
	}
	if snapshot.Error == "" && snapshot.Description != "" {
		node.Message = snapshot.Description
	}
	if snapshot.MaxValue > 0 {
		node.Progress = &parsers.TestProgress{Phase: snapshot.Description, Done: snapshot.Progress, Total: snapshot.MaxValue}
	}
	var waiting []string
	for _, id := range snapshot.DependencyIDs {
		dependency, ok := byID[id]
		if !ok {
			panic("test UI: virtual task dependency is absent: " + id)
		}
		if dependency.Status == "pending" || dependency.Status == "running" {
			waiting = append(waiting, dependency.Name)
		}
	}
	if snapshot.Status == "pending" && len(waiting) > 0 {
		node.Message = "Waiting for " + strings.Join(waiting, ", ")
	}
	return node
}
