package native

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Priority is the durable operator-assigned importance of an issue.
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// IssueStatus is durable workflow state. It is intentionally independent of
// the transient Captain-derived ExecutionState.
type IssueStatus string

const (
	StatusDraft     IssueStatus = "draft"
	StatusOpen      IssueStatus = "open"
	StatusVerified  IssueStatus = "verified"
	StatusClosed    IssueStatus = "closed"
	StatusCancelled IssueStatus = "cancelled"
)

// ExecutionState is the latest projection of linked Captain execution.
type ExecutionState string

const (
	ExecutionIdle               ExecutionState = "idle"
	ExecutionPlanning           ExecutionState = "planning"
	ExecutionRunning            ExecutionState = "running"
	ExecutionWaiting            ExecutionState = "waiting"
	ExecutionStalled            ExecutionState = "stalled"
	ExecutionFailed             ExecutionState = "failed"
	ExecutionVerifying          ExecutionState = "verifying"
	ExecutionVerificationFailed ExecutionState = "verification_failed"
)

// RelationshipKind is the stored direction of an issue relationship. Blocks
// is presented as the reverse of RelationshipDependsOn and is never stored.
type RelationshipKind string

const (
	RelationshipDependsOn RelationshipKind = "depends_on"
	RelationshipRelatedTo RelationshipKind = "related_to"
	// RelationshipBlocks is the read-only reverse presentation of depends_on.
	// It is returned by relationship queries and is never persisted.
	RelationshipBlocks RelationshipKind = "blocks"
)

// StepKind identifies the purpose of a linked Captain prompt run.
type StepKind string

const (
	StepPlan   StepKind = "plan"
	StepRun    StepKind = "run"
	StepVerify StepKind = "verify"
)

// Workspace is the durable repository identity containing native issues.
type Workspace struct {
	ID          uuid.UUID `json:"id"`
	RepoKey     string    `json:"repoKey"`
	RootPath    string    `json:"rootPath,omitempty"`
	DisplayName string    `json:"displayName,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// WorkspacePath is one normalized filesystem location known for a workspace.
// A workspace with paths has exactly one primary path; older locations remain
// available so moved workspaces continue to resolve.
type WorkspacePath struct {
	WorkspaceID uuid.UUID `json:"workspaceId"`
	Path        string    `json:"path"`
	IsPrimary   bool      `json:"isPrimary"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Issue is durable Gavel-owned state. Captain-owned execution details are
// referenced by ID and are not duplicated here.
type Issue struct {
	ID                uuid.UUID      `json:"id"`
	WorkspaceID       uuid.UUID      `json:"workspaceId"`
	Title             string         `json:"title"`
	Body              string         `json:"body"`
	Verification      string         `json:"verification"`
	Labels            []string       `json:"labels"`
	Priority          Priority       `json:"priority"`
	Status            IssueStatus    `json:"status"`
	ExecutionState    ExecutionState `json:"executionState"`
	ActivePromptRunID *uuid.UUID     `json:"activePromptRunId,omitempty"`
	SelectedPlanID    *uuid.UUID     `json:"selectedPlanId,omitempty"`
	Version           int64          `json:"version"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
}

// IssueStatusCount is one group of CountIssuesByStatus: the durable status, the
// projected execution state, active step and selected plan's approval state —
// the inputs the derived TODO status is computed from — plus how many issues
// share them. ApprovalState is empty when the issue has no selected plan.
type IssueStatusCount struct {
	Status         IssueStatus    `json:"status"`
	ExecutionState ExecutionState `json:"executionState"`
	StepKind       StepKind       `json:"stepKind"`
	ApprovalState  string         `json:"approvalState"`
	Count          int            `json:"count"`
}

// Alias is a normalized workspace-scoped reference to an issue.
type Alias struct {
	WorkspaceID uuid.UUID `json:"workspaceId"`
	Alias       string    `json:"alias"`
	IssueID     uuid.UUID `json:"issueId"`
	Kind        string    `json:"kind,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Event is the sole append-only audit/history record for an issue. Comments
// are represented by Kind == "comment".
type Event struct {
	ID        uuid.UUID       `json:"id"`
	IssueID   uuid.UUID       `json:"issueId"`
	Sequence  int64           `json:"sequence"`
	Kind      string          `json:"kind"`
	Actor     string          `json:"actor,omitempty"`
	Body      string          `json:"body,omitempty"`
	Payload   json.RawMessage `json:"payload"`
	Source    string          `json:"source"`
	SourceID  string          `json:"sourceId,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
}

// Relationship is the canonical stored representation of an issue edge.
type Relationship struct {
	WorkspaceID   uuid.UUID        `json:"workspaceId"`
	IssueID       uuid.UUID        `json:"issueId"`
	TargetIssueID uuid.UUID        `json:"targetIssueId"`
	Relation      RelationshipKind `json:"relation"`
	CreatedAt     time.Time        `json:"createdAt"`
}

// PromptRunLink associates one Captain prompt run with an issue.
type PromptRunLink struct {
	IssueID     uuid.UUID `json:"issueId"`
	PromptRunID uuid.UUID `json:"promptRunId"`
	StepKind    StepKind  `json:"stepKind"`
	Ordinal     int       `json:"ordinal"`
	CreatedAt   time.Time `json:"createdAt"`
}

// PlanLink associates a durable Captain plan with an issue.
type PlanLink struct {
	IssueID   uuid.UUID `json:"issueId"`
	PlanID    uuid.UUID `json:"planId"`
	Ordinal   int       `json:"ordinal"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateWorkspaceInput struct {
	ID          uuid.UUID
	RepoKey     string
	RootPath    string
	DisplayName string
}

type UpdateWorkspaceInput struct {
	RepoKey     *string
	RootPath    *string
	DisplayName *string
}

type AliasInput struct {
	Alias string
	Kind  string
}

type CreateIssueInput struct {
	ID           uuid.UUID
	WorkspaceID  uuid.UUID
	Aliases      []AliasInput
	Title        string
	Body         string
	Verification string
	Labels       []string
	Priority     Priority
	Status       IssueStatus
	Actor        string
}

// IssuePatch uses pointers so callers can distinguish no change from setting
// a field to its zero value.
type IssuePatch struct {
	Title        *string
	Body         *string
	Verification *string
	Labels       *[]string
	Priority     *Priority
	Status       *IssueStatus
	Actor        string
}

type EventInput struct {
	Kind     string
	Actor    string
	Body     string
	Payload  any
	Source   string
	SourceID string
}
