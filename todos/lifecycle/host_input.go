package lifecycle

import (
	"context"
	"fmt"
	"sync"

	captainai "github.com/flanksource/captain/pkg/ai"
	capsetup "github.com/flanksource/captain/pkg/ai/agent/setup"
	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	gavelai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/tools"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/utils"
	"github.com/google/uuid"

	_ "github.com/flanksource/gavel/fixtures/verifier"
)

type stepInput struct {
	input     promptrun.Input
	meta      todos.RunStartMetadata
	execution *todos.ExecutionResult
	sawResult bool
	eventMu   sync.Mutex
	// hooks are gavel's own hooks, built from runHooks; setup is the plugin
	// gavel adds itself when a supplied provider skips Captain's, which must
	// trail every other hook.
	hooks    []any
	runHooks RunHooks
	// admitted runs once Captain committed a recorded run's admission.
	admitted func() error
	setup    *capsetup.Plugin
	// tools is the state the step's todo tools share; nil without tools.
	tools *tools.Run
}

// runInput constructs the same real hooks and callbacks for preview and dispatch.
// Only start publishes execution metadata.
func (h *Host) runInput(exec *todos.ExecutorContext, todo *types.TODO, prepared *preparedStep, opts RunOptions) (*stepInput, error) {
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
	state := &stepInput{meta: h.runMetadata(req, providerSessionID, todo, prepared)}
	state.execution = &todos.ExecutionResult{ExecutorName: h.executorName(prepared), Runtime: state.meta, Transcript: exec.GetTranscript()}
	defs, err := h.stepTools(context.WithoutCancel(exec), todo, prepared, state)
	if err != nil {
		return nil, err
	}
	state.runHooks = h.runHooks(todo, req, HookOptions{Meta: state.meta, pr: prepared.pr, tools: state.tools})
	state.hooks = state.runHooks.build(prepared.pr.forRun(state.tools))
	if opts.Provider != nil {
		// A supplied provider skips Captain's setup hook, and gavel's
		// supplied-provider seam still needs local setup.
		state.setup = &capsetup.Plugin{BaseDir: prepared.workDir}
	}
	state.input = promptrun.Input{
		Resolved: api.ResolvedSpec{
			Spec: req, Trace: prepared.trace, Provenance: prepared.provenance, Warnings: prepared.warnings,
		},
		RuntimePresets: prepared.runtimePresets,
		RuntimeProfile: prepared.runtimeProfile,
		Config:         captainai.Config{Model: req.Model, Budget: req.Budget, NoCache: req.NoCache, SessionID: providerSessionID},
		Provider:       opts.Provider, Hooks: state.hookList(nil),
		ContextSources:    prepared.contextSources,
		CallerOwnsCommits: req.Workflow != nil && len(req.Workflow.Commits) > 0,
		Verify:            capverify.Options{Timeout: prepared.timeout},
		OnEvent: func(_ int, ev captainai.Event) {
			state.eventMu.Lock()
			defer state.eventMu.Unlock()
			h.handleEvent(exec, ev, state.execution, todo, &state.sawResult, state.meta)
		},
		Timeout: prepared.timeout,
		// Changes are relative to the repository, even when a TODO runs in a subdirectory.
		Repo: utils.GitRoot(prepared.workDir),
	}
	state.input.Config.Tools = defs
	switch {
	case opts.OnApproval != nil && opts.Approvals:
		return nil, fmt.Errorf("run options set both OnApproval and Approvals; a run has one approval broker")
	case opts.OnApproval != nil:
		state.input.Config.OnApproval = opts.OnApproval
	case opts.Approvals || tools.NeedsApprovals(defs):
		state.input.Approvals = &promptrun.ApprovalOptions{RequestedBy: "gavel-dashboard"}
	}
	return state, nil
}

// stepTools builds the tool sets the step declares, bound to this run's todo
// and to the run state the host reads back once the run is over.
func (h *Host) stepTools(ctx context.Context, todo *types.TODO, prepared *preparedStep, state *stepInput) ([]api.ToolDefinition, error) {
	if len(prepared.tools) == 0 {
		return nil, nil
	}
	state.tools = &tools.Run{}
	defs, err := tools.For(prepared.tools, tools.Options{
		Provider: h.Provider, Subject: todo, WorkDir: prepared.workDir, Labels: todos.AllowedLabels(ctx, h.Provider),
		Preview: h.Preview, Run: state.tools, PlanStep: prepared.class == types.ModePlan,
		PRWrite: prepared.pr != nil && prepared.pr.write,
	})
	if err != nil {
		return nil, fmt.Errorf("step tools: %w", err)
	}
	return defs, nil
}

// hookList is the hook order promptrun is handed: gavel's hooks, the admission
// hook of a recorded run, then the setup plugin gavel supplies itself.
func (in *stepInput) hookList(admitted *admittedHook) []any {
	hooks := append([]any(nil), in.hooks...)
	if admitted != nil {
		hooks = append(hooks, admitted)
	}
	if in.setup != nil {
		hooks = append(hooks, in.setup)
	}
	return hooks
}

// record files the run in Captain's store under the admission the runtime
// cleared, and runs admitted once that admission is committed.
func (in *stepInput) record(recording *promptrun.Recording, admitted func() error) {
	in.input.Record = recording
	in.input.Hooks = in.hookList(&admittedHook{admitted: admitted})
	in.admitted = admitted
}

func (in *stepInput) start(exec *todos.ExecutorContext, todo *types.TODO, prepared *preparedStep) {
	if sessionID := in.input.Config.SessionID; sessionID != "" {
		setSessionID(todo, sessionID)
		exec.RecordSessionID(sessionID)
	}
	exec.Logger.Infof("Resolved TODO runtime: step=%s mode=%s agent=%s provider=%s model=%s effort=%s cwd=%s",
		prepared.definition.Name, in.meta.Driver, in.meta.Agent, firstNonEmpty(in.meta.Provider, "unknown"),
		firstNonEmpty(in.meta.ResolvedModel, "default"), firstNonEmpty(in.meta.Effort, "default"), prepared.workDir)
	gavelai.NormalizeEnv()
}
