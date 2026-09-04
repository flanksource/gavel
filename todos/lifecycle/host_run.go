package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	capsetup "github.com/flanksource/captain/pkg/ai/agent/setup"
	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/commons/logger"
	gavelai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/claude"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/utils"
	"github.com/google/uuid"
)

// RunOptions is what the caller decides about one step run; everything else
// comes from the lifecycle, the project's configuration and the todo.
type RunOptions struct {
	// Exec carries the run's logger, transcript and notification sink. Nil gets
	// a plain context with the standard logger.
	Exec *todos.ExecutorContext
	// Request is the caller's explicit spec — parsed flags or the dashboard
	// payload — folded as the top layer.
	Request api.Spec
	// Prior are the layers a continuation inherits from the run it continues.
	Prior []api.SpecLayer
	// Resume continues the todo's recorded agent session instead of opening one.
	Resume bool
	// Message is the user turn a resumed session continues with — an answer, a
	// reviewer's feedback — sent in place of the rendered prompt the session has
	// already acted on. It requires Resume.
	Message string
	// Concurrent admits the run alongside a live one the caller has confirmed.
	Concurrent bool
	// Broker builds the tool-approval callback; nil for a host that cannot answer.
	Broker todos.ApprovalBroker
	// Provider, when set, replaces the one promptrun would build: the seam a test
	// drives a scripted event stream through. promptrun then adds no setup plugin,
	// so the host adds one itself.
	Provider captainai.Provider
}

// StepOutcome is one finished step run: the facts the outcome predicates saw,
// the status they chose, and the execution record the provider persists.
type StepOutcome struct {
	Step Step
	// Status is the status the lifecycle's outcomes chose, or OutcomeKeep.
	Status string
	Result StepResult
	// Execution is the run as the provider records it: the attempt, the
	// transcript, the envelope, and the definition-of-done verdict.
	Execution *todos.ExecutionResult
	// Admission is the durable identity Captain admitted the run under.
	Admission todos.RunPreparationResult
	// Request is the spec the run was dispatched with.
	Request api.Spec
}

// preparedStep is a step resolved down to one dispatchable request.
type preparedStep struct {
	definition   todoprompt.Definition
	class        types.RunMode
	request      captainai.Request
	timeout      time.Duration
	workDir      string
	template     string
	existingPlan string
	agent        string
	// trace is captain's provenance for the spec fold, lowest precedence first.
	trace []api.SpecLayer
}

// RunStep runs one step of the lifecycle for a todo: the prompt rendered, the
// spec layers folded, the run admitted, dispatched through captain's promptrun,
// and its result classified by the step's outcomes. It writes no status — that
// is OnOutcome's — so a caller can decide what to persist when the run itself
// could not be classified.
func (h *Host) RunStep(ctx context.Context, todo *types.TODO, step Step, opts RunOptions) (*StepOutcome, error) {
	resolution, err := h.Resolve(ctx, todo, step, opts)
	if err != nil {
		return nil, err
	}
	return h.Dispatch(ctx, todo, resolution, opts)
}

// Dispatch runs an already-resolved step: the run admitted under its durable
// identity, dispatched through captain's promptrun, and its result classified by
// the step's outcomes. It writes no status — that is OnOutcome's — so a caller
// can decide what to persist when the run itself could not be classified.
//
// It takes a Resolution rather than a Step so a caller that showed the user a
// preview dispatches that exact request, not a second fold of it.
func (h *Host) Dispatch(ctx context.Context, todo *types.TODO, resolution *Resolution, opts RunOptions) (*StepOutcome, error) {
	if resolution == nil || resolution.prepared == nil {
		return nil, fmt.Errorf("lifecycle dispatch: step was not resolved")
	}
	start := time.Now()
	exec := opts.Exec
	if exec == nil {
		exec = todos.NewExecutorContext(ctx, logger.StandardLogger(), nil)
	}
	step, prepared := resolution.Step, resolution.prepared
	admission, err := h.admit(exec, todo, step, prepared, opts)
	if err != nil {
		return nil, err
	}
	d := h.dispatch(exec, todo, prepared, opts)
	h.recordIterations(exec, admission.PromptRunID, &d)
	outcome := h.collect(exec, todo, step, prepared, d, start)
	outcome.Admission = admission
	outcome.Request = prepared.request
	status, err := h.Def.Outcome(step, resolution.lc, outcome.Result)
	if err != nil {
		return outcome, err
	}
	outcome.Status = status
	return outcome, nil
}

// resolveStep is the shared front half of a run: the todo projected onto the
// lifecycle's variables, and the step prepared into a dispatchable request.
// RunStep dispatches what it returns; Resolve reports it.
func (h *Host) resolveStep(ctx context.Context, todo *types.TODO, step Step, opts RunOptions) (Context, *preparedStep, error) {
	if h.Def == nil {
		return Context{}, nil, fmt.Errorf("lifecycle host: no lifecycle loaded")
	}
	if _, ok := h.Def.Definition().Step(step.Name); !ok {
		return Context{}, nil, fmt.Errorf("step %q is not part of lifecycle %s", step.Name, h.Def.Definition().Name)
	}
	lc, err := h.Context(ctx, todo)
	if err != nil {
		return Context{}, nil, err
	}
	prepared, err := h.prepare(ctx, todo, step, lc, opts)
	if err != nil {
		return Context{}, nil, err
	}
	return lc, prepared, nil
}

// prepare resolves the step to the request captain will run: the prompt
// reference, the step's spec with its placeholders expanded, every spec layer
// folded, and the template rendered with the step's inputs.
func (h *Host) prepare(ctx context.Context, todo *types.TODO, step Step, lc Context, opts RunOptions) (*preparedStep, error) {
	workDir := h.stepWorkDir(todo)
	definition, err := h.promptFor(step)
	if err != nil {
		return nil, err
	}
	class := definition.Class
	stepSpec, err := expandSpec(step.Spec, map[string]any{VarSubject: lc.Subject})
	if err != nil {
		return nil, fmt.Errorf("step %s: %w", step.Name, err)
	}
	var frontmatter []api.SpecLayer
	var template string
	if class != types.ModeVerify {
		if frontmatter, template, err = PromptLayers(workDir, []*types.TODO{todo}, definition); err != nil {
			return nil, err
		}
	}
	resolved, err := ResolveLayers(LayerInput{
		Config: h.Config, Step: step.Name, Frontmatter: frontmatter, StepSpec: stepSpec,
		Todos: []*types.TODO{todo}, Prior: opts.Prior, Host: h.Kind, Request: opts.Request,
	})
	if err != nil {
		return nil, err
	}
	spec := resolved.Spec
	// A verify step runs no agent turn: it executes the definition of done. The
	// only thing there that needs a model is the acceptance-criteria grader, so a
	// project that configures no model still verifies todos whose definition of
	// done is fixture steps — the same rule graderSpec applies to the document.
	if class != types.ModeVerify || len(todo.AcceptanceCriteria) > 0 {
		if err := ApplyModel(&spec, opts.Request.Effort); err != nil {
			return nil, err
		}
		if err := RequireModel(spec); err != nil {
			return nil, fmt.Errorf("step %s: %w", step.Name, err)
		}
	}
	timeout, err := ApplyTimeout(&spec)
	if err != nil {
		return nil, err
	}
	ApplyClassInvariants(&spec, class)
	if err := ValidateSpec(spec); err != nil {
		return nil, fmt.Errorf("step %s: %w", step.Name, err)
	}
	// A verify step IS its verifiers: it runs no agent turn, so a workflow that
	// declares none has nothing to judge with. Refuse before anything is
	// dispatched — a run admitted here would write an attempt and a status for a
	// verdict it could never reach, which is how a todo with no definition of
	// done came to read as checked.
	if class == types.ModeVerify && !declaresVerify(spec.Workflow) {
		return nil, fmt.Errorf(
			"step %s: todo %s has no definition of done (no verification fixture, acceptance criteria, or configured checks)",
			step.Name, todoName(todo))
	}
	prepared := &preparedStep{
		definition: definition, class: class, request: spec, timeout: timeout,
		workDir: workDir, template: template, trace: resolved.Trace,
	}
	prepared.agent, _ = claude.ResolveAgent(spec.Name)
	if class != types.ModeVerify {
		if err := h.render(ctx, todo, step, lc, prepared); err != nil {
			return nil, err
		}
	}
	if message := strings.TrimSpace(opts.Message); message != "" {
		if !opts.Resume {
			return nil, fmt.Errorf("step %s: a message continues a resumed session; this run resumes nothing", step.Name)
		}
		prepared.request.Prompt.User = message
	}
	if prepared.request.Setup == nil {
		prepared.request.Setup = &shell.Setup{}
	}
	if prepared.request.Setup.Cwd == "" {
		prepared.request.Setup.Cwd = workDir
	}
	return prepared, nil
}

// render fills the prompt template with the todo and the step's declared
// inputs. A verify-only step renders nothing: it runs the definition of done.
func (h *Host) render(ctx context.Context, todo *types.TODO, step Step, lc Context, prepared *preparedStep) error {
	inputs, err := h.Def.Inputs(step, lc)
	if err != nil {
		return err
	}
	if existing, ok := inputs["existingPlan"].(string); ok {
		prepared.existingPlan = existing
	}
	backlog := ""
	if prepared.definition.Envelope == todoprompt.EnvelopeTriage {
		backlog = h.backlog(ctx, todo)
	}
	req, _, err := todoprompt.Render([]*types.TODO{todo}, todoprompt.Options{
		WorkDir:      prepared.workDir,
		Prompt:       prepared.definition.Name,
		Envelope:     prepared.definition.Envelope,
		Mode:         prepared.class,
		Spec:         prepared.request,
		Template:     prepared.template,
		ExistingPlan: prepared.existingPlan,
		Backlog:      backlog,
		Inputs:       inputs,
	})
	if err != nil {
		return fmt.Errorf("step %s: render prompt %s: %w", step.Name, prepared.definition.Name, err)
	}
	prepared.request = req
	return nil
}

// backlog is the duplicate-detection index a triage run compares against. A
// backlog that cannot be listed degrades duplicate detection but not the other
// four verdicts, so it is logged rather than fatal.
func (h *Host) backlog(ctx context.Context, todo *types.TODO) string {
	if h.Provider == nil {
		return ""
	}
	candidates, err := h.Provider.List(ctx, todos.DiscoveryFilters{
		ExcludeStatuses: []types.Status{types.StatusCompleted},
	})
	if err != nil {
		logger.Warnf("triage duplicate detection is degraded: could not list the backlog: %v", err)
		return ""
	}
	return todos.BuildBacklogIndex(candidates, []*types.TODO{todo})
}

// admit allocates the run's durable Captain identity before anything is
// dispatched, and wires the provider's persistence to the execution context so
// the run's session id, runtime, progress and notices land as they happen.
func (h *Host) admit(exec *todos.ExecutorContext, todo *types.TODO, step Step, prepared *preparedStep, opts RunOptions) (todos.RunPreparationResult, error) {
	lifecycleProvider, ok := h.Provider.(todos.RunLifecycleProvider)
	if !ok {
		return todos.RunPreparationResult{}, nil
	}
	spec := prepared.request
	runtime := api.RuntimeOf(spec.Provider, spec.Mode)
	persistCtx, cancel := todos.PersistenceContext(exec)
	defer cancel()
	admission, err := lifecycleProvider.PrepareRun(persistCtx, todo, todos.RunPreparation{
		Mode: prepared.class, Prompt: step.Name, ExecutorName: h.executorName(prepared),
		Resume: opts.Resume, Concurrent: opts.Concurrent,
		Requested: captaindb.PromptRunRuntimeSelection{
			Provider: runtime.Provider, Mode: string(spec.Mode), Model: spec.Name, Effort: string(spec.Effort),
		},
		Spec: spec,
	})
	if err != nil {
		return todos.RunPreparationResult{}, fmt.Errorf("prepare native TODO run: %w", err)
	}
	exec.RecordRunPrepared(admission)
	exec.SetSessionIDHook(func(sessionID string) {
		setSessionID(todo, sessionID)
		h.updateState(exec, todo, todos.StateUpdate{SessionID: &sessionID})
	})
	exec.SetRunStartHook(h.runStartRecorder(exec, todo, prepared.class))
	exec.SetNoticesHook(func(sessionID string, notices []api.Notice) {
		provider, ok := h.Provider.(todos.RunNoticeProvider)
		if !ok {
			return
		}
		// A verify-only step never opens an agent session, so there is no
		// transcript for its notices to ride on; its verdict reaches the
		// dashboard as the iteration's VerifyReport instead.
		if strings.TrimSpace(sessionID) == "" {
			return
		}
		persistCtx, cancel := todos.PersistenceContext(exec)
		defer cancel()
		if err := provider.RecordRunNotices(persistCtx, sessionID, notices); err != nil {
			exec.Logger.Errorf("failed to record run notices: %v", err)
		}
	})
	return admission, nil
}

func (h *Host) executorName(prepared *preparedStep) string {
	return string(prepared.request.Mode) + "-" + prepared.agent
}

// runStartRecorder persists the resolved runtime when the run starts, and
// comments it on the todo once the session is known.
func (h *Host) runStartRecorder(exec *todos.ExecutorContext, todo *types.TODO, class types.RunMode) func(todos.RunStartMetadata) {
	commented := false
	return func(meta todos.RunStartMetadata) {
		if meta.Mode == "" {
			meta.Mode = string(class)
		}
		persistCtx, cancel := todos.PersistenceContext(exec)
		defer cancel()
		if lifecycleProvider, ok := h.Provider.(todos.RunLifecycleProvider); ok {
			if err := lifecycleProvider.RecordRunStart(persistCtx, todo, meta); err != nil {
				exec.Logger.Errorf("failed to record native TODO run start: %v", err)
			}
		}
		if commented || meta.SessionID == "" || h.Provider == nil {
			return
		}
		if err := h.Provider.Comment(persistCtx, todo, todos.RenderRunStartComment(meta)); err != nil {
			exec.Logger.Errorf("failed to comment TODO run metadata: %v", err)
		}
		commented = true
	}
}

// updateState persists a mid-run state change. A failure is logged on the
// run's own logger rather than failing the run: the session id or runtime it
// carries is bookkeeping the outcome does not depend on.
func (h *Host) updateState(exec *todos.ExecutorContext, todo *types.TODO, update todos.StateUpdate) {
	if h.Provider == nil {
		return
	}
	persistCtx, cancel := todos.PersistenceContext(exec)
	defer cancel()
	if err := h.Provider.UpdateState(persistCtx, todo, update); err != nil {
		exec.Logger.Errorf("failed to update TODO state: %v", err)
	}
}

// dispatched is what came back from captain, with the two facts only the
// dispatching context can tell: whether the run was cancelled by its caller,
// and whether it hit its deadline before reporting a result.
type dispatched struct {
	out       promptrun.Result
	err       error
	cancelled bool
	timedOut  bool
	execution *todos.ExecutionResult
}

// dispatch drives the request through captain's promptrun seam — provider
// construction, tool-policy enforcement, the workflow's checks, the setup plugin
// and the generate→verify loop. What stays here is what is gavel's: the todo's
// identity on the run, the commit pipeline, the transcript, and the progress
// sink.
func (h *Host) dispatch(exec *todos.ExecutorContext, todo *types.TODO, prepared *preparedStep, opts RunOptions) dispatched {
	req := prepared.request
	requested := req.SessionID
	req.SessionID = ""
	providerSessionID := ""
	if opts.Resume {
		req.SessionID = firstNonEmpty(priorSessionID(todo), requested)
	}
	// The cmux runtime is handed a fresh claude session id up front so it launches
	// `--session-id <id>` and the host can follow the session log live.
	if req.Mode == api.ModeCmux && !opts.Resume && prepared.agent == "claude" {
		providerSessionID = firstNonEmpty(requested, uuid.NewString())
		setSessionID(todo, providerSessionID)
		exec.RecordSessionID(providerSessionID)
	}
	meta := h.runMetadata(req, providerSessionID, todo, prepared)
	exec.RecordRunStart(meta)
	exec.Logger.Infof("Resolved TODO runtime: step=%s mode=%s agent=%s provider=%s model=%s effort=%s cwd=%s",
		prepared.definition.Name, meta.Driver, meta.Agent, firstNonEmpty(meta.Provider, "unknown"),
		firstNonEmpty(meta.ResolvedModel, "default"), firstNonEmpty(meta.Effort, "default"), prepared.workDir)
	gavelai.NormalizeEnv()

	execution := &todos.ExecutionResult{ExecutorName: h.executorName(prepared), Runtime: meta, Transcript: exec.GetTranscript()}
	var canUseTool api.PermissionFunc
	if opts.Broker != nil {
		broker, err := opts.Broker(exec)
		if err != nil {
			return dispatched{err: err, execution: execution}
		}
		canUseTool = broker
	}
	progress := h.progressSink(exec, todo)
	exec.SetVerifyProgressHook(progress.record)

	hooks := h.Hooks(todo, req, meta, exec.RecordRunStart)
	if opts.Provider != nil {
		// promptrun adds no setup plugin for a caller-supplied provider, which it
		// takes to own its own workspace. The test seam does not, so the host adds
		// the plugin itself — before the recorder, so the recorder still trails it.
		recorder := hooks[len(hooks)-1]
		hooks = append(hooks[:len(hooks)-1], &capsetup.Plugin{BaseDir: prepared.workDir}, recorder)
	}
	runCtx, cancel := context.WithTimeout(exec, prepared.timeout)
	defer cancel()
	sawResult := false
	out, err := promptrun.Run(runCtx, promptrun.Input{
		Request: req,
		Config: captainai.Config{
			Model: req.Model, Budget: req.Budget, NoCache: req.NoCache,
			SessionID: providerSessionID, CanUseTool: canUseTool,
		},
		Provider:          opts.Provider,
		Hooks:             hooks,
		CallerOwnsCommits: req.Workflow != nil && len(req.Workflow.Commits) > 0,
		Verify:            capverify.Options{Timeout: prepared.timeout, Progress: exec.RecordVerifyProgress},
		OnEvent: func(_ int, ev captainai.Event) {
			h.handleEvent(exec, ev, execution, todo, &sawResult, meta)
		},
		Timeout: prepared.timeout,
		// Repo is the root of the tree workDir sits in, not workDir itself: a todo
		// carrying a subdirectory CWD still has its edits recorded relative to the
		// root, which is the namespace the commit hooks compare against.
		Repo: utils.GitRoot(prepared.workDir),
	})
	if progress.err != nil {
		err = errors.Join(err, progress.err)
	}
	return dispatched{
		out: out, err: err, execution: execution,
		cancelled: errors.Is(context.Cause(runCtx), todos.ErrExecutionCancelled),
		timedOut:  errors.Is(runCtx.Err(), context.DeadlineExceeded) && !sawResult,
	}
}

func (h *Host) runMetadata(req captainai.Request, providerSessionID string, todo *types.TODO, prepared *preparedStep) todos.RunStartMetadata {
	runtime := api.RuntimeOf(req.Provider, req.Mode)
	return todos.RunStartMetadata{
		SessionID:     firstNonEmpty(firstNonEmpty(req.SessionID, providerSessionID), priorSessionID(todo)),
		Mode:          string(prepared.class),
		Driver:        string(req.Mode),
		Agent:         prepared.agent,
		Provider:      runtime.Provider,
		RuntimeMode:   string(runtime.Mode),
		ResolvedModel: req.Name,
		Effort:        string(req.Effort),
	}
}
