package ui

import (
	"context"
	"time"

	cmuxprov "github.com/flanksource/captain/pkg/ai/provider/cmux"
	captaindb "github.com/flanksource/captain/pkg/database"
)

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
	awaitingAnswer := false
	if resolved.run != nil {
		runs, err := store.ListPromptRuns(ctx, captaindb.PromptRunFilter{SessionID: &resolved.run.ID})
		if err != nil {
			return todoSessionStatsResponse{}, err
		}
		awaitingAnswer = len(runs) > 0 && runs[0].State == captaindb.PromptRunStateWaiting && runs[0].ResultJSON["endStatus"] == "ask"
	}
	if resolved.transcript == nil {
		if awaitingAnswer {
			stats.State = "ask"
		}
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
	if awaitingAnswer {
		stats.State = "ask"
		stats.InProgress = false
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

// projectThreadStatus is the settled state the stats header reports for a
// session thread with no live activity signal of its own.
func projectThreadStatus(sessions []captaindb.SessionOverview) string {
	if len(sessions) == 0 {
		return "waiting"
	}
	for _, session := range sessions {
		if session.ProcessActive || session.ActivityState == string(captaindb.SessionActivityWorking) || session.ActivityState == string(captaindb.SessionActivityThinking) {
			return "working"
		}
	}
	switch sessions[0].LifecycleStatus {
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

func intValue(value *int64) int {
	if value == nil {
		return 0
	}
	return int(*value)
}
