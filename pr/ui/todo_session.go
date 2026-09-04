package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	cmuxprov "github.com/flanksource/captain/pkg/ai/provider/cmux"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
)

// captainSessionStore is the native session read seam used by the TODO UI.
// Runtime providers expose Captain's handle over the same pool Gavel owns;
// tests and non-native compatibility providers can omit it and retain the
// legacy on-disk Claude stream below.
type captainSessionStore interface {
	GetSessionByIdentity(context.Context, string, string, string, string) (*captaindb.Session, error)
	GetSessionOverviewByIdentity(context.Context, string) (*captaindb.SessionOverview, error)
	ListTranscriptMessages(context.Context, captaindb.TranscriptPage) ([]captaindb.TranscriptMessage, error)
}

type captainThreadSessionStore interface {
	ListSessionOverviewsByIdentity(context.Context, string) ([]captaindb.SessionOverview, error)
	ListThreadSessionOverviews(context.Context, uuid.UUID) ([]captaindb.SessionOverview, error)
	ListThreadTranscriptMessages(context.Context, uuid.UUID) ([]captaindb.TranscriptMessage, error)
}

type captainSessionProvider interface {
	Captain() *captaindb.DB
}

type captainSessionResolution struct {
	run         *captaindb.Session
	transcript  *captaindb.Session
	known       bool
	diagnostics []sessionDiagnostic
}

func todoCaptainSessionStore(ctx context.Context, dir string) (captainSessionStore, bool) {
	provider, err := openTodoProvider(ctx, dir)
	if err != nil {
		return nil, false
	}
	native, ok := provider.(captainSessionProvider)
	if !ok || native.Captain() == nil {
		return nil, false
	}
	return native.Captain(), true
}

// resolveCaptainSession finds the admission root and the monitored agent row.
// The root is only a provider-identity bridge; live state, timing, messages and
// accounting all come from the Claude/Codex session.
func resolveCaptainSession(ctx context.Context, store captainSessionStore, sessionID string) (captainSessionResolution, error) {
	var resolved captainSessionResolution
	run, err := store.GetSessionByIdentity(ctx, sessionID, "gavel", "", "")
	if err == nil {
		resolved.run = run
		resolved.known = true
	} else if !errors.Is(err, captaindb.ErrSessionNotFound) {
		return resolved, err
	}

	sources := []string{"codex", "claude"}
	if run != nil && strings.Contains(strings.ToLower(run.Provider), "claude") {
		sources[0], sources[1] = sources[1], sources[0]
	}
	if threadStore, ok := store.(captainThreadSessionStore); ok {
		candidates, candidateErr := threadStore.ListSessionOverviewsByIdentity(ctx, sessionID)
		if candidateErr == nil {
			selected, diagnostics, conflict := selectLegacyThreadCandidate(sessionID, candidates)
			resolved.diagnostics = diagnostics
			if conflict {
				return resolved, fmt.Errorf("%w: provider session ID %q has multiple transcript-bearing sessions", captaindb.ErrSessionConflict, sessionID)
			}
			if selected != nil {
				transcript, getErr := store.GetSessionByIdentity(ctx, selected.ID.String(), "", "", "")
				if getErr != nil {
					return resolved, getErr
				}
				resolved.transcript = transcript
				resolved.known = true
				return resolved, nil
			}
		} else if !errors.Is(candidateErr, captaindb.ErrSessionNotFound) {
			return resolved, candidateErr
		}
	}
	for _, source := range sources {
		transcript, transcriptErr := store.GetSessionByIdentity(ctx, sessionID, source, "", "")
		if transcriptErr == nil {
			resolved.transcript = transcript
			resolved.known = true
			return resolved, nil
		}
		if !errors.Is(transcriptErr, captaindb.ErrSessionNotFound) {
			return resolved, transcriptErr
		}
	}
	return resolved, nil
}

// handleTodoSessionStats returns the rolled-up stats for a TODO's agent session
// — agent/model/effort, elapsed time, token usage and derived cost. Live runs are
// served from the in-memory cache the cmux tailer feeds; sessions no tailer is
// watching are read (and cached) from the on-disk log. A session that never
// produced a log is reported as found=false, not an error, so the dashboard
// simply hides the timer.
func (s *Server) handleTodoSessionStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := strings.TrimSpace(r.URL.Query().Get("sessionId"))
	if sessionID == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("sessionId is required"))
		return
	}
	dir := s.resolveTodoDir(strings.TrimSpace(r.URL.Query().Get("dir")))
	if store, ok := todoCaptainSessionStore(r.Context(), dir); ok {
		resolved, err := resolveCaptainSession(r.Context(), store, sessionID)
		if err != nil {
			writeTodoError(w, http.StatusInternalServerError, err)
			return
		}
		if resolved.known {
			resp, err := captainSessionStats(r.Context(), store, sessionID, resolved)
			if err != nil {
				writeTodoError(w, http.StatusInternalServerError, err)
				return
			}
			if err := attachPendingApproval(r.Context(), store, resolved, &resp); err != nil {
				writeTodoError(w, http.StatusInternalServerError, err)
				return
			}
			json.NewEncoder(w).Encode(resp) //nolint:errcheck
			return
		}
	}
	path, err := cmuxprov.SessionLogPath(dir, sessionID)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	stats, err := cmuxprov.GlobalSessionStats().Get(sessionID, path)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	// A session with no Captain row has no durable approval table behind it, and
	// therefore nothing that could be awaiting a decision.
	json.NewEncoder(w).Encode(todoSessionStatsResponse{SessionStats: stats}) //nolint:errcheck
}

// attachPendingApproval overrides the derived state when a tool request is
// outstanding, so the dashboard renders the "needs approval" affordance and its
// Approve/Deny controls. The request is read from Captain's durable table, so a
// dashboard that reconnected — or one in a different process from the run —
// sees it just the same.
func attachPendingApproval(
	ctx context.Context,
	store captainSessionStore,
	resolved captainSessionResolution,
	resp *todoSessionStatsResponse,
) error {
	approvals, ok := store.(approvalStore)
	if !ok || resolved.run == nil {
		return nil
	}
	pending, err := pendingApprovals(ctx, approvals, resolved.run.ID, nil)
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		return nil
	}
	resp.State = "approval"
	resp.Approval = &pending[0]
	resp.Approvals = pending
	return nil
}

// todoSessionStatsResponse is the session-stats payload plus any pending
// tool-permission requests awaiting a decision. Approval is the oldest of them,
// kept as a single field for the control that answers one at a time; Approvals
// is the whole queue.
type todoSessionStatsResponse struct {
	cmuxprov.SessionStats
	Approval  *todoApproval  `json:"approval,omitempty"`
	Approvals []todoApproval `json:"approvals,omitempty"`
}

func captainSessionStats(ctx context.Context, store captainSessionStore, sessionID string, resolved captainSessionResolution) (todoSessionStatsResponse, error) {
	stats := cmuxprov.SessionStats{SessionID: sessionID, Found: resolved.transcript != nil}
	if resolved.transcript == nil {
		return todoSessionStatsResponse{SessionStats: stats}, nil
	}
	overview, err := store.GetSessionOverviewByIdentity(ctx, resolved.transcript.ID.String())
	if err != nil {
		return todoSessionStatsResponse{}, err
	}
	overviews := []captaindb.SessionOverview{*overview}
	if threadStore, ok := store.(captainThreadSessionStore); ok {
		if rows, threadErr := threadStore.ListThreadSessionOverviews(ctx, resolved.transcript.ID); threadErr != nil {
			return todoSessionStatsResponse{}, threadErr
		} else if len(rows) > 0 {
			overviews = rows
			overview = &overviews[0]
		}
	}
	stats.Found = true
	stats.Agent = resolved.transcript.AgentType
	if stats.Agent == "" && resolved.run != nil {
		stats.Agent = resolved.run.Provider
	}
	stats.StartedAt = resolved.transcript.CreatedAt
	if overview.StartedAt != nil {
		stats.StartedAt = *overview.StartedAt
	}
	for _, row := range overviews {
		if row.LastActivityAt != nil && row.LastActivityAt.After(stats.UpdatedAt) {
			stats.UpdatedAt = *row.LastActivityAt
		}
	}
	if stats.UpdatedAt.IsZero() {
		stats.UpdatedAt = stats.StartedAt
	}
	for _, row := range overviews {
		stats.InProgress = stats.InProgress || row.LifecycleStatus == string(captaindb.SessionLifecycleRunning) || row.ProcessActive
	}
	if stats.InProgress {
		stats.UpdatedAt = time.Now().UTC()
	}
	if !stats.StartedAt.IsZero() && stats.UpdatedAt.After(stats.StartedAt) {
		stats.DurationMs = stats.UpdatedAt.Sub(stats.StartedAt).Milliseconds()
	} else if overview.DurationSeconds != nil {
		stats.DurationMs = int64(*overview.DurationSeconds * 1000)
	}
	switch {
	case overview.LifecycleStatus == string(captaindb.SessionLifecycleFailed) || overview.HealthState == string(captaindb.SessionHealthZombie):
		stats.State, stats.Error = "error", resolved.transcript.StateReason
	case overview.HealthState == string(captaindb.SessionHealthStalled):
		stats.State = "stalled"
	case overview.ActivityState != "" && overview.ActivityState != string(captaindb.SessionActivityIdle):
		stats.State = overview.ActivityState
	case stats.InProgress:
		stats.State = "working"
	default:
		stats.State = projectThreadStatus(overviews)
	}
	if overview.Model != nil {
		stats.Model = *overview.Model
	}
	if overview.Effort != nil {
		stats.Effort = *overview.Effort
	}
	for _, row := range overviews {
		stats.InputTokens += int(row.InputTokens)
		stats.OutputTokens += int(row.OutputTokens)
		stats.CacheReadTokens += int(row.CacheReadTokens)
		stats.CacheCreationTokens += int(row.CacheWriteTokens)
		stats.TotalTokens += int(row.TotalTokens)
		stats.Turns += int(row.TurnCount)
		stats.CostUSD += row.CostUSD
	}
	stats.ContextTokens = intValue(overview.ContextTokens)
	stats.ContextWindow = intValue(overview.ContextWindowTokens)
	return todoSessionStatsResponse{SessionStats: stats}, nil
}

func intValue(value *int64) int {
	if value == nil {
		return 0
	}
	return int(*value)
}

// todoSessionApprovePayload is what the dashboard's approval controls POST.
//
// ApprovalID names the durable request rather than the session, because a run
// can have more than one outstanding and a session id cannot tell them apart.
type todoSessionApprovePayload struct {
	ApprovalID string `json:"approvalId"`
	// Action is approve, deny or respond. `respond` runs the call with Input
	// substituted; `deny` refuses it and feeds Message back as the reason.
	Action  string         `json:"action"`
	Message string         `json:"message,omitempty"`
	Input   map[string]any `json:"input,omitempty"`
}

// handleTodoSessionApprove answers one pending tool-permission request — the
// dashboard's Approve/Deny/Respond controls POST here, which unblocks the
// broker awaiting the decision.
//
// The decision is written to captain's durable approval table, not to an
// in-process registry: the run may have been started by a different process, and
// an approval outstanding across a dashboard restart must still be answerable.
func (s *Server) handleTodoSessionApprove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoSessionApprovePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	requestID, err := uuid.Parse(strings.TrimSpace(payload.ApprovalID))
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("approvalId must be the pending approval's id: %w", err))
		return
	}
	action, err := parseApprovalAction(payload.Action)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	dir := s.resolveTodoDir(strings.TrimSpace(r.URL.Query().Get("dir")))
	store, err := todoApprovalStore(r.Context(), dir)
	if err != nil {
		writeTodoError(w, http.StatusServiceUnavailable, err)
		return
	}
	sessionID, err := approvalSessionID(r.Context(), store, strings.TrimSpace(r.URL.Query().Get("sessionId")), requestID)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	resolvedRequest, err := resolveApproval(r.Context(), store, sessionID, requestID, action, payload.Message, payload.Input)
	if err != nil {
		writeTodoError(w, http.StatusConflict, err)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
		"resolved":   true,
		"approvalId": resolvedRequest.ID.String(),
		"state":      string(resolvedRequest.State),
	})
}

// approvalSessionID is the session the approval belongs to. A client that knows
// the session sends it; one that only has the approval id from a notification
// does not, and the row itself carries the answer.
func approvalSessionID(ctx context.Context, store approvalStore, requested string, requestID uuid.UUID) (uuid.UUID, error) {
	if requested != "" {
		if sessionID, err := uuid.Parse(requested); err == nil {
			return sessionID, nil
		}
	}
	request, err := store.GetTurnRequest(ctx, requestID)
	if err != nil {
		return uuid.Nil, err
	}
	return request.SessionID, nil
}

// handleTodoSessionApprovals lists a run's unanswered tool requests, so a
// dashboard that reconnected can show what is blocking it.
func (s *Server) handleTodoSessionApprovals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID, err := uuid.Parse(strings.TrimSpace(r.URL.Query().Get("sessionId")))
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("sessionId must be Captain's admitted session id: %w", err))
		return
	}
	var promptRunID *uuid.UUID
	if raw := strings.TrimSpace(r.URL.Query().Get("promptRunId")); raw != "" {
		parsed, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid promptRunId: %w", parseErr))
			return
		}
		promptRunID = &parsed
	}
	store, err := todoApprovalStore(r.Context(), s.resolveTodoDir(strings.TrimSpace(r.URL.Query().Get("dir"))))
	if err != nil {
		writeTodoError(w, http.StatusServiceUnavailable, err)
		return
	}
	pending, err := pendingApprovals(r.Context(), store, sessionID, promptRunID)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	if pending == nil {
		pending = []todoApproval{}
	}
	json.NewEncoder(w).Encode(map[string]any{"approvals": pending}) //nolint:errcheck
}

// handleTodoSessionFocus switches cmux to the workspace running a TODO's agent
// session, so the dashboard's "focus" control brings the live terminal to the
// front. The workspace is identified by the run's working directory and agent
// (claude/codex), matching how the cmux executor names it. A closed terminal or
// a stopped cmux yields a 4xx with the reason rather than a silent no-op.
func (s *Server) handleTodoSessionFocus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dir := s.resolveTodoDir(strings.TrimSpace(r.URL.Query().Get("dir")))
	agent := strings.TrimSpace(r.URL.Query().Get("agent"))
	if agent == "" {
		agent = "claude"
	}
	if err := cmuxprov.FocusSession(r.Context(), cmuxprov.NewClient(""), dir, agent); err != nil {
		writeTodoError(w, http.StatusBadGateway, err)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"focused": true}) //nolint:errcheck
}

// todoSessionCmuxResponse locates the cmux terminal a session maps to. It is
// best-effort: a stopped cmux or a closed terminal reports found=false with the
// reason rather than an error, so the dashboard's cmux control still renders (and
// can offer Resume) instead of the whole request failing.
type todoSessionCmuxResponse struct {
	Found     bool   `json:"found"`
	Workspace string `json:"workspace,omitempty"`
	Surface   string `json:"surface,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// handleTodoSessionCmux resolves the cmux workspace and surface a TODO's agent
// session lives on, so the dashboard can label the "Focus / Resume in cmux"
// control with the concrete terminal it targets (e.g. "workspace:2 surface:1").
// Resolution reuses the same workspace naming the cmux executor uses, so no cmux
// reference has to be tracked on the issue.
func (s *Server) handleTodoSessionCmux(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dir := s.resolveTodoDir(strings.TrimSpace(r.URL.Query().Get("dir")))
	agent := strings.TrimSpace(r.URL.Query().Get("agent"))
	if agent == "" {
		agent = "claude"
	}
	workspace, surface, err := resolveCmuxSurface(r.Context(), dir, agent)
	resp := todoSessionCmuxResponse{Found: err == nil}
	if err != nil {
		resp.Reason = err.Error()
	} else {
		resp.Workspace = workspace
		resp.Surface = surface
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// resolveCmuxSurface finds the cmux workspace running the agent session for
// workDir — named the same way the cmux executor names it — and the surface its
// terminal currently shows. The surface read is best-effort: a workspace whose
// tree can't be read still yields the workspace reference.
func resolveCmuxSurface(ctx context.Context, workDir, agent string) (workspace, surface string, err error) {
	client := cmuxprov.NewClient("")
	if err := client.Available(ctx); err != nil {
		return "", "", err
	}
	name := cmuxprov.AgentWorkspaceName(workDir, agent)
	ref, found, err := client.FindWorkspace(ctx, name, workDir)
	if err != nil {
		return "", "", err
	}
	if !found {
		return "", "", fmt.Errorf("no cmux workspace %q for %s; the session terminal may have been closed", name, workDir)
	}
	surface, _ = client.ResolveSurface(ctx, ref.String())
	return ref.String(), surface, nil
}
