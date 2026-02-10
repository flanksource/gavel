package todos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

// CheckOptions configures the TODO check operation.
type CheckOptions struct {
	WorkDir string        // Working directory for test execution
	Timeout time.Duration // Timeout for each test execution
	Logger  logger.Logger // Logger for output
}

// CheckTODOs executes verification tests for the given TODOs in parallel and returns results.
// For each TODO:
//  1. Runs the Verification section tests using the TODO executor
//  2. Updates frontmatter with execution results (last_run, status, attempts)
//  3. Returns a CheckResult with test outcomes
//
// The function uses task.StartGroup for parallel execution, following the pattern
// established in fixtures/runner.go.
func CheckTODOs(ctx context.Context, todoList []*types.TODO, opts CheckOptions) ([]*types.CheckResult, error) {
	// Create task group for parallel TODO checking
	todoGroup := task.StartGroup[*types.CheckResult]("TODO Checks")

	// Map tasks to TODOs for result correlation
	taskToTODO := make(map[task.TypedTask[*types.CheckResult]]*types.TODO)

	for _, todo := range todoList {
		todoRef := todo // capture for closure
		workDir := opts.WorkDir
		if todoRef.CWD != "" {
			workDir = todoRef.CWD
		}

		executor := &TODOExecutor{workDir: workDir}

		typedTask := todoGroup.Add(
			todoRef.Filename(),
			func(fCtx flanksourceContext.Context, t *task.Task) (*types.CheckResult, error) {
				result := checkSingleTODO(ctx, executor, todoRef, opts)

				// Set task status based on result
				if result.AllPassed {
					t.Success()
				} else {
					t.Failed()
				}
				return result, nil
			},
			task.WithTaskTimeout(opts.Timeout),
		)
		taskToTODO[typedTask] = todoRef
	}

	// Wait for all tasks
	groupResult := todoGroup.WaitFor()
	if groupResult.Error != nil {
		opts.Logger.Warnf("Some TODO checks failed: %v", groupResult.Error)
	}

	// Collect results and update state
	taskResults, err := todoGroup.GetResults()
	if err != nil {
		return nil, fmt.Errorf("failed to get results: %w", err)
	}

	results := make([]*types.CheckResult, 0, len(taskResults))
	for typedTask, result := range taskResults {
		results = append(results, result)

		// Update TODO state
		if todo, exists := taskToTODO[typedTask]; exists {
			updateTODOAfterCheck(todo, result, opts.Logger)
		}
	}

	return results, nil
}

// updateTODOAfterCheck updates the TODO frontmatter and persists state after a check completes.
func updateTODOAfterCheck(todo *types.TODO, result *types.CheckResult, log logger.Logger) {
	now := time.Now()
	todo.LastRun = &now
	attempts := todo.Attempts + 1
	todo.Attempts = attempts

	var status types.Status
	if result.AllPassed {
		status = types.StatusCompleted
	} else {
		status = types.StatusFailed
	}
	todo.Status = status

	updates := StateUpdate{
		LastRun:  &now,
		Attempts: &attempts,
		Status:   &status,
	}
	if err := UpdateTODOState(todo, updates); err != nil {
		log.Warnf("Failed to update TODO state for %s: %v", todo.FilePath, err)
	}

	if result.TestResult != nil {
		if err := UpdateLatestFailure(todo, result.TestResult); err != nil {
			log.Warnf("Failed to update Latest Failure section for %s: %v", todo.FilePath, err)
		}
	}
}

// checkSingleTODO runs verification tests for a single TODO and returns the result.
func checkSingleTODO(ctx context.Context, executor *TODOExecutor, todo *types.TODO, opts CheckOptions) *types.CheckResult {
	start := time.Now()

	// Get git info before running tests
	gitBranch, gitCommit, gitDirty, _ := GetGitInfo(executor.workDir)

	// Collect tests from FileNode (new architecture) or fall back to Verification (legacy)
	var testNodes []*fixtures.FixtureNode
	if todo.FileNode != nil {
		testNodes = types.CollectTests(todo.FileNode)
	} else if len(todo.Verification) > 0 {
		// Legacy: use Verification field for backwards compatibility
		testNodes = todo.Verification
	}

	// If no tests found, treat as failed
	if len(testNodes) == 0 {
		return &types.CheckResult{
			TODO:      todo,
			Results:   []fixtures.FixtureResult{},
			AllPassed: false,
			Duration:  time.Since(start),
			Error:     fmt.Errorf("no tests defined"),
		}
	}

	// Create context with timeout if specified
	execCtx := ctx
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Set CWD on test nodes if not already set
	for _, node := range testNodes {
		if node.Test != nil && node.Test.CWD == "" {
			node.Test.CWD = todo.CWD
		}
	}

	// Execute tests using the TODO executor
	testResults := executor.ExecuteSection(execCtx, testNodes)

	// Check if all tests passed
	allPassed := AllPassed(testResults)
	duration := time.Since(start)

	// Build TestResultInfo from test results
	testResultInfo := buildTestResultInfo(testResults, types.BuildTestResultInfoOptions{
		CWD:       executor.workDir,
		GitBranch: gitBranch,
		GitCommit: gitCommit,
		GitDirty:  gitDirty,
		Timestamp: start,
		Passed:    allPassed,
		Duration:  duration,
	})

	return &types.CheckResult{
		TODO:       todo,
		Results:    testResults,
		AllPassed:  allPassed,
		Duration:   duration,
		TestResult: testResultInfo,
	}
}

// buildTestResultInfo creates a TestResultInfo from fixture results.
func buildTestResultInfo(results []fixtures.FixtureResult, opts types.BuildTestResultInfoOptions) *types.TestResultInfo {
	// Build combined output and command from results
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
	// Truncate output if too long
	if len(combinedOutput) > 2000 {
		combinedOutput = combinedOutput[:2000] + "\n... (output truncated)"
	}

	command := strings.Join(commands, " && ")
	if command == "" && len(results) > 0 {
		// Try to get command from test name
		command = fmt.Sprintf("fixtures check (tests: %d)", len(results))
	}

	return &types.TestResultInfo{
		Command:   command,
		CWD:       opts.CWD,
		GitBranch: opts.GitBranch,
		GitCommit: opts.GitCommit,
		GitDirty:  opts.GitDirty,
		Timestamp: opts.Timestamp,
		Passed:    opts.Passed,
		Output:    combinedOutput,
		Duration:  opts.Duration,
	}
}
