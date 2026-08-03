package todos

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	captainapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

const providerPersistenceTimeout = 30 * time.Second

var ErrExecutionCancelled = errors.New("todo run stopped by user")

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
	Cancelled        bool
	ExecutorName     string        // Which executor was used
	TokensUsed       int           // Total tokens consumed
	CostUSD          float64       // Cost in USD
	Duration         time.Duration // Total execution time
	NumTurns         int           // Number of interaction rounds
	ActionsPerformed []string      // List of actions taken (tool uses, etc.)
	ErrorMessage     string
	CommitSHA        string
	Runtime          RunStartMetadata
	Transcript       *ExecutionTranscript
	// Envelope fields — the agent's structured final result. EndStatus is empty
	// when no envelope was captured (the executor logged the degradation and the
	// transport Success decides).
	Summary   string
	EndStatus types.EndStatus
	Questions []types.AgentQuestion
	Plan      *types.PlanResult
	// DoD is the definition-of-done verdict for an implement run: nil when the
	// todo had no verifiers (Verification fixture / configured checks), else Ran
	// is true and Passed reports whether every verifier passed within the
	// iteration budget. It drives the verified/unverified terminal status.
	DoD *DoDOutcome
}

// DoDOutcome is the terminal verdict of a run's definition-of-done verifiers:
// Passed is true only when the agent loop stopped because all verifiers passed
// (captain "condition-met"), false when the iteration budget ran out with a
// verifier still failing.
type DoDOutcome struct {
	Ran    bool                      `json:"ran"`
	Passed bool                      `json:"passed"`
	Output *types.VerificationOutput `json:"output,omitempty"`
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
	resume    bool
	// mode drives the envelope→status mapping (see applyOutcome). Post-run
	// checks live in the run loop itself now: fixture-backed verify plugins
	// built by BuildCheckVerifiers and threaded through AgentRunConfig.
	mode types.RunMode
}

// SetResume marks a normal Execute call as a continuation of the prior agent
// session. Native storage can then reuse the Captain session while allocating
// a new prompt run only when the prior operation is already terminal.
func (e *TODOExecutor) SetResume(resume bool) {
	e.resume = resume
}

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
	ctx.Logger.Infof("Starting TODO execution: title=%q id=%s mode=%s", todo.Title, todo.ID, e.Mode())

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

	// Check if test already passes (skip if so). A plan run changes nothing, so
	// neither the pre-check nor post-verification applies to it.
	if e.Mode() == types.ModeRun && len(todo.StepsToReproduce) > 0 {
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

	// Execute with the configured driver. The driver logs its resolved agent,
	// backend, model, and effort after Captain constructs the provider.
	if err := e.prepareRun(ctx, todo); err != nil {
		if errors.Is(err, ErrRunDispatchAlreadyClaimed) || errors.Is(err, ErrRunResumeModeMismatch) {
			return nil, fmt.Errorf("prepare native TODO run: %w", err)
		}
		todo.Status = types.StatusFailed
		todo.Attempts++
		e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, Attempts: &todo.Attempts})
		return nil, fmt.Errorf("prepare native TODO run: %w", err)
	}
	ctx.SetSessionIDHook(e.sessionIDPersister(ctx, []*types.TODO{todo}))
	ctx.SetRunStartHook(e.runStartPersister(ctx, []*types.TODO{todo}))
	var progress *progressPersister
	if provider, ok := e.activeProvider().(RunProgressProvider); ok {
		progress = newProgressPersister(ctx, provider, todo)
		ctx.Context = fixtures.WithProgressSink(ctx.Context, progress.Sink)
	}
	result, err := e.executor.Execute(ctx, todo)
	if progress != nil {
		if progressErr := progress.Close(); progressErr != nil {
			err = errors.Join(err, progressErr)
			if result == nil {
				result = &ExecutionResult{ExecutorName: e.executor.Name()}
			}
			result.Success = false
			result.ErrorMessage = progressErr.Error()
		}
	}
	// The executor may generate its own session id (e.g. cmux's claude
	// --session-id); persist it now that it is known, regardless of outcome.
	e.persistSessionID(ctx, todo)
	if err != nil {
		if result != nil && result.Cancelled {
			ctx.Logger.Infof("Execution cancelled: %v", err)
			todo.Status = types.StatusPending
			todo.Attempts++
			if saveErr := e.saveAttempt(ctx, todo, result); saveErr != nil {
				return result, fmt.Errorf("save cancelled attempt: %w", saveErr)
			}
			e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, Attempts: &todo.Attempts})
			return result, err
		}
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

	// The todo's `## Verification` fixture is its definition of done: it runs
	// inside the agent loop as a verifier (BuildCheckVerifiers), so the run
	// iterates against it and its terminal verdict lands the todo verified or
	// unverified in applyOutcome — no separate post-run gate.

	// Map the run's envelope onto the todo (status/plan/summary/questions).
	ctx.Logger.Infof("TODO execution completed successfully")
	if applyErr := e.applyOutcome(ctx, todo, result); applyErr != nil {
		todo.Status = types.StatusFailed
		todo.Attempts++
		e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, Attempts: &todo.Attempts})
		result.Success = false
		result.ErrorMessage = applyErr.Error()
		return result, applyErr
	}

	return result, nil
}

// GroupExecutor is implemented by executors that support combined group execution.
type GroupExecutor interface {
	ExecuteGroup(ctx *ExecutorContext, todosInGroup []*types.TODO) (*ExecutionResult, error)
}

// ExecuteGroup orchestrates group execution: one AI session for multiple TODOs,
// then independent verification per TODO.
func (e *TODOExecutor) ExecuteGroup(ctx *ExecutorContext, todosInGroup []*types.TODO) ([]*ExecutionResult, error) {
	if policy, ok := e.activeProvider().(GroupExecutionPolicy); ok && !policy.SupportsGroupedExecution() {
		return nil, fmt.Errorf("grouped TODO execution is not supported by the native PostgreSQL runtime; run one issue at a time")
	}
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

	// Pre-check: filter out TODOs whose steps already pass (never for plan runs —
	// they change nothing and always need the investigation).
	var needsExecution []*types.TODO
	results := make(map[string]*ExecutionResult)
	for _, todo := range todosInGroup {
		if e.Mode() != types.ModePlan && len(todo.StepsToReproduce) > 0 && e.stepsAlreadyPass(ctx, todo.StepsToReproduce) {
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
		ctx.SetRunStartHook(e.runStartPersister(ctx, needsExecution))
		groupResult, err = groupExec.ExecuteGroup(ctx, needsExecution)
		// The group executor may generate a session id per todo (e.g. cmux's
		// claude --session-id); persist it now that it is known.
		for _, todo := range needsExecution {
			e.persistSessionID(ctx, todo)
		}
		if err != nil {
			for _, todo := range needsExecution {
				todo.Status = types.StatusFailed
				if groupResult != nil && groupResult.Cancelled {
					todo.Status = types.StatusPending
				}
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

		// Map each TODO's envelope + definition-of-done verdict onto its status.
		// Verification runs in-loop as a DoD verifier (see the single-todo path),
		// so there is no separate post-run verification gate here either.
		for _, todo := range needsExecution {
			perTodo := e.splitResult(groupResult, len(needsExecution))

			if applyErr := e.applyOutcome(ctx, todo, perTodo); applyErr != nil {
				todo.Status = types.StatusFailed
				e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, Attempts: &todo.Attempts})
				perTodo.Success = false
				perTodo.ErrorMessage = applyErr.Error()
			}
			results[todo.FilePath] = perTodo
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
		Cancelled:    groupResult.Cancelled,
		ExecutorName: groupResult.ExecutorName,
		TokensUsed:   groupResult.TokensUsed / count,
		CostUSD:      groupResult.CostUSD / float64(count),
		Duration:     groupResult.Duration,
		NumTurns:     groupResult.NumTurns,
		CommitSHA:    groupResult.CommitSHA,
		Runtime:      groupResult.Runtime,
		Transcript:   groupResult.Transcript,
		Summary:      groupResult.Summary,
		EndStatus:    groupResult.EndStatus,
		Questions:    groupResult.Questions,
		Plan:         groupResult.Plan,
		DoD:          groupResult.DoD,
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
func (e *TODOExecutor) activeProvider() Provider {
	return e.provider
}

func (e *TODOExecutor) saveAttempt(ctx context.Context, todo *types.TODO, result *ExecutionResult) error {
	provider := e.activeProvider()
	if provider == nil {
		// Pure outcome tests and embedders may intentionally omit persistence.
		// Unlike the retired behavior, omission never falls back to `.todos`.
		return nil
	}
	persistCtx, cancel := providerPersistenceContext(ctx)
	defer cancel()
	return provider.SaveAttempt(persistCtx, todo, result)
}

func (e *TODOExecutor) prepareRun(ctx *ExecutorContext, todo *types.TODO) error {
	lifecycle, ok := e.activeProvider().(RunLifecycleProvider)
	if !ok {
		return nil
	}
	persistCtx, cancel := providerPersistenceContext(ctx)
	defer cancel()
	preparation := RunPreparation{
		Mode:         e.Mode(),
		ExecutorName: e.executor.Name(),
		Resume:       e.resume,
	}
	if runtimeProvider, ok := e.executor.(RunRuntimeProvider); ok {
		preparation.Requested = runtimeProvider.RunRuntimeSelection()
	}
	if renderer, ok := e.executor.(RunSpecProvider); ok {
		spec, err := renderer.RenderRunSpec(ctx, todo)
		if err != nil {
			return fmt.Errorf("render native TODO spec: %w", err)
		}
		preparation.Spec = spec
	}
	return lifecycle.PrepareRun(persistCtx, todo, preparation)
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

func (e *TODOExecutor) runStartPersister(ctx context.Context, todoList []*types.TODO) func(RunStartMetadata) {
	commented := false
	return func(meta RunStartMetadata) {
		if strings.TrimSpace(meta.Mode) == "" {
			meta.Mode = string(e.Mode())
		}
		if strings.TrimSpace(meta.Driver) == "" {
			meta.Driver = e.executor.Name()
		}
		persistCtx, cancel := providerPersistenceContext(ctx)
		defer cancel()
		if lifecycle, ok := e.activeProvider().(RunLifecycleProvider); ok {
			for _, todo := range todoList {
				if todo == nil {
					continue
				}
				if err := lifecycle.RecordRunStart(persistCtx, todo, meta); err != nil {
					fmt.Fprintf(os.Stderr, "failed to record native TODO run start: %v\n", err)
				}
			}
		}
		if commented || strings.TrimSpace(meta.SessionID) == "" {
			return
		}
		body := renderRunStartComment(meta)
		provider := e.activeProvider()
		if provider == nil {
			fmt.Fprintln(os.Stderr, "failed to comment TODO run metadata: native PostgreSQL provider is required")
			return
		}
		for _, todo := range todoList {
			if todo == nil {
				continue
			}
			if err := provider.Comment(persistCtx, todo, body); err != nil {
				fmt.Fprintf(os.Stderr, "failed to comment TODO run metadata: %v\n", err)
			}
		}
		commented = true
	}
}

func renderRunStartComment(meta RunStartMetadata) string {
	var b strings.Builder
	b.WriteString("**Todo run started**\n\n")
	b.WriteString("- **Session ID:** `" + commentValue(meta.SessionID, "unknown") + "`\n")
	b.WriteString("- **Mode:** `" + commentValue(meta.Mode, "run") + "`\n")
	b.WriteString("- **Driver:** `" + commentValue(meta.Driver, "unknown") + "`\n")
	b.WriteString("- **Agent:** `" + commentValue(meta.Agent, "unknown") + "`\n")
	b.WriteString("- **Provider:** `" + commentValue(meta.Provider, "unknown") + "`\n")
	b.WriteString("- **Backend:** `" + commentValue(meta.Backend, "default") + "`\n")
	b.WriteString("- **Resolved Model:** `" + commentValue(meta.ResolvedModel, "default") + "`\n")
	b.WriteString("- **Effort:** `" + commentValue(meta.Effort, "default") + "`")
	return b.String()
}

func commentValue(value, fallback string) string {
	if v := strings.TrimSpace(value); v != "" {
		return v
	}
	return fallback
}

func (e *TODOExecutor) updateProviderState(ctx context.Context, todo *types.TODO, updates StateUpdate) {
	provider := e.activeProvider()
	if provider == nil {
		applyStateUpdate(&todo.TODOFrontmatter, updates)
		return
	}
	persistCtx, cancel := providerPersistenceContext(ctx)
	defer cancel()
	if err := provider.UpdateState(persistCtx, todo, updates); err != nil {
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

// ExecuteSection runs all fixture nodes in a section.
// Returns the results of executing each node using the fixtures runner infrastructure.
func (e *TODOExecutor) ExecuteSection(ctx context.Context, nodes []*fixtures.FixtureNode) []fixtures.FixtureResult {
	return runFixtureSection(ctx, nodes, e.workDir, nil)
}

// runFixtureSection runs a list of fixture nodes in workDir via the fixtures
// runner infrastructure, returning one result per test node. Shared by the
// reproduction/verification section runners and the in-loop DoD verifier.
func runFixtureSection(ctx context.Context, nodes []*fixtures.FixtureNode, workDir string, spec *captainapi.Spec) []fixtures.FixtureResult {
	reporter := fixtures.NewExecutionReporter(nodes, workDir, nil, fixtures.ProgressSinkFromContext(ctx))
	if err := reporter.Publish(ctx); err != nil {
		return []fixtures.FixtureResult{fixtureErr(fixtures.FixtureTest{Name: "verification progress"}, err.Error())}
	}
	results, err := runFixtureSectionWithProgress(ctx, nodes, workDir, spec, reporter)
	if err != nil {
		return append(results, fixtureErr(fixtures.FixtureTest{Name: "verification progress"}, err.Error()))
	}
	return results
}

func runFixtureSectionWithProgress(
	ctx context.Context,
	nodes []*fixtures.FixtureNode,
	workDir string,
	spec *captainapi.Spec,
	reporter *fixtures.ExecutionReporter,
) ([]fixtures.FixtureResult, error) {
	evaluator, err := fixtures.NewCELEvaluator()
	if err != nil {
		return []fixtures.FixtureResult{{
			Status: "error",
			Error:  fmt.Sprintf("failed to create CEL evaluator: %v", err),
		}}, nil
	}
	opts := fixtures.RunOptions{Context: ctx, WorkDir: workDir, Evaluator: evaluator, Spec: spec}

	var results []fixtures.FixtureResult
	var walk func(*fixtures.FixtureNode) error
	walk = func(node *fixtures.FixtureNode) error {
		if node == nil {
			return nil
		}
		if node.Test != nil {
			if err := reporter.StartFixture(ctx, node); err != nil {
				return err
			}
			opts.Progress = func(done, total int) error {
				return reporter.UpdateFixture(ctx, node, done, total)
			}
			result := dispatchFixture(ctx, *node.Test, opts)
			results = append(results, result)
			if err := reporter.CompleteFixture(ctx, node, result); err != nil {
				return err
			}
		}
		for _, child := range node.Children {
			if err := walk(child); err != nil {
				return err
			}
		}
		return nil
	}
	for _, node := range nodes {
		if err := walk(node); err != nil {
			return results, err
		}
	}
	return results, nil
}

// dispatchFixture runs one fixture test with the same node-type dispatch the
// fixture engine's (unexported) executeFixture uses: `ai` steps via the
// AIStepRunner hook, `yaml test`/`yaml lint` fences via the runner-step hooks,
// everything else via the type registry. Replicated here because the section
// runner works on a node list, not a whole fixture file.
func dispatchFixture(ctx context.Context, test fixtures.FixtureTest, opts fixtures.RunOptions) fixtures.FixtureResult {
	if reason := test.ShouldSkip(); reason != "" {
		return fixtures.FixtureResult{Name: test.Name, Status: task.StatusSKIP, Test: test, Error: reason}
	}
	if test.IsAIStep() {
		if fixtures.AIStepRunner == nil {
			return fixtureErr(test, "AI step runner not registered; import _ \"github.com/flanksource/gavel/fixtures/types\"")
		}
		return fixtures.AIStepRunner(test, opts)
	}
	if test.IsRunnerStep() {
		runner := fixtures.TestStepRunner
		if test.IsLintStep() {
			runner = fixtures.LintStepRunner
		}
		if runner == nil {
			return fixtureErr(test, "runner step hook not registered; import _ \"github.com/flanksource/gavel/fixtures/types\"")
		}
		return runner(test, opts)
	}
	fixtureType, err := fixtures.DefaultRegistry.GetForFixture(test)
	if err != nil {
		return fixtureErr(test, err.Error())
	}
	return fixtureType.Run(ctx, test, opts)
}

func fixtureErr(test fixtures.FixtureTest, msg string) fixtures.FixtureResult {
	return fixtures.FixtureResult{Name: test.Name, Status: task.StatusERR, Test: test, Error: msg}
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
