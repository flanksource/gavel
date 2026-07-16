package todos

import (
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/cel-go/cel"
)

// retryEnv declares the retry predicate's variables with concrete types so the
// full CEL surface — including the list comprehension macros (`exists`, `all`,
// `filter`) over changed_files/session_log — is available, not just the dynamic
// field access gomplate's AnyType binding allows.
var retryEnv = newRetryEnv()

func newRetryEnv() *cel.Env {
	env, err := cel.NewEnv(
		cel.Variable("results", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("test_results", cel.ListType(cel.MapType(cel.StringType, cel.DynType))),
		cel.Variable("changed_files", cel.ListType(cel.StringType)),
		cel.Variable("session_log", cel.ListType(cel.MapType(cel.StringType, cel.DynType))),
		cel.Variable("iteration", cel.IntType),
	)
	if err != nil {
		panic(fmt.Sprintf("retry CEL env: %v", err))
	}
	return env
}

// The definition-of-done verdict is a CEL predicate. After each agent iteration
// the aggregate DoD fixture is run to a TestResults tree; the predicate (default
// types.DefaultRetryExpr — "results.failed > 0 || results.warned > 0") decides
// whether to re-run the agent (retry, with the failing/warned nodes as feedback)
// or stop (verified). The predicate reads {results, test_results, changed_files,
// session_log, iteration}, evaluated via the same gomplate CEL path fixture
// expectations use. changed_files only binds when the hook runs scoped
// (HookContext.Scope == agent.ScopeChanged); session_log is always empty now —
// see sessionLog's doc comment.

// celVerifier is the aggregate definition-of-done verifier: it runs every DoD
// source (configured checks test/lint step fixtures, and the todo's
// `## Verification` section nodes), aggregates them into one TestResults view,
// and evaluates the retry CEL. VerifyResult.Valid = !retry.
type celVerifier struct {
	name      string
	retryExpr string
	steps     []stepFixture
	nodes     []*fixtures.FixtureNode
	// aiStep, when set, is the acceptance-criteria checklist: an LLM executes
	// each item against the change and returns {item, passed, message} per item,
	// exposed to the predicate as results.checklist (so a rule can require
	// `results.checklist.all(i, i.passed)`).
	aiStep  *stepFixture
	workDir string
}

// stepFixture is one synthetic step fixture (checks:test / checks:lint / ai) plus
// its registered runner.
type stepFixture struct {
	fixture fixtures.FixtureTest
	runner  func(fixtures.FixtureTest, fixtures.RunOptions) fixtures.FixtureResult
}

func (v *celVerifier) Name() string { return v.name }

func (v *celVerifier) Verify(hc *agent.HookContext) (agent.VerifyResult, error) {
	results := v.runDeterministic(hc)
	checklist := v.runChecklist(hc)
	summary := resultsSummary(results, checklist)
	output := types.VerificationOutput{Results: results, Checklist: checklist, Summary: summary}
	retry, err := evalRetry(v.retryExpr, results, checklist, hc)
	if err != nil {
		return agent.VerifyResult{}, err
	}
	if !retry {
		return agent.VerifyResult{Valid: true, Output: output}, nil
	}
	// Build the next iteration's request explicitly: the same rendered request
	// (permissions/model/setup unchanged) with a fresh feedback prompt, threading
	// the initial turn's envelope schema (same bytes, not a recompute) so the
	// claude-agent per-turn byte-equality guard holds. The new Runner has no
	// SessionReuse of its own, so resuming the same session is this hook's job.
	retryReq := *hc.Request
	retryReq.Prompt = api.Prompt{
		User:       verifierFeedback(results, checklist),
		SchemaJSON: hc.Request.Prompt.SchemaJSON,
		Source:     "todos.check",
	}
	retryReq.SessionID = hc.Workspace().SessionID
	return agent.VerifyResult{Valid: false, Retry: &retryReq, Output: output}, nil
}

// runDeterministic executes the configured checks and the todo's `## Verification`
// section — the cheap, deterministic part of the definition of done.
func (v *celVerifier) runDeterministic(hc *agent.HookContext) []fixtures.FixtureResult {
	var results []fixtures.FixtureResult
	for _, s := range v.steps {
		results = append(results, s.runner(s.fixture, fixtures.RunOptions{WorkDir: v.workDir}))
	}
	if len(v.nodes) > 0 {
		results = append(results, runFixtureSection(hc, v.nodes, v.workDir)...)
	}
	return results
}

// runChecklist executes the acceptance-criteria checklist (the LLM step) and
// returns one {item, passed, message} entry per criterion — nil when the todo
// has no criteria.
func (v *celVerifier) runChecklist(hc *agent.HookContext) []map[string]any {
	if v.aiStep == nil {
		return nil
	}
	res := v.aiStep.runner(v.aiStep.fixture, fixtures.RunOptions{WorkDir: v.workDir})
	return checklistFromResult(res)
}

// evalRetry builds the CEL activation from the run's results, the checklist, and
// the hook context, and evaluates the retry predicate to a bool. An empty
// expression uses the default; a compile/eval error or non-bool result is
// surfaced (never a silent pass).
func evalRetry(expr string, results []fixtures.FixtureResult, checklist []map[string]any, hc *agent.HookContext) (bool, error) {
	if strings.TrimSpace(expr) == "" {
		expr = types.DefaultRetryExpr
	}
	ast, issues := retryEnv.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("retry predicate %q: %w", expr, issues.Err())
	}
	prg, err := retryEnv.Program(ast)
	if err != nil {
		return false, fmt.Errorf("retry predicate %q: %w", expr, err)
	}
	out, _, err := prg.Eval(map[string]any{
		"results":       resultsSummary(results, checklist),
		"test_results":  resultLeaves(results),
		"changed_files": changedFiles(hc),
		"session_log":   sessionLog(),
		"iteration":     hc.Iteration,
	})
	if err != nil {
		return false, fmt.Errorf("retry predicate %q: %w", expr, err)
	}
	b, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("retry predicate %q did not return a bool: got %T(%v)", expr, out.Value(), out.Value())
	}
	return b, nil
}

// resultsSummary tallies every deterministic leaf into the `results` CEL
// variable: {total, passed, failed, warned, skipped} plus `checklist` — the
// acceptance-criteria per-item results ([]{item, passed, message}), always a
// list so `results.checklist.all(i, i.passed)` holds vacuously without criteria.
func resultsSummary(results []fixtures.FixtureResult, checklist []map[string]any) map[string]any {
	var total, passed, failed, warned, skipped int
	for i := range results {
		walkLeaves(&results[i], func(status task.Status) {
			total++
			switch status {
			case task.StatusFAIL, task.StatusERR, task.StatusFailed:
				failed++
			case task.StatusWarning:
				warned++
			case task.StatusSKIP, task.StatusCancelled:
				skipped++
			default:
				passed++
			}
		})
	}
	if checklist == nil {
		checklist = []map[string]any{}
	}
	return map[string]any{
		"total":     total,
		"passed":    passed,
		"failed":    failed,
		"warned":    warned,
		"skipped":   skipped,
		"checklist": checklist,
	}
}

// resultLeaves flattens the DoD results into the `test_results` CEL variable: one
// {name, status, error} entry per leaf, for per-test predicates.
func resultLeaves(results []fixtures.FixtureResult) []map[string]any {
	var leaves []map[string]any
	for i := range results {
		res := &results[i]
		if len(res.Children) == 0 {
			leaves = append(leaves, leafMap(res))
			continue
		}
		for _, child := range res.Children {
			if child != nil && child.Results != nil {
				leaves = append(leaves, leafMap(child.Results))
			}
		}
	}
	return leaves
}

func leafMap(res *fixtures.FixtureResult) map[string]any {
	return map[string]any{
		"name":   res.Name,
		"status": strings.ToLower(string(res.Status)),
		"error":  res.Error,
	}
}

// walkLeaves visits every leaf status in a result tree (nodes with no children).
func walkLeaves(res *fixtures.FixtureResult, visit func(task.Status)) {
	if res == nil {
		return
	}
	if len(res.Children) == 0 {
		visit(res.Status)
		return
	}
	for _, child := range res.Children {
		if child != nil && child.Results != nil {
			walkLeaves(child.Results, visit)
		}
	}
}

// changedFiles is the `changed_files` CEL variable — the repo-relative paths the
// agent's file-mutating tool uses touched this run (Bash-driven edits are not
// tracked), scoped to hooks the runner asked to act only on changed files
// (HookContext.Scope == agent.ScopeChanged). A ScopeAll run leaves it empty:
// "changed" isn't a meaningful restriction when the hook is meant to act on the
// whole tree. Always a slice so CEL `.exists`/`.size()` never hit a null.
func changedFiles(hc *agent.HookContext) []string {
	if hc == nil || hc.Scope != agent.ScopeChanged {
		return []string{}
	}
	if changed := hc.Workspace().Changed; changed != nil {
		return changed
	}
	return []string{}
}

// sessionLog is the `session_log` CEL variable. The old captain Runner exposed
// the completed LoopIteration, so this turn's raw events (assistant text, tool
// uses, the result) were available directly; the current HookContext only
// carries the run's folded state (Response.Text/Usage, Workspace), not the
// per-event stream.
//
// DEVIATION: session_log is always empty now. Predicates keying off it (e.g.
// matching a specific tool call) will see []; `results`, `test_results`, and
// scoped changed_files remain the reliable retry signals.
func sessionLog() []map[string]any {
	return []map[string]any{}
}

// verifierFeedback renders the failing/warned deterministic nodes plus any
// unmet checklist items as the feedback fed back to the agent on a retry.
func verifierFeedback(results []fixtures.FixtureResult, checklist []map[string]any) string {
	var parts []string
	for i := range results {
		if fb := fixtureFeedback(results[i]); fb != "" {
			parts = append(parts, fb)
		}
	}
	for _, item := range checklist {
		if passed, _ := item["passed"].(bool); passed {
			continue
		}
		line := "- criterion not met: " + fmt.Sprint(item["item"])
		if msg, _ := item["message"].(string); msg != "" {
			line += ": " + msg
		}
		parts = append(parts, line)
	}
	return strings.Join(parts, "\n")
}

// checklistFromResult maps the ai checklist step's per-criterion verdicts into
// the CEL `results.checklist` shape ([]{item, passed, message}). A step that
// errored before producing verdicts surfaces as a single failed item so the
// predicate never verifies on a missing checklist.
func checklistFromResult(res fixtures.FixtureResult) []map[string]any {
	verdicts := checklistVerdicts(res)
	if verdicts == nil {
		return []map[string]any{{"item": res.Name, "passed": false, "message": res.Error}}
	}
	checklist := make([]map[string]any, 0, len(verdicts))
	for _, c := range verdicts {
		checklist = append(checklist, map[string]any{
			"item":    c.Item,
			"passed":  c.Passed,
			"message": c.Message,
		})
	}
	return checklist
}

// checklistVerdicts pulls the ai-step's per-item results back off the fixture
// result (stored under Metadata["checklist"] by RunAIStep). nil when the step did
// not produce a checklist (e.g. it errored building the agent).
func checklistVerdicts(res fixtures.FixtureResult) []fixtures.ChecklistResult {
	if res.Metadata == nil {
		return nil
	}
	if verdicts, ok := res.Metadata["checklist"].([]fixtures.ChecklistResult); ok {
		return verdicts
	}
	return nil
}
