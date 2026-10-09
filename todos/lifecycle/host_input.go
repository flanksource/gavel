package lifecycle

import (
	"sync"

	captainai "github.com/flanksource/captain/pkg/ai"
	capsetup "github.com/flanksource/captain/pkg/ai/agent/setup"
	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	gavelai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/triagenew"
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
	// hooks are gavel's own hooks; setup is the plugin gavel adds itself when a
	// supplied provider skips Captain's, which must trail every other hook.
	hooks     []any
	setup     *capsetup.Plugin
	triageNew *triagenew.Review
}

// runInput constructs the same real hooks and callbacks for preview and dispatch.
// Only start publishes execution metadata.
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
	state := &stepInput{meta: h.runMetadata(req, providerSessionID, todo, prepared)}
	state.execution = &todos.ExecutionResult{ExecutorName: h.executorName(prepared), Runtime: state.meta, Transcript: exec.GetTranscript()}
	state.hooks = h.Hooks(todo, req, state.meta)
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
	if prepared.definition.Name == "triage.new" {
		state.triageNew = h.newTriageReview(state, todo)
		state.input.Config.Tools = state.triageNew.Tools()
	}
	if opts.Approvals || prepared.definition.Name == "triage.new" {
		state.input.Approvals = &promptrun.ApprovalOptions{RequestedBy: "gavel-dashboard"}
	}
	return state
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
