package todos

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

// Executor represents any AI system that can execute TODOs.
// Implementations include ClaudeExecutor, and potentially OpenAI, Anthropic API, etc.
type Executor interface {
	// Execute runs a TODO with the given interactive context.
	// Returns execution result with tokens, cost, and other metadata.
	Execute(ctx *ExecutorContext, todo *types.TODO) (*ExecutionResult, error)

	// Name returns the executor name (e.g., "claude-code", "openai-gpt4")
	Name() string
}

// ExecutionResult contains the outcome from any executor.
// This is executor-agnostic - all executors return this structure.
type ExecutionResult struct {
	Success          bool
	Skipped          bool
	ExecutorName     string        // Which executor was used
	TokensUsed       int           // Total tokens consumed
	CostUSD          float64       // Cost in USD
	Duration         time.Duration // Total execution time
	NumTurns         int           // Number of interaction rounds
	ActionsPerformed []string      // List of actions taken (tool uses, etc.)
	ErrorMessage     string
	Transcript       *ExecutionTranscript
}

func (e ExecutionResult) Pretty() api.Text {
	result := clicky.Text(" Executed with ", "text-gray-500").Append(e.ExecutorName, "text-blue-600 font-bold")

	if e.Success {
		result = result.Add(icons.Pass)
	} else if e.Skipped {
		result = result.Add(icons.Skip)
	} else {
		result = result.Add(icons.Fail)
	}

	if e.TokensUsed > 0 {
		result = result.Append(fmt.Sprintf("   Tokens: %d", e.TokensUsed), "text-gray-500")
	}

	if e.CostUSD > 0 {
		result = result.Append(fmt.Sprintf("   Cost: $%.4f", e.CostUSD), "text-gray-500")
	}

	if e.Duration > 0 {
		result = result.Append(fmt.Sprintf("   Duration: %s", e.Duration.String()), "text-gray-500")
	}

	if e.NumTurns > 0 {
		result = result.Append(fmt.Sprintf("   Turns: %d", e.NumTurns), "text-gray-500")
	}

	if len(e.ActionsPerformed) > 0 {
		result = result.Append("   Actions: ", "text-gray-500").Append(fmt.Sprintf("%v", e.ActionsPerformed), "text-gray-500")
	}

	for _, msg := range e.Transcript.Entries {
		result = result.NewLine().Add(msg.Pretty().Indent(2))
	}
	return result

}

// TODOExecutor orchestrates TODO execution with any AI executor.
// It handles pre-checks, verification, and frontmatter updates.
type TODOExecutor struct {
	workDir   string
	executor  Executor // Pluggable executor implementation
	sessionID string   // Session ID for resumption across runs
}

// NewTODOExecutor creates a TODO executor with the specified AI backend.
func NewTODOExecutor(workDir string, executor Executor, sessionID string) *TODOExecutor {
	return &TODOExecutor{
		workDir:   workDir,
		executor:  executor,
		sessionID: sessionID,
	}
}

// Execute runs a TODO using the configured executor with interactive context.
// It performs pre-checks, delegates to the executor, runs verification, and updates metadata.
func (e *TODOExecutor) Execute(ctx *ExecutorContext, todo *types.TODO) (*ExecutionResult, error) {
	ctx.Logger.Infof("Starting TODO execution: %s", todo.FilePath)

	// Update status to in_progress and record start time
	todo.Status = types.StatusInProgress
	now := time.Now()
	todo.LastRun = &now

	// Initialize LLM config if needed and save session ID immediately
	if todo.LLM == nil {
		todo.LLM = &types.LLM{}
	}
	if e.sessionID != "" {
		todo.LLM.SessionId = e.sessionID
	}

	// Check if test already passes (skip if so)
	if len(todo.StepsToReproduce) > 0 {
		ctx.Logger.Debugf("Checking if test already passes")
		ctx.Notify(Notification{
			Type:    NotifyProgress,
			Message: "Checking if test already passes",
		})

		if e.stepsAlreadyPass(ctx, todo.StepsToReproduce) {
			ctx.Logger.Infof("Test already passes, skipping execution")
			todo.Status = types.StatusSkipped
			return &ExecutionResult{
				Skipped:      true,
				ExecutorName: e.executor.Name(),
				Transcript:   ctx.GetTranscript(),
			}, nil
		}
	}

	// Execute with configured executor (Claude, OpenAI, etc.)
	ctx.Logger.Infof("Executing with %s", e.executor.Name())
	result, err := e.executor.Execute(ctx, todo)
	if err != nil {
		ctx.Logger.Errorf("Execution failed: %v", err)
		todo.Status = types.StatusFailed
		return result, err
	}

	// Verify the fix
	if len(todo.Verification) > 0 {
		ctx.Logger.Debugf("Running verification tests")
		ctx.Notify(Notification{
			Type:    NotifyProgress,
			Message: "Running verification tests",
		})

		if !e.verificationPasses(ctx, todo.Verification) {
			ctx.Logger.Errorf("Verification tests failed")
			todo.Status = types.StatusFailed
			result.Success = false
			result.ErrorMessage = "Verification tests failed"
			return result, fmt.Errorf("verification failed")
		}
	}

	// Update frontmatter with results
	ctx.Logger.Infof("TODO execution completed successfully")
	e.updateFrontmatter(todo, result)

	return result, nil
}

// updateFrontmatter updates the TODO's frontmatter with execution results.
func (e *TODOExecutor) updateFrontmatter(todo *types.TODO, result *ExecutionResult) {
	todo.Status = types.StatusCompleted
	todo.Attempts++

	// Update LLM usage metrics
	if todo.LLM == nil {
		todo.LLM = &types.LLM{}
	}
	todo.LLM.Model = result.ExecutorName
	todo.LLM.TokensUsed = result.TokensUsed
	todo.LLM.CostIncurred = result.CostUSD
}

// stepsAlreadyPass checks if reproduction steps already pass.
func (e *TODOExecutor) stepsAlreadyPass(ctx *ExecutorContext, steps []*fixtures.FixtureNode) bool {
	results := e.ExecuteSection(ctx, steps)
	return AllPassed(results)
}

// verificationPasses checks if verification tests pass.
func (e *TODOExecutor) verificationPasses(ctx *ExecutorContext, verification []*fixtures.FixtureNode) bool {
	results := e.ExecuteSection(ctx, verification)
	return AllPassed(results)
}

// ExecuteSection runs all fixture nodes in a section.
// Returns the results of executing each node using the fixtures runner infrastructure.
func (e *TODOExecutor) ExecuteSection(ctx context.Context, nodes []*fixtures.FixtureNode) []fixtures.FixtureResult {
	var results []fixtures.FixtureResult

	// Create CEL evaluator for fixture execution
	evaluator, err := fixtures.NewCELEvaluator()
	if err != nil {
		return []fixtures.FixtureResult{{
			Status: "error",
			Error:  fmt.Sprintf("failed to create CEL evaluator: %v", err),
		}}
	}

	// Prepare run options
	opts := fixtures.RunOptions{
		WorkDir:   e.workDir,
		Verbose:   false,
		NoCache:   false,
		Evaluator: evaluator,
	}

	// Execute each test node
	for _, node := range nodes {
		if node.Test == nil {
			continue
		}

		// Get the appropriate fixture type from registry
		fixtureType, err := fixtures.DefaultRegistry.GetForFixture(*node.Test)
		if err != nil {
			results = append(results, fixtures.FixtureResult{
				Name:   node.Test.Name,
				Status: "error",
				Test:   *node.Test,
				Error:  err.Error(),
			})
			continue
		}

		// Run the fixture test
		result := fixtureType.Run(ctx, *node.Test, opts)
		results = append(results, result)
	}

	return results
}

// AllPassed checks if all fixture results passed.
func AllPassed(results []fixtures.FixtureResult) bool {
	for _, r := range results {
		if !r.IsOK() {
			return false
		}
	}
	return true
}
