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

// StepKind is the lifecycle step a linked Captain prompt run was dispatched as.
//
// It is an OPEN string, not an enum: the steps that exist are declared by the
// project's lifecycle definition (todos/lifecycle/todos.yaml, overridable from
// .gavel.yaml), so storage accepts any lower-case name and the host — which is
// the only thing that has loaded the definition — is what rejects a name the
// lifecycle does not declare.
//
// The constants below are the built-in lifecycle's steps, named here because
// storage genuinely reasons about two of them: `verify` decides the phase a run
// is recorded under, and the others are what historical rows contain.
type StepKind string

const (
	StepPlan   StepKind = "plan"
	StepRun    StepKind = "run"
	StepVerify StepKind = "verify"
	StepTriage StepKind = "triage"
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

// IssuePhaseRun is the latest Captain prompt run one issue has for one phase:
// what a backlog needs to show a per-phase status, progress and elapsed time
// without opening the todo.
//
// DurationSeconds comes from captain_prompt_run_overview, which measures an
// unfinished run against clock_timestamp() — so a live phase's value ticks, and
// is therefore only a starting point for a UI that should tick locally rather
// than re-poll.
type IssuePhaseRun struct {
	IssueID    uuid.UUID  `json:"issueId"`
	Phase      StepKind   `json:"phase"`
	State      string     `json:"state"`
	RunPhase   string     `json:"runPhase"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	// DurationSeconds is null until the run starts.
	DurationSeconds *float64 `json:"durationSeconds,omitempty"`
	// Iterations/Succeeded/Failed are the progress of a plan, run or triage
	// pass. Verification counts its own fixture results instead, from
	// VerificationResult.
	Iterations int `json:"iterations"`
	Succeeded  int `json:"succeeded"`
	Failed     int `json:"failed"`
	// VerificationResult is captain's stored api.VerifyReport for the run's
	// newest verified iteration, as JSON — the ONE record of a verification.
	// Empty means the run has produced no verdict.
	VerificationResult string  `json:"verificationResult,omitempty"`
	CostUSD            float64 `json:"costUsd"`
	// Active marks the run the issue currently points at, so a caller can tell a
	// phase that is running now from one that merely ran last.
	Active bool `json:"active"`
}

// IssueRunRecord is one linked Captain prompt run of an issue as the lifecycle's
// run history reads it. Phase is the step the row is listed under: the link's
// step kind, or `verify` for the second listing of a run step whose run verified
// its own work (see ListIssueRunHistory).
type IssueRunRecord struct {
	IssueID     uuid.UUID  `json:"issueId"`
	PromptRunID uuid.UUID  `json:"promptRunId"`
	Phase       StepKind   `json:"phase"`
	Ordinal     int        `json:"ordinal"`
	State       string     `json:"state"`
	QueuedAt    time.Time  `json:"queuedAt"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	FinishedAt  *time.Time `json:"finishedAt,omitempty"`
	// Outcome is the status the lifecycle recorded for this run in its outcome
	// event; empty while the run is live, or when the step kept the status.
	Outcome string `json:"outcome,omitempty"`
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
	// The dispatching process, when one currently holds the run. See
	// ownership.go: a link without an owner is a run nothing is driving.
	OwnerHostID      *string    `gorm:"column:owner_host_id" json:"ownerHostId,omitempty"`
	OwnerPID         *int64     `gorm:"column:owner_pid" json:"ownerPid,omitempty"`
	OwnerStartedAt   *time.Time `gorm:"column:owner_started_at" json:"ownerStartedAt,omitempty"`
	OwnerToken       *uuid.UUID `gorm:"column:owner_token" json:"ownerToken,omitempty"`
	OwnerHeartbeatAt *time.Time `gorm:"column:owner_heartbeat_at" json:"ownerHeartbeatAt,omitempty"`
}

// Owner returns the claim recorded on the link, or nil when the run is unowned.
func (l PromptRunLink) Owner() *RunOwner {
	if l.OwnerHostID == nil || l.OwnerPID == nil || l.OwnerStartedAt == nil || l.OwnerToken == nil {
		return nil
	}
	return &RunOwner{
		HostID: *l.OwnerHostID, PID: *l.OwnerPID, StartedAt: *l.OwnerStartedAt,
		Token: *l.OwnerToken, HeartbeatAt: l.OwnerHeartbeatAt,
	}
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
