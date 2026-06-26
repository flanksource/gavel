package todos

import (
	"context"
	"testing"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
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

func TestTODOExecutorPersistsSessionBeforeExecutorReturns(t *testing.T) {
	execCtx := NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	provider := &recordingProvider{}
	exec := &sessionHookExecutor{sessionID: "mid-run-session"}
	runner := NewTODOExecutor(".", exec, "", provider)
	todo := &types.TODO{
		ID:              "todo-1",
		FilePath:        "todo-1",
		TODOFrontmatter: types.TODOFrontmatter{Title: "Persist session early"},
	}

	if _, err := runner.Execute(execCtx, todo); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	// The executor observed the session id on the todo immediately after calling
	// RecordSessionID — i.e. the hook persisted it before the executor returned.
	if !exec.sawSessionPersisted {
		t.Fatal("session id was not recorded on the todo during the run")
	}
	if todo.LLM == nil || todo.LLM.SessionId != "mid-run-session" {
		t.Fatalf("todo session id = %+v, want mid-run-session", todo.LLM)
	}
	found := false
	for _, sid := range provider.sessionIDs {
		if sid == "mid-run-session" {
			found = true
		}
	}
	if !found {
		t.Fatalf("provider never received the session id, got %#v", provider.sessionIDs)
	}
}

type sessionHookExecutor struct {
	sessionID           string
	sawSessionPersisted bool
}

func (e *sessionHookExecutor) Name() string { return "session-hook" }

func (e *sessionHookExecutor) Execute(ctx *ExecutorContext, todo *types.TODO) (*ExecutionResult, error) {
	ctx.RecordSessionID(e.sessionID)
	// RecordSessionID runs the hook synchronously, so by now the id is on the todo.
	e.sawSessionPersisted = todo.LLM != nil && todo.LLM.SessionId == e.sessionID
	return &ExecutionResult{ExecutorName: e.Name()}, nil
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
	sessionIDs    []string
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

func (p *recordingProvider) Edit(context.Context, *types.TODO, EditRequest) error {
	return nil
}

func (p *recordingProvider) Comment(context.Context, *types.TODO, string) error {
	return nil
}

func (p *recordingProvider) UpdateState(ctx context.Context, _ *types.TODO, updates StateUpdate) error {
	p.updateCtxErrs = append(p.updateCtxErrs, ctx.Err())
	if updates.SessionID != nil {
		p.sessionIDs = append(p.sessionIDs, *updates.SessionID)
	}
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

func (p *recordingProvider) SaveVerification(context.Context, *types.TODO, *verify.VerifyResult) error {
	return nil
}
