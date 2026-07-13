package todos

import (
	"context"
	"errors"

	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
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
	Get(ctx context.Context, ref string) (*types.TODO, error)
	Create(ctx context.Context, req CreateRequest) (*types.TODO, error)
	Delete(ctx context.Context, todo *types.TODO) error
	// Edit updates a TODO's title and/or body in place.
	Edit(ctx context.Context, todo *types.TODO, edit EditRequest) error
	// Comment appends a free-form comment to a TODO's history.
	Comment(ctx context.Context, todo *types.TODO, body string) error
	UpdateState(ctx context.Context, todo *types.TODO, updates StateUpdate) error
	UpdateLatestFailure(ctx context.Context, todo *types.TODO, result *types.TestResultInfo) error
	SaveAttempt(ctx context.Context, todo *types.TODO, result *ExecutionResult) error
	// SaveVerification records an issue-verification verdict as a persistent
	// "## Verification Result" section/comment, replacing any prior result.
	SaveVerification(ctx context.Context, todo *types.TODO, result *verify.VerifyResult) error
}

// RunPreparation is the durable identity needed before an external agent is
// dispatched. PostgreSQL-backed providers use it to create and attach the
// authoritative Captain prompt run before any provider session starts.
type RunPreparation struct {
	Mode         types.RunMode
	ExecutorName string
	Resume       bool
	// PromptMarkdown is the exact user prompt the executor will dispatch. Native
	// storage persists it on Captain's prompt run before external execution.
	PromptMarkdown string
}

// RunPromptProvider exposes the exact initial user prompt an executor will
// dispatch. TODOExecutor asks for it before native admission so Captain stores
// the same prompt rather than a lossy reconstruction from the issue body.
type RunPromptProvider interface {
	RenderRunPrompt(ctx *ExecutorContext, todo *types.TODO) (string, error)
}

// RunLifecycleProvider is implemented by the native PostgreSQL runtime. Native
// execution state is owned by Captain and projected into issues.
type RunLifecycleProvider interface {
	PrepareRun(ctx context.Context, todo *types.TODO, preparation RunPreparation) error
	RecordRunStart(ctx context.Context, todo *types.TODO, metadata RunStartMetadata) error
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

// PlanContentProvider resolves the durable Captain plan content that should be
// fed to an agent. Plan mode receives the latest selected revision so it can
// revise it; run mode receives only the explicitly approved revision.
type PlanContentProvider interface {
	PlanMarkdown(ctx context.Context, todo *types.TODO, mode types.RunMode) (string, error)
}

// PlanRevisionProvider appends a human-edited immutable Captain revision and
// keeps it selected on the native issue. Implementations must not rewrite an
// agent-owned plan file.
type PlanRevisionProvider interface {
	SavePlanRevision(ctx context.Context, todo *types.TODO, markdown, actor string) (*types.TODO, error)
}

type CreateRequest struct {
	Title    string
	Body     string
	Priority types.Priority
	Status   types.Status
	Path     types.StringOrSlice
	Labels   []string
	Metadata map[string]any
}

// EditRequest is a partial update to a TODO's title and/or body. A nil field is
// left unchanged, mirroring StateUpdate's pointer semantics.
type EditRequest struct {
	Title    *string
	Body     *string
	Path     *types.StringOrSlice
	Labels   []string
	Metadata map[string]any
}

// IsEmpty reports whether the edit would change nothing.
func (e EditRequest) IsEmpty() bool {
	return e.Title == nil && e.Body == nil && e.Path == nil && len(e.Labels) == 0 && len(e.Metadata) == 0
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
