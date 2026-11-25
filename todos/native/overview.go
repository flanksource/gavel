package native

import (
	"time"

	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

// IssueOverview is the database's joined execution and selected-plan projection.
type IssueOverview struct {
	SessionSummary    *types.SessionSummary `gorm:"serializer:json"`
	OverviewRunID     *uuid.UUID
	ProviderSessionID string
	ResultText        string
	RunAt             *time.Time
	StepKind          StepKind
	Attempts          int
	OverviewPlanID    *uuid.UUID
	ApprovalState     string
	PlanRevision      *int
}
