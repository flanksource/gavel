package todos

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

// CheckOptions configures the TODO check operation.
type CheckOptions struct {
	WorkDir  string        // Working directory for test execution
	Timeout  time.Duration // Timeout for each verification run
	Logger   logger.Logger // Logger for output
	Provider Provider
	Spec     *api.Spec
}

// CheckTODOs executes each issue's fixture-backed definition of done in
// parallel. Every item is a verify-only Captain issue run, so persistence and
// status transitions are shared with verification after implementation runs.
func CheckTODOs(ctx context.Context, todoList []*types.TODO, opts CheckOptions) ([]*types.CheckResult, error) {
	if opts.Logger == nil {
		opts.Logger = logger.StandardLogger()
	}
	todoGroup := task.StartGroup[*types.CheckResult]("TODO Checks")

	for _, todo := range todoList {
		todoRef := todo
		todoGroup.Add(
			todoRef.Filename(),
			func(_ flanksourceContext.Context, t *task.Task) (*types.CheckResult, error) {
				result := CheckTODO(ctx, todoRef, CheckOptions{
					WorkDir:  opts.WorkDir,
					Timeout:  opts.Timeout,
					Logger:   opts.Logger,
					Provider: opts.Provider,
					Spec:     opts.Spec,
				})
				if result.AllPassed {
					t.Success()
				} else {
					t.Failed()
				}
				return result, nil
			},
			task.WithTaskTimeout(opts.Timeout),
		)
	}

	groupResult := todoGroup.WaitFor()
	if groupResult.Error != nil {
		opts.Logger.Warnf("Some TODO checks failed: %v", groupResult.Error)
	}
	taskResults, err := todoGroup.GetResults()
	if err != nil {
		return nil, fmt.Errorf("failed to get results: %w", err)
	}
	results := make([]*types.CheckResult, 0, len(taskResults))
	for _, result := range taskResults {
		results = append(results, result)
	}
	return results, nil
}

// CheckTODO runs one issue's complete definition of done: configured test/lint
// steps, its persisted Verification fixture, and its acceptance-criteria AI
// checklist. It is the shared entrypoint for `todos check` and the dashboard.
func CheckTODO(ctx context.Context, todo *types.TODO, opts CheckOptions) *types.CheckResult {
	start := time.Now()
	if opts.Logger == nil {
		opts.Logger = logger.StandardLogger()
	}
	if todo == nil {
		err := fmt.Errorf("todo is required")
		return &types.CheckResult{AllPassed: false, Duration: time.Since(start), Error: err, ErrorText: err.Error()}
	}
	timeout, err := todoCheckTimeout(opts)
	if err != nil {
		result := &types.CheckResult{
			TODO: todo, Results: []fixtures.FixtureResult{}, AllPassed: false,
			Duration: time.Since(start), Error: err, ErrorText: err.Error(),
		}
		updateTODOAfterCheck(ctx, opts.Provider, todo, result, opts.Logger)
		return result
	}
	workDir := todoCheckWorkDir(opts.WorkDir, todo)
	gitBranch, gitCommit, gitDirty, _ := GetGitInfo(workDir)

	verifier, err := newVerificationExecutor(workDir, []*types.TODO{todo}, opts.Spec)
	if err != nil {
		result := &types.CheckResult{
			TODO: todo, Results: []fixtures.FixtureResult{}, AllPassed: false,
			Duration: time.Since(start), Error: err, ErrorText: err.Error(),
		}
		updateTODOAfterCheck(ctx, opts.Provider, todo, result, opts.Logger)
		return result
	}

	execCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	runner := NewTODOExecutor(workDir, verifier, "", opts.Provider)
	runner.SetMode(types.ModeVerify)
	execution, runErr := runner.Execute(NewExecutorContext(execCtx, opts.Logger, nil), todo)

	var output *types.VerificationOutput
	allPassed := false
	if execution != nil && execution.DoD != nil {
		allPassed = execution.DoD.Ran && execution.DoD.Passed
		output = execution.DoD.Output
	}
	if runErr != nil {
		allPassed = false
	}
	var testResults []fixtures.FixtureResult
	if output != nil {
		testResults = output.Results
	}
	duration := time.Since(start)
	testResultInfo := buildTestResultInfo(testResults, types.BuildTestResultInfoOptions{
		CWD: workDir, GitBranch: gitBranch, GitCommit: gitCommit, GitDirty: gitDirty,
		Timestamp: start, Passed: allPassed, Duration: duration,
	})
	result := &types.CheckResult{
		TODO: todo, Results: testResults, AllPassed: allPassed, Duration: duration,
		Output: output, TestResult: testResultInfo,
	}
	if runErr != nil {
		result.Error = runErr
		result.ErrorText = runErr.Error()
	}
	if !allPassed && opts.Provider != nil {
		if err := opts.Provider.UpdateLatestFailure(ctx, todo, testResultInfo); err != nil {
			opts.Logger.Warnf("Failed to update Latest Failure section for %s: %v", todo.FilePath, err)
		}
	}
	return result
}

func todoCheckTimeout(opts CheckOptions) (time.Duration, error) {
	if opts.Spec == nil || strings.TrimSpace(opts.Spec.Budget.Timeout) == "" {
		return opts.Timeout, nil
	}
	timeout, err := time.ParseDuration(opts.Spec.Budget.Timeout)
	if err != nil {
		return 0, fmt.Errorf("budget.timeout: %w", err)
	}
	if timeout <= 0 {
		return 0, fmt.Errorf("budget.timeout must be greater than zero")
	}
	return timeout, nil
}

func todoCheckWorkDir(base string, todo *types.TODO) string {
	if todo == nil || strings.TrimSpace(todo.CWD) == "" {
		return base
	}
	if filepath.IsAbs(todo.CWD) {
		return todo.CWD
	}
	return filepath.Join(base, todo.CWD)
}

// updateTODOAfterCheck handles a pre-execution failure such as a missing
// definition of done. Executed verify-only runs persist through TODOExecutor.
func updateTODOAfterCheck(ctx context.Context, provider Provider, todo *types.TODO, result *types.CheckResult, log logger.Logger) {
	now := time.Now()
	attempts := todo.Attempts + 1
	status := types.StatusUnverified
	todo.LastRun = &now
	todo.Attempts = attempts
	todo.Status = status
	if provider == nil {
		return
	}
	if err := provider.UpdateState(ctx, todo, StateUpdate{LastRun: &now, Attempts: &attempts, Status: &status}); err != nil {
		log.Warnf("Failed to update TODO state for %s: %v", todo.FilePath, err)
	}
	if result.TestResult != nil {
		if err := provider.UpdateLatestFailure(ctx, todo, result.TestResult); err != nil {
			log.Warnf("Failed to update Latest Failure section for %s: %v", todo.FilePath, err)
		}
	}
}

// buildTestResultInfo creates a TestResultInfo from fixture results.
func buildTestResultInfo(results []fixtures.FixtureResult, opts types.BuildTestResultInfoOptions) *types.TestResultInfo {
	var commands []string
	var outputs []string
	for _, r := range results {
		if r.Command != "" {
			commands = append(commands, r.Command)
		}
		output := strings.TrimSpace(r.Stdout + r.Stderr)
		if output != "" {
			outputs = append(outputs, output)
		}
		if r.Error != "" {
			outputs = append(outputs, "Error: "+r.Error)
		}
	}
	combinedOutput := strings.Join(outputs, "\n---\n")
	if len(combinedOutput) > 2000 {
		combinedOutput = combinedOutput[:2000] + "\n... (output truncated)"
	}
	command := strings.Join(commands, " && ")
	if command == "" && len(results) > 0 {
		command = fmt.Sprintf("fixtures check (tests: %d)", len(results))
	}
	return &types.TestResultInfo{
		Command: command, CWD: opts.CWD, GitBranch: opts.GitBranch,
		GitCommit: opts.GitCommit, GitDirty: opts.GitDirty, Timestamp: opts.Timestamp,
		Passed: opts.Passed, Output: combinedOutput, Duration: opts.Duration,
	}
}
