package todos

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/labels"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
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

// ErrRunOwnedElsewhere reports that a TODO already has a live run, driven by a
// process that is still running. It is not a failure of the new dispatch so
// much as a decision for the caller: run alongside the incumbent, or don't.
// The dispatcher is named because "still running" is only meaningful with the
// process that is doing the running.
type ErrRunOwnedElsewhere struct {
	IssueID     uuid.UUID
	PromptRunID uuid.UUID
	StepKind    string
	Owner       string
	Since       time.Duration
}

func (e *ErrRunOwnedElsewhere) Error() string {
	return fmt.Sprintf(
		"todo %s already has an active %s run %s, driven by %s for %s — rerun with --force to run alongside it",
		e.IssueID, e.StepKind, e.PromptRunID, e.Owner, e.Since.Round(time.Second),
	)
}

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
	Mode types.RunMode
	// Prompt is the name of the prompt being dispatched. Several prompts share a
	// Mode, so it — not Mode — distinguishes two runs of the same behaviour class
	// against one issue version. Empty means the prompt matching Mode.
	Prompt       string
	ExecutorName string
	Resume       bool
	// Concurrent dispatches alongside a run that is still live and owned by a
	// running process, instead of refusing. The caller has confirmed it (the
	// --force flag, or the dashboard's confirmation), so the two runs get
	// distinct Captain identities and each reports its own outcome.
	Concurrent bool
	Requested  captaindb.PromptRunRuntimeSelection
	// Spec is the exact request the executor will dispatch — model, budget,
	// permissions, setup, workflow and the rendered user prompt. Native storage
	// persists it on Captain's prompt run before external execution, so a later
	// continuation replays what actually ran instead of reconstructing it from
	// the run's resolved model/backend labels.
	Spec api.Spec
}

// RunPreparationResult is the durable Captain identity allocated before an
// external agent is dispatched. SessionID is Captain's admission session UUID,
// not the provider-specific session identity reported after launch.
type RunPreparationResult struct {
	SessionID string
	// PromptRunID is the run this execution owns. With concurrent runs allowed
	// on one TODO, the issue's active-run pointer is no longer enough to tell
	// an executor's callbacks which run they are reporting on.
	PromptRunID uuid.UUID
}

// RunRuntimeProvider exposes the configuration known before dispatch.
type RunRuntimeProvider interface {
	RunRuntimeSelection() captaindb.PromptRunRuntimeSelection
}

// RunPromptProvider reports which named prompt an executor is about to run. It
// is derived from the executor rather than set alongside the mode because the
// two must agree, and a caller that remembers one and forgets the other would
// give the run an identity that collides with a different prompt's.
type RunPromptProvider interface {
	RunPromptName() string
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
	PrepareRun(ctx context.Context, todo *types.TODO, preparation RunPreparation) (RunPreparationResult, error)
	RecordRunStart(ctx context.Context, todo *types.TODO, metadata RunStartMetadata) error
}

type RunProgressProvider interface {
	RecordRunProgress(ctx context.Context, todo *types.TODO, snapshot fixtures.ExecutionSnapshot) error
}

// RunNoticeProvider persists what a run's lifecycle hooks did — the commits they
// cut between turns — into the session transcript. Flushed once the run is over,
// because a hook firing mid-turn cannot yet know the transcript session's id:
// that row only exists after the provider's log has been ingested.
type RunNoticeProvider interface {
	RecordRunNotices(ctx context.Context, sessionID string, notices []api.Notice) error
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

// TodoAlias is one workspace-scoped reference that resolves to a TODO. Kind
// names the system the reference belongs to, e.g. "github" for an issue pushed
// to a GitHub tracker.
type TodoAlias struct {
	Alias string `json:"alias"`
	Kind  string `json:"kind,omitempty"`
}

// AliasProvider exposes a TODO's external references and appends new ones.
// AddAlias must preserve the aliases already recorded — the underlying storage
// replaces the whole set, so implementations read before they write.
type AliasProvider interface {
	Aliases(ctx context.Context, todo *types.TODO) ([]TodoAlias, error)
	AddAlias(ctx context.Context, todo *types.TODO, alias TodoAlias) error
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

// LabelDefinitionProvider exposes the editable label taxonomy — the colour,
// glyph and description a raw label string renders as. A definition is global
// (every workspace) or workspace-scoped; a workspace row shadows the global one
// of the same name, which shadows the built-in default. Labels themselves stay
// plain strings on the TODO: this is presentation, not content.
type LabelDefinitionProvider interface {
	LabelDefinitions(ctx context.Context) (labels.Definitions, error)
	SetLabelDefinition(ctx context.Context, definition labels.Definition, global bool) (labels.Definition, error)
	// DeleteLabelDefinition retires a label. Removing it from this workspace
	// also strips it from every TODO here; removing the global definition drops
	// only the shared presentation, since that scope spans every project.
	DeleteLabelDefinition(ctx context.Context, name string, global bool) (labels.Removal, error)
	LabelCounts(ctx context.Context) (map[string]int, error)
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
	// Labels replaces the TODO's whole label set. A nil pointer leaves them
	// unchanged; a non-nil empty slice clears every label. The two are different
	// requests, which a bare slice cannot express — a plain []string made
	// "remove the last label" indistinguishable from "don't touch labels".
	Labels   *[]string
	Metadata map[string]any
}

// IsEmpty reports whether the edit would change nothing.
func (e EditRequest) IsEmpty() bool {
	return e.Title == nil && e.Body == nil && e.Verification == nil && e.Path == nil && e.Labels == nil && len(e.Metadata) == 0
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
