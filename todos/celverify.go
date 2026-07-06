package todos

import (
	"fmt"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
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
// expectations use.

// celVerifier is the aggregate definition-of-done verifier: it runs every DoD
// source (configured checks test/lint step fixtures, and the todo's
// `## Verification` section nodes), aggregates them into one TestResults view,
// and evaluates the retry CEL. Verdict.OK = !retry.
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

func (v *celVerifier) Verify(rc *agent.RunContext, iter *captainai.LoopIteration) (agent.Verdict, error) {
	results := v.runDeterministic(rc)
	checklist := v.runChecklist(rc)
	retry, err := evalRetry(v.retryExpr, results, checklist, rc, iter)
	if err != nil {
		return agent.Verdict{}, err
	}
	if !retry {
		return agent.Verdict{OK: true}, nil
	}
	return agent.Verdict{
		OK:       false,
		Reason:   "definition of done not met",
		Feedback: verifierFeedback(results, checklist),
	}, nil
}

// runDeterministic executes the configured checks and the todo's `## Verification`
// section — the cheap, deterministic part of the definition of done.
func (v *celVerifier) runDeterministic(rc *agent.RunContext) []fixtures.FixtureResult {
	var results []fixtures.FixtureResult
	for _, s := range v.steps {
		results = append(results, s.runner(s.fixture, fixtures.RunOptions{WorkDir: v.workDir}))
	}
	if len(v.nodes) > 0 {
		results = append(results, runFixtureSection(rc.Ctx, v.nodes, v.workDir)...)
	}
	return results
}

// runChecklist executes the acceptance-criteria checklist (the LLM step) and
// returns one {item, passed, message} entry per criterion — nil when the todo
// has no criteria.
func (v *celVerifier) runChecklist(rc *agent.RunContext) []map[string]any {
	if v.aiStep == nil {
		return nil
	}
	res := v.aiStep.runner(v.aiStep.fixture, fixtures.RunOptions{WorkDir: v.workDir})
	return checklistFromResult(res)
}

// evalRetry builds the CEL activation from the run's results, the checklist, and
// the run context, and evaluates the retry predicate to a bool. An empty
// expression uses the default; a compile/eval error or non-bool result is
// surfaced (never a silent pass).
func evalRetry(expr string, results []fixtures.FixtureResult, checklist []map[string]any, rc *agent.RunContext, iter *captainai.LoopIteration) (bool, error) {
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
		"changed_files": changedFiles(rc),
		"session_log":   sessionLog(iter),
		"iteration":     iteration(iter),
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
// tracked). Always a slice so CEL `.exists`/`.size()` never hit a null.
func changedFiles(rc *agent.RunContext) []string {
	if rc == nil || rc.ChangedFiles == nil {
		return []string{}
	}
	return rc.ChangedFiles
}

// sessionLog is the `session_log` CEL variable — this turn's events (assistant
// text, tool uses, the result) from the in-memory iteration, so a predicate can
// react to what the agent did without reading the log file.
func sessionLog(iter *captainai.LoopIteration) []map[string]any {
	if iter == nil {
		return []map[string]any{}
	}
	events := make([]map[string]any, 0, len(iter.Events))
	for _, ev := range iter.Events {
		events = append(events, map[string]any{
			"kind": string(ev.Kind),
			"text": ev.Text,
			"tool": ev.Tool,
		})
	}
	return events
}

func iteration(iter *captainai.LoopIteration) int {
	if iter == nil {
		return 0
	}
	return iter.Iteration
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

// checklistFromResult maps the ai checklist step's per-criterion verify result
// into the CEL `results.checklist` shape ([]{item, passed, message}). A step that
// errored before producing results surfaces as a single failed item so the
// predicate never verifies on a missing checklist.
func checklistFromResult(res fixtures.FixtureResult) []map[string]any {
	vr := verifyResultFrom(res)
	if vr == nil {
		return []map[string]any{{"item": res.Name, "passed": false, "message": res.Error}}
	}
	checklist := make([]map[string]any, 0, len(vr.AcceptanceCriteria))
	for _, c := range vr.AcceptanceCriteria {
		checklist = append(checklist, map[string]any{
			"item":    c.Criteria,
			"passed":  c.Pass,
			"message": c.Comments,
		})
	}
	return checklist
}

func verifyResultFrom(res fixtures.FixtureResult) *verify.VerifyResult {
	if res.Metadata == nil {
		return nil
	}
	if vr, ok := res.Metadata["verify"].(*verify.VerifyResult); ok {
		return vr
	}
	return nil
}
