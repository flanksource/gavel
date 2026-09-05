package lifecycle

import (
	"context"
	"fmt"

	captainai "github.com/flanksource/captain/pkg/ai"
	capsetup "github.com/flanksource/captain/pkg/ai/agent/setup"
	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	gavelai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/utils"
	"github.com/google/uuid"

	_ "github.com/flanksource/gavel/fixtures/verifier"
)

type stepInput struct {
	input         promptrun.Input
	meta          todos.RunStartMetadata
	execution     *todos.ExecutionResult
	progress      *progressSink
	sawResult     bool
	brokerFactory todos.ApprovalBroker
	broker        api.PermissionFunc
}

// runInput constructs the same real hooks and callbacks for preview and dispatch.
// Only start invokes the broker factory or publishes execution metadata.
func (h *Host) runInput(exec *todos.ExecutorContext, todo *types.TODO, prepared *preparedStep, opts RunOptions) *stepInput {
	req := prepared.request
	requested := req.SessionID
	req.SessionID = ""
	if opts.Resume {
		req.SessionID = firstNonEmpty(priorSessionID(todo), requested)
	}
	providerSessionID := ""
	if req.Mode == api.ModeCmux && !opts.Resume && prepared.agent == "claude" {
		providerSessionID = firstNonEmpty(requested, uuid.NewString())
	}
	state := &stepInput{
		meta:     h.runMetadata(req, providerSessionID, todo, prepared),
		progress: h.progressSink(exec, todo), brokerFactory: opts.Broker,
	}
	state.execution = &todos.ExecutionResult{ExecutorName: h.executorName(prepared), Runtime: state.meta, Transcript: exec.GetTranscript()}
	hooks := h.Hooks(todo, req, state.meta, exec.RecordRunStart)
	if opts.Provider != nil {
		// A supplied provider skips Captain's setup hook. Gavel's supplied-provider
		// seam still needs local setup, before its final spec recorder.
		recorder := hooks[len(hooks)-1]
		hooks = append(hooks[:len(hooks)-1], &capsetup.Plugin{BaseDir: prepared.workDir}, recorder)
	}
	state.input = promptrun.Input{
		Request:  req,
		Config:   captainai.Config{Model: req.Model, Budget: req.Budget, NoCache: req.NoCache, SessionID: providerSessionID},
		Provider: opts.Provider, Hooks: hooks,
		CallerOwnsCommits: req.Workflow != nil && len(req.Workflow.Commits) > 0,
		Verify:            capverify.Options{Timeout: prepared.timeout, Progress: exec.RecordVerifyProgress},
		OnEvent: func(_ int, ev captainai.Event) {
			h.handleEvent(exec, ev, state.execution, todo, &state.sawResult, state.meta)
		},
		Timeout: prepared.timeout, Constraints: prepared.constraints,
		// Changes are relative to the repository, even when a TODO runs in a subdirectory.
		Repo: utils.GitRoot(prepared.workDir),
	}
	if opts.Broker != nil {
		state.input.Config.CanUseTool = state.canUseTool
	}
	return state
}

func (in *stepInput) canUseTool(ctx context.Context, req api.PermissionRequest) (api.PermissionDecision, error) {
	if in.broker == nil {
		return api.PermissionDecision{}, fmt.Errorf("tool approval broker is not initialized for this run")
	}
	return in.broker(ctx, req)
}

func (in *stepInput) start(exec *todos.ExecutorContext, todo *types.TODO, prepared *preparedStep) error {
	if sessionID := in.input.Config.SessionID; sessionID != "" {
		setSessionID(todo, sessionID)
		exec.RecordSessionID(sessionID)
	}
	exec.RecordRunStart(in.meta)
	exec.Logger.Infof("Resolved TODO runtime: step=%s mode=%s agent=%s provider=%s model=%s effort=%s cwd=%s",
		prepared.definition.Name, in.meta.Driver, in.meta.Agent, firstNonEmpty(in.meta.Provider, "unknown"),
		firstNonEmpty(in.meta.ResolvedModel, "default"), firstNonEmpty(in.meta.Effort, "default"), prepared.workDir)
	gavelai.NormalizeEnv()
	if in.brokerFactory != nil {
		broker, err := in.brokerFactory(exec)
		if err != nil {
			return err
		}
		if broker == nil {
			return fmt.Errorf("tool approval broker factory returned no callback")
		}
		in.broker = broker
	}
	exec.SetVerifyProgressHook(in.progress.record)
	return nil
}
