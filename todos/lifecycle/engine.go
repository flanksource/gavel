package lifecycle

import (
	"fmt"
	"time"

	"github.com/flanksource/captain/pkg/api"
)

// Context is what the applicability predicates see: the todo as the host
// projects it, and the runs it has had.
type Context struct {
	Subject map[string]any
	Runs    []StepRun
}

// StepRun is one recorded run of a step, oldest first in Context.Runs. Last is
// derived from it: the latest run per step.
type StepRun struct {
	Step        string
	State       string // pending, running, waiting, succeeded, failed, cancelled
	Outcome     string // the status the run landed, when it finished
	PromptRunID string
	StartedAt   *time.Time
	FinishedAt  *time.Time
}

// StepResult is what a finished step run reports back for the outcome
// predicates. Every field is normalised into a map with all keys present, so a
// predicate reads an absent fact as its zero value rather than as an error.
type StepResult struct {
	Run       RunFacts
	Verify    *api.VerifyReport
	Envelope  Envelope
	Plan      *PlanFacts
	Questions []any
}

// RunFacts is the run's own fate, independent of what it produced.
type RunFacts struct {
	State      string // succeeded, failed, cancelled, waiting
	Error      string
	StopReason string
	Iterations int
	CostUSD    float64
}

// Envelope is the structured result the prompt returned, reduced to the fields
// every prompt's envelope shares. Extra is the raw structured output for a
// custom prompt whose outcomes read its own fields.
type Envelope struct {
	Summary   string
	EndStatus string
	Extra     map[string]any
}

// PlanFacts is a plan-envelope result.
type PlanFacts struct {
	Status  string
	Path    string
	Content string
}

// Engine is a compiled lifecycle.
type Engine struct {
	def   Lifecycle
	env   *Env
	steps map[string]*compiledStep
}

type compiledStep struct {
	when     *Program
	inputs   map[string]*Program
	outcomes []*Program
}

// New validates the definition and compiles every predicate, so a broken
// lifecycle fails when it is loaded rather than when a todo reaches the step.
func New(def Lifecycle) (*Engine, error) {
	if err := def.Validate(); err != nil {
		return nil, err
	}
	env, err := NewEnv(def.Subject)
	if err != nil {
		return nil, fmt.Errorf("lifecycle %s: %w", def.Name, err)
	}
	engine := &Engine{def: def, env: env, steps: map[string]*compiledStep{}}
	for _, step := range def.Steps {
		compiled, err := compileStep(env, step)
		if err != nil {
			return nil, fmt.Errorf("lifecycle %s step %s: %w", def.Name, step.Name, err)
		}
		engine.steps[step.Name] = compiled
	}
	return engine, nil
}

func compileStep(env *Env, step Step) (*compiledStep, error) {
	compiled := &compiledStep{inputs: map[string]*Program{}}
	if step.When != "" {
		when, err := env.CompileWhen(step.When)
		if err != nil {
			return nil, fmt.Errorf("when %w", err)
		}
		compiled.when = when
	}
	for name, expr := range step.Inputs {
		program, err := env.CompileWhen(expr)
		if err != nil {
			return nil, fmt.Errorf("input %s %w", name, err)
		}
		compiled.inputs[name] = program
	}
	for i, outcome := range step.Outcomes {
		program, err := env.CompileOutcome(outcome.When)
		if err != nil {
			return nil, fmt.Errorf("outcome %d (%s) %w", i, outcome.Status, err)
		}
		compiled.outcomes = append(compiled.outcomes, program)
	}
	return compiled, nil
}

// Definition is the lifecycle this engine was compiled from.
func (e *Engine) Definition() Lifecycle { return e.def }

// Applicable returns every step whose predicate holds, in definition order,
// auxiliary steps included — a caller listing what a todo can do wants them;
// Next does not.
func (e *Engine) Applicable(c Context) ([]Step, error) {
	vars, err := e.whenVars(c)
	if err != nil {
		return nil, err
	}
	var steps []Step
	for _, step := range e.def.Steps {
		ok, err := e.applies(step, vars)
		if err != nil {
			return nil, err
		}
		if ok {
			steps = append(steps, step)
		}
	}
	return steps, nil
}

// Next is the first non-auxiliary step whose predicate holds. ok is false when
// no step applies — a todo in review or ask is a person's to move.
func (e *Engine) Next(c Context) (Step, bool, error) {
	vars, err := e.whenVars(c)
	if err != nil {
		return Step{}, false, err
	}
	for _, step := range e.def.Steps {
		if step.Auxiliary {
			continue
		}
		ok, err := e.applies(step, vars)
		if err != nil {
			return Step{}, false, err
		}
		if ok {
			return step, true, nil
		}
	}
	return Step{}, false, nil
}

// Explain reports, per step, whether it applies — the reasons a CLI prints next
// to `gavel todos steps`.
func (e *Engine) Explain(c Context) (map[string]bool, error) {
	vars, err := e.whenVars(c)
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(e.def.Steps))
	for _, step := range e.def.Steps {
		ok, err := e.applies(step, vars)
		if err != nil {
			return nil, err
		}
		out[step.Name] = ok
	}
	return out, nil
}

func (e *Engine) applies(step Step, vars map[string]any) (bool, error) {
	compiled := e.steps[step.Name]
	if compiled.when == nil {
		return true, nil
	}
	ok, err := compiled.when.Bool(vars)
	if err != nil {
		return false, fmt.Errorf("step %s when %w", step.Name, err)
	}
	return ok, nil
}

// Outcome is the status the finished step lands the todo in: the first outcome
// whose predicate holds. No outcome holding is an error — a run the definition
// cannot classify must not silently keep the status it started under.
//
// The predicates evaluate against the PRE-RUN Context: `subject`, `runs` and
// `last` are the todo as it was when the step was chosen, not as the run left
// it. What the run produced arrives only through result — `run`, `verify`,
// `envelope`, `plan` and `questions` — so an outcome that needs a fact the run
// changed reads it from there, never from the subject.
//
// Only the step's name is read from the caller: its outcomes are the compiled
// definition's, so a caller holding a stale or partial Step value cannot make
// the engine return a status the lifecycle never declared.
func (e *Engine) Outcome(step Step, c Context, result StepResult) (string, error) {
	compiled, ok := e.steps[step.Name]
	if !ok {
		return "", fmt.Errorf("step %q is not part of lifecycle %s", step.Name, e.def.Name)
	}
	declared, _ := e.def.Step(step.Name)
	vars, err := e.outcomeVars(c, result)
	if err != nil {
		return "", err
	}
	for i, program := range compiled.outcomes {
		holds, err := program.Bool(vars)
		if err != nil {
			return "", fmt.Errorf("step %s outcome %d (%s) %w", step.Name, i, declared.Outcomes[i].Status, err)
		}
		if holds {
			return declared.Outcomes[i].Status, nil
		}
	}
	return "", fmt.Errorf("step %s: no outcome matched run state %q, envelope endStatus %q, verify ran=%v passed=%v",
		step.Name, result.Run.State, result.Envelope.EndStatus, result.Verify != nil && result.Verify.Ran,
		result.Verify != nil && result.Verify.Passed)
}

// Inputs evaluates the step's input expressions.
func (e *Engine) Inputs(step Step, c Context) (map[string]any, error) {
	compiled, ok := e.steps[step.Name]
	if !ok {
		return nil, fmt.Errorf("step %q is not part of lifecycle %s", step.Name, e.def.Name)
	}
	vars, err := e.whenVars(c)
	if err != nil {
		return nil, err
	}
	inputs := make(map[string]any, len(compiled.inputs))
	for name, program := range compiled.inputs {
		value, err := program.Value(vars)
		if err != nil {
			return nil, fmt.Errorf("step %s input %s %w", step.Name, name, err)
		}
		inputs[name] = value
	}
	return inputs, nil
}

func (e *Engine) whenVars(c Context) (map[string]any, error) {
	if err := e.env.CheckSubject(c.Subject); err != nil {
		return nil, err
	}
	runs := make([]any, 0, len(c.Runs))
	last := map[string]any{}
	for _, run := range c.Runs {
		entry := run.vars()
		runs = append(runs, entry)
		last[run.Step] = entry
	}
	return map[string]any{VarSubject: c.Subject, VarRuns: runs, VarLast: last}, nil
}

func (r StepRun) vars() map[string]any {
	entry := map[string]any{
		"step": r.Step, "state": r.State, "outcome": r.Outcome, "prompt_run_id": r.PromptRunID,
	}
	if r.StartedAt != nil {
		entry["started_at"] = *r.StartedAt
	}
	if r.FinishedAt != nil {
		entry["finished_at"] = *r.FinishedAt
	}
	return entry
}

func (e *Engine) outcomeVars(c Context, result StepResult) (map[string]any, error) {
	vars, err := e.whenVars(c)
	if err != nil {
		return nil, err
	}
	verify := api.VerifyReport{}
	if result.Verify != nil {
		verify = *result.Verify
	}
	verifyVars, err := verify.CELVars()
	if err != nil {
		return nil, fmt.Errorf("verify report: %w", err)
	}
	plan := PlanFacts{}
	if result.Plan != nil {
		plan = *result.Plan
	}
	questions := result.Questions
	if questions == nil {
		questions = []any{}
	}
	vars[VarRun] = map[string]any{
		"state": result.Run.State, "error": result.Run.Error, "stop_reason": result.Run.StopReason,
		"iterations": result.Run.Iterations, "cost_usd": result.Run.CostUSD,
	}
	vars[VarVerify] = verifyVars
	vars[VarEnvelope] = result.Envelope.vars()
	vars[VarPlan] = map[string]any{"status": plan.Status, "path": plan.Path, "content": plan.Content}
	vars[VarQuestions] = questions
	return vars, nil
}

func (env Envelope) vars() map[string]any {
	out := make(map[string]any, len(env.Extra)+2)
	for key, value := range env.Extra {
		out[key] = value
	}
	out["summary"] = env.Summary
	out["endStatus"] = env.EndStatus
	return out
}
