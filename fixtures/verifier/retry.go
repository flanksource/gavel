package verifier

import (
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/google/cel-go/cel"

	"github.com/flanksource/gavel/fixtures"
)

// retryEnv declares the predicate's one variable with a concrete type so the
// whole CEL surface — including the list macros (`exists`, `all`, `filter`) over
// `verify.tests` and `verify.checklist` — is available.
//
// A name this env does not declare fails to compile, which is the point: an
// undeclared variable is a loud error at the predicate rather than a silent
// empty value that lets a definition of done pass vacuously.
var retryEnv = newRetryEnv()

func newRetryEnv() *cel.Env {
	env, err := cel.NewEnv(cel.Variable("verify", cel.MapType(cel.StringType, cel.DynType)))
	if err != nil {
		panic(fmt.Sprintf("verify retry CEL env: %v", err))
	}
	return env
}

// retryPredicate is a compiled `verify.retry` expression.
type retryPredicate struct {
	expr    string
	program cel.Program
}

// compileRetry compiles the document's retry predicate, falling back to
// fixtures.DefaultRetryExpr. It runs at hook-assembly time so a broken
// predicate stops the run before an agent turn is spent, not after.
func compileRetry(expr string) (*retryPredicate, error) {
	if strings.TrimSpace(expr) == "" {
		expr = fixtures.DefaultRetryExpr
	}
	ast, issues := retryEnv.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("verify retry predicate %q: %w", expr, issues.Err())
	}
	program, err := retryEnv.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("verify retry predicate %q: %w", expr, err)
	}
	return &retryPredicate{expr: expr, program: program}, nil
}

// Eval evaluates the predicate against a report. A non-bool result or an
// evaluation error is surfaced, never treated as "do not retry".
func (p *retryPredicate) Eval(report api.VerifyReport) (bool, error) {
	vars, err := report.CELVars()
	if err != nil {
		return false, err
	}
	out, _, err := p.program.Eval(map[string]any{"verify": vars})
	if err != nil {
		return false, fmt.Errorf("verify retry predicate %q: %w", p.expr, err)
	}
	retry, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("verify retry predicate %q did not return a bool: got %T(%v)",
			p.expr, out.Value(), out.Value())
	}
	return retry, nil
}

// applyRetry folds the predicate's answer into the report.
//
// The predicate can only make the verdict stricter. A report's verdict is the
// tree it ran — captain's Validate rejects a report whose state disagrees with
// its leaves — so a predicate cannot talk a red tree into passing. It can add a
// reason to re-run a green one, and when it does the reason becomes a failing
// node of its own, so the tree still justifies the verdict and the agent's
// feedback names the predicate rather than reporting a mysterious retry.
func (p *retryPredicate) applyRetry(report *api.VerifyReport) error {
	retry, err := p.Eval(*report)
	if err != nil {
		return err
	}
	if !retry || !report.Passed {
		return nil
	}
	node := api.VerifyNode{
		Name:      "verify.retry",
		Framework: Kind,
		Failed:    true,
		Message:   fmt.Sprintf("the retry predicate %q holds on an otherwise green run", p.expr),
		Context:   &api.VerifyNodeContext{CELExpression: p.expr},
	}
	report.Tests = append(report.Tests, node)
	report.Summary = api.SummarizeNodes(report.Tests)
	report.State = api.StateForReport(report.Tests)
	report.Passed = false
	report.Reason = node.Message
	report.Feedback = Feedback(report.Tests, report.Checklist)
	return nil
}
