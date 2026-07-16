package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/session"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
)

type sessionDetailProvider interface {
	Captain() *captaindb.DB
	Repository() *native.Repository
}

type providerThreadStore interface {
	GetSessionByIdentity(context.Context, string, string, string, string) (*captaindb.Session, error)
	ListThreadSessionOverviews(context.Context, uuid.UUID) ([]captaindb.SessionOverview, error)
	ListThreadTranscriptMessages(context.Context, uuid.UUID) ([]captaindb.TranscriptMessage, error)
	ListThreadTurns(context.Context, uuid.UUID) ([]captaindb.SessionTurn, error)
	ListThreadAgents(context.Context, uuid.UUID) ([]captaindb.SessionAgent, error)
	ListThreadCosts(context.Context, uuid.UUID) ([]captaindb.SessionCost, error)
}

type sessionDiagnostic struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Details  any    `json:"details,omitempty"`
}

type todoAttemptDetail struct {
	PromptRunID       uuid.UUID                 `json:"promptRunId"`
	Ordinal           int                       `json:"ordinal"`
	Step              string                    `json:"step"`
	Mode              string                    `json:"mode,omitempty"`
	Driver            string                    `json:"driver,omitempty"`
	Requested         todos.RunRuntimeSelection `json:"requested"`
	Resolved          todos.RunRuntimeSelection `json:"resolved"`
	State             captaindb.PromptRunState  `json:"state"`
	Phase             captaindb.PromptRunPhase  `json:"phase"`
	QueuedAt          time.Time                 `json:"queuedAt"`
	StartedAt         *time.Time                `json:"startedAt,omitempty"`
	FinishedAt        *time.Time                `json:"finishedAt,omitempty"`
	DurationMS        *int64                    `json:"durationMs,omitempty"`
	Error             string                    `json:"error,omitempty"`
	ResultText        string                    `json:"resultText,omitempty"`
	ResultJSON        map[string]any            `json:"resultJson,omitempty"`
	AdmissionSession  uuid.UUID                 `json:"admissionSessionId"`
	ExecutionSession  *uuid.UUID                `json:"executionSessionId,omitempty"`
	ProviderSessionID string                    `json:"providerSessionId,omitempty"`
	CanStop           bool                      `json:"canStop,omitempty"`
	Stopping          bool                      `json:"stopping,omitempty"`
	CreatedAt         time.Time                 `json:"createdAt"`
}

type todoProviderThread struct {
	ID             uuid.UUID                   `json:"id"`
	ProviderID     string                      `json:"providerSessionId,omitempty"`
	Status         string                      `json:"status"`
	Root           captaindb.SessionOverview   `json:"root"`
	Sessions       []captaindb.SessionOverview `json:"sessions"`
	Turns          []captaindb.SessionTurn     `json:"turns"`
	Agents         []captaindb.SessionAgent    `json:"agents"`
	Costs          []captaindb.SessionCost     `json:"costs"`
	Messages       []session.Message           `json:"messages"`
	StartedAt      *time.Time                  `json:"startedAt,omitempty"`
	LastActivityAt *time.Time                  `json:"lastActivityAt,omitempty"`
	DurationMS     int64                       `json:"durationMs"`
	InputTokens    int64                       `json:"inputTokens"`
	OutputTokens   int64                       `json:"outputTokens"`
	TotalTokens    int64                       `json:"totalTokens"`
	CostUSD        float64                     `json:"costUsd"`
}

type todoSessionDetailResponse struct {
	Attempts                 []todoAttemptDetail `json:"attempts"`
	SelectedPromptRunID      *uuid.UUID          `json:"selectedPromptRunId,omitempty"`
	SelectedExecutionSession *uuid.UUID          `json:"selectedExecutionSessionId,omitempty"`
	Thread                   *todoProviderThread `json:"thread,omitempty"`
	Diagnostics              []sessionDiagnostic `json:"diagnostics"`
}

func (s *Server) handleTodoSessionDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if ref == "" {
		writeTodoError(w, http.StatusBadRequest, errors.New("ref is required"))
		return
	}
	dir := s.resolveTodoDir(strings.TrimSpace(r.URL.Query().Get("dir")))
	provider, err := openTodoProvider(r.Context(), dir)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	detailProvider, ok := provider.(sessionDetailProvider)
	if !ok || detailProvider.Captain() == nil || detailProvider.Repository() == nil {
		writeTodoError(w, http.StatusNotImplemented, errors.New("session details require native TODO storage"))
		return
	}
	todo, err := provider.Get(r.Context(), ref)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	issueID, err := uuid.Parse(todo.ID)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, fmt.Errorf("native TODO has invalid ID %q: %w", todo.ID, err))
		return
	}
	response, conflict, err := buildTodoSessionDetail(r.Context(), detailProvider, issueID, strings.TrimSpace(r.URL.Query().Get("sessionId")))
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	issue, err := detailProvider.Repository().GetIssue(r.Context(), issueID)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	if issue.ActivePromptRunID != nil {
		status := s.todoRuns.status(issueID)
		for index := range response.Attempts {
			if response.Attempts[index].PromptRunID == *issue.ActivePromptRunID {
				response.Attempts[index].CanStop = status.CanStop
				response.Attempts[index].Stopping = status.Stopping
			}
		}
	}
	if conflict {
		w.WriteHeader(http.StatusConflict)
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(fmt.Errorf("encode TODO session detail: %w", err))
	}
}

func buildTodoSessionDetail(ctx context.Context, provider sessionDetailProvider, issueID uuid.UUID, requestedSessionID string) (todoSessionDetailResponse, bool, error) {
	links, err := provider.Repository().ListPromptRuns(ctx, issueID)
	if err != nil {
		return todoSessionDetailResponse{}, false, err
	}
	sort.SliceStable(links, func(i, j int) bool { return links[i].CreatedAt.Before(links[j].CreatedAt) })
	response := todoSessionDetailResponse{Attempts: []todoAttemptDetail{}, Diagnostics: []sessionDiagnostic{}}
	for index, link := range links {
		run, runErr := provider.Captain().GetPromptRun(ctx, link.PromptRunID)
		if runErr != nil {
			return response, false, runErr
		}
		admission, sessionErr := provider.Captain().GetSession(ctx, run.SessionID)
		if sessionErr != nil {
			return response, false, sessionErr
		}
		response.Attempts = append(response.Attempts, attemptDetail(index+1, link, run, admission))
	}
	for left, right := 0, len(response.Attempts)-1; left < right; left, right = left+1, right-1 {
		response.Attempts[left], response.Attempts[right] = response.Attempts[right], response.Attempts[left]
	}

	selected := selectAttempt(response.Attempts, requestedSessionID)
	var rootID *uuid.UUID
	providerID := requestedSessionID
	if selected != nil {
		response.SelectedPromptRunID = &selected.PromptRunID
		if selected.ExecutionSession != nil {
			rootID = selected.ExecutionSession
			response.SelectedExecutionSession = selected.ExecutionSession
		}
		if providerID == "" {
			providerID = selected.ProviderSessionID
		}
	}
	conflict := false
	if rootID == nil && providerID != "" {
		candidates, listErr := provider.Captain().ListSessionOverviewsByIdentity(ctx, providerID)
		if listErr != nil && !errors.Is(listErr, captaindb.ErrSessionNotFound) {
			return response, false, listErr
		}
		candidate, diagnostics, ambiguous := selectLegacyThreadCandidate(providerID, candidates)
		response.Diagnostics = append(response.Diagnostics, diagnostics...)
		conflict = ambiguous
		if candidate != nil {
			rootID = &candidate.ID
			response.SelectedExecutionSession = rootID
		}
	}
	if rootID != nil {
		thread, loadErr := loadProviderThread(ctx, provider.Captain(), *rootID)
		if loadErr != nil {
			return response, false, loadErr
		}
		response.Thread = thread
		if selected != nil && selected.FinishedAt != nil && thread.LastActivityAt != nil && thread.LastActivityAt.After(*selected.FinishedAt) {
			response.Diagnostics = append(response.Diagnostics, sessionDiagnostic{
				Severity: "warning", Code: "thread_continued_after_attempt",
				Message: fmt.Sprintf("Provider thread continued for %s after attempt %d ended.", thread.LastActivityAt.Sub(*selected.FinishedAt).Round(time.Second), selected.Ordinal),
				Details: map[string]any{"attemptFinishedAt": selected.FinishedAt, "threadLastActivityAt": thread.LastActivityAt},
			})
		}
	}
	return response, conflict, nil
}

func attemptDetail(ordinal int, link native.PromptRunLink, run *captaindb.PromptRun, admission *captaindb.Session) todoAttemptDetail {
	runtimeSpec := mapValue(run.RenderedSpec["runtime"])
	mode, _ := runtimeSpec["mode"].(string)
	driver, _ := runtimeSpec["driver"].(string)
	if mode == "" {
		mode = run.SpecProfile
	}
	if driver == "" {
		driver = admission.Provider
	}
	detail := todoAttemptDetail{
		PromptRunID: run.ID, Ordinal: ordinal, Step: string(link.StepKind), Mode: mode, Driver: driver,
		Requested: selectionValue(runtimeSpec["requested"]), Resolved: selectionValue(runtimeSpec["resolved"]),
		State: run.State, Phase: run.Phase, QueuedAt: run.QueuedAt, StartedAt: run.StartedAt, FinishedAt: run.FinishedAt,
		Error: run.Error, ResultText: run.ResultText, ResultJSON: run.ResultJSON, AdmissionSession: run.SessionID,
		ExecutionSession: run.ExecutionSessionID, ProviderSessionID: admission.ProviderSessionID, CreatedAt: run.CreatedAt,
	}
	if run.StartedAt != nil {
		end := time.Now().UTC()
		if run.FinishedAt != nil {
			end = *run.FinishedAt
		}
		duration := end.Sub(*run.StartedAt).Milliseconds()
		detail.DurationMS = &duration
	}
	return detail
}

func selectionValue(value any) todos.RunRuntimeSelection {
	values := mapValue(value)
	return todos.RunRuntimeSelection{
		Provider: stringValue(values["provider"]), Backend: stringValue(values["backend"]),
		Model: stringValue(values["model"]), Effort: stringValue(values["effort"]),
	}
}

func selectAttempt(attempts []todoAttemptDetail, sessionID string) *todoAttemptDetail {
	for index := range attempts {
		attempt := &attempts[index]
		if sessionID == "" || attempt.ProviderSessionID == sessionID || (attempt.ExecutionSession != nil && attempt.ExecutionSession.String() == sessionID) {
			return attempt
		}
	}
	return nil
}

func selectLegacyThreadCandidate(providerID string, candidates []captaindb.SessionOverview) (*captaindb.SessionOverview, []sessionDiagnostic, bool) {
	exact := make([]captaindb.SessionOverview, 0, len(candidates))
	transcripts := make([]captaindb.SessionOverview, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.ProviderSessionID == nil || *candidate.ProviderSessionID != providerID {
			continue
		}
		exact = append(exact, candidate)
		if (candidate.Source == "claude" || candidate.Source == "codex") &&
			(candidate.MessageCount > 0 || candidate.TurnCount > 0 || candidate.Path != nil || candidate.HistoryFile != nil) {
			transcripts = append(transcripts, candidate)
		}
	}
	if len(transcripts) == 1 {
		selected := transcripts[0]
		return &selected, []sessionDiagnostic{{
			Severity: "warning", Code: "legacy_session_identity_resolved",
			Message: "A legacy provider session had multiple Captain rows; the only transcript-bearing row was selected without modifying history.",
			Details: exact,
		}}, false
	}
	if len(transcripts) > 1 {
		return nil, []sessionDiagnostic{{
			Severity: "error", Code: "ambiguous_transcript_sessions",
			Message: fmt.Sprintf("Provider session %q has %d transcript-bearing Captain sessions.", providerID, len(transcripts)),
			Details: exact,
		}}, true
	}
	return nil, nil, false
}

func loadProviderThread(ctx context.Context, db providerThreadStore, rootID uuid.UUID) (*todoProviderThread, error) {
	sessions, err := db.ListThreadSessionOverviews(ctx, rootID)
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, fmt.Errorf("%w: thread %s", captaindb.ErrSessionNotFound, rootID)
	}
	turns, err := db.ListThreadTurns(ctx, rootID)
	if err != nil {
		return nil, err
	}
	agents, err := db.ListThreadAgents(ctx, rootID)
	if err != nil {
		return nil, err
	}
	costs, err := db.ListThreadCosts(ctx, rootID)
	if err != nil {
		return nil, err
	}
	messages, err := loadThreadMessages(ctx, db, rootID)
	if err != nil {
		return nil, err
	}
	thread := &todoProviderThread{ID: rootID, Root: sessions[0], Sessions: sessions, Turns: turns, Agents: agents, Costs: costs, Messages: messages, Status: projectThreadStatus(sessions)}
	if sessions[0].ProviderSessionID != nil {
		thread.ProviderID = *sessions[0].ProviderSessionID
	}
	for _, session := range sessions {
		if session.StartedAt != nil && (thread.StartedAt == nil || session.StartedAt.Before(*thread.StartedAt)) {
			thread.StartedAt = session.StartedAt
		}
		activity := session.LastActivityAt
		if activity == nil {
			activity = session.EndedAt
		}
		if activity != nil && (thread.LastActivityAt == nil || activity.After(*thread.LastActivityAt)) {
			thread.LastActivityAt = activity
		}
	}
	for _, agent := range agents {
		thread.InputTokens += agent.InputTokens
		thread.OutputTokens += agent.OutputTokens
		thread.TotalTokens += agent.TotalTokens
		thread.CostUSD += agent.CostUSD
	}
	if thread.StartedAt != nil && thread.LastActivityAt != nil && thread.LastActivityAt.After(*thread.StartedAt) {
		thread.DurationMS = thread.LastActivityAt.Sub(*thread.StartedAt).Milliseconds()
	}
	return thread, nil
}

func loadThreadMessages(ctx context.Context, db providerThreadStore, rootID uuid.UUID) ([]session.Message, error) {
	rows, err := db.ListThreadTranscriptMessages(ctx, rootID)
	if err != nil {
		return nil, err
	}
	owners := make(map[uuid.UUID]*captaindb.Session)
	messages := make([]session.Message, 0, len(rows))
	for _, row := range rows {
		owner := owners[row.SessionID]
		if owner == nil {
			owner, err = db.GetSessionByIdentity(ctx, row.SessionID.String(), "", "", "")
			if err != nil {
				return nil, fmt.Errorf("load transcript owner %s: %w", row.SessionID, err)
			}
			owners[row.SessionID] = owner
		}
		message, _, messageErr := captainTranscriptMessage(row, owner)
		if messageErr != nil {
			return nil, fmt.Errorf("project transcript message %s: %w", row.ID, messageErr)
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func projectThreadStatus(sessions []captaindb.SessionOverview) string {
	if len(sessions) == 0 {
		return "waiting"
	}
	for _, session := range sessions {
		if session.ProcessActive || session.ActivityState == string(captaindb.SessionActivityWorking) || session.ActivityState == string(captaindb.SessionActivityThinking) {
			return "working"
		}
	}
	root := sessions[0]
	switch root.LifecycleStatus {
	case string(captaindb.SessionLifecycleFailed):
		return "failed"
	case string(captaindb.SessionLifecycleCancelled):
		return "cancelled"
	case string(captaindb.SessionLifecycleInterrupted):
		return "interrupted"
	case string(captaindb.SessionLifecycleSucceeded):
		return "completed"
	}
	for _, session := range sessions {
		if session.MessageCount > 0 || session.TurnCount > 0 {
			return "idle"
		}
	}
	return "waiting"
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func mapValue(value any) map[string]any {
	values, _ := value.(map[string]any)
	if values == nil {
		return map[string]any{}
	}
	return values
}
