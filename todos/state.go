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
