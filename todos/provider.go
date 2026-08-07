package todos

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

const ProviderDB = "db"

var (
	// ErrRunDispatchAlreadyClaimed means another caller already won the durable
	// admission for this exact Captain prompt run. The caller must not dispatch a
	// second external agent for the idempotently resolved database record.
	ErrRunDispatchAlreadyClaimed = errors.New("native TODO run dispatch already claimed")

	// ErrRunResumeModeMismatch means a caller tried to resume one nonterminal
	// Captain operation using a different Gavel step kind.
	ErrRunResumeModeMismatch = errors.New("native TODO run resume mode mismatch")
)

// Provider is the persistence boundary for TODO storage.
type Provider interface {
	List(ctx context.Context, filters DiscoveryFilters) (types.TODOS, error)
	// CountByStatus reports how many TODOs resolve to each status without
	// materializing them. Callers that only need counts must use this rather
	// than bucketing List, which decodes every body to read one field.
	CountByStatus(ctx context.Context) (map[types.Status]int, error)
	Get(ctx context.Context, ref string) (*types.TODO, error)
	Create(ctx context.Context, req CreateRequest) (*types.TODO, error)
	Delete(ctx context.Context, todo *types.TODO) error
	// Edit updates a TODO's content fields in place.
	Edit(ctx context.Context, todo *types.TODO, edit EditRequest) error
	// Comment appends a free-form comment to a TODO's history.
	Comment(ctx context.Context, todo *types.TODO, body string) error
	UpdateState(ctx context.Context, todo *types.TODO, updates StateUpdate) error
	UpdateLatestFailure(ctx context.Context, todo *types.TODO, result *types.TestResultInfo) error
	SaveAttempt(ctx context.Context, todo *types.TODO, result *ExecutionResult) error
}

// RunPreparation is the durable identity needed before an external agent is
// dispatched. PostgreSQL-backed providers use it to create and attach the
// authoritative Captain prompt run before any provider session starts.
type RunPreparation struct {
	Mode         types.RunMode
	ExecutorName string
	Resume       bool
	Requested    captaindb.PromptRunRuntimeSelection
	// Spec is the exact request the executor will dispatch — model, budget,
	// permissions, setup, workflow and the rendered user prompt. Native storage
	// persists it on Captain's prompt run before external execution, so a later
	// continuation replays what actually ran instead of reconstructing it from
	// the run's resolved model/backend labels.
	Spec api.Spec
}

// RunRuntimeProvider exposes the configuration known before dispatch.
type RunRuntimeProvider interface {
	RunRuntimeSelection() captaindb.PromptRunRuntimeSelection
}

// RuntimeProviderForBackend maps Captain's built-in backend families to their
// provider. Unknown custom backends remain unreported instead of being guessed.
func RuntimeProviderForBackend(backend string) string {
	backend = strings.ToLower(strings.TrimSpace(backend))
	switch {
	case strings.Contains(backend, "claude"):
		return "anthropic"
	case strings.Contains(backend, "codex"), strings.Contains(backend, "openai"):
		return "openai"
	default:
		return ""
	}
}

// RunSpecProvider exposes the exact request an executor will dispatch.
// TODOExecutor asks for it before native admission so Captain stores the real
// runtime and prompt rather than a lossy reconstruction from the issue body.
type RunSpecProvider interface {
	RenderRunSpec(ctx *ExecutorContext, todo *types.TODO) (api.Spec, error)
}

// RunLifecycleProvider is implemented by the native PostgreSQL runtime. Native
// execution state is owned by Captain and projected into issues.
type RunLifecycleProvider interface {
	PrepareRun(ctx context.Context, todo *types.TODO, preparation RunPreparation) error
	RecordRunStart(ctx context.Context, todo *types.TODO, metadata RunStartMetadata) error
}

type RunProgressProvider interface {
	RecordRunProgress(ctx context.Context, todo *types.TODO, snapshot fixtures.ExecutionSnapshot) error
}

// GroupExecutionPolicy lets a persistence boundary reject execution shapes
// that its data model cannot represent. Native prompt runs are single-issue.
type GroupExecutionPolicy interface {
	SupportsGroupedExecution() bool
}

// GlobalReferenceProvider resolves a native UUID or imported alias without a
// caller first listing every workspace. Ambiguous cross-workspace aliases must
// be reported rather than guessed.
type GlobalReferenceProvider interface {
	GetGlobal(ctx context.Context, ref string) (*types.TODO, error)
}

// GlobalSessionReferenceProvider resolves a Captain session UUID to the native
// TODO that owns its linked prompt run. The returned session ID is the
// canonical provider identity when Captain knows one, otherwise the native
// Captain session UUID. Keeping it separate from GlobalReferenceProvider makes
// ordinary issue aliases authoritative when the same UUID-shaped token could
// identify both records.
type GlobalSessionReferenceProvider interface {
	GetGlobalBySession(ctx context.Context, ref string) (*types.TODO, string, error)
}

// TransferProvider preserves native identity/history when moving an issue
// between database workspaces.
type TransferProvider interface {
	MoveTo(ctx context.Context, todo *types.TODO, target Provider) (*types.TODO, error)
}

// PlanReviewProvider persists review decisions against Captain's durable plan
// revisions while keeping Gavel's issue link/version in the same operation.
type PlanReviewProvider interface {
	ApprovePlan(ctx context.Context, todo *types.TODO, actor, comment string) (*types.TODO, error)
	RejectPlan(ctx context.Context, todo *types.TODO, actor, comment string) (*types.TODO, error)
	RequestPlanRevision(ctx context.Context, todo *types.TODO, actor, feedback string) (*types.TODO, error)
}

// PlanRecoveryProvider backfills the durable Captain plan selected by a TODO
// from its most recent plan-step session without changing that run's outcome.
type PlanRecoveryProvider interface {
	RecoverPlan(ctx context.Context, todo *types.TODO) (*types.TODO, error)
}

// PlanContentProvider resolves the durable Captain plan content that should be
// fed to an agent. Plan mode receives the latest selected revision so it can
// revise it; run mode receives only the explicitly approved revision.
type PlanContentProvider interface {
	PlanMarkdown(ctx context.Context, todo *types.TODO, mode types.RunMode) (string, error)
}

// Link is one issue-to-issue relationship seen from the queried TODO. The
// target's identity and status travel with it so callers render a link without
// a second lookup.
type Link struct {
	Relation      types.RelationKind `json:"relation"`
	TargetID      string             `json:"target_id"`
	TargetShortID string             `json:"target_short_id,omitempty"`
	TargetTitle   string             `json:"target_title,omitempty"`
	TargetStatus  types.Status       `json:"target_status,omitempty"`
	CreatedAt     time.Time          `json:"created_at,omitempty"`
}

// RelationshipProvider exposes issue-to-issue links: depends_on for blocking
// work and related_to for duplicates and overlapping scope. Links reports both,
// plus the derived read-only blocks relation for incoming dependencies.
type RelationshipProvider interface {
	Link(ctx context.Context, todo *types.TODO, targetRef string, relation types.RelationKind) (*Link, error)
	Unlink(ctx context.Context, todo *types.TODO, targetRef string, relation types.RelationKind) error
	Links(ctx context.Context, todo *types.TODO) ([]Link, error)
}

// PlanRevisionProvider appends a human-edited immutable Captain revision and
// keeps it selected on the native issue. Implementations must not rewrite an
// agent-owned plan file.
type PlanRevisionProvider interface {
	SavePlanRevision(ctx context.Context, todo *types.TODO, markdown, actor string) (*types.TODO, error)
}

type CreateRequest struct {
	Title        string
	Body         string
	Verification string
	Plan         *CreatePlanRequest
	Priority     types.Priority
	Status       types.Status
	Path         types.StringOrSlice
	Labels       []string
	Metadata     map[string]any
}

type CreatePlanRequest struct {
	Markdown string
	Approved bool
}

// EditRequest is a partial update to a TODO's content. A nil field is left
// unchanged, mirroring StateUpdate's pointer semantics.
type EditRequest struct {
	Title        *string
	Body         *string
	Verification *string
	Path         *types.StringOrSlice
	Labels       []string
	Metadata     map[string]any
}

// IsEmpty reports whether the edit would change nothing.
func (e EditRequest) IsEmpty() bool {
	return e.Title == nil && e.Body == nil && e.Verification == nil && e.Path == nil && len(e.Labels) == 0 && len(e.Metadata) == 0
}

func TODOReference(todo *types.TODO) string {
	if todo == nil {
		return ""
	}
	if todo.FilePath != "" {
		return todo.FilePath
	}
	return todo.ID
}
