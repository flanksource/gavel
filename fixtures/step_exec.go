package fixtures

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strconv"
	"time"

	"github.com/flanksource/clicky/task"
)

// StepRef identifies a step on its result: Metadata["step"].
type StepRef struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

// runStep runs one fixture test through its Kind. It is the one execution path
// for every frontend: render the step's config, Execute it (or Start it in the
// background) under its retry policy, wait on its Observer, evaluate its
// `cel:`, record its outputs under steps.<id>, apply its exports and register
// its cleanup — all within the step's timeout.
func runStep(ctx context.Context, test FixtureTest, opts RunOptions) FixtureResult {
	spec := test.StepSpec()
	if spec.Kind == "" {
		return nodeError(test, fmt.Sprintf("unable to determine fixture type for %q: it declares no command, query, runner fence, ai step or kind", test.Name))
	}
	kind, err := DefaultKinds.Lookup(spec.Kind)
	if err != nil {
		return nodeError(test, err.Error())
	}
	if opts.DaemonPort > 0 {
		test.TemplateVars = maps.Clone(test.TemplateVars)
		if test.TemplateVars == nil {
			test.TemplateVars = map[string]any{}
		}
		test.TemplateVars["port"] = strconv.Itoa(opts.DaemonPort)
	}
	step, err := newStepContext(test, spec, opts)
	if err != nil {
		return nodeError(test, err.Error())
	}
	if test.Step == nil {
		return runKind(ctx, kind, step)
	}
	// A typed step depends on the ones before it (their services, exports and
	// todo state), so once one fails the rest are skipped, naming the cause,
	// instead of each failing on a consequence of it.
	if failed := step.Scope.FailedStep(); failed != "" {
		return normalizeResult(FixtureResult{Type: spec.Kind, Status: task.StatusSKIP, Error: fmt.Sprintf("skipped: step %s failed", failed)}, step)
	}
	result := runKind(ctx, kind, step)
	if result.Status == task.StatusFAIL || result.Status == task.StatusERR {
		step.Scope.MarkFailed(stepLabel(step))
	}
	return result
}

// errStepTimeout is the cancellation cause of a step's own deadline, telling it
// apart from its run being cancelled or reaching a deadline of its own.
var errStepTimeout = errors.New("step timed out")

// runKind runs the step under its own declared timeout; a step that did not
// pass once that deadline fired is failed as timed out.
func runKind(ctx context.Context, kind Kind, step *StepContext) FixtureResult {
	timeout := step.Test.declaredTimeout()
	if timeout == nil {
		return runKindUntimed(ctx, kind, step)
	}
	ctx, cancel := context.WithTimeoutCause(ctx, *timeout, errStepTimeout)
	defer cancel()
	result := runKindUntimed(ctx, kind, step)
	if isPassingStatus(result.Status) || !errors.Is(context.Cause(ctx), errStepTimeout) {
		return result
	}
	result.TimedOut = true
	message := fmt.Sprintf("step %s timed out after %s", stepLabel(step), *timeout)
	if result.Error != "" {
		message += ": " + result.Error
	}
	result = result.Failf("%s", message)
	if attempts, ok := result.Metadata["attempts"].([]AttemptReport); ok {
		result = recordFinalAttempt(result, attempts)
	}
	return result
}

// declaredTimeout is the deadline a node declares for itself: a typed step's
// `timeout:` (or the document default), a markdown fixture's `timeout`.
func (fixture FixtureTest) declaredTimeout() *time.Duration {
	if fixture.Step != nil {
		if timeout := fixture.Step.EffectiveTimeout(); timeout != nil {
			d := timeout.Std()
			return &d
		}
		return nil
	}
	if fixture.Expected.Timeout != nil {
		return fixture.Expected.Timeout
	}
	return fixture.Timeout
}

func runKindUntimed(ctx context.Context, kind Kind, step *StepContext) FixtureResult {
	spec := step.Spec
	if _, raw := kind.(SelfTemplating); !raw && len(spec.Config) > 0 {
		rendered, err := renderConfig(step)
		if err != nil {
			return stepError(step, fmt.Errorf("step %s: %w", stepLabel(step), err))
		}
		step.Spec.Config = rendered
	}
	if spec.Background {
		return startBackground(ctx, kind, step)
	}
	return executeStep(ctx, kind, step)
}

func newStepContext(test FixtureTest, spec StepSpec, opts RunOptions) (*StepContext, error) {
	if opts.Scope == nil {
		return nil, fmt.Errorf("fixtures: step %q has no fixture scope", test.Name)
	}
	sourceDir := test.SourceDir
	if sourceDir == "" {
		sourceDir = opts.WorkDir
	}
	step := &StepContext{Test: test, Spec: spec, Opts: opts, Scope: opts.Scope, WorkDir: opts.Setup.Dir(sourceDir)}
	if test.Step != nil {
		tmp, err := opts.Scope.TempDir()
		if err != nil {
			return nil, err
		}
		step.TmpDir = tmp
	}
	return step, nil
}

func executeStep(ctx context.Context, kind Kind, step *StepContext) FixtureResult {
	scope := step.Scope
	policy, err := effectiveWait(kind, step.Spec)
	if err != nil {
		return stepError(step, err)
	}
	retry := retrier{policy: effectiveRetry(step.Spec), check: evalStepCEL, clock: realClock{}, rand: defaultRand,
		vars: func(outcome StepOutcome, err error) map[string]any {
			vars := retryVarsFor(outcome, err)
			vars["env"], vars["steps"] = scope.Env.Map(), scope.StepOutputs()
			return vars
		}}
	outcome, attempts := retry.run(ctx, func(ctx context.Context) (StepOutcome, error) { return kind.Execute(ctx, step) })
	finalStarted := time.Now()
	result := normalizeResult(outcome.Result, step)
	if outcome.Cleanup != nil {
		if err := scope.AddCleanup(stepLabel(step), outcome.Cleanup); err != nil {
			result = result.Errorf(err, "register cleanup")
		}
	}
	outputs := outcome.Outputs
	var waited *WaitReport
	if policy != nil && isPassingStatus(result.Status) {
		report, last := waitOn(ctx, kind.(Observer), step, outcome.Handle, *policy)
		waited = &report
		if last != nil {
			outputs = last
		}
		result = applyWait(result, report)
	}
	if step.Spec.CEL != "" && isPassingStatus(result.Status) {
		result = applyStepCEL(result, step.Spec.CEL, stepCELVarsWithTraces(outputs, scope, step.Opts.Traces.Collect(ctx).CELVars))
	}
	if err := scope.SetOutputs(step.Spec.ID, outputs); err != nil {
		result = result.Errorf(err, "record outputs")
	}
	exports := map[string]ExportRecord{}
	if isPassingStatus(result.Status) {
		if exports, err = applyExports(step, outcome.Exports, outputs); err != nil {
			result = result.Errorf(err, "export")
		}
	}
	if step.Test.Step != nil {
		for i := range attempts {
			attempts[i].Result.Type = step.Spec.Kind
			attempts[i].Result.Name = step.Test.Name
			attempts[i].Result.Metadata["step"] = StepRef{ID: step.Spec.ID, Kind: step.Spec.Kind}
		}
		if len(attempts) > 0 {
			attempts[len(attempts)-1].DurationMs += time.Since(finalStarted).Milliseconds()
		}
		stampStepMetadata(&result, step.Spec, outputs, exports, waited, attempts)
		result = recordFinalAttempt(result, attempts)
	}
	return result
}

// effectiveWait merges the step's wait over the document defaults over the
// kind's DefaultWait over the built-in timing, and checks the kind can honour
// it. Nil when neither the step nor its kind asks for a wait.
func effectiveWait(kind Kind, spec StepSpec) (*WaitPolicy, error) {
	var kindDefault, file *WaitPolicy
	if waiter, ok := kind.(DefaultWaiter); ok {
		kindDefault = waiter.DefaultWait(spec)
	}
	if spec.Wait == nil && kindDefault == nil {
		return nil, nil
	}
	if _, ok := kind.(Observer); !ok {
		return nil, fmt.Errorf("kind %q cannot wait: it does not implement Observe", kind.Name())
	}
	if spec.Defaults != nil {
		file = spec.Defaults.Wait
	}
	policy := mergeWait(spec.Wait, file, mergeWait(kindDefault, nil, builtinWait(spec.EffectiveTimeout())))
	if err := validateWait(*policy); err != nil {
		return nil, fmt.Errorf("step %q: %w", spec.ID, err)
	}
	return policy, nil
}

func effectiveRetry(spec StepSpec) RetryPolicy {
	var file *RetryPolicy
	if spec.Defaults != nil {
		file = spec.Defaults.Retry
	}
	if merged := mergeRetry(spec.Retry, file); merged != nil {
		return *merged
	}
	return RetryPolicy{Attempts: 1}
}

func waitOn(ctx context.Context, observer Observer, step *StepContext, handle any, policy WaitPolicy) (WaitReport, map[string]any) {
	return waiter{
		policy:  policy,
		observe: func(ctx context.Context) (map[string]any, error) { return observer.Observe(ctx, step, handle) },
		vars:    func(outputs map[string]any) map[string]any { return stepCELVars(outputs, step.Scope) },
		check:   evalStepCEL,
		clock:   realClock{},
		rand:    defaultRand,
	}.run(ctx)
}

func applyWait(result FixtureResult, report WaitReport) FixtureResult {
	if report.Outcome == WaitSatisfied {
		return result
	}
	result.CELExpression = report.Policy.Until
	result.CELTrace = report.Trace
	result.TimedOut = report.Outcome == WaitTimeout
	return result.Failf("%s", report.Error)
}

// stepCELVarsWithTraces are a step's CEL roots with its armed traces' summaries
// over its outputs, each under its kind (`http_proxy`, `terminal`, …).
func stepCELVarsWithTraces(outputs map[string]any, scope *StepScope, traceVars map[string]any) map[string]any {
	vars := stepCELVars(outputs, scope)
	maps.Copy(vars, traceVars)
	return vars
}

func applyStepCEL(result FixtureResult, expression string, vars map[string]any) FixtureResult {
	ok, trace, err := evalStepCEL(expression, vars)
	result.CELExpression = expression
	if err == nil && ok {
		result.CELResult = true
		return result
	}
	// The fixture Env is the process environment: keep it out of the recorded
	// variables, which reach reports and PR comments.
	delete(vars, "env")
	result.CELVars, result.CELTrace = vars, trace
	if err != nil {
		return result.Errorf(err, "evaluate cel %s", expression)
	}
	return result.Failf("%s is false", expression)
}

func startBackground(ctx context.Context, kind Kind, step *StepContext) FixtureResult {
	background, ok := kind.(BackgroundKind)
	if !ok {
		return stepError(step, fmt.Errorf("kind %q cannot run in the background: it does not implement Start", kind.Name()))
	}
	handle, err := background.Start(ctx, step)
	if err != nil {
		return stepError(step, fmt.Errorf("start background step %s: %w", stepLabel(step), err))
	}
	if stopped := step.Scope.AddBackground(stepLabel(step), stampedBackground{handle: handle, spec: step.Spec}); stopped != nil {
		return stepError(step, fmt.Errorf("background step %s started after fixture teardown: %s", stepLabel(step), stopped.Error))
	}
	result := normalizeResult(FixtureResult{Status: task.StatusPASS}, step)
	stampStepMetadata(&result, step.Spec, nil, map[string]ExportRecord{}, nil, nil)
	return result
}

// stampedBackground stamps the step's record onto a background verdict, merged
// over whatever the kind recorded (a watcher's events), so the teardown verdict
// carries the same step/outputs/exports contract as the step that started it.
type stampedBackground struct {
	handle BackgroundHandle
	spec   StepSpec
}

func (b stampedBackground) Stop(ctx context.Context) FixtureResult {
	result := b.handle.Stop(ctx)
	stampStepMetadata(&result, b.spec, nil, map[string]ExportRecord{}, nil, nil)
	return result
}

func normalizeResult(result FixtureResult, step *StepContext) FixtureResult {
	if result.Name == "" {
		result.Name = step.Test.Name
	}
	if result.Test.Name == "" {
		result.Test = step.Test
	}
	if result.Status == "" {
		result.Status = task.StatusERR
		result.Error = fmt.Sprintf("kind %q returned no status for step %s", step.Spec.Kind, stepLabel(step))
	}
	return result
}

func stepError(step *StepContext, err error) FixtureResult {
	result := nodeError(step.Test, err.Error())
	if step.Test.Step != nil {
		stampStepMetadata(&result, step.Spec, nil, map[string]ExportRecord{}, nil, nil)
	}
	return result
}

// stampStepMetadata writes the result contract the UI renders a typed step from.
func stampStepMetadata(result *FixtureResult, spec StepSpec, outputs map[string]any, exports map[string]ExportRecord, wait *WaitReport, attempts []AttemptReport) {
	result.Type = spec.Kind
	if result.Metadata == nil {
		result.Metadata = map[string]interface{}{}
	}
	if outputs == nil {
		outputs = map[string]any{}
	}
	if attempts == nil {
		attempts = []AttemptReport{}
	}
	result.Metadata["step"] = StepRef{ID: spec.ID, Kind: spec.Kind}
	result.Metadata["outputs"] = outputs
	result.Metadata["exports"] = exports
	result.Metadata["attempts"] = attempts
	result.Metadata["background"] = spec.Background
	if wait != nil {
		result.Metadata["wait"] = *wait
	}
}

// stepLabel names a step in errors, export origins and teardown: its id, else
// its test name.
func stepLabel(step *StepContext) string {
	if step.Spec.ID != "" {
		return step.Spec.ID
	}
	return step.Test.Name
}
