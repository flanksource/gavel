package fixtures

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/flanksource/clicky/task"
)

type executionTracker struct {
	mu        sync.Mutex
	root      *ExecutionNode
	fixture   map[*FixtureNode]*ExecutionNode
	byKey     map[string]*ExecutionNode
	sink      ProgressSink
	iteration int64
}

func newExecutionTracker(root *FixtureNode, workDir string, steps []ExecutionStep, sink ProgressSink) *executionTracker {
	t := &executionTracker{
		root:    &ExecutionNode{Key: "fixtures", Name: "Fixtures", Kind: ExecutionKindRoot, State: ExecutionQueued},
		fixture: make(map[*FixtureNode]*ExecutionNode),
		byKey:   make(map[string]*ExecutionNode),
		sink:    sink,
	}
	t.byKey[t.root.Key] = t.root
	for _, step := range steps {
		node := &ExecutionNode{Key: step.Key, Name: step.Name, Kind: step.Kind, State: ExecutionQueued, Origin: cloneOrigin(step.Origin)}
		t.root.Children = append(t.root.Children, node)
		t.byKey[node.Key] = node
	}
	for i, child := range root.Children {
		t.root.Children = append(t.root.Children, t.fromFixture(child, t.root.Key, workDir, i))
	}
	t.rollup()
	return t
}

func (t *executionTracker) fromFixture(source *FixtureNode, parentKey, workDir string, ordinal int) *ExecutionNode {
	key := executionKey(source, parentKey, workDir, ordinal)
	node := &ExecutionNode{Key: key, Name: source.Name, Kind: executionKind(source), State: ExecutionQueued, Origin: cloneOrigin(source.Origin)}
	t.fixture[source] = node
	t.byKey[key] = node
	for i, child := range source.Children {
		node.Children = append(node.Children, t.fromFixture(child, key, workDir, i))
	}
	return node
}

func (t *executionTracker) Snapshot() ExecutionSnapshot {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.snapshotLocked()
}

func (t *executionTracker) Publish(ctx context.Context) error {
	return t.mutateAndPublish(ctx, nil)
}

func (t *executionTracker) Start(ctx context.Context, fixture *FixtureNode) error {
	return t.mutateAndPublish(ctx, func(now time.Time) error {
		node, ok := t.fixture[fixture]
		if !ok {
			return fmt.Errorf("fixture progress node not found: %q", fixture.Name)
		}
		startNode(node, now)
		return nil
	})
}

func (t *executionTracker) Complete(ctx context.Context, fixture *FixtureNode, result FixtureResult) error {
	return t.mutateAndPublish(ctx, func(now time.Time) error {
		node, ok := t.fixture[fixture]
		if !ok {
			return fmt.Errorf("fixture progress node not found: %q", fixture.Name)
		}
		completeNode(node, executionState(result.Status), result.Error, result.Duration, now)
		return nil
	})
}

func (t *executionTracker) Update(ctx context.Context, fixture *FixtureNode, done, total int) error {
	return t.mutateAndPublish(ctx, func(now time.Time) error {
		node, ok := t.fixture[fixture]
		if !ok {
			return fmt.Errorf("fixture progress node not found: %q", fixture.Name)
		}
		if node.State == ExecutionQueued {
			startNode(node, now)
		}
		node.Done = done
		node.Total = total
		return nil
	})
}

func (t *executionTracker) StartStep(ctx context.Context, key string) error {
	return t.mutateAndPublish(ctx, func(now time.Time) error {
		node, ok := t.byKey[key]
		if !ok {
			return fmt.Errorf("execution progress node not found: %q", key)
		}
		startNode(node, now)
		return nil
	})
}

func (t *executionTracker) UpdateStep(ctx context.Context, key string, done, total int) error {
	return t.mutateAndPublish(ctx, func(now time.Time) error {
		node, ok := t.byKey[key]
		if !ok {
			return fmt.Errorf("execution progress node not found: %q", key)
		}
		if node.State == ExecutionQueued {
			startNode(node, now)
		}
		node.Done = done
		node.Total = total
		return nil
	})
}

func (t *executionTracker) CompleteStep(ctx context.Context, key string, state ExecutionState, stepErr error) error {
	return t.mutateAndPublish(ctx, func(now time.Time) error {
		node, ok := t.byKey[key]
		if !ok {
			return fmt.Errorf("execution progress node not found: %q", key)
		}
		message := ""
		if stepErr != nil {
			message = stepErr.Error()
		}
		completeNode(node, state, message, 0, now)
		return nil
	})
}

func (t *executionTracker) CancelQueued(ctx context.Context, cause error) error {
	return t.mutateAndPublish(ctx, func(now time.Time) error {
		walkExecution(t.root, func(node *ExecutionNode) {
			if isWorkNode(node) && node.State == ExecutionQueued {
				completeNode(node, ExecutionCancelled, cause.Error(), 0, now)
			}
		})
		return nil
	})
}

func (t *executionTracker) mutateAndPublish(ctx context.Context, mutate func(time.Time) error) error {
	t.mu.Lock()
	if mutate != nil {
		if err := mutate(time.Now()); err != nil {
			t.mu.Unlock()
			return err
		}
	}
	t.rollup()
	t.iteration++
	snapshot := t.snapshotLocked()
	t.mu.Unlock()
	if t.sink == nil {
		return nil
	}
	return t.sink(ctx, snapshot)
}

func (t *executionTracker) snapshotLocked() ExecutionSnapshot {
	root := cloneExecutionNode(t.root)
	return ExecutionSnapshot{
		Version: ExecutionSnapshotVersion, Iteration: t.iteration, State: root.State,
		StartedAt: root.StartedAt, EndedAt: root.FinishedAt, Summary: summarizeExecution(root), Root: root,
	}
}

func (t *executionTracker) rollup() {
	rollupExecutionNode(t.root)
}

func startNode(node *ExecutionNode, now time.Time) {
	node.State = ExecutionRunning
	node.StartedAt = &now
	node.FinishedAt = nil
	node.Error = ""
}

func completeNode(node *ExecutionNode, state ExecutionState, message string, duration time.Duration, now time.Time) {
	if node.StartedAt == nil {
		node.StartedAt = &now
	}
	node.State = state
	node.FinishedAt = &now
	node.Error = message
	if duration > 0 {
		node.Duration = duration
	} else {
		node.Duration = now.Sub(*node.StartedAt)
	}
}

func executionState(status task.Status) ExecutionState {
	switch status {
	case task.StatusPASS, task.StatusSuccess:
		return ExecutionPassed
	case task.StatusFAIL, task.StatusFailed:
		return ExecutionFailed
	case task.StatusERR:
		return ExecutionErrored
	case task.StatusWarning:
		return ExecutionWarned
	case task.StatusSKIP:
		return ExecutionSkipped
	case task.StatusCancelled:
		return ExecutionCancelled
	case task.StatusRunning:
		return ExecutionRunning
	default:
		return ExecutionQueued
	}
}

func executionKind(node *FixtureNode) ExecutionKind {
	switch node.Type {
	case FileNode:
		return ExecutionKindFile
	case SectionNode:
		return ExecutionKindSection
	case TableNode:
		return ExecutionKindTable
	}
	if node.Test == nil {
		return ExecutionKindCommand
	}
	if node.Test.IsTestStep() {
		return ExecutionKindTest
	}
	if node.Test.IsLintStep() {
		return ExecutionKindLint
	}
	if node.Test.IsAIStep() {
		return ExecutionKindAI
	}
	return ExecutionKindCommand
}

func executionKey(node *FixtureNode, parentKey, workDir string, ordinal int) string {
	if node.Origin == nil {
		return fmt.Sprintf("%s/%s:%d", parentKey, executionKind(node), ordinal)
	}
	file := filepath.Clean(node.Origin.File)
	if workDir != "" {
		if relative, err := filepath.Rel(workDir, file); err == nil {
			file = relative
		}
	}
	file = filepath.ToSlash(file)
	return fmt.Sprintf("%s/%s|%s|%s|table:%d|row:%d|line:%d|ordinal:%d",
		parentKey, file, node.Origin.SectionPath, node.Origin.Kind, node.Origin.TableIndex, node.Origin.RowIndex, node.Origin.Line, ordinal)
}

func rollupExecutionNode(node *ExecutionNode) ExecutionState {
	if len(node.Children) == 0 {
		return node.State
	}
	states := make([]ExecutionState, 0, len(node.Children))
	for _, child := range node.Children {
		states = append(states, rollupExecutionNode(child))
	}
	node.State = rolledState(states)
	node.StartedAt, node.FinishedAt = executionBounds(node.Children)
	if node.StartedAt != nil && node.FinishedAt != nil {
		node.Duration = node.FinishedAt.Sub(*node.StartedAt)
	}
	return node.State
}

func rolledState(states []ExecutionState) ExecutionState {
	if containsExecutionState(states, ExecutionRunning) ||
		(containsTerminalState(states) && containsExecutionState(states, ExecutionQueued)) {
		return ExecutionRunning
	}
	if allExecutionState(states, ExecutionQueued) {
		return ExecutionQueued
	}
	for _, state := range []ExecutionState{ExecutionErrored, ExecutionFailed, ExecutionTimedOut, ExecutionCancelled, ExecutionWarned} {
		if containsExecutionState(states, state) {
			return state
		}
	}
	if allExecutionState(states, ExecutionSkipped) {
		return ExecutionSkipped
	}
	return ExecutionPassed
}

func executionBounds(children []*ExecutionNode) (*time.Time, *time.Time) {
	var started, finished *time.Time
	allFinished := true
	for _, child := range children {
		if child.StartedAt != nil && (started == nil || child.StartedAt.Before(*started)) {
			value := *child.StartedAt
			started = &value
		}
		if child.FinishedAt == nil {
			allFinished = false
			continue
		}
		if finished == nil || child.FinishedAt.After(*finished) {
			value := *child.FinishedAt
			finished = &value
		}
	}
	if !allFinished {
		finished = nil
	}
	return started, finished
}

func summarizeExecution(root *ExecutionNode) ExecutionSummary {
	var summary ExecutionSummary
	walkExecution(root, func(node *ExecutionNode) {
		if !isWorkNode(node) {
			return
		}
		summary.Total++
		switch node.State {
		case ExecutionQueued:
			summary.Queued++
		case ExecutionRunning:
			summary.Running++
		case ExecutionPassed:
			summary.Passed++
		case ExecutionFailed:
			summary.Failed++
		case ExecutionErrored:
			summary.Errored++
		case ExecutionWarned:
			summary.Warned++
		case ExecutionSkipped:
			summary.Skipped++
		case ExecutionCancelled:
			summary.Cancelled++
		case ExecutionTimedOut:
			summary.TimedOut++
		}
	})
	return summary
}

func isWorkNode(node *ExecutionNode) bool {
	return len(node.Children) == 0 && node.Kind != ExecutionKindFile && node.Kind != ExecutionKindSection && node.Kind != ExecutionKindTable && node.Kind != ExecutionKindRoot
}

func containsTerminalState(states []ExecutionState) bool {
	for _, state := range states {
		if state != ExecutionQueued && state != ExecutionRunning {
			return true
		}
	}
	return false
}

func containsExecutionState(states []ExecutionState, target ExecutionState) bool {
	for _, state := range states {
		if state == target {
			return true
		}
	}
	return false
}

func allExecutionState(states []ExecutionState, target ExecutionState) bool {
	if len(states) == 0 {
		return false
	}
	for _, state := range states {
		if state != target {
			return false
		}
	}
	return true
}

func walkExecution(node *ExecutionNode, visit func(*ExecutionNode)) {
	visit(node)
	for _, child := range node.Children {
		walkExecution(child, visit)
	}
}

func cloneExecutionNode(node *ExecutionNode) *ExecutionNode {
	clone := *node
	clone.Origin = cloneOrigin(node.Origin)
	clone.StartedAt = cloneTime(node.StartedAt)
	clone.FinishedAt = cloneTime(node.FinishedAt)
	clone.Children = make([]*ExecutionNode, len(node.Children))
	for i, child := range node.Children {
		clone.Children[i] = cloneExecutionNode(child)
	}
	return &clone
}

func cloneOrigin(origin *FixtureOrigin) *FixtureOrigin {
	if origin == nil {
		return nil
	}
	clone := *origin
	return &clone
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}
