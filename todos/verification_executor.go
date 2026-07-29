package todos

import (
	"fmt"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
)

// verificationExecutor runs an issue's definition-of-done without generating
// or editing code. The empty request makes captain's Runner execute its Verify
// hooks exactly once, which is the canonical manual "check" behaviour.
type verificationExecutor struct {
	workDir   string
	verifiers []agent.Verify
}

// newVerificationExecutor builds a check-only run. spec is already resolved for
// types.ModeVerify, so it is both the run spec (whether to verify) and the
// grader — there is no implementer here whose runtime could contaminate it.
func newVerificationExecutor(workDir string, todoList []*types.TODO, spec *api.Spec) (*verificationExecutor, error) {
	var grader api.Spec
	if spec != nil {
		grader = *spec
	}
	verifiers, _, err := BuildCheckVerifiers(CheckVerifierOptions{
		WorkDir: workDir,
		Todos:   todoList,
		Run:     spec,
		Grader:  grader,
	})
	if err != nil {
		return nil, err
	}
	if len(verifiers) == 0 {
		return nil, fmt.Errorf("no verification fixture, acceptance criteria, or configured checks")
	}
	return &verificationExecutor{workDir: workDir, verifiers: verifiers}, nil
}

func (e *verificationExecutor) Name() string { return "gavel-fixtures" }

// RenderRunSpec deliberately returns the empty spec that Execute dispatches: a
// check-only run carries no prompt, and its definition of done is stamped onto
// the persisted spec from the issue's verification markdown.
func (e *verificationExecutor) RenderRunSpec(_ *ExecutorContext, _ *types.TODO) (api.Spec, error) {
	return api.Spec{}, nil
}

func (e *verificationExecutor) Execute(ctx *ExecutorContext, _ *types.TODO) (*ExecutionResult, error) {
	start := time.Now()
	ctx.RecordRunStart(RunStartMetadata{Mode: string(types.ModeVerify), ResolvedModel: e.Name()})

	request := captainai.Request{}
	request.SetCwd(e.workDir)
	hooks := make([]any, len(e.verifiers))
	for i, verifier := range e.verifiers {
		hooks[i] = verifier
	}
	runner := agent.Runner[any]{
		Request: request,
		Hooks:   hooks,
		Repo:    e.workDir,
		Cwd:     e.workDir,
		Scope:   agent.ScopeAll,
	}
	run, err := runner.Run(ctx)
	result := &ExecutionResult{
		Success:      err == nil,
		ExecutorName: e.Name(),
		Duration:     time.Since(start),
	}
	if err != nil {
		result.ErrorMessage = err.Error()
		return result, err
	}

	passed := len(run.Verdicts) > 0
	var output *types.VerificationOutput
	for _, verdict := range run.Verdicts {
		if !verdict.Valid {
			passed = false
		}
		switch value := verdict.Output.(type) {
		case types.VerificationOutput:
			copy := value
			output = &copy
		case *types.VerificationOutput:
			output = value
		}
	}
	result.DoD = &DoDOutcome{Ran: true, Passed: passed, Output: output}
	return result, nil
}
