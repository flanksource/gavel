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

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/session"
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
	captaindb.PromptRunOverview
	PromptRunID      uuid.UUID `json:"promptRunId"`
	AdmissionSession uuid.UUID `json:"admissionSessionId"`
	Ordinal          int       `json:"ordinal"`
	Step             string    `json:"step"`
	// Verification is the newest definition-of-done report the attempt's
	// verifiers produced, read from captain's per-iteration record. It is
	// always present — null for an attempt that was never verified — so a
	// reader can tell "the server knows there is none" from "not sent".
	Verification *api.VerifyReport `json:"verification"`
	CanStop      bool              `json:"canStop,omitempty"`
	Stopping     bool              `json:"stopping,omitempty"`
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

	// AttemptsOnly marks a response whose provider thread was deliberately not
	// resolved, so a skipped thread is never mistaken for a missing one.
	AttemptsOnly bool `json:"attemptsOnly,omitempty"`
}

// todoSessionDetailQuery is the parsed request. AttemptsOnly serves the
// Verification tab and its badge, which need the attempt list (and the
// verification report captain stores on each attempt's latest iteration,
// read from captain_prompt_run_iterations.verification_result) but never the
// transcript — resolving the full thread means every session, turn, agent,
// cost and message on a poll that runs while the tab is closed.
type todoSessionDetailQuery struct {
	SessionID    string
	AttemptsOnly bool
}

func parseTodoSessionDetailQuery(r *http.Request) (todoSessionDetailQuery, error) {
	query := todoSessionDetailQuery{SessionID: strings.TrimSpace(r.URL.Query().Get("sessionId"))}
	switch attempts := strings.TrimSpace(r.URL.Query().Get("attempts")); attempts {
	case "":
	case "only":
		query.AttemptsOnly = true
	default:
		return query, fmt.Errorf("unknown attempts value %q, want \"only\"", attempts)
	}
	return query, nil
}

func (s *Server) handleTodoSessionDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if ref == "" {
		writeTodoError(w, http.StatusBadRequest, errors.New("ref is required"))
		return
	}
	query, err := parseTodoSessionDetailQuery(r)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
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
	response, conflict, err := buildTodoSessionDetail(r.Context(), detailProvider, issueID, query)
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
		status := todoRuns().Status(issueID)
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

func buildTodoSessionDetail(ctx context.Context, provider sessionDetailProvider, issueID uuid.UUID, query todoSessionDetailQuery) (todoSessionDetailResponse, bool, error) {
	requestedSessionID := query.SessionID
	links, err := provider.Repository().ListPromptRuns(ctx, issueID)
	if err != nil {
		return todoSessionDetailResponse{}, false, err
	}
	sort.SliceStable(links, func(i, j int) bool { return links[i].CreatedAt.Before(links[j].CreatedAt) })
	response := todoSessionDetailResponse{Attempts: []todoAttemptDetail{}, Diagnostics: []sessionDiagnostic{}}
	ids := make([]uuid.UUID, len(links))
	for index := range links {
		ids[index] = links[index].PromptRunID
	}
	overviews, err := provider.Captain().ListPromptRunOverviews(ctx, captaindb.PromptRunOverviewFilter{IDs: ids})
	if err != nil {
		return response, false, err
	}
	byID := make(map[uuid.UUID]captaindb.PromptRunOverview, len(overviews))
	for index := range overviews {
		byID[overviews[index].ID] = overviews[index]
	}
	verifications, err := provider.Captain().LatestPromptRunVerifications(ctx, ids)
	if err != nil {
		return response, false, err
	}
	for index, link := range links {
		overview, ok := byID[link.PromptRunID]
		if !ok {
			return response, false, fmt.Errorf("%w: linked prompt run %s", captaindb.ErrPromptRunNotFound, link.PromptRunID)
		}
		attempt := attemptDetail(index+1, link, overview)
		if verification, ok := verifications[link.PromptRunID]; ok {
			attempt.Verification = verification.Report
		}
		response.Attempts = append(response.Attempts, attempt)
	}
	for left, right := 0, len(response.Attempts)-1; left < right; left, right = left+1, right-1 {
		response.Attempts[left], response.Attempts[right] = response.Attempts[right], response.Attempts[left]
	}

	if query.AttemptsOnly {
		response.AttemptsOnly = true
		return response, false, nil
	}

	selected := selectAttempt(response.Attempts, requestedSessionID)
	var rootID *uuid.UUID
	providerID := requestedSessionID
	if selected != nil {
		response.SelectedPromptRunID = &selected.PromptRunID
		if selected.ExecutionSessionID != nil {
			rootID = selected.ExecutionSessionID
			response.SelectedExecutionSession = selected.ExecutionSessionID
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

func attemptDetail(ordinal int, link native.PromptRunLink, overview captaindb.PromptRunOverview) todoAttemptDetail {
	return todoAttemptDetail{
		PromptRunOverview: overview, PromptRunID: overview.ID, AdmissionSession: overview.SessionID,
		Ordinal: ordinal, Step: string(link.StepKind),
	}
}

func selectAttempt(attempts []todoAttemptDetail, sessionID string) *todoAttemptDetail {
	for index := range attempts {
		attempt := &attempts[index]
		if sessionID == "" || attempt.ProviderSessionID == sessionID || (attempt.ExecutionSessionID != nil && attempt.ExecutionSessionID.String() == sessionID) {
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
