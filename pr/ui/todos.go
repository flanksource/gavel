package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/prwatch"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/drivers"
	"github.com/flanksource/gavel/todos/labels"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/run"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	todospec "github.com/flanksource/gavel/todos/spec"
	"github.com/flanksource/gavel/todos/types"
	"github.com/ghodss/yaml"
	"github.com/google/uuid"
)

type todoCounts struct {
	Total      int `json:"total"`
	Open       int `json:"open"`
	Draft      int `json:"draft"`
	Pending    int `json:"pending"`
	InProgress int `json:"inProgress"`
	Review     int `json:"review"`
	Ask        int `json:"ask"`
	Failed     int `json:"failed"`
	Verified   int `json:"verified"`
	Completed  int `json:"completed"`
	Skipped    int `json:"skipped"`
}

type todoListResponse struct {
	Dir    string        `json:"dir,omitempty"`
	Counts todoCounts    `json:"counts"`
	Items  []todoSummary `json:"items"`
}

type todoSummary struct {
	Ref            string         `json:"ref"`
	ID             string         `json:"id,omitempty"`
	ShortID        string         `json:"shortId,omitempty"`
	Version        int64          `json:"version,omitempty"`
	WorkspaceID    string         `json:"workspaceId,omitempty"`
	ExecutionState string         `json:"executionState,omitempty"`
	Title          string         `json:"title"`
	Status         types.Status   `json:"status"`
	Priority       types.Priority `json:"priority"`
	CWD            string         `json:"cwd,omitempty"`
	Labels         []string       `json:"labels,omitempty"`
	Attempts       int            `json:"attempts,omitempty"`
	Created        *time.Time     `json:"created,omitempty"`
	LastRun        *time.Time     `json:"lastRun,omitempty"`
	SessionID      string         `json:"sessionId,omitempty"`
	// LookupSessionID is set only when a global detail request used a Captain
	// session UUID instead of an issue reference. The client uses it to open the
	// exact historical session without replacing the Todo's active SessionID.
	LookupSessionID string                `json:"lookupSessionId,omitempty"`
	Body            string                `json:"body,omitempty"`
	Implementation  string                `json:"implementation,omitempty"`
	Events          []types.ProviderEvent `json:"events,omitempty"`
	// Criteria are the todo's acceptance criteria, parsed from its
	// "## Acceptance Criteria" section; populated on detail responses so the
	// dashboard can render and edit them structurally.
	Criteria []types.AcceptanceCriterion `json:"criteria,omitempty"`
	// VerificationMarkdown is the raw fixture source stored separately from the
	// issue body; populated on detail responses for the Verification editor.
	VerificationMarkdown string `json:"verificationMarkdown,omitempty"`
	// Diff is the aggregated git diff footprint of the todo's commits (those
	// carrying its Gavel-Issue-Id trailer); nil when no commits reference it.
	Diff *todoDiffStat `json:"diff,omitempty"`
	// ExternalIssue is the tracker issue this todo was pushed to; nil when it
	// has never been pushed. Populated on list responses too, so the dashboard
	// can filter linked from unlinked work without a second request.
	ExternalIssue *types.ExternalIssue `json:"externalIssue,omitempty"`
	// HasPlan/HasVerification are lightweight availability flags for the list
	// row's indicators — unlike PlanPath/VerificationMarkdown they're populated
	// on both list and detail responses since they cost no extra parsing.
	HasPlan         bool `json:"hasPlan,omitempty"`
	HasVerification bool `json:"hasVerification,omitempty"`
	// Phases is the latest run per lifecycle phase (plan/triage/run/verify).
	// Populated on list responses too: the provider reads the whole workspace's
	// phases in one query, so the dashboard's phase columns cost nothing per
	// row and need no follow-up request.
	Phases types.PhaseRuns `json:"phases,omitempty"`
	// Envelope-driven fields from the last agent run: the native plan file the
	// agent reported (rendered by the Plan tab / review mode), its status, the
	// run mode an answer-resume continues in, the final summary, and the
	// questions blocking an ask todo.
	PlanPath       string                `json:"planPath,omitempty"`
	PlanStatus     types.PlanStatus      `json:"planStatus,omitempty"`
	RunMode        types.RunMode         `json:"runMode,omitempty"`
	LastRunSummary string                `json:"lastRunSummary,omitempty"`
	Questions      []types.AgentQuestion `json:"questions,omitempty"`
}

// todoDiffStat is the JSON shape of a todo's aggregated git diff footprint,
// mirroring git.DiffStat for the dashboard.
type todoDiffStat struct {
	Commits int `json:"commits"`
	Files   int `json:"files"`
	Adds    int `json:"adds"`
	Dels    int `json:"dels"`
}

type todoSource struct {
	Dir string
}

type todoPRVerificationPayload struct {
	PRNumber   int      `json:"prNumber,omitempty"`
	Repo       string   `json:"repo,omitempty"`
	CommentIDs []int64  `json:"commentIds,omitempty"`
	Actions    []string `json:"actions,omitempty"`
}

type todoCreatePayload struct {
	Dir      string         `json:"dir,omitempty"`
	Title    string         `json:"title"`
	Body     string         `json:"body,omitempty"`
	Priority types.Priority `json:"priority,omitempty"`
	Status   types.Status   `json:"status,omitempty"`
	// Criteria, when set, are folded into the body as a "## Acceptance Criteria"
	// checklist on create — used by the dashboard's "create todo from PR" flow to
	// seed a todo with the PR's failing tests and lint violations as criteria.
	Criteria []types.AcceptanceCriterion `json:"criteria,omitempty"`
	// PRVerification, when set, is converted server-side into a generated exec
	// fixture in "## Verification" so GitHub comments/actions are checked by the
	// fixture runner rather than duplicated as AI-scored acceptance criteria.
	PRVerification *todoPRVerificationPayload `json:"prVerification,omitempty"`
	// Labels tags the new todo. Presentation for each label is resolved from the
	// label definitions, not stored here.
	Labels []string `json:"labels,omitempty"`
}

type todoNewPayload struct {
	todoCreatePayload
	AutoSave *bool `json:"autoSave,omitempty"`
}

type todoAttachmentSummary struct {
	Field       string `json:"field,omitempty"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType,omitempty"`
	Size        int64  `json:"size"`
	ID          string `json:"id,omitempty"`
	URL         string `json:"url,omitempty"`
	IsImage     bool   `json:"isImage,omitempty"`
}

type todoNewResponse struct {
	Todo        todoSummary             `json:"todo"`
	AutoSave    bool                    `json:"autoSave"`
	Attachments []todoAttachmentSummary `json:"attachments,omitempty"`
}

type todoUpdatePayload struct {
	Dir      string         `json:"dir,omitempty"`
	Ref      string         `json:"ref,omitempty"`
	Status   types.Status   `json:"status,omitempty"`
	Priority types.Priority `json:"priority,omitempty"`
	// Title/Body edit the TODO's content; a nil pointer leaves the field
	// unchanged (an explicit empty body is allowed, an empty title is not).
	Title *string `json:"title,omitempty"`
	Body  *string `json:"body,omitempty"`
	// Comment, when set, appends a comment. Combined with status it reopens (or
	// closes) the TODO with a comment in one request.
	Comment string `json:"comment,omitempty"`
	// Labels replaces the TODO's whole label set. A nil pointer leaves them
	// unchanged; a non-nil empty array clears every label.
	Labels *[]string `json:"labels,omitempty"`
}

// todoTransferPayload moves the todo at Ref from one native workspace to another.
type todoTransferPayload struct {
	Ref     string `json:"ref"`
	FromDir string `json:"fromDir,omitempty"`
	ToDir   string `json:"toDir"`
}

type todoTransferResponse struct {
	Dir  string      `json:"dir"`
	Todo todoSummary `json:"todo"`
}

type todoRunPayload struct {
	Dir  string   `json:"dir,omitempty"`
	Ref  string   `json:"ref,omitempty"`
	Refs []string `json:"refs,omitempty"`
	// Agent and Mode are retained only to reject the removed legacy payload shape
	// with a useful boundary error.
	Agent string `json:"agent,omitempty"`
	Mode  string `json:"mode,omitempty"`
	// Driver is the canonical execution mode: api, agent, cli, or cmux.
	Driver string `json:"driver,omitempty"`

	// Spec carries the model/mode/effort/prompt/budget/permissions/session
	// knobs — the same request shape captain's prompt run editor produces
	// (model, mode, effort, prompt.user, budget.{cost,maxTurns,timeout},
	// permissions.{mode,tools}, sessionId, ...).
	//
	// It is a named field under its own `spec` key, not embedded. api.Spec
	// declares value-receiver MarshalJSON/MarshalYAML to omit its empty sections;
	// embedding promoted them onto the payload, so marshaling emitted a bare spec
	// and silently dropped ref/agent/mode/driver/runMode/plan/resume — including
	// in the review API's `options` object. Nesting also dissolves the wire
	// collision between this payload's `mode`/`resume` and api.Model's own.
	Spec api.Spec `json:"spec,omitempty"`

	// RunMode selects the behaviour class: run (implement and commit) or plan
	// (neither). It supersedes the legacy Plan bool, which is still accepted until
	// the dashboard sends runMode.
	RunMode string `json:"runMode,omitempty"`
	Plan    bool   `json:"plan,omitempty"`
	// Prompt names the template to run — run, plan, triage, or a name declared in
	// .gavel.yaml todos.prompts. It is a separate axis from RunMode: each prompt
	// declares the class it runs as, so several prompts share one. Supplying both
	// and disagreeing is rejected rather than silently resolved.
	Prompt string `json:"prompt,omitempty"`
	// Resume continues the todo's prior agent session (claude --resume) instead of
	// starting fresh. It stays a sibling flag rather than a Spec field because it is
	// a session-identity decision: a fresh run also carries a (minted) sessionId, so
	// resume cannot be inferred from Spec.SessionID.
	Resume bool `json:"resume,omitempty"`
	// Force dispatches even when the todo already has a live run owned by a
	// running process: the two runs proceed in parallel. Without it such a
	// dispatch is refused and the client is told which run is in the way.
	Force bool `json:"force,omitempty"`
	// Dirty/DryRun/Commit/Check are no longer sibling flags — the run's dirty-tree
	// carry-across, dry-run, auto-commit, and checks all come from Spec
	// (Setup.Checkout.Worktree.Uncommitted and Workflow.Verify/Commits), surfaced
	// by clicky's Workspace/Verify/Commit sections.
}

type todoRunResponse struct {
	Status    string   `json:"status"`
	Ref       string   `json:"ref"`
	Refs      []string `json:"refs,omitempty"`
	Count     int      `json:"count"`
	Dir       string   `json:"dir"`
	Provider  string   `json:"provider,omitempty"`
	Model     string   `json:"model,omitempty"`
	Mode      string   `json:"mode,omitempty"`
	Effort    string   `json:"effort,omitempty"`
	Driver    string   `json:"driver,omitempty"`
	RunMode   string   `json:"runMode,omitempty"`
	Plan      bool     `json:"plan,omitempty"`
	Resume    bool     `json:"resume,omitempty"`
	SessionID string   `json:"sessionId,omitempty"`
	Timeout   string   `json:"timeout"`
	MaxBudget float64  `json:"maxBudget,omitempty"`
	MaxTurns  int      `json:"maxTurns,omitempty"`
	Commit    bool     `json:"commit"`
	Message   string   `json:"message"`
}

type todoRunPreviewResponse struct {
	Prompt   string `json:"prompt"`
	SpecYAML string `json:"specYaml"`
	Provider string `json:"provider,omitempty"`
	Mode     string `json:"mode,omitempty"`
	Effort   string `json:"effort,omitempty"`
	RunMode  string `json:"runMode,omitempty"`
	Plan     bool   `json:"plan,omitempty"`
	Count    int    `json:"count"`
}

// Run execution lives in todos/run, not here: executing a TODO is not an HTTP
// concern, and the CLI and the clicky entity need the same seam. What stays in
// this package is the dashboard's own resolution — the (driver, mode)
// catalog the run dialog offers, which the other entrypoints have no equivalent
// of — plus the handlers. These aliases keep the existing call sites reading in
// the dashboard's vocabulary.
type (
	todoRunStartResult = run.StartResult
	todoRunRequest     = run.Request
	todoRunOptions     = run.Options
)

// run.Start is deliberately NOT aliased here. An alias copies the function
// value at init, so a test replacing the copy would leave every run started
// through the todos entity — which calls run.Start directly — going to the real
// driver. One seam, called by its own name.

var (
	specCommit            = run.Commit
	specDryRun            = run.DryRun
	specDirty             = run.Dirty
	runIsStoppable        = run.IsStoppable
	todoRunRefs           = run.Refs
	todoRunStartedMessage = run.StartedMessage
	todoRunLabel          = run.Label
	resolveRunSessionID   = run.ResolveSessionID

	newTodoRunExecutor        = run.NewExecutor
	newTodoRunExecutorContext = run.NewExecutorContext
)

func (s *Server) handleTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		s.handleTodosList(w, r)
	case http.MethodPost:
		s.handleTodoCreate(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTodoItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		s.handleTodoGet(w, r)
	case http.MethodPatch:
		s.handleTodoPatch(w, r)
	case http.MethodDelete:
		s.handleTodoDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTodosList(w http.ResponseWriter, r *http.Request) {
	source := todoSourceFromRequest(r)
	provider, source, err := s.todoProviderContext(r.Context(), source)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	filters, err := todoFiltersFromRequest(r)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	items, err := provider.List(r.Context(), filters)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	resp := todoListResponse{
		Dir:    source.Dir,
		Counts: summarizeTodos(items),
		Items:  make([]todoSummary, 0, len(items)),
	}
	stats := commitDiffStats(r.Context(), source.Dir)
	for _, item := range items {
		sum := summarizeTodo(item, false)
		sum.Diff = diffStatFor(stats, item.ID)
		resp.Items = append(resp.Items, sum)
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// diffStatFor builds the JSON diff stat for a todo id from the workspace's
// computed map, returning nil when the todo has no linked commits so the field
// is omitted rather than serialized as an all-zero object.
func diffStatFor(stats map[string]gavelgit.DiffStat, id string) *todoDiffStat {
	d, ok := stats[strings.TrimSpace(id)]
	if !ok || (d.Commits == 0 && d.Files == 0) {
		return nil
	}
	return &todoDiffStat{Commits: d.Commits, Files: d.Files, Adds: d.Adds, Dels: d.Dels}
}

func (s *Server) handleTodoGet(w http.ResponseWriter, r *http.Request) {
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if ref == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	requestedSource := todoSourceFromRequest(r)
	_, source, todo, lookupSessionID, err := s.resolveTodoGetReference(r.Context(), requestedSource, ref)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	sum := summarizeTodo(todo, true)
	sum.LookupSessionID = lookupSessionID
	sum.Diff = diffStatFor(commitDiffStats(r.Context(), source.Dir), todo.ID)
	json.NewEncoder(w).Encode(sum) //nolint:errcheck
}

// resolveTodoGetReference extends the ordinary global issue lookup with an
// exact UUID-only session fallback. Issue UUIDs and UUID-shaped aliases remain
// authoritative; only a genuine issue miss is allowed to resolve through
// Captain's prompt-run links. The canonical issue is then re-read through its
// owning workspace provider so all detail fields retain their normal boundary.
func (s *Server) resolveTodoGetReference(
	ctx context.Context,
	requested todoSource,
	ref string,
) (todos.Provider, todoSource, *types.TODO, string, error) {
	provider, source, todo, err := s.resolveTodoReference(ctx, requested, ref)
	if err == nil {
		return provider, source, todo, "", nil
	}
	if strings.TrimSpace(requested.Dir) != "" || !errors.Is(err, native.ErrNotFound) {
		return nil, source, nil, "", err
	}
	if _, parseErr := uuid.Parse(strings.TrimSpace(ref)); parseErr != nil {
		return nil, source, nil, "", err
	}

	global, globalErr := openGlobalTodoProvider(ctx)
	if globalErr != nil {
		return nil, source, nil, "", err
	}
	sessions, ok := global.(todos.GlobalSessionReferenceProvider)
	if !ok {
		return nil, source, nil, "", err
	}
	sessionTodo, sessionID, sessionErr := sessions.GetGlobalBySession(ctx, ref)
	if sessionErr != nil {
		if errors.Is(sessionErr, native.ErrNotFound) {
			return nil, source, nil, "", fmt.Errorf("%w: TODO or session UUID %q", native.ErrNotFound, ref)
		}
		return nil, source, nil, "", sessionErr
	}
	ownerDir := strings.TrimSpace(sessionTodo.CWD)
	if ownerDir == "" {
		return nil, source, nil, "", fmt.Errorf("resolved session %q has no owning workspace path", ref)
	}
	provider, source, err = s.todoProviderContext(ctx, todoSource{Dir: ownerDir})
	if err != nil {
		return nil, source, nil, "", err
	}
	todo, err = provider.Get(ctx, sessionTodo.ID)
	if err != nil {
		return nil, source, nil, "", err
	}
	return provider, source, todo, sessionID, nil
}

// resolveTodoReference loads one issue and returns a provider scoped to its
// owning workspace. With an explicit dir the lookup is workspace-local. With
// no dir it first resolves a UUID/short UUID/imported alias globally, then
// reopens the provider for the authoritative CWD so later mutations and run
// lifecycle writes cannot accidentally target the server's default workspace.
func (s *Server) resolveTodoReference(ctx context.Context, requested todoSource, ref string) (todos.Provider, todoSource, *types.TODO, error) {
	globalLookup := strings.TrimSpace(requested.Dir) == ""
	if !globalLookup {
		provider, source, err := s.todoProviderContext(ctx, requested)
		if err != nil {
			return nil, source, nil, err
		}
		todo, err := provider.Get(ctx, ref)
		return provider, source, todo, err
	}

	global, err := openGlobalTodoProvider(ctx)
	if err != nil {
		return nil, requested, nil, err
	}
	todo, err := global.GetGlobal(ctx, ref)
	if err != nil {
		return nil, requested, nil, err
	}
	ownerDir := strings.TrimSpace(todo.CWD)
	if ownerDir == "" {
		return nil, requested, nil, fmt.Errorf("resolved TODO %q has no owning workspace path", ref)
	}
	provider, source, err := s.todoProviderContext(ctx, todoSource{Dir: ownerDir})
	if err != nil {
		return nil, source, nil, err
	}
	// Re-read through the owning repository so the returned object and provider
	// share the same optimistic-version and workspace boundary.
	todo, err = provider.Get(ctx, todo.ID)
	if err != nil {
		return nil, source, nil, err
	}
	return provider, source, todo, nil
}

func (s *Server) handleTodoCreate(w http.ResponseWriter, r *http.Request) {
	var payload todoCreatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	if payload.Status != "" {
		if err := types.ValidateAssignableStatus(payload.Status); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
	}
	if payload.Priority != "" {
		if err := types.ValidatePriority(payload.Priority); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
	}
	source := todoSourceFromRequest(r)
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	provider, _, err := s.todoProviderContext(r.Context(), source)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	todo, err := provider.Create(r.Context(), todos.CreateRequest{
		Title:    payload.Title,
		Body:     bodyWithCreateSections(payload.Body, payload.Criteria, payload.PRVerification),
		Priority: payload.Priority,
		Status:   payload.Status,
		Labels:   payload.Labels,
	})
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(summarizeTodo(todo, true)) //nolint:errcheck
}

// bodyWithCriteria folds an "## Acceptance Criteria" checklist into the todo body
// when criteria are supplied (e.g. a todo created from a PR's failing tests and
// lint violations), so they round-trip through the provider's parse as the todo's
// acceptance criteria. An empty list leaves the body untouched.
func bodyWithCriteria(body string, criteria []types.AcceptanceCriterion) string {
	if len(criteria) == 0 {
		return body
	}
	return todos.UpsertCriteriaSection(body, criteria)
}

func bodyWithCreateSections(body string, criteria []types.AcceptanceCriterion, verification *todoPRVerificationPayload) string {
	body = bodyWithCriteria(body, criteria)
	if verification == nil {
		return body
	}
	return prwatch.UpsertPRStatusVerification(body, prwatch.PRStatusVerification{
		PRNumber:   verification.PRNumber,
		Repo:       verification.Repo,
		CommentIDs: verification.CommentIDs,
		Actions:    verification.Actions,
	})
}

func (s *Server) handleTodoNew(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	payload, attachments, err := parseTodoNewPayload(r)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	autoSave := false
	if payload.AutoSave != nil {
		autoSave = *payload.AutoSave
	}
	if payload.Status == "" {
		if autoSave {
			payload.Status = types.StatusPending
		} else {
			payload.Status = types.StatusDraft
		}
	}
	if payload.Status != "" {
		if err := types.ValidateAssignableStatus(payload.Status); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
	}
	if payload.Priority != "" {
		if err := types.ValidatePriority(payload.Priority); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
	}

	source := todoSourceFromRequest(r)
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	provider, _, err := s.todoProviderContext(r.Context(), source)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	todo, err := provider.Create(r.Context(), todos.CreateRequest{
		Title:    payload.Title,
		Body:     bodyWithCreateSections(todoBodyWithAttachments(payload.Body, attachments), payload.Criteria, payload.PRVerification),
		Priority: payload.Priority,
		Status:   payload.Status,
	})
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todoNewResponse{ //nolint:errcheck
		Todo:        summarizeTodo(todo, true),
		AutoSave:    autoSave,
		Attachments: attachments,
	})
}

func (s *Server) handleTodoPatch(w http.ResponseWriter, r *http.Request) {
	payload, attachments, err := parseTodoUpdatePayload(r)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	ref := strings.TrimSpace(payload.Ref)
	if ref == "" {
		ref = strings.TrimSpace(r.URL.Query().Get("ref"))
	}
	if ref == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	// A PATCH may edit content (title/body), change state (status/priority),
	// add a comment, or any combination; at least one operation is required.
	var update todos.StateUpdate
	if payload.Status != "" {
		if err := types.ValidateAssignableStatus(payload.Status); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
		update.Status = &payload.Status
	}
	if payload.Priority != "" {
		if err := types.ValidatePriority(payload.Priority); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
		update.Priority = &payload.Priority
	}

	var edit todos.EditRequest
	if payload.Title != nil {
		title := strings.TrimSpace(*payload.Title)
		if title == "" {
			writeTodoError(w, http.StatusBadRequest, fmt.Errorf("title cannot be empty"))
			return
		}
		edit.Title = &title
	}
	if payload.Body != nil {
		edit.Body = payload.Body
	}
	if payload.Labels != nil {
		normalized := make([]string, 0, len(*payload.Labels))
		for _, label := range *payload.Labels {
			if label = labels.Normalize(label); label != "" {
				normalized = append(normalized, label)
			}
		}
		edit.Labels = &normalized
	}
	comment := strings.TrimSpace(payload.Comment)
	if len(attachments) > 0 {
		comment = todoBodyWithAttachments(comment, attachments)
	}

	hasState := update.Status != nil || update.Priority != nil
	if !hasState && edit.IsEmpty() && comment == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("status, priority, title, body, labels, or comment is required"))
		return
	}

	source := todoSourceFromRequest(r)
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	provider, _, todo, err := s.resolveTodoReference(r.Context(), source, ref)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	// Order: edit content, then reopen/close, then comment, so a reopen-with-comment
	// posts the comment against the now-open TODO and it lands last in the timeline.
	if !edit.IsEmpty() {
		if err := provider.Edit(r.Context(), todo, edit); err != nil {
			writeTodoError(w, http.StatusInternalServerError, err)
			return
		}
	}
	if hasState {
		if err := provider.UpdateState(r.Context(), todo, update); err != nil {
			writeTodoError(w, http.StatusInternalServerError, err)
			return
		}
	}
	if comment != "" {
		if err := provider.Comment(r.Context(), todo, comment); err != nil {
			writeTodoError(w, http.StatusInternalServerError, err)
			return
		}
	}
	// Edits and comments mutate the body/event history; re-read so the response
	// reflects the provider's authoritative state (new event, rewritten body).
	if !edit.IsEmpty() || comment != "" {
		if refreshed, gerr := provider.Get(r.Context(), todo.ID); gerr == nil {
			todo = refreshed
		}
	}
	json.NewEncoder(w).Encode(summarizeTodo(todo, true)) //nolint:errcheck
}

func (s *Server) handleTodoDelete(w http.ResponseWriter, r *http.Request) {
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if ref == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	provider, _, todo, err := s.resolveTodoReference(r.Context(), todoSourceFromRequest(r), ref)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	if err := provider.Delete(r.Context(), todo); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	fmt.Fprint(w, `{"status":"ok"}`)
}

func (s *Server) handleTodoTransfer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload todoTransferPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	if strings.TrimSpace(payload.Ref) == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	if strings.TrimSpace(payload.ToDir) == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("toDir is required"))
		return
	}
	requestedSource := todoSource{Dir: payload.FromDir}
	source, src, resolvedTodo, err := s.resolveTodoReference(r.Context(), requestedSource, payload.Ref)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	// Transfer through the canonical native id after a legacy alias lookup.
	payload.Ref = resolvedTodo.ID
	target, dst, err := s.todoProviderContext(r.Context(), todoSource{Dir: payload.ToDir})
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	// Refuse a no-op self-transfer, which would create a duplicate and then
	// delete the original in the same native workspace.
	if src.Dir == dst.Dir {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("source and target are the same workspace"))
		return
	}
	created, err := todos.Transfer(r.Context(), source, target, payload.Ref)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(todoTransferResponse{ //nolint:errcheck
		Dir:  dst.Dir,
		Todo: summarizeTodo(created, true),
	})
}

// resolveTodoRunRequest decodes a run/preview payload and resolves its options,
// provider, and todos. handleTodoRun and handleTodoRunPreview share it so both
// interpret the same request identically; the returned status is the HTTP code
// to report when err is non-nil.
func (s *Server) resolveTodoRunRequest(r *http.Request) (todos.Provider, todoSource, []*types.TODO, todoRunOptions, int, error) {
	var payload todoRunPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, todoSource{}, nil, todoRunOptions{}, http.StatusBadRequest, fmt.Errorf("invalid json")
	}
	refs := normalizeTodoRunRefs(payload, r)
	return s.resolveTodoRunPayload(r.Context(), payload, refs, todoSourceFromRequest(r), requestOrigin(r))
}

// resolveTodoRunPayload turns an already-decoded payload into everything a run
// needs. It is separate from resolveTodoRunRequest so a fan-out handler can
// resolve one synthetic payload per selected TODO through exactly the same path
// a single run takes — a second resolution path would be one more place for the
// dashboard's runs to diverge from each other.
func (s *Server) resolveTodoRunPayload(
	ctx context.Context,
	payload todoRunPayload,
	refs []string,
	source todoSource,
	origin string,
) (todos.Provider, todoSource, []*types.TODO, todoRunOptions, int, error) {
	if len(refs) == 0 {
		return nil, todoSource{}, nil, todoRunOptions{}, http.StatusBadRequest, fmt.Errorf("ref is required")
	}
	// Wire validation runs before anything is looked up, so a malformed payload
	// still fails as a 400 without opening a workspace. Resolution itself needs
	// the workspace and the todos — it folds .gavel.yaml and each todo's `llm:`
	// frontmatter — so it happens last.
	if _, err := validateTodoRunPayload(payload); err != nil {
		return nil, todoSource{}, nil, todoRunOptions{}, http.StatusBadRequest, err
	}
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	globalLookup := strings.TrimSpace(source.Dir) == ""
	if err := validateTodoRunCardinality(len(refs)); err != nil {
		return nil, source, nil, todoRunOptions{}, http.StatusBadRequest, err
	}

	var provider todos.Provider
	var todoList []*types.TODO
	if globalLookup {
		p, owningSource, todo, err := s.resolveTodoReference(ctx, source, refs[0])
		if err != nil {
			return p, owningSource, nil, todoRunOptions{}, http.StatusNotFound, err
		}
		provider, source, todoList = p, owningSource, []*types.TODO{todo}
	} else {
		p, resolvedSource, err := s.todoProviderContext(ctx, source)
		if err != nil {
			return nil, resolvedSource, nil, todoRunOptions{}, http.StatusBadRequest, err
		}
		provider, source = p, resolvedSource
		todoList = make([]*types.TODO, 0, len(refs))
		for _, ref := range refs {
			todo, err := provider.Get(ctx, ref)
			if err != nil {
				return provider, source, nil, todoRunOptions{}, http.StatusNotFound, err
			}
			todoList = append(todoList, todo)
		}
	}
	for _, todo := range todoList {
		todo.MarkdownBody = todos.AbsolutizeAttachmentURLs(todo.MarkdownBody, origin)
	}

	opts, err := normalizeTodoRunOptions(source.Dir, todoList, payload)
	if err != nil {
		return provider, source, todoList, todoRunOptions{}, http.StatusBadRequest, err
	}
	return provider, source, todoList, opts, http.StatusOK, nil
}

func validateTodoRunCardinality(count int) error {
	if count <= 1 {
		return nil
	}
	return fmt.Errorf("grouped TODO execution is not supported by the native PostgreSQL runtime; run one issue at a time")
}

func (s *Server) handleTodoRun(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	provider, source, todoList, opts, status, err := s.resolveTodoRunRequest(r)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	backend := todos.ProviderDB
	// Resolve the run's session id once, up front, so it is stable across the
	// validation and start calls below and can be returned to the client to
	// follow the session log live (see handleTodoSessionStream).
	opts.Spec.SessionID = resolveRunSessionID(opts, todoList)
	req := todoRunRequest{
		Provider: provider,
		Registry: todoRuns(),
		Todos:    todoList,
		Dir:      source.Dir,
		Backend:  backend,
		Options:  opts,
	}
	if _, _, err := newTodoRunExecutorContext(r.Context(), req); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	resp := todoRunResponse{
		Ref:       todos.TODOReference(todoList[0]),
		Refs:      todoRunRefs(todoList),
		Count:     len(todoList),
		Dir:       source.Dir,
		Provider:  providerKey(opts.Spec.Model),
		Driver:    opts.Driver,
		Mode:      string(opts.Spec.Mode),
		Model:     opts.Spec.Name,
		Effort:    string(opts.Spec.Effort),
		RunMode:   string(opts.RunMode),
		Plan:      opts.RunMode == types.ModePlan,
		Resume:    opts.Resume,
		Timeout:   opts.Timeout.String(),
		MaxBudget: opts.Spec.Budget.Cost,
		MaxTurns:  opts.Spec.Budget.MaxTurns,
		Commit:    specCommit(opts.Spec) && !specDryRun(opts.Spec),
	}
	// A dry run still executes the agent; Captain's commit hook reports rather
	// than cuts the declared commit. Prompt-only inspection uses the preview API.
	started, err := run.Start(req)
	if err != nil {
		// A todo that is already running is a question for the user, not a bad
		// request: answer it by retrying with force, and the two runs proceed in
		// parallel. The dialog needs the incumbent's identity to say so.
		var owned *todos.ErrRunOwnedElsewhere
		if errors.As(err, &owned) {
			writeTodoRunConflict(w, owned)
			return
		}
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	if started.Status == "started" && strings.TrimSpace(started.SessionID) == "" {
		writeTodoError(w, http.StatusInternalServerError, errors.New("todo run was admitted without a Captain session id"))
		return
	}
	resp.Status = started.Status
	resp.SessionID = started.SessionID
	resp.Message = started.Message
	if resp.Message == "" && started.Status == "started" {
		resp.Message = todoRunStartedMessage(len(todoList))
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// handleTodoRunPreview renders the exact prompt a run would dispatch, without
// starting it, so the advanced run dialog can show the prompt that will be sent
// before the user commits to a run. It accepts the same payload as handleTodoRun.
func (s *Server) handleTodoRunPreview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	provider, source, todoList, opts, status, err := s.resolveTodoRunRequest(r)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	renderedSpec, specYAML, err := buildTodoRunSpecPreview(r.Context(), provider, source, todoList, opts)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	resp := todoRunPreviewResponse{
		Prompt:   renderedSpec.Prompt.User,
		SpecYAML: specYAML,
		Provider: providerKey(opts.Spec.Model),
		Mode:     string(opts.Spec.Mode),
		Effort:   string(opts.Spec.Effort),
		RunMode:  string(opts.RunMode),
		Plan:     opts.RunMode == types.ModePlan,
		Count:    len(todoList),
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// buildTodoRunPromptPreview renders the exact prompt a run would dispatch:
// framing, TODO sections, effort directive, and the mode's structured-output
// schema instruction. The editable prompt override replaces the body but the
// schema instruction is always re-appended, so the preview shows the full
// contract the agent receives.
func buildTodoRunSpecPreview(ctx context.Context, provider todos.Provider, source todoSource, todoList []*types.TODO, opts todoRunOptions) (api.Spec, string, error) {
	executor, _, err := newTodoRunExecutorContext(ctx, todoRunRequest{
		Provider: provider,
		Todos:    todoList,
		Dir:      source.Dir,
		Backend:  todos.ProviderDB,
		Options:  opts,
	})
	if err != nil {
		return api.Spec{}, "", err
	}
	renderer, ok := executor.(todos.RunSpecProvider)
	if !ok {
		return api.Spec{}, "", fmt.Errorf("todo run driver %q cannot render its Captain spec", opts.Driver)
	}
	rendered, err := renderer.RenderRunSpec(todos.NewExecutorContext(ctx, logger.StandardLogger(), nil), todoList[0])
	if err != nil {
		return api.Spec{}, "", err
	}
	encoded, err := yaml.Marshal(rendered)
	if err != nil {
		return api.Spec{}, "", fmt.Errorf("marshal rendered Captain spec as YAML: %w", err)
	}
	return rendered, string(encoded), nil
}

// resolveTodoDir turns a request's dir param into an absolute workspace path,
// defaulting to the server's work dir and joining relative dirs onto it.
func (s *Server) resolveTodoDir(dir string) string {
	workDir := s.todoWorkDir()
	if dir == "" {
		return workDir
	}
	if !filepath.IsAbs(dir) {
		return filepath.Join(workDir, dir)
	}
	return dir
}

func (s *Server) todoProviderContext(ctx context.Context, source todoSource) (todos.Provider, todoSource, error) {
	source.Dir = s.resolveTodoDir(source.Dir)
	provider, err := openTodoProvider(ctx, source.Dir)
	if err != nil {
		return nil, source, err
	}
	return provider, source, nil
}

// ProviderForProject resolves a stored project to the PostgreSQL runtime.
func ProviderForProject(ctx context.Context, p Project) (todos.Provider, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return todoruntime.Open(ctx, p.WorkspaceOptions())
}

// todoRuns is the in-flight run registry the dashboard starts and stops runs
// through. It is the process-wide one rather than Server state because the CLI
// and the todos entity start runs through the same registry, and a run the
// dashboard cannot see is a run it cannot stop.
func todoRuns() *run.Registry { return run.Shared() }

// openTodoProvider is the single API/UI native runtime seam. It is a variable so
// package tests can inject an in-memory implementation without opening PostgreSQL.
var openTodoProvider = func(ctx context.Context, dir string) (todos.Provider, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	project, err := ProjectForDir(dir)
	if err != nil {
		return nil, err
	}
	return todoruntime.Open(ctx, project.WorkspaceOptions())
}

var openGlobalTodoProvider = func(ctx context.Context) (todos.GlobalReferenceProvider, error) {
	return todoruntime.OpenGlobal(ctx)
}

func (s *Server) todoWorkDir() string {
	if s != nil && s.ghOpts.WorkDir != "" {
		return s.ghOpts.WorkDir
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

func todoSourceFromRequest(r *http.Request) todoSource {
	return todoSource{
		Dir: strings.TrimSpace(r.URL.Query().Get("dir")),
	}
}

func todoFiltersFromRequest(r *http.Request) (todos.DiscoveryFilters, error) {
	status := types.Status(strings.TrimSpace(r.URL.Query().Get("status")))
	if status == "" {
		return todos.DiscoveryFilters{}, nil
	}
	// Filtering accepts every known status, including the run projections a
	// caller may not write.
	if !types.IsKnownStatus(status) {
		return todos.DiscoveryFilters{}, fmt.Errorf("invalid status %q", status)
	}
	return todos.DiscoveryFilters{IncludeStatuses: []types.Status{status}}, nil
}

func summarizeTodo(todo *types.TODO, detail bool) todoSummary {
	if todo == nil {
		return todoSummary{}
	}
	title := todo.Title
	if title == "" {
		title = todo.Filename()
	}
	out := todoSummary{
		Ref:            todos.TODOReference(todo),
		ID:             todo.ID,
		ShortID:        todo.DisplayID(),
		Version:        todo.Version,
		WorkspaceID:    todo.WorkspaceID,
		ExecutionState: todo.ExecutionState,
		Title:          title,
		Status:         todo.Status,
		Priority:       todo.Priority,
		CWD:            todo.CWD,
		Labels:         todo.Labels,
		Attempts:       todo.Attempts,
		Created:        todo.Created,
		LastRun:        todo.LastRun,
		ExternalIssue:  todo.ExternalIssue,
		Phases:         todo.PhaseRuns,
	}
	if todo.LLM != nil {
		out.SessionID = todo.LLM.SessionId
	}
	if out.Ref == "" {
		out.Ref = todo.FilePath
	}
	out.HasPlan = todos.HasPlan(todo)
	out.HasVerification = strings.TrimSpace(todo.VerificationMarkdown) != ""
	if detail {
		out.Body = strings.TrimSpace(todo.MarkdownBody)
		out.Implementation = strings.TrimSpace(todo.Implementation)
		out.Events = todo.ProviderEvents
		out.Criteria = todo.AcceptanceCriteria
		out.VerificationMarkdown = strings.TrimSpace(todo.VerificationMarkdown)
		if out.Body == "" {
			out.Body = out.Implementation
		}
		out.PlanPath = todo.PlanPath
		out.PlanStatus = todo.PlanStatus
		out.RunMode = todo.RunMode
		out.LastRunSummary = todo.LastRunSummary
		out.Questions = todo.Questions
	}
	return out
}

func summarizeTodos(items types.TODOS) todoCounts {
	var counts todoCounts
	for _, item := range items {
		addTodoStatus(&counts, item.Status, 1)
	}
	return counts
}

// addTodoStatus folds n todos of one status into counts. It is the single
// status→bucket mapping: summarizeTodos walks materialized todos one at a time,
// countProjectTodos folds the provider's SQL aggregate, and neither owns a
// private copy of the switch.
func addTodoStatus(counts *todoCounts, status types.Status, n int) {
	counts.Total += n
	switch status {
	case types.StatusCompleted:
		counts.Completed += n
	case types.StatusDraft:
		counts.Open += n
		counts.Draft += n
	case types.StatusInProgress:
		counts.Open += n
		counts.InProgress += n
	case types.StatusReview:
		counts.Open += n
		counts.Review += n
	case types.StatusAsk:
		counts.Open += n
		counts.Ask += n
	case types.StatusFailed:
		counts.Open += n
		counts.Failed += n
	case types.StatusVerified:
		counts.Open += n
		counts.Verified += n
	case types.StatusSkipped:
		counts.Open += n
		counts.Skipped += n
	default:
		counts.Open += n
		counts.Pending += n
	}
}

func writeTodoError(w http.ResponseWriter, status int, err error) {
	switch {
	case errors.Is(err, ErrProjectNotFound):
		status = http.StatusNotFound
	case errors.Is(err, database.ErrUnavailable):
		status = http.StatusServiceUnavailable
	case errors.Is(err, native.ErrAmbiguousReference), errors.Is(err, native.ErrVersionConflict), errors.Is(err, native.ErrAliasConflict):
		status = http.StatusConflict
	case errors.Is(err, native.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, native.ErrInvalidInput):
		status = http.StatusBadRequest
	}
	w.WriteHeader(status)
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}

// writeTodoRunConflict answers a dispatch that lost to a live run with what the
// client needs to decide: which run is in the way, who is driving it, and that
// retrying with force runs the two in parallel.
func writeTodoRunConflict(w http.ResponseWriter, owned *todos.ErrRunOwnedElsewhere) {
	w.WriteHeader(http.StatusConflict)
	json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
		"error":       owned.Error(),
		"reason":      "run_owned_elsewhere",
		"promptRunId": owned.PromptRunID,
		"stepKind":    owned.StepKind,
		"owner":       owned.Owner,
		"runningFor":  owned.Since.Round(time.Second).String(),
		"retryWith":   map[string]bool{"force": true},
	})
}

func parseTodoNewPayload(r *http.Request) (todoNewPayload, []todoAttachmentSummary, error) {
	var payload todoNewPayload
	var attachments []todoAttachmentSummary
	contentType := strings.ToLower(r.Header.Get("Content-Type"))

	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"):
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return payload, nil, fmt.Errorf("invalid multipart form: %w", err)
		}
		if r.MultipartForm != nil {
			if err := applyTodoNewValues(&payload, r.MultipartForm.Value, true); err != nil {
				return payload, nil, err
			}
			stored, err := persistMultipartAttachments(r.MultipartForm)
			if err != nil {
				return payload, nil, err
			}
			attachments = stored
		}
	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		if err := r.ParseForm(); err != nil {
			return payload, nil, fmt.Errorf("invalid form: %w", err)
		}
		if err := applyTodoNewValues(&payload, r.PostForm, true); err != nil {
			return payload, nil, err
		}
	case strings.HasPrefix(contentType, "application/json"):
		if r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				return payload, nil, fmt.Errorf("invalid json")
			}
		}
	case contentType == "":
		// Query-only create requests are valid.
	default:
		return payload, nil, fmt.Errorf("unsupported content type %q", r.Header.Get("Content-Type"))
	}

	if err := applyTodoNewValues(&payload, r.URL.Query(), false); err != nil {
		return payload, nil, err
	}
	return payload, attachments, nil
}

func parseTodoUpdatePayload(r *http.Request) (todoUpdatePayload, []todoAttachmentSummary, error) {
	var payload todoUpdatePayload
	var attachments []todoAttachmentSummary
	contentType := strings.ToLower(r.Header.Get("Content-Type"))

	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"):
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return payload, nil, fmt.Errorf("invalid multipart form: %w", err)
		}
		if r.MultipartForm != nil {
			applyTodoUpdateValues(&payload, r.MultipartForm.Value)
			stored, err := persistMultipartAttachments(r.MultipartForm)
			if err != nil {
				return payload, nil, err
			}
			attachments = stored
		}
	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		if err := r.ParseForm(); err != nil {
			return payload, nil, fmt.Errorf("invalid form: %w", err)
		}
		applyTodoUpdateValues(&payload, r.PostForm)
	case strings.HasPrefix(contentType, "application/json"), contentType == "":
		// A body with no explicit Content-Type is treated as JSON — many clients
		// (and httptest.NewRequest) omit the header. A bodyless request decodes
		// nothing and falls through to the later "operation is required"
		// validation, so query-only PATCHes still work.
		if r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				return payload, nil, fmt.Errorf("invalid json")
			}
		}
	default:
		return payload, nil, fmt.Errorf("unsupported content type %q", r.Header.Get("Content-Type"))
	}
	return payload, attachments, nil
}

func applyTodoUpdateValues(payload *todoUpdatePayload, values map[string][]string) {
	assignString := func(target *string, keys ...string) {
		if value, ok := firstTodoUpdateValue(values, keys...); ok {
			*target = strings.TrimSpace(value)
		}
	}
	assignPointer := func(target **string, trim bool, keys ...string) {
		if value, ok := firstTodoUpdateValue(values, keys...); ok {
			if trim {
				value = strings.TrimSpace(value)
			}
			*target = &value
		}
	}

	assignString(&payload.Dir, "dir")
	assignString(&payload.Ref, "ref")
	if value, ok := firstTodoUpdateValue(values, "status"); ok {
		payload.Status = types.Status(strings.TrimSpace(value))
	}
	if value, ok := firstTodoUpdateValue(values, "priority", "severity"); ok {
		payload.Priority = types.Priority(strings.TrimSpace(value))
	}
	assignPointer(&payload.Title, true, "title", "name")
	assignPointer(&payload.Body, false, "body", "description", "text")
	assignString(&payload.Comment, "comment")
}

func firstTodoUpdateValue(values map[string][]string, keys ...string) (string, bool) {
	for _, key := range keys {
		if vals, ok := values[key]; ok {
			if len(vals) == 0 {
				return "", true
			}
			return vals[0], true
		}
	}
	return "", false
}

func applyTodoNewValues(payload *todoNewPayload, values map[string][]string, overwrite bool) error {
	assignString := func(target *string, keys ...string) {
		if !overwrite && strings.TrimSpace(*target) != "" {
			return
		}
		if value := firstTodoNewValue(values, keys...); value != "" {
			*target = value
		}
	}
	assignPriority := func(target *types.Priority, keys ...string) {
		if !overwrite && *target != "" {
			return
		}
		if value := firstTodoNewValue(values, keys...); value != "" {
			*target = types.Priority(value)
		}
	}
	assignStatus := func(target *types.Status, keys ...string) {
		if !overwrite && *target != "" {
			return
		}
		if value := firstTodoNewValue(values, keys...); value != "" {
			*target = types.Status(value)
		}
	}

	assignString(&payload.Dir, "dir")
	assignString(&payload.Title, "title", "name")
	assignString(&payload.Body, "body", "description", "text")
	assignPriority(&payload.Priority, "priority", "severity")
	assignStatus(&payload.Status, "status")
	if !overwrite && payload.AutoSave != nil {
		return nil
	}
	if raw := firstTodoNewValue(values, "autoSave", "autosave", "auto_save"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("invalid autoSave %q", raw)
		}
		payload.AutoSave = &parsed
	}
	return nil
}

func firstTodoNewValue(values map[string][]string, keys ...string) string {
	for _, key := range keys {
		for _, value := range values[key] {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

// normalizeTodoRunRefs collects the todo refs to run, de-duplicated and in order:
// the explicit refs[] (multi-select), then the single ref, then the ?ref query
// param. Multiple refs run together in a single agent session.
func normalizeTodoRunRefs(payload todoRunPayload, r *http.Request) []string {
	seen := make(map[string]bool)
	var refs []string
	add := func(ref string) {
		ref = strings.TrimSpace(ref)
		if ref == "" || seen[ref] {
			return
		}
		seen[ref] = true
		refs = append(refs, ref)
	}
	for _, ref := range payload.Refs {
		add(ref)
	}
	add(payload.Ref)
	if len(refs) == 0 {
		add(r.URL.Query().Get("ref"))
	}
	return refs
}

// resolution fold, plus the sibling flags that are deliberately not spec fields.
type todoRunWire struct {
	// Override is the payload's spec as the HIGHEST resolution layer — a knob the
	// dashboard did not send must stay zero here so .gavel.yaml and the mode's
	// .prompt frontmatter can supply it.
	Override api.Spec
	// Driver is the requested mechanism, empty when the payload named none and
	// .gavel.yaml todos.driver should decide.
	Driver string
	// RunMode is empty when the client named a prompt but asserted no class, so
	// the prompt's own declared class decides.
	RunMode types.RunMode
	// Prompt is the requested prompt name, empty when the client named none.
	Prompt string
	Resume bool
	// Force is the client's answer to "this todo already has a live run": run
	// alongside it. It is not a spec knob — it decides admission, not behaviour.
	Force bool
}

// validateTodoRunPayload rejects a malformed payload at the wire boundary,
// before anything touches the database. It validates only what the client sent;
// the resolved spec — which also carries the .gavel.yaml layers — is validated
// again by todos/spec.Resolve.
//
// Nothing is defaulted here. Defaults belong to the resolution seam, and a
// default applied at this layer would outrank the configuration it claims to
// defer to.
func validateTodoRunPayload(payload todoRunPayload) (todoRunWire, error) {
	driver, err := payloadDriver(payload)
	if err != nil {
		return todoRunWire{}, err
	}
	// RunMode selects the built-in prompt; the legacy plan bool maps to plan
	// until the dashboard sends runMode. Plan works on every driver now (the
	// plan template's frontmatter carries the plan permission posture).
	runModeValue := payload.RunMode
	if runModeValue == "" && payload.Plan {
		runModeValue = string(types.ModePlan)
	}
	// A named prompt declares its own class. Only a class the client actually
	// asserted is forwarded, so todos/spec can check the two for agreement
	// instead of a defaulted "run" contradicting every non-run prompt.
	promptName := strings.TrimSpace(payload.Prompt)
	var runMode types.RunMode
	if runModeValue != "" || promptName == "" {
		parsed, err := types.ParseRunMode(runModeValue)
		if err != nil {
			return todoRunWire{}, err
		}
		runMode = parsed
	}

	spec := payload.Spec
	if effort := strings.ToLower(strings.TrimSpace(string(spec.Effort))); effort != "" {
		if !validTodoRunEffort(effort) {
			return todoRunWire{}, fmt.Errorf("invalid effort %q", payload.Spec.Effort)
		}
		spec.Effort = api.Effort(effort)
	}
	if raw := strings.TrimSpace(spec.Budget.Timeout); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return todoRunWire{}, fmt.Errorf("invalid timeout %q", payload.Spec.Budget.Timeout)
		}
		if parsed <= 0 {
			return todoRunWire{}, fmt.Errorf("timeout must be greater than zero")
		}
		spec.Budget.Timeout = parsed.String()
	}
	if err := spec.Budget.Validate(); err != nil {
		return todoRunWire{}, err
	}
	if err := spec.Permissions.Validate(); err != nil {
		return todoRunWire{}, err
	}
	// Validate the run's Workflow (verify scope/maxIterations); nil is allowed.
	if err := spec.Workflow.Validate(); err != nil {
		return todoRunWire{}, fmt.Errorf("workflow: %w", err)
	}

	return todoRunWire{
		Override: spec, Driver: driver, RunMode: runMode, Prompt: promptName,
		Resume: payload.Resume, Force: payload.Force,
	}, nil
}

// normalizeTodoRunOptions resolves the run a payload asks for. The payload is
// the top layer of todos/spec.Resolve, not the base: a dashboard run therefore
// picks up `.gavel.yaml` ai:/todos: exactly as `gavel todos run` does, instead
// of silently ignoring project configuration.
//
// dir is the workspace the .gavel.yaml layers are read from, and todoList
// contributes each todo's `llm:` frontmatter — so both must be resolved before
// this is called.
func normalizeTodoRunOptions(dir string, todoList []*types.TODO, payload todoRunPayload) (todoRunOptions, error) {
	wire, err := validateTodoRunPayload(payload)
	if err != nil {
		return todoRunOptions{}, err
	}
	resolved, err := todospec.Resolve(todospec.Input{
		WorkDir:  dir,
		Mode:     wire.RunMode,
		Prompt:   wire.Prompt,
		Todos:    todoList,
		Override: wire.Override,
		Driver:   wire.Driver,
		// The dashboard serves /api/todos/session/approve, so a run that asks for
		// a Bash approval has something to answer it.
		CanApprove: true,
	})
	if err != nil {
		return todoRunOptions{}, err
	}

	// Mode selection is the dashboard's own concern: it offers a (driver,
	// mode) catalog the CLI has no equivalent of, and it runs against the
	// resolved model rather than the payload's, so a model that arrived from
	// .gavel.yaml gets the same treatment as one the dialog picked.
	spec := resolved.Spec
	mode, model, err := resolveTodoRunRuntime(resolved.Driver, string(spec.Mode), spec.Name)
	if err != nil {
		return todoRunOptions{}, err
	}
	spec.Name = model
	spec.Mode = api.RuntimeMode(mode)

	return todoRunOptions{
		Spec:       spec,
		Driver:     string(resolved.Driver),
		RunMode:    resolved.Mode,
		Prompt:     resolved.Prompt,
		Envelope:   resolved.Envelope,
		Resume:     wire.Resume,
		Concurrent: wire.Force,
		Template:   resolved.Template,
		Approvals:  resolved.Approvals,
		Timeout:    resolved.Timeout,
	}, nil
}

// payloadDriver validates the canonical mechanism-only driver. Provider and
// model identity live in Spec.Model and are resolved by Captain.
func payloadDriver(p todoRunPayload) (string, error) {
	if strings.TrimSpace(p.Agent) != "" || strings.TrimSpace(p.Mode) != "" {
		return "", fmt.Errorf("invalid run configuration: agent and mode are not supported; use driver api, agent, cli, or cmux")
	}
	s := strings.TrimSpace(p.Driver)
	if s == "" {
		return "", nil
	}
	kind, err := drivers.Parse(s)
	if err != nil {
		return "", err
	}
	return string(kind), nil
}

// backlog that cannot be listed degrades duplicate detection but not the other
// four verdicts, so it is logged rather than fatal.
// has one and otherwise lets claude manage its own id.

func countProjectTodos(ctx context.Context, project Project) (todoCounts, error) {
	if project.ResolvedDir() == "" {
		return todoCounts{}, nil
	}
	nativeProvider, err := ProviderForProject(ctx, project)
	if err != nil {
		return todoCounts{}, err
	}
	byStatus, err := nativeProvider.CountByStatus(ctx)
	if err != nil {
		return todoCounts{}, err
	}
	var counts todoCounts
	for status, n := range byStatus {
		addTodoStatus(&counts, status, n)
	}
	return counts, nil
}
