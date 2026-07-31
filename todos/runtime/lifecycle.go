package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

var (
	_ todos.RunLifecycleProvider = (*Provider)(nil)
	_ todos.PlanContentProvider  = (*Provider)(nil)
)

type activeRun struct {
	issue   *native.Issue
	link    *native.PromptRunLink
	run     *captaindb.PromptRun
	session *captaindb.Session
}

// PrepareRun creates Captain's authoritative session and prompt-run records,
// links that exact run to one issue, and activates it before an external agent
// can be dispatched. The issue version is part of the deterministic admission
// identity, so an exact retry resolves the same launch while a later attempt
// receives a new identity.
func (p *Provider) PrepareRun(ctx context.Context, todo *types.TODO, preparation todos.RunPreparation) error {
	issueID, err := p.todoID(todo)
	if err != nil {
		return err
	}
	mode := preparation.Mode
	if mode == "" {
		mode = types.ModeRun
	}
	step, err := stepForMode(mode)
	if err != nil {
		return err
	}
	issue, err := p.repository.GetIssue(ctx, issueID)
	if err != nil {
		return err
	}
	if issue.WorkspaceID != p.workspace.ID {
		return fmt.Errorf("%w: TODO %s belongs to workspace %s, provider owns %s", native.ErrCrossWorkspace, issue.ID, issue.WorkspaceID, p.workspace.ID)
	}
	seed := fmt.Sprintf("%s:%s:%d", issue.ID, step, todo.Version)
	sessionID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("gavel-todo-session:"+seed))
	promptRunID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("gavel-todo-prompt-run:"+seed))
	if todo.Version != issue.Version {
		// A contender may read after the winning admission has already advanced
		// the issue version. Recognize that exact deterministic launch as an owned
		// dispatch instead of misclassifying it as an unrelated stale mutation.
		if issue.ActivePromptRunID != nil && *issue.ActivePromptRunID == promptRunID {
			links, linkErr := p.repository.ListPromptRuns(ctx, issue.ID)
			if linkErr != nil {
				return linkErr
			}
			for _, link := range links {
				if link.PromptRunID == promptRunID && link.StepKind == step {
					return fmt.Errorf(
						"%w: Captain prompt run %s for issue %s already has an external dispatcher",
						todos.ErrRunDispatchAlreadyClaimed, promptRunID, issue.ID,
					)
				}
			}
		}
		return fmt.Errorf("%w: issue %s expected version %d, current version %d", native.ErrVersionConflict, issue.ID, todo.Version, issue.Version)
	}

	ordinal, err := p.nextPromptOrdinal(ctx, issue.ID, step)
	if err != nil {
		return err
	}
	executorName := strings.TrimSpace(preparation.ExecutorName)
	if executorName == "" {
		executorName = "unknown"
	}
	promptMarkdown := preparation.Spec.Prompt.User
	if strings.TrimSpace(promptMarkdown) == "" && mode != types.ModeVerify {
		promptMarkdown = issue.Body
	}
	cwd := strings.TrimSpace(todo.CWD)
	if cwd == "" {
		cwd = p.workDir
	}
	sessionInput := p.todoOperationSessionInput(issue, todoOperationSessionOptions{
		ID: sessionID, Operation: string(step), Provider: executorName, CWD: cwd, Prompt: promptMarkdown,
	})
	if preparation.Resume && issue.ActivePromptRunID != nil {
		previousRun, runErr := p.captain.GetPromptRun(ctx, *issue.ActivePromptRunID)
		if runErr != nil {
			return runErr
		}
		previousSession, sessionErr := p.captain.GetSession(ctx, previousRun.SessionID)
		if sessionErr != nil {
			return sessionErr
		}
		if !terminalPromptRun(previousRun.State) {
			links, linkErr := p.repository.ListPromptRuns(ctx, issue.ID)
			if linkErr != nil {
				return linkErr
			}
			activeStep := native.StepKind("")
			for _, link := range links {
				if link.PromptRunID == previousRun.ID {
					activeStep = link.StepKind
					break
				}
			}
			if activeStep == "" {
				return fmt.Errorf("%w: active prompt run %s is not linked to issue %s", native.ErrLinkConflict, previousRun.ID, issue.ID)
			}
			if activeStep != step {
				return fmt.Errorf(
					"%w: active Captain prompt run %s is step %q and cannot resume as %q",
					todos.ErrRunResumeModeMismatch, previousRun.ID, activeStep, step,
				)
			}
			// A waiting/running prompt run is one interactive operation. Resume it
			// in place rather than manufacturing a second active root operation.
			p.markPrepared(issue.ID, previousRun.ID)
			return p.replaceTODO(ctx, todo, issue, cwd)
		}
		// A terminal operation gets a new prompt run, but continuation keeps the
		// authoritative Captain/provider session identity.
		sessionInput = captaindb.CreateSessionInput{
			ID: previousSession.ID, ProviderSessionID: previousSession.ProviderSessionID,
			Source: previousSession.Source, Provider: previousSession.Provider,
			HostID: previousSession.HostID, ParentSessionID: previousSession.ParentSessionID,
			RootSessionID: previousSession.RootSessionID, Path: previousSession.Path,
			Project: previousSession.Project, CWD: previousSession.CWD,
			Title: previousSession.Title, InitialPrompt: previousSession.InitialPrompt,
			Slug: previousSession.Slug, AgentType: previousSession.AgentType,
			Description: previousSession.Description, CLIVersion: previousSession.CLIVersion,
		}
	}

	rendered, err := renderedSpec(preparation.Spec, issue.Verification)
	if err != nil {
		return err
	}
	promptInput := captaindb.CreatePromptRunInput{
		ID:           promptRunID,
		Origin:       "gavel.todos",
		SpecProfile:  string(mode),
		AdmissionKey: "gavel-todo:" + seed,
		RenderedSpec: rendered,
		Runtime: captaindb.PromptRunRuntime{
			Mode: string(mode), Driver: executorName, Requested: preparation.Requested,
		},
		PromptMarkdown:       promptMarkdown,
		VerificationMarkdown: issue.Verification,
	}
	if err := p.attachInputPlan(ctx, issue, mode, &promptInput); err != nil {
		return err
	}
	launch, err := p.coordinator.LaunchPromptRun(ctx, native.PromptRunLaunchInput{
		RootSession: p.todoRootSessionInput(native.CreateIssueInput{
			ID: issue.ID, Title: issue.Title, Body: issue.Body,
		}),
		Session:   sessionInput,
		PromptRun: promptInput,
		Attachment: native.PromptRunLaunchAttachment{
			IssueID:              issue.ID,
			StepKind:             step,
			Ordinal:              ordinal,
			ExpectedIssueVersion: issue.Version,
			Actor:                mutationActor,
		},
	})
	if err != nil {
		return err
	}
	if !launch.DispatchOwned {
		return fmt.Errorf(
			"%w: Captain prompt run %s for issue %s already has an external dispatcher",
			todos.ErrRunDispatchAlreadyClaimed, launch.PromptRun.ID, issue.ID,
		)
	}
	p.markPrepared(issue.ID, launch.PromptRun.ID)
	return p.replaceTODO(ctx, todo, launch.Issue, cwd)
}

// RecordRunStart binds the external provider identity and execution thread to
// the prompt run. Provider-thread lifecycle remains monitor-owned.
func (p *Provider) RecordRunStart(ctx context.Context, todo *types.TODO, metadata todos.RunStartMetadata) error {
	active, err := p.loadActiveRun(ctx, todo)
	if err != nil {
		return err
	}
	p.markPrepared(active.issue.ID, active.run.ID)

	sessionUpdate := captaindb.UpdateSessionStateInput{
		ID: active.session.ID, ExpectedVersion: active.session.StateVersion,
	}
	if sessionID := strings.TrimSpace(metadata.SessionID); sessionID != "" {
		sessionUpdate.ProviderSessionID = &sessionID
	}
	if sessionUpdate.ProviderSessionID != nil {
		root, err := p.captain.UpdateSessionState(ctx, sessionUpdate)
		if err != nil {
			return fmt.Errorf("bind Captain admission session: %w", err)
		}
		active.session = root
	}
	metadata.Provider = runStartProvider(active.run.Runtime, metadata)

	// A prompt run binds its execution thread once. Re-resolving the agent
	// identity on a resumed turn would fork a second session — provider is part
	// of Captain's session identity key — and the new id then loses the update
	// to the execution-session guard.
	var agentSession *captaindb.Session
	if active.run.ExecutionSessionID == nil {
		agentSession, err = p.ensureAgentSession(ctx, active.session, metadata)
		if err != nil {
			return err
		}
	}

	state := captaindb.PromptRunStateRunning
	phase := captaindb.PromptRunPhaseGenerate
	if active.link.StepKind == native.StepVerify {
		phase = captaindb.PromptRunPhaseVerify
	}
	if !terminalPromptRun(active.run.State) {
		runtime := mergeRunStartRuntime(active.run.Runtime, metadata)
		update := captaindb.UpdatePromptRunInput{
			ID: active.run.ID, ExpectedVersion: active.run.Version,
			State: &state, Phase: &phase, Runtime: &runtime,
		}
		if agentSession != nil {
			update.ExecutionSessionID = &agentSession.ID
		}
		// Only the report that trails setup carries a spec. Reports that cannot
		// see it (the session-id report, the verify executor) leave it nil rather
		// than overwriting the transformed spec with the request it started as.
		if metadata.Spec != nil {
			rendered, err := renderedSpec(*metadata.Spec, active.issue.Verification)
			if err != nil {
				return err
			}
			update.RenderedSpec = &rendered
		}
		if _, err := p.captain.UpdatePromptRun(ctx, update); err != nil {
			return fmt.Errorf("record Captain prompt-run start: %w", err)
		}
	}
	return p.reloadTODO(ctx, todo, todo.CWD)
}

// PlanMarkdown returns Captain-owned immutable plan content. Runtime callers
// never read a portable-file plan pointer or silently execute an unapproved
// revision.
func (p *Provider) PlanMarkdown(ctx context.Context, todo *types.TODO, mode types.RunMode) (string, error) {
	issueID, err := p.todoID(todo)
	if err != nil {
		return "", err
	}
	issue, err := p.repository.GetIssue(ctx, issueID)
	if err != nil {
		return "", err
	}
	if issue.SelectedPlanID == nil {
		return "", nil
	}
	plan, err := p.captain.GetPlan(ctx, *issue.SelectedPlanID)
	if err != nil {
		return "", err
	}
	switch mode {
	case "", types.ModeRun:
		if plan.ApprovalState != captaindb.PlanApprovalApproved || plan.ApprovedRevision == nil {
			return "", fmt.Errorf("selected Captain plan %s is %s; approve an immutable revision before implementation", plan.ID, plan.ApprovalState)
		}
		return plan.ApprovedRevision.PlanMarkdown, nil
	case types.ModePlan:
		if plan.LatestRevision == nil {
			return "", nil
		}
		return plan.LatestRevision.PlanMarkdown, nil
	case types.ModeVerify:
		return "", nil
	default:
		return "", fmt.Errorf("unsupported TODO run mode %q", mode)
	}
}

// finishAttempt projects one executor result into Captain. It returns false
// only for compatibility calls that have no active Captain run.
func (p *Provider) finishAttempt(ctx context.Context, todo *types.TODO, result *todos.ExecutionResult) (bool, error) {
	active, err := p.loadActiveRun(ctx, todo)
	if err != nil {
		if errors.Is(err, native.ErrNotFound) || errors.Is(err, captaindb.ErrPromptRunNotFound) {
			return false, nil
		}
		return false, err
	}

	if active.link.StepKind == native.StepPlan && successfulPlanAttempt(result) && result.Plan != nil {
		switch result.Plan.Status {
		case types.PlanNew, types.PlanUpdated:
			if err := p.persistPlanResult(ctx, todo, active, result); err != nil {
				_ = p.failPromptRun(ctx, active, err.Error())
				p.clearPrepared(active.issue.ID, active.run.ID)
				return true, err
			}
			active, err = p.loadActiveRun(ctx, todo)
			if err != nil {
				return true, err
			}
		case types.PlanUnchanged:
			if active.issue.SelectedPlanID == nil {
				return true, fmt.Errorf("plan run reported unchanged but issue %s has no selected Captain plan", active.issue.ID)
			}
			// Codex plan mode keeps the full Markdown in its native plan file and
			// may return only a short summary in plan.content. An "unchanged"
			// result therefore still needs to reconcile Captain when an earlier
			// attempt persisted that summary instead of the referenced file.
			path := strings.TrimSpace(result.Plan.Path)
			if path != "" {
				fileMarkdown, _, exists, readErr := todos.ReadPlanFile(path)
				if readErr != nil {
					return true, readErr
				}
				if exists && strings.TrimSpace(fileMarkdown) != "" {
					plan, getErr := p.captain.GetPlan(ctx, *active.issue.SelectedPlanID)
					if getErr != nil {
						return true, getErr
					}
					if plan.LatestRevision == nil || normalizePlanResultMarkdown(plan.LatestRevision.PlanMarkdown) != normalizePlanResultMarkdown(fileMarkdown) {
						if err := p.persistPlanResult(ctx, todo, active, result); err != nil {
							_ = p.failPromptRun(ctx, active, err.Error())
							p.clearPrepared(active.issue.ID, active.run.ID)
							return true, err
						}
						active, err = p.loadActiveRun(ctx, todo)
						if err != nil {
							return true, err
						}
					}
				}
			}
		}
	}

	state, phase, _, _, reason := terminalState(result, active.link.StepKind)
	resultText := ""
	resultJSON := executionResultJSON(result)
	errorText := ""
	if result != nil {
		resultText = strings.TrimSpace(result.Summary)
		errorText = strings.TrimSpace(result.ErrorMessage)
	}
	if state == captaindb.PromptRunStateFailed && errorText == "" {
		errorText = reason
	}
	if !terminalPromptRun(active.run.State) || active.run.State == captaindb.PromptRunStateWaiting {
		updatedRun, updateErr := p.captain.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{
			ID: active.run.ID, ExpectedVersion: active.run.Version,
			State: &state, Phase: &phase, ResultText: &resultText,
			ResultJSON: &resultJSON, Error: &errorText,
		})
		if updateErr != nil {
			return true, fmt.Errorf("finish Captain prompt run: %w", updateErr)
		}
		active.run = updatedRun
	}
	if state != captaindb.PromptRunStateWaiting {
		p.clearPrepared(active.issue.ID, active.run.ID)
	}
	return true, p.reloadTODO(ctx, todo, todo.CWD)
}

func successfulPlanAttempt(result *todos.ExecutionResult) bool {
	if result == nil || !result.Success || result.EndStatus == types.EndFailed || result.EndStatus == types.EndAsk {
		return false
	}
	return strings.TrimSpace(result.ErrorMessage) == ""
}

func (p *Provider) failPreparedRun(ctx context.Context, todo *types.TODO, reason string) error {
	issueID, err := p.todoID(todo)
	if err != nil {
		return err
	}
	runID, ok := p.preparedRun(issueID)
	if !ok {
		return nil
	}
	active, err := p.loadActiveRun(ctx, todo)
	if err != nil {
		return err
	}
	if active.run.ID != runID {
		return nil
	}
	if err := p.failPromptRun(ctx, active, reason); err != nil {
		return err
	}
	p.clearPrepared(issueID, runID)
	return p.reloadTODO(ctx, todo, todo.CWD)
}

func (p *Provider) failPromptRun(ctx context.Context, active *activeRun, reason string) error {
	state := captaindb.PromptRunStateFailed
	phase := active.run.Phase
	if phase == captaindb.PromptRunPhaseQueued || phase == captaindb.PromptRunPhasePreRun {
		phase = captaindb.PromptRunPhaseGenerate
	}
	if !terminalPromptRun(active.run.State) {
		if _, err := p.captain.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{
			ID: active.run.ID, ExpectedVersion: active.run.Version,
			State: &state, Phase: &phase, Error: &reason,
		}); err != nil {
			return err
		}
	}
	return nil
}

func agentSessionSource(executor string) string {
	executor = strings.ToLower(strings.TrimSpace(executor))
	switch {
	case strings.Contains(executor, "codex"):
		return "codex"
	case strings.Contains(executor, "claude"):
		return "claude"
	default:
		return ""
	}
}

// runStartProvider recovers the LLM provider a turn does not report. A resumed
// turn knows only its session and mode, and a blank provider resolves to a
// different Captain session because provider is part of the session identity
// key — so fall back to what the run already resolved.
func runStartProvider(runtime captaindb.PromptRunRuntime, metadata todos.RunStartMetadata) string {
	return firstNonBlank(
		metadata.Provider,
		runtime.Resolved.Provider,
		runtime.Requested.Provider,
		todos.RuntimeProviderForBackend(metadata.Backend),
		todos.RuntimeProviderForBackend(runtime.Resolved.Backend),
	)
}

// mergeRunStartRuntime layers a turn's reported runtime over what the prompt run
// already recorded, so a resumed turn that cannot name its driver or model keeps
// the original run's provenance instead of erasing it.
func mergeRunStartRuntime(current captaindb.PromptRunRuntime, metadata todos.RunStartMetadata) captaindb.PromptRunRuntime {
	merged := current
	merged.Mode = firstNonBlank(metadata.Mode, current.Mode)
	merged.Driver = firstNonBlank(metadata.Driver, current.Driver)
	merged.Resolved = captaindb.PromptRunRuntimeSelection{
		Provider: firstNonBlank(metadata.Provider, current.Resolved.Provider),
		Backend:  firstNonBlank(metadata.Backend, current.Resolved.Backend),
		Model:    firstNonBlank(metadata.ResolvedModel, current.Resolved.Model),
		Effort:   firstNonBlank(metadata.Effort, current.Resolved.Effort),
	}
	return merged
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// ensureAgentSession resolves or creates the monitor-owned session identity.
// Captain's transcript ingestor uses the same (source, host, provider ID)
// identity, so later ingest enriches this row rather than creating another
// agent record. The admission root remains a separate bookkeeping row.
func (p *Provider) ensureAgentSession(ctx context.Context, admission *captaindb.Session, metadata todos.RunStartMetadata) (*captaindb.Session, error) {
	if admission == nil || strings.TrimSpace(admission.ProviderSessionID) == "" {
		return nil, nil
	}
	source := agentSessionSource(admission.Provider)
	if source == "" {
		return nil, nil
	}
	session, err := p.captain.CreateOrGetSession(ctx, captaindb.CreateSessionInput{
		ProviderSessionID: admission.ProviderSessionID,
		Source:            source,
		Provider:          strings.TrimSpace(metadata.Provider),
		HostID:            captaindb.LocalHostID(),
		ParentSessionID:   &admission.ID,
		Project:           admission.Project,
		CWD:               admission.CWD,
		Title:             admission.Title,
		InitialPrompt:     admission.InitialPrompt,
		AgentType:         admission.Provider,
		Description:       "Agent session for " + admission.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve monitored Captain session: %w", err)
	}
	return session, nil
}

func (p *Provider) persistPlanResult(ctx context.Context, todo *types.TODO, active *activeRun, result *todos.ExecutionResult) error {
	markdown, path, err := planResultContent(todo, result)
	if err != nil {
		return err
	}
	currentIssue, err := p.repository.GetIssue(ctx, active.issue.ID)
	if err != nil {
		return err
	}
	planInput := captaindb.CreatePlanInput{
		SourceSessionID:   active.run.SessionID,
		SourcePromptRunID: &active.run.ID,
		Title:             currentIssue.Title,
		Path:              path,
		Variant:           "primary",
		SpecProfile:       "gavel.todo.plan",
	}
	ordinal := 0
	if currentIssue.SelectedPlanID != nil {
		existing, err := p.captain.GetPlan(ctx, *currentIssue.SelectedPlanID)
		if err != nil {
			return err
		}
		planInput = captaindb.CreatePlanInput{
			ID: existing.ID, SourceSessionID: existing.SourceSessionID,
			Title: existing.Title, Path: path, SpecProfile: existing.SpecProfile,
		}
		links, err := p.repository.ListPlans(ctx, currentIssue.ID)
		if err != nil {
			return err
		}
		for _, link := range links {
			if link.PlanID == existing.ID {
				ordinal = link.Ordinal
				break
			}
		}
	}
	persisted, err := p.coordinator.PersistAndSelectPlan(ctx, planInput, captaindb.AppendPlanRevisionInput{
		PlanMarkdown: markdown,
		CreatedBy:    strings.TrimSpace(result.ExecutorName),
	}, native.PlanSelectionAttachment{
		IssueID: currentIssue.ID, Ordinal: ordinal,
		ExpectedIssueVersion: currentIssue.Version, Actor: mutationActor,
	})
	if err != nil {
		return fmt.Errorf("persist Captain plan revision: %w", err)
	}
	return p.replaceTODO(ctx, todo, persisted.Issue, todo.CWD)
}

func planResultContent(todo *types.TODO, result *todos.ExecutionResult) (content, path string, err error) {
	if result == nil || result.Plan == nil {
		return "", "", fmt.Errorf("plan result is missing")
	}
	path = strings.TrimSpace(result.Plan.Path)
	// The native plan file is authoritative when the agent supplies one.
	// Codex commonly puts the detailed plan there while plan.content is only a
	// short completion summary, so preferring inline content truncates the
	// immutable Captain revision and the dashboard's Plan tab.
	if path != "" {
		read, _, exists, readErr := todos.ReadPlanFile(path)
		if readErr != nil {
			return "", path, readErr
		}
		if exists && strings.TrimSpace(read) != "" {
			return strings.TrimSpace(read), path, nil
		}
	}
	if content = strings.TrimSpace(result.Plan.Content); content != "" {
		return content, path, nil
	}
	resolvedPath, resolved := todos.ResolveSessionPlan(todo)
	if strings.TrimSpace(resolved) != "" {
		if path == "" {
			path = resolvedPath
		}
		return strings.TrimSpace(resolved), path, nil
	}
	return "", path, fmt.Errorf("plan run produced no durable markdown content")
}

func normalizePlanResultMarkdown(markdown string) string {
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")
	markdown = strings.ReplaceAll(markdown, "\r", "\n")
	return strings.TrimSpace(markdown)
}

func (p *Provider) attachInputPlan(ctx context.Context, issue *native.Issue, mode types.RunMode, input *captaindb.CreatePromptRunInput) error {
	if issue.SelectedPlanID == nil || mode == types.ModeVerify {
		return nil
	}
	plan, err := p.captain.GetPlan(ctx, *issue.SelectedPlanID)
	if err != nil {
		return err
	}
	var revision *captaindb.PlanRevision
	switch mode {
	case types.ModeRun:
		if plan.ApprovalState != captaindb.PlanApprovalApproved || plan.ApprovedRevision == nil {
			return fmt.Errorf("selected Captain plan %s is %s; approve an immutable revision before implementation", plan.ID, plan.ApprovalState)
		}
		revision = plan.ApprovedRevision
	case types.ModePlan:
		revision = plan.LatestRevision
	}
	if revision != nil {
		planID := plan.ID
		revisionID := revision.ID
		input.InputPlanID = &planID
		input.InputPlanRevisionID = &revisionID
	}
	return nil
}

// decorateExecution projects Captain-owned details into the temporary legacy
// TODO view. The database records remain authoritative; these fields exist only
// so current CLI/API/UI consumers can render the cutover without a second
// provider or a filesystem plan pointer.
func (p *Provider) decorateExecution(ctx context.Context, issue *native.Issue, todo *types.TODO) error {
	if p == nil || p.captain == nil || p.repository == nil || issue == nil || todo == nil {
		return nil
	}
	if issue.ActivePromptRunID != nil {
		run, err := p.captain.GetPromptRun(ctx, *issue.ActivePromptRunID)
		if err != nil {
			return err
		}
		session, err := p.captain.GetSession(ctx, run.SessionID)
		if err != nil {
			return err
		}
		if todo.LLM == nil {
			todo.LLM = &types.LLM{}
		}
		// session.Provider is the executor/driver identity (e.g. "cmux-claude"),
		// never an LLM model — assigning it to LLM.Model poisons the next run's
		// --model resolution. Only the session id is a legitimate LLM field here.
		if strings.TrimSpace(session.ProviderSessionID) != "" {
			todo.LLM.SessionId = session.ProviderSessionID
		}
		lastRun := run.QueuedAt
		if run.StartedAt != nil {
			lastRun = *run.StartedAt
		}
		if run.FinishedAt != nil {
			lastRun = *run.FinishedAt
		}
		if issue.UpdatedAt.After(lastRun) {
			lastRun = issue.UpdatedAt
		}
		todo.LastRun = &lastRun
		todo.LastRunSummary = strings.TrimSpace(run.ResultText)
		if questions, ok := run.ResultJSON["questions"]; ok {
			todo.Questions = decodeQuestions(questions)
		}
		links, err := p.repository.ListPromptRuns(ctx, issue.ID)
		if err != nil {
			return err
		}
		todo.Attempts = len(links)
		for _, link := range links {
			if link.PromptRunID != run.ID {
				continue
			}
			switch link.StepKind {
			case native.StepPlan:
				todo.RunMode = types.ModePlan
			case native.StepVerify:
				todo.RunMode = types.ModeVerify
			default:
				todo.RunMode = types.ModeRun
			}
			break
		}
	}

	if issue.SelectedPlanID == nil {
		return nil
	}
	plan, err := p.captain.GetPlan(ctx, *issue.SelectedPlanID)
	if err != nil {
		return err
	}
	if plan.LatestRevision != nil {
		if plan.LatestRevision.Revision <= 1 {
			todo.PlanStatus = types.PlanNew
		} else {
			todo.PlanStatus = types.PlanUpdated
		}
	}
	// Plan paths are source metadata only in native storage.
	todo.PlanPath = ""
	todo.Status = todoStatusWithPlan(issue.Status, issue.ExecutionState, plan.ApprovalState)
	return nil
}

func todoStatusWithPlan(
	status native.IssueStatus,
	execution native.ExecutionState,
	approval captaindb.PlanApprovalState,
) types.Status {
	projected := todoStatus(status, execution)
	if (status != native.StatusOpen && status != native.StatusDraft) || execution != native.ExecutionIdle {
		return projected
	}
	switch approval {
	case captaindb.PlanApprovalPending, captaindb.PlanApprovalRevisionRequested:
		return types.StatusReview
	case captaindb.PlanApprovalRejected, captaindb.PlanApprovalApproved:
		return types.StatusPending
	default:
		return projected
	}
}

// ActivePromptRun returns the Captain prompt run backing the todo's current
// attempt, or nil when it has none. Callers deciding how to continue a run — the
// runtime it resolved, whether an agent is still live — read it rather than
// re-deriving those facts from the projected status.
func (p *Provider) ActivePromptRun(ctx context.Context, todo *types.TODO) (*captaindb.PromptRun, error) {
	active, err := p.loadActiveRun(ctx, todo)
	if err != nil {
		if errors.Is(err, native.ErrNotFound) || errors.Is(err, captaindb.ErrPromptRunNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return active.run, nil
}

func (p *Provider) loadActiveRun(ctx context.Context, todo *types.TODO) (*activeRun, error) {
	issueID, err := p.todoID(todo)
	if err != nil {
		return nil, err
	}
	issue, err := p.repository.GetIssue(ctx, issueID)
	if err != nil {
		return nil, err
	}
	if issue.ActivePromptRunID == nil {
		return nil, fmt.Errorf("%w: issue %s has no active Captain prompt run", native.ErrNotFound, issue.ID)
	}
	run, err := p.captain.GetPromptRun(ctx, *issue.ActivePromptRunID)
	if err != nil {
		return nil, err
	}
	session, err := p.captain.GetSession(ctx, run.SessionID)
	if err != nil {
		return nil, err
	}
	links, err := p.repository.ListPromptRuns(ctx, issue.ID)
	if err != nil {
		return nil, err
	}
	for i := range links {
		if links[i].PromptRunID == run.ID {
			link := links[i]
			return &activeRun{issue: issue, link: &link, run: run, session: session}, nil
		}
	}
	return nil, fmt.Errorf("%w: active prompt run %s is not linked to issue %s", native.ErrLinkConflict, run.ID, issue.ID)
}

func (p *Provider) nextPromptOrdinal(ctx context.Context, issueID uuid.UUID, step native.StepKind) (int, error) {
	links, err := p.repository.ListPromptRuns(ctx, issueID)
	if err != nil {
		return 0, err
	}
	next := 0
	for _, link := range links {
		if link.StepKind == step && link.Ordinal >= next {
			next = link.Ordinal + 1
		}
	}
	return next, nil
}

func (p *Provider) todoID(todo *types.TODO) (uuid.UUID, error) {
	if todo == nil {
		return uuid.Nil, fmt.Errorf("native TODO is nil")
	}
	id, err := uuid.Parse(strings.TrimSpace(todo.ID))
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid native TODO ID %q: %w", todo.ID, err)
	}
	if todo.WorkspaceID != "" {
		workspaceID, err := uuid.Parse(todo.WorkspaceID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid native TODO workspace ID %q: %w", todo.WorkspaceID, err)
		}
		if workspaceID != p.workspace.ID {
			return uuid.Nil, fmt.Errorf("%w: TODO %s belongs to workspace %s, provider owns %s", native.ErrCrossWorkspace, id, workspaceID, p.workspace.ID)
		}
	}
	return id, nil
}

func stepForMode(mode types.RunMode) (native.StepKind, error) {
	switch mode {
	case types.ModePlan:
		return native.StepPlan, nil
	case types.ModeRun:
		return native.StepRun, nil
	case types.ModeVerify:
		return native.StepVerify, nil
	default:
		return "", fmt.Errorf("unsupported TODO run mode %q", mode)
	}
}

func terminalState(result *todos.ExecutionResult, step native.StepKind) (
	captaindb.PromptRunState,
	captaindb.PromptRunPhase,
	captaindb.SessionLifecycleStatus,
	captaindb.SessionActivityState,
	string,
) {
	phase := captaindb.PromptRunPhaseFinished
	if result == nil {
		return captaindb.PromptRunStateFailed, captaindb.PromptRunPhaseGenerate,
			captaindb.SessionLifecycleFailed, captaindb.SessionActivityIdle, "agent run returned no result"
	}
	if result.Cancelled {
		reason := strings.TrimSpace(result.Summary)
		if reason == "" {
			reason = strings.TrimSpace(result.ErrorMessage)
		}
		if reason == "" {
			reason = todos.ErrExecutionCancelled.Error()
		}
		return captaindb.PromptRunStateCancelled, phase,
			captaindb.SessionLifecycleCancelled, captaindb.SessionActivityIdle, reason
	}
	if result.EndStatus == types.EndAsk {
		return captaindb.PromptRunStateWaiting, captaindb.PromptRunPhaseGenerate,
			captaindb.SessionLifecycleRunning, captaindb.SessionActivityAsk, strings.TrimSpace(result.Summary)
	}
	failedVerification := result.DoD != nil && result.DoD.Ran && !result.DoD.Passed
	if failedVerification || !result.Success || result.EndStatus == types.EndFailed || strings.TrimSpace(result.ErrorMessage) != "" {
		if failedVerification || step == native.StepVerify {
			phase = captaindb.PromptRunPhaseVerify
		} else {
			phase = captaindb.PromptRunPhaseGenerate
		}
		reason := strings.TrimSpace(result.ErrorMessage)
		if reason == "" {
			reason = strings.TrimSpace(result.Summary)
		}
		if reason == "" {
			reason = "agent run failed"
		}
		return captaindb.PromptRunStateFailed, phase,
			captaindb.SessionLifecycleFailed, captaindb.SessionActivityIdle, reason
	}
	return captaindb.PromptRunStateSucceeded, phase,
		captaindb.SessionLifecycleSucceeded, captaindb.SessionActivityIdle, strings.TrimSpace(result.Summary)
}

func executionResultJSON(result *todos.ExecutionResult) map[string]any {
	if result == nil {
		return map[string]any{}
	}
	out := map[string]any{
		"success": result.Success, "skipped": result.Skipped, "cancelled": result.Cancelled,
		"executor": result.ExecutorName, "tokens": result.TokensUsed,
		"costUsd": result.CostUSD, "turns": result.NumTurns,
		"summary": result.Summary, "endStatus": result.EndStatus,
		"commit": result.CommitSHA, "questions": result.Questions,
	}
	if result.Plan != nil {
		out["plan"] = map[string]any{"status": result.Plan.Status, "path": result.Plan.Path}
	}
	if result.DoD != nil {
		out["definitionOfDone"] = result.DoD
	}
	return out
}

func terminalPromptRun(state captaindb.PromptRunState) bool {
	return state == captaindb.PromptRunStateSucceeded ||
		state == captaindb.PromptRunStateFailed ||
		state == captaindb.PromptRunStateCancelled
}

func (p *Provider) markPrepared(issueID, runID uuid.UUID) {
	p.preparedMu.Lock()
	defer p.preparedMu.Unlock()
	if p.prepared == nil {
		p.prepared = map[uuid.UUID]uuid.UUID{}
	}
	p.prepared[issueID] = runID
}

func (p *Provider) preparedRun(issueID uuid.UUID) (uuid.UUID, bool) {
	p.preparedMu.RLock()
	defer p.preparedMu.RUnlock()
	runID, ok := p.prepared[issueID]
	return runID, ok
}

func (p *Provider) clearPrepared(issueID, runID uuid.UUID) {
	p.preparedMu.Lock()
	defer p.preparedMu.Unlock()
	if p.prepared[issueID] == runID {
		delete(p.prepared, issueID)
	}
}

func decodeQuestions(value any) []types.AgentQuestion {
	data, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var questions []types.AgentQuestion
	if json.Unmarshal(data, &questions) != nil {
		return nil
	}
	return questions
}
