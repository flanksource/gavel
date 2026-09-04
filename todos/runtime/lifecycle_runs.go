package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

type activeRun struct {
	issue   *native.Issue
	link    *native.PromptRunLink
	run     *captaindb.PromptRun
	session *captaindb.Session
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

func (p *Provider) RecordRunProgress(ctx context.Context, todo *types.TODO, report api.VerifyReport) error {
	active, err := p.loadActiveRun(ctx, todo)
	if err != nil {
		return err
	}
	if terminalPromptRun(active.run.State) {
		return fmt.Errorf("record verification progress: Captain prompt run %s is already %s", active.run.ID, active.run.State)
	}
	resultJSON := progressResultJSON(active.run.ResultJSON, report)
	phase := captaindb.PromptRunPhaseVerify
	if _, err := p.captain.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{
		ID: active.run.ID, ExpectedVersion: active.run.Version,
		Phase: &phase, ResultJSON: &resultJSON,
	}); err != nil {
		return fmt.Errorf("record Captain verification progress: %w", err)
	}
	return nil
}

// RecordRunNotices writes the run's lifecycle notices into the transcript of the
// session the agent actually ran in.
//
// sessionID is the provider's own id, which names several rows here: the gavel
// admission root, and the claude/codex transcript the monitor ingested from the
// on-disk log. The notices belong on the transcript — that is the row the
// dashboard streams messages from — so a run whose log has not been ingested yet
// has nowhere to put them, and says so rather than writing them somewhere they
// would never be read.
func (p *Provider) RecordRunNotices(ctx context.Context, sessionID string, notices []api.Notice) error {
	if len(notices) == 0 {
		return nil
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return fmt.Errorf("record run notices: session ID is required")
	}
	for _, source := range []string{"claude", "codex"} {
		transcript, err := p.captain.GetSessionByIdentity(ctx, sessionID, source, "", "")
		if err == nil {
			return p.captain.PutSessionNotices(ctx, transcript.ID, notices)
		}
		if !errors.Is(err, captaindb.ErrSessionNotFound) {
			return fmt.Errorf("resolve transcript session %s: %w", sessionID, err)
		}
	}
	return fmt.Errorf("record run notices: no ingested transcript for session %s", sessionID)
}

func cloneResultJSON(source map[string]any) map[string]any {
	clone := make(map[string]any, len(source)+1)
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func progressResultJSON(source map[string]any, report api.VerifyReport) map[string]any {
	resultJSON := cloneResultJSON(source)
	definitionOfDone := map[string]any{}
	if existing, ok := resultJSON["definitionOfDone"].(map[string]any); ok {
		for key, value := range existing {
			definitionOfDone[key] = value
		}
	}
	definitionOfDone["progress"] = report
	resultJSON["definitionOfDone"] = definitionOfDone
	return resultJSON
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
//
// The two composite fallbacks this used to end with are gone: a mode never
// carried a provider, so deriving one from it was always a guess.
func runStartProvider(runtime captaindb.PromptRunRuntime, metadata todos.RunStartMetadata) string {
	return firstNonBlank(
		metadata.Provider,
		runtime.Resolved.Provider,
		runtime.Requested.Provider,
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
		Mode:     firstNonBlank(metadata.RuntimeMode, current.Resolved.Mode),
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

// loadActiveRun resolves the run a caller is acting on. Inside an execution
// that is the run this execution was prepared for — a TODO may have several
// runs in flight, and each has to report its own outcome. Outside one (read
// surfaces, recovery commands) it is the TODO's current active run.
func (p *Provider) loadActiveRun(ctx context.Context, todo *types.TODO) (*activeRun, error) {
	issueID, err := p.todoID(todo)
	if err != nil {
		return nil, err
	}
	issue, err := p.repository.GetIssue(ctx, issueID)
	if err != nil {
		return nil, err
	}
	runID := todos.PromptRunFromContext(ctx)
	if runID == uuid.Nil {
		if issue.ActivePromptRunID == nil {
			return nil, fmt.Errorf("%w: issue %s has no active Captain prompt run", native.ErrNotFound, issue.ID)
		}
		runID = *issue.ActivePromptRunID
	}
	run, err := p.captain.GetPromptRun(ctx, runID)
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
	return nil, fmt.Errorf("%w: prompt run %s is not linked to issue %s", native.ErrLinkConflict, run.ID, issue.ID)
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

func terminalPromptRun(state captaindb.PromptRunState) bool {
	return state == captaindb.PromptRunStateSucceeded ||
		state == captaindb.PromptRunStateFailed ||
		state == captaindb.PromptRunStateCancelled
}

// markPrepared records that this process dispatched a run for an issue. It is a
// set per issue, not one entry: a TODO can have several runs in flight and each
// of them is separately this process's to finish.
func (p *Provider) markPrepared(issueID, runID uuid.UUID) {
	p.preparedMu.Lock()
	defer p.preparedMu.Unlock()
	if p.prepared == nil {
		p.prepared = map[uuid.UUID]map[uuid.UUID]struct{}{}
	}
	if p.prepared[issueID] == nil {
		p.prepared[issueID] = map[uuid.UUID]struct{}{}
	}
	p.prepared[issueID][runID] = struct{}{}
}

func (p *Provider) isPrepared(issueID, runID uuid.UUID) bool {
	p.preparedMu.RLock()
	defer p.preparedMu.RUnlock()
	_, ok := p.prepared[issueID][runID]
	return ok
}

func (p *Provider) hasPrepared(issueID uuid.UUID) bool {
	p.preparedMu.RLock()
	defer p.preparedMu.RUnlock()
	return len(p.prepared[issueID]) > 0
}

// clearPrepared drops this process's binding to a finished run. It is the one
// choke point every terminal path goes through, so it is also where the durable
// ownership claim is released: a run nothing is driving must not look owned.
func (p *Provider) clearPrepared(issueID, runID uuid.UUID) {
	p.preparedMu.Lock()
	delete(p.prepared[issueID], runID)
	if len(p.prepared[issueID]) == 0 {
		delete(p.prepared, issueID)
	}
	p.preparedMu.Unlock()
	p.releaseRun(context.Background(), runID)
}
