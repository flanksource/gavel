package todos

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

const providerPersistenceTimeout = 30 * time.Second

// Executor represents any AI system that can execute TODOs.
// Implementations include ClaudeExecutor, and potentially OpenAI, Anthropic API, etc.
type Executor interface {
	// Execute runs a TODO with the given interactive context.
	// Returns execution result with tokens, cost, and other metadata.
	Execute(ctx *ExecutorContext, todo *types.TODO) (*ExecutionResult, error)

	// Name returns the executor name (e.g., "claude-code", "openai-gpt4")
	Name() string
}

// FeedbackExecutor is implemented by executors that can resume the agent with a
// follow-up message. The post-completion check loop uses it to hand failing
// test/lint output back to the agent so it can fix the issues. Executors that
// do not implement it (or whose SendFeedback returns an error) make the loop
// report the failures without iterating.
type FeedbackExecutor interface {
	SendFeedback(ctx *ExecutorContext, todos []*types.TODO, feedback string) (*ExecutionResult, error)
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
	CommitSHA        string
	Transcript       *ExecutionTranscript
}

// ShouldCommitAfter reports whether a post-run `gavel commit` should run after a
// TODO's agent completes: only when enabled, the run succeeded, and the executor
// did not already commit. The inline claude executor sets CommitSHA after its own
// commit, so committing again would either duplicate that change set or sweep up
// the user's restored working-tree changes; the cmux executor leaves CommitSHA
// empty — that is the path auto-commit is meant to cover.
func ShouldCommitAfter(result *ExecutionResult, enabled bool) bool {
	return enabled && result != nil && result.Success && result.CommitSHA == ""
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

	return result
}

// TODOExecutor orchestrates TODO execution with any AI executor.
// It handles pre-checks, verification, and frontmatter updates.
type TODOExecutor struct {
	workDir   string
	executor  Executor // Pluggable executor implementation
	sessionID string   // Session ID for resumption across runs
	provider  Provider
	// forceChecks turns the post-completion check loop on regardless of config
	// (the --check flag / dashboard toggle). The loop also runs when .gavel.yaml
	// or a TODO's frontmatter enables `checks` — see runCheckLoop.
	forceChecks bool
}

// EnableChecks forces the post-completion check loop on for this executor,
// matching `gavel todos run --check`. Without it, the loop still runs when
// .gavel.yaml or a TODO's frontmatter enables `checks`.
func (e *TODOExecutor) EnableChecks(force bool) { e.forceChecks = force }

// NewTODOExecutor creates a TODO executor with the specified AI backend.
func NewTODOExecutor(workDir string, executor Executor, sessionID string, provider ...Provider) *TODOExecutor {
	var p Provider
	if len(provider) > 0 {
		p = provider[0]
	}
	return &TODOExecutor{
		workDir:   workDir,
		executor:  executor,
		sessionID: sessionID,
		provider:  p,
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
	e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, LastRun: &now, SessionID: &todo.LLM.SessionId})

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
			e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, LastRun: &now})
			return &ExecutionResult{
				Skipped:      true,
				ExecutorName: e.executor.Name(),
				Transcript:   ctx.GetTranscript(),
			}, nil
		}
	}

	// Execute with configured executor (Claude, OpenAI, etc.)
	ctx.Logger.Infof("Executing with %s", e.executor.Name())
	ctx.SetSessionIDHook(e.sessionIDPersister(ctx, []*types.TODO{todo}))
	result, err := e.executor.Execute(ctx, todo)
	// The executor may generate its own session id (e.g. cmux's claude
	// --session-id); persist it now that it is known, regardless of outcome.
	e.persistSessionID(ctx, todo)
	if err != nil {
		ctx.Logger.Errorf("Execution failed: %v", err)
		todo.Status = types.StatusFailed
		todo.Attempts++
		if result != nil {
			if saveErr := e.saveAttempt(ctx, todo, result); saveErr != nil {
				fmt.Fprintf(os.Stderr, "failed to save attempt: %v\n", saveErr)
			}
		}
		e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, Attempts: &todo.Attempts})
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
			todo.Attempts++
			if saveErr := e.saveAttempt(ctx, todo, result); saveErr != nil {
				fmt.Fprintf(os.Stderr, "failed to save attempt: %v\n", saveErr)
			}
			e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, Attempts: &todo.Attempts})
			return result, fmt.Errorf("verification failed")
		}
	}

	// Post-completion check loop: run configured tests/lint and feed failures
	// back to the agent until they pass (opt-in; no-op when not enabled). Only
	// gate a run the executor reported successful — runCheckLoop flips Success to
	// false only when the checks themselves fail.
	if result.Success {
		e.runCheckLoop(ctx, []*types.TODO{todo}, result)
		if !result.Success {
			ctx.Logger.Errorf("Post-completion checks failed: %s", result.ErrorMessage)
			todo.Status = types.StatusFailed
			todo.Attempts++
			if saveErr := e.saveAttempt(ctx, todo, result); saveErr != nil {
				fmt.Fprintf(os.Stderr, "failed to save attempt: %v\n", saveErr)
			}
			e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, Attempts: &todo.Attempts})
			return result, fmt.Errorf("post-completion checks failed: %s", result.ErrorMessage)
		}
	}

	// Update frontmatter with results
	ctx.Logger.Infof("TODO execution completed successfully")
	e.updateFrontmatter(ctx, todo, result)

	return result, nil
}

// GroupExecutor is implemented by executors that support combined group execution.
type GroupExecutor interface {
	ExecuteGroup(ctx *ExecutorContext, todosInGroup []*types.TODO) (*ExecutionResult, error)
}

// ExecuteGroup orchestrates group execution: one AI session for multiple TODOs,
// then independent verification per TODO.
func (e *TODOExecutor) ExecuteGroup(ctx *ExecutorContext, todosInGroup []*types.TODO) ([]*ExecutionResult, error) {
	groupExec, ok := e.executor.(GroupExecutor)
	if !ok {
		return nil, fmt.Errorf("executor %s does not support group execution", e.executor.Name())
	}

	now := time.Now()
	for _, todo := range todosInGroup {
		todo.Status = types.StatusInProgress
		todo.LastRun = &now
		if todo.LLM == nil {
			todo.LLM = &types.LLM{}
		}
		if e.sessionID != "" {
			todo.LLM.SessionId = e.sessionID
		}
		e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, LastRun: &now, SessionID: &todo.LLM.SessionId})
	}

	// Pre-check: filter out TODOs whose steps already pass
	var needsExecution []*types.TODO
	results := make(map[string]*ExecutionResult)
	for _, todo := range todosInGroup {
		if len(todo.StepsToReproduce) > 0 && e.stepsAlreadyPass(ctx, todo.StepsToReproduce) {
			ctx.Logger.Infof("TODO %s already passes, skipping", todo.Filename())
			todo.Status = types.StatusSkipped
			e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, LastRun: &now})
			results[todo.FilePath] = &ExecutionResult{
				Skipped:      true,
				ExecutorName: e.executor.Name(),
				Transcript:   ctx.GetTranscript(),
			}
		} else {
			needsExecution = append(needsExecution, todo)
		}
	}

	// Run combined session if any TODOs need work
	var groupResult *ExecutionResult
	if len(needsExecution) > 0 {
		var err error
		ctx.SetSessionIDHook(e.sessionIDPersister(ctx, needsExecution))
		groupResult, err = groupExec.ExecuteGroup(ctx, needsExecution)
		// The group executor may generate a session id per todo (e.g. cmux's
		// claude --session-id); persist it now that it is known.
		for _, todo := range needsExecution {
			e.persistSessionID(ctx, todo)
		}
		if err != nil {
			for _, todo := range needsExecution {
				todo.Status = types.StatusFailed
				todo.Attempts++
				if groupResult != nil {
					perTodo := e.splitResult(groupResult, len(needsExecution))
					if saveErr := e.saveAttempt(ctx, todo, perTodo); saveErr != nil {
						fmt.Fprintf(os.Stderr, "failed to save attempt: %v\n", saveErr)
					}
				}
				e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, Attempts: &todo.Attempts})
			}
			return e.collectResults(todosInGroup, results), err
		}

		// Verify each TODO independently
		for _, todo := range needsExecution {
			perTodo := e.splitResult(groupResult, len(needsExecution))

			if len(todo.Verification) > 0 {
				ctx.Notify(Notification{
					Type:    NotifyProgress,
					Message: fmt.Sprintf("Verifying %s", todo.Filename()),
				})
				if !e.verificationPasses(ctx, todo.Verification) {
					todo.Status = types.StatusFailed
					perTodo.Success = false
					perTodo.ErrorMessage = "Verification tests failed"
					todo.Attempts++
					if saveErr := e.saveAttempt(ctx, todo, perTodo); saveErr != nil {
						fmt.Fprintf(os.Stderr, "failed to save attempt: %v\n", saveErr)
					}
					e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, Attempts: &todo.Attempts})
					results[todo.FilePath] = perTodo
					continue
				}
			}

			e.updateFrontmatter(ctx, todo, perTodo)
			results[todo.FilePath] = perTodo
		}

		// Group-level post-completion checks: run the suite once for the shared
		// agent session and feed failures back to it, but only when every todo
		// that needed work verified cleanly. On failure, flip the whole group.
		if e.allGroupResultsOK(needsExecution, results) {
			checkResult := &ExecutionResult{Success: true, ExecutorName: e.executor.Name()}
			e.runCheckLoop(ctx, needsExecution, checkResult)
			if !checkResult.Success {
				e.markGroupCheckFailure(ctx, needsExecution, results, checkResult)
			}
		}
	}

	return e.collectResults(todosInGroup, results), nil
}

func (e *TODOExecutor) splitResult(groupResult *ExecutionResult, count int) *ExecutionResult {
	if count == 0 {
		count = 1
	}
	return &ExecutionResult{
		Success:      groupResult.Success,
		ExecutorName: groupResult.ExecutorName,
		TokensUsed:   groupResult.TokensUsed / count,
		CostUSD:      groupResult.CostUSD / float64(count),
		Duration:     groupResult.Duration,
		NumTurns:     groupResult.NumTurns,
		CommitSHA:    groupResult.CommitSHA,
		Transcript:   groupResult.Transcript,
	}
}

func (e *TODOExecutor) collectResults(todosInGroup []*types.TODO, resultMap map[string]*ExecutionResult) []*ExecutionResult {
	out := make([]*ExecutionResult, len(todosInGroup))
	for i, todo := range todosInGroup {
		if r, ok := resultMap[todo.FilePath]; ok {
			out[i] = r
		} else {
			out[i] = &ExecutionResult{ExecutorName: e.executor.Name()}
		}
	}
	return out
}

// updateFrontmatter updates the TODO's frontmatter with execution results.
func (e *TODOExecutor) updateFrontmatter(ctx context.Context, todo *types.TODO, result *ExecutionResult) {
	todo.Status = types.StatusCompleted
	todo.Attempts++

	// Update LLM usage metrics
	if todo.LLM == nil {
		todo.LLM = &types.LLM{}
	}
	todo.LLM.Model = result.ExecutorName
	todo.LLM.TokensUsed = result.TokensUsed
	todo.LLM.CostIncurred = result.CostUSD

	if err := e.saveAttempt(ctx, todo, result); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save attempt: %v\n", err)
	}
	e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, Attempts: &todo.Attempts})
}

func (e *TODOExecutor) activeProvider() Provider {
	if e.provider != nil {
		return e.provider
	}
	return &FileProvider{}
}

func (e *TODOExecutor) saveAttempt(ctx context.Context, todo *types.TODO, result *ExecutionResult) error {
	persistCtx, cancel := providerPersistenceContext(ctx)
	defer cancel()
	return e.activeProvider().SaveAttempt(persistCtx, todo, result)
}

// persistSessionID records the executor's session id (e.g. the cmux/claude
// --session-id) on the provider so the issue carries a session:<id> label.
func (e *TODOExecutor) persistSessionID(ctx context.Context, todo *types.TODO) {
	if todo.LLM == nil || todo.LLM.SessionId == "" {
		return
	}
	e.updateProviderState(ctx, todo, StateUpdate{SessionID: &todo.LLM.SessionId})
}

// sessionIDPersister builds the SetSessionIDHook callback for a run: when the
// executor reports its session id (before launching the agent), record it on
// each todo and persist the session:<id> label immediately, so an interrupted
// run still carries — and can resume — this session.
func (e *TODOExecutor) sessionIDPersister(ctx context.Context, todoList []*types.TODO) func(string) {
	return func(sessionID string) {
		if sessionID == "" {
			return
		}
		for _, todo := range todoList {
			if todo == nil {
				continue
			}
			if todo.LLM == nil {
				todo.LLM = &types.LLM{}
			}
			todo.LLM.SessionId = sessionID
			e.updateProviderState(ctx, todo, StateUpdate{SessionID: &sessionID})
		}
	}
}

func (e *TODOExecutor) updateProviderState(ctx context.Context, todo *types.TODO, updates StateUpdate) {
	persistCtx, cancel := providerPersistenceContext(ctx)
	defer cancel()
	if err := e.activeProvider().UpdateState(persistCtx, todo, updates); err != nil {
		fmt.Fprintf(os.Stderr, "failed to update TODO state: %v\n", err)
	}
}

func providerPersistenceContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return context.WithTimeout(context.Background(), providerPersistenceTimeout)
	}
	if ctx.Err() == nil {
		return ctx, func() {}
	}
	return context.WithTimeout(context.WithoutCancel(ctx), providerPersistenceTimeout)
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
