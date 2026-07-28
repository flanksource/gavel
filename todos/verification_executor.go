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

func newVerificationExecutor(workDir string, todoList []*types.TODO, spec *api.Spec) (*verificationExecutor, error) {
	verifiers, _, err := BuildCheckVerifiers(workDir, todoList, spec)
	if err != nil {
		return nil, err
	}
	if len(verifiers) == 0 {
		return nil, fmt.Errorf("no verification fixture, acceptance criteria, or configured checks")
	}
	return &verificationExecutor{workDir: workDir, verifiers: verifiers}, nil
}

func (e *verificationExecutor) Name() string { return "gavel-fixtures" }

// RenderRunPrompt deliberately returns an empty prompt. Native admission keeps
// verify-only prompt runs empty and stores the issue's verification markdown in
// the workflow verification spec instead.
func (e *verificationExecutor) RenderRunPrompt(_ *ExecutorContext, _ *types.TODO) (string, error) {
	return "", nil
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
