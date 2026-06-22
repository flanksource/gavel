package todos

import (
	"context"
	"testing"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

func TestAllPassed_AllSuccess(t *testing.T) {
	// Input: []fixtures.FixtureResult with all IsOK() == true
	results := []fixtures.FixtureResult{
		{Status: task.StatusPASS},
		{Status: task.StatusSuccess},
	}

	if !AllPassed(results) {
		t.Error("Expected AllPassed to return true for all successful results")
	}
}

func TestAllPassed_OneFailed(t *testing.T) {
	// Input: []fixtures.FixtureResult with one IsOK() == false
	results := []fixtures.FixtureResult{
		{Status: task.StatusPASS},
		{Status: task.StatusFailed},
	}

	if AllPassed(results) {
		t.Error("Expected AllPassed to return false when one result failed")
	}
}

func TestAllPassed_Empty(t *testing.T) {
	// Input: Empty slice
	results := []fixtures.FixtureResult{}

	if !AllPassed(results) {
		t.Error("Expected AllPassed to return true for empty results")
	}
}

func TestTODOExecutorPersistsFailureAfterContextCanceled(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	execCtx := NewExecutorContext(parent, logger.StandardLogger(), nil)
	provider := &recordingProvider{}
	executor := NewTODOExecutor(".", cancelingExecutor{cancel: cancel}, "", provider)
	todo := &types.TODO{
		ID:       "todo-1",
		FilePath: "todo-1",
		TODOFrontmatter: types.TODOFrontmatter{
			Title: "Fix canceled cleanup",
		},
	}

	result, err := executor.Execute(execCtx, todo)
	if err != context.Canceled {
		t.Fatalf("Execute() error = %v, want context.Canceled", err)
	}
	if result == nil || result.ErrorMessage == "" {
		t.Fatalf("expected failed result with error message, got %#v", result)
	}
	if provider.saveCalls != 1 {
		t.Fatalf("SaveAttempt calls = %d, want 1", provider.saveCalls)
	}
	if provider.saveCtxErr != nil {
		t.Fatalf("SaveAttempt used canceled context: %v", provider.saveCtxErr)
	}
	if len(provider.updateCtxErrs) != 2 {
		t.Fatalf("UpdateState calls = %d, want 2", len(provider.updateCtxErrs))
	}
	for i, err := range provider.updateCtxErrs {
		if err != nil {
			t.Fatalf("UpdateState call %d used canceled context: %v", i+1, err)
		}
	}
	if todo.Status != types.StatusFailed || todo.Attempts != 1 {
		t.Fatalf("todo state = (%s, attempts=%d), want failed/1", todo.Status, todo.Attempts)
	}
}

type cancelingExecutor struct {
	cancel context.CancelFunc
}

func (e cancelingExecutor) Name() string { return "test-executor" }

func (e cancelingExecutor) Execute(_ *ExecutorContext, _ *types.TODO) (*ExecutionResult, error) {
	e.cancel()
	return &ExecutionResult{
		ExecutorName: e.Name(),
		ErrorMessage: "interrupted",
	}, context.Canceled
}

type recordingProvider struct {
	saveCalls     int
	saveCtxErr    error
	updateCtxErrs []error
}

func (p *recordingProvider) List(context.Context, DiscoveryFilters) (types.TODOS, error) {
	return nil, nil
}

func (p *recordingProvider) Get(context.Context, string) (*types.TODO, error) {
	return nil, nil
}

func (p *recordingProvider) Create(context.Context, CreateRequest) (*types.TODO, error) {
	return nil, nil
}

func (p *recordingProvider) Delete(context.Context, *types.TODO) error {
	return nil
}

func (p *recordingProvider) UpdateState(ctx context.Context, _ *types.TODO, _ StateUpdate) error {
	p.updateCtxErrs = append(p.updateCtxErrs, ctx.Err())
	return ctx.Err()
}

func (p *recordingProvider) UpdateLatestFailure(context.Context, *types.TODO, *types.TestResultInfo) error {
	return nil
}

func (p *recordingProvider) SaveAttempt(ctx context.Context, _ *types.TODO, _ *ExecutionResult) error {
	p.saveCalls++
	p.saveCtxErr = ctx.Err()
	return ctx.Err()
}
