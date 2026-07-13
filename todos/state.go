package todos

import (
	"time"

	"github.com/flanksource/gavel/todos/types"
)

// StateUpdate represents a partial update to a TODO's frontmatter.
// Only non-nil fields will be updated in the TODO file.
type StateUpdate struct {
	SessionID *string
	Status    *types.Status
	Priority  *types.Priority
	Attempts  *int
	LastRun   *time.Time
	// Envelope-driven fields from an agent run's structured result: the native
	// plan-mode file the agent reported, the plan/run mode bookkeeping an
	// answer-resume needs, and the summary/questions surfaced in the dashboard.
	PlanPath       *string
	PlanStatus     *types.PlanStatus
	RunMode        *types.RunMode
	LastRunSummary *string
	Questions      *[]types.AgentQuestion
}

// applyStateUpdate updates only the in-memory portable representation. Runtime
// persistence belongs to the native PostgreSQL provider; this helper also lets
// pure outcome tests exercise state transitions without a filesystem fallback.
func applyStateUpdate(frontmatter *types.TODOFrontmatter, updates StateUpdate) {
	if frontmatter == nil {
		return
	}
	if updates.Status != nil {
		frontmatter.Status = *updates.Status
	}
	if updates.Priority != nil {
		frontmatter.Priority = *updates.Priority
	}
	if updates.Attempts != nil {
		frontmatter.Attempts = *updates.Attempts
	}
	if updates.LastRun != nil {
		frontmatter.LastRun = updates.LastRun
	}
	if updates.SessionID != nil {
		if frontmatter.LLM == nil {
			frontmatter.LLM = &types.LLM{}
		}
		frontmatter.LLM.SessionId = *updates.SessionID
	}
	if updates.PlanPath != nil {
		frontmatter.PlanPath = *updates.PlanPath
	}
	if updates.PlanStatus != nil {
		frontmatter.PlanStatus = *updates.PlanStatus
	}
	if updates.RunMode != nil {
		frontmatter.RunMode = *updates.RunMode
	}
	if updates.LastRunSummary != nil {
		frontmatter.LastRunSummary = *updates.LastRunSummary
	}
	if updates.Questions != nil {
		frontmatter.Questions = *updates.Questions
	}
}
