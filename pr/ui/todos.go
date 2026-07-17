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

	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/commit"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/prwatch"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/claude"
	"github.com/flanksource/gavel/todos/drivers"
	"github.com/flanksource/gavel/todos/native"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
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
	// VerificationMarkdown is the raw fixture markdown inside the todo's
	// "## Verification" section (excluding the heading); populated on detail
	// responses so the Verification tab's FixtureEditor can be seeded with it.
	VerificationMarkdown string `json:"verificationMarkdown,omitempty"`
	// Diff is the aggregated git diff footprint of the todo's commits (those
	// carrying its Gavel-Issue-Id trailer); nil when no commits reference it.
	Diff *todoDiffStat `json:"diff,omitempty"`
	// HasPlan/HasVerification are lightweight availability flags for the list
	// row's indicators — unlike PlanPath/VerificationMarkdown they're populated
	// on both list and detail responses since they cost no extra parsing.
	HasPlan         bool `json:"hasPlan,omitempty"`
	HasVerification bool `json:"hasVerification,omitempty"`
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
	Dir   string   `json:"dir,omitempty"`
	Ref   string   `json:"ref,omitempty"`
	Refs  []string `json:"refs,omitempty"`
	Agent string   `json:"agent,omitempty"`
	Mode  string   `json:"mode,omitempty"`
	// Driver selects the agent driver (claude-cmux, claude-headless, claude-sdk,
	// claude-api, codex-cmux, codex-headless). When empty it is derived from the
	// legacy agent+mode pair for backward compatibility.
	Driver string `json:"driver,omitempty"`

	// Spec inlines the model/backend/effort/prompt/budget/permissions/session
	// knobs — the same request shape captain's prompt run editor produces
	// (model, backend, effort, prompt.user, budget.{cost,maxTurns,timeout},
	// permissions.{mode,tools}, sessionId, ...).
	api.Spec

	// RunMode selects the built-in prompt: run (implement) or plan (read-only
	// investigation producing a reviewable plan). It supersedes the legacy Plan
	// bool, which is still accepted until the dashboard sends runMode.
	RunMode string `json:"runMode,omitempty"`
	Plan    bool   `json:"plan,omitempty"`
	// Resume continues the todo's prior agent session (claude --resume) instead of
	// starting fresh. It stays a sibling flag rather than a Spec field because it is
	// a session-identity decision: a fresh run also carries a (minted) sessionId, so
	// resume cannot be inferred from Spec.SessionID.
	Resume bool `json:"resume,omitempty"`
	// Dirty/DryRun/Commit/Check are no longer sibling flags — the run's dirty-tree
	// stash, dry-run, auto-commit, and checks all come from the embedded api.Spec
	// (Setup.Checkout.Dirty and Workflow.Verify/PostRun), surfaced by clicky's
	// Workspace/Verify/Commit sections.
}

type todoRunResponse struct {
	Status    string   `json:"status"`
	Ref       string   `json:"ref"`
	Refs      []string `json:"refs,omitempty"`
	Count     int      `json:"count"`
	Dir       string   `json:"dir"`
	Agent     string   `json:"agent"`
	Mode      string   `json:"mode"`
	Model     string   `json:"model,omitempty"`
	Backend   string   `json:"backend,omitempty"`
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
	Prompt  string `json:"prompt"`
	Mode    string `json:"mode"`
	Agent   string `json:"agent"`
	Backend string `json:"backend,omitempty"`
	Effort  string `json:"effort,omitempty"`
	RunMode string `json:"runMode,omitempty"`
	Plan    bool   `json:"plan,omitempty"`
	Count   int    `json:"count"`
}

type todoRunRequest struct {
	Provider todos.Provider
	Registry *todoRunRegistry
	// Todos are executed together in a single agent session (multi-select run);
	// a single-element slice is the ordinary one-todo run.
	Todos   []*types.TODO
	Source  todoSource
	Backend string
	Options todoRunOptions
}

type todoRunOptions struct {
	// Spec carries the model/backend/effort/budget/prompt/permissions/session knobs
	// plus the run's Setup (dirty/checkout) and Workflow (verify/postRun), resolved
	// and validated by normalizeTodoRunOptions. dirty/checks/commit/dryRun are read
	// from Spec.Setup/Spec.Workflow (see specDirty/specVerify/specCommit/specDryRun),
	// not sibling flags.
	api.Spec
	Agent   string
	Mode    string
	Driver  string
	RunMode types.RunMode
	Resume  bool
}

// specCommit reports whether the run's Workflow.PostRun asks for an auto-commit.
// Absent Workflow/PostRun means no commit; the run dialog seeds postRun.commit=true
// on the Commit section so a default dashboard run still auto-commits.
func specCommit(spec api.Spec) bool {
	return spec.Workflow != nil && spec.Workflow.PostRun != nil && spec.Workflow.PostRun.Commit
}

// specDryRun reports whether the run is a dry run (Workflow.PostRun.DryRun): the
// agent runs normally but the post-run auto-commit is suppressed.
func specDryRun(spec api.Spec) bool {
	return spec.Workflow != nil && spec.Workflow.PostRun != nil && spec.Workflow.PostRun.DryRun
}

// specVerify returns the run's Workflow.Verify (nil when unset). Its presence
// force-enables the checks suite and its MaxIterations caps the verify loop.
func specVerify(spec api.Spec) *api.Verify {
	if spec.Workflow == nil {
		return nil
	}
	return spec.Workflow.Verify
}

// specDirty reports whether the run's Setup.Checkout carries a dirty-tree stash
// config (surfaced by the Workspace section).
func specDirty(spec api.Spec) bool {
	return spec.Setup != nil && spec.Setup.Checkout != nil && spec.Setup.Checkout.Dirty != nil
}

// timeout parses the run's budget timeout, defaulting to 30 minutes when
// unset. normalizeTodoRunOptions always writes a canonical time.Duration
// string here, so a parse failure indicates a construction bug rather than
// bad user input.
func (o todoRunOptions) timeout() time.Duration {
	if o.Budget.Timeout == "" {
		return 30 * time.Minute
	}
	d, err := time.ParseDuration(o.Budget.Timeout)
	if err != nil {
		panic(fmt.Sprintf("todoRunOptions: invalid budget timeout %q (should have been validated by normalizeTodoRunOptions): %v", o.Budget.Timeout, err))
	}
	return d
}

var startTodoRun = defaultStartTodoRun

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
	if payload.Status != "" && !validTodoStatus(payload.Status) {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid status %q", payload.Status))
		return
	}
	if payload.Priority != "" && !validTodoPriority(payload.Priority) {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid priority %q", payload.Priority))
		return
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
		if !validTodoStatus(payload.Status) {
			writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid status %q", payload.Status))
			return
		}
		update.Status = &payload.Status
	}
	if payload.Priority != "" {
		if !validTodoPriority(payload.Priority) {
			writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid priority %q", payload.Priority))
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
	comment := strings.TrimSpace(payload.Comment)
	if len(attachments) > 0 {
		comment = todoBodyWithAttachments(comment, attachments)
	}

	hasState := update.Status != nil || update.Priority != nil
	if !hasState && edit.IsEmpty() && comment == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("status, priority, title, body, or comment is required"))
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
	if len(refs) == 0 {
		return nil, todoSource{}, nil, todoRunOptions{}, http.StatusBadRequest, fmt.Errorf("ref is required")
	}
	opts, err := normalizeTodoRunOptions(payload)
	if err != nil {
		return nil, todoSource{}, nil, todoRunOptions{}, http.StatusBadRequest, err
	}
	source := todoSourceFromRequest(r)
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	globalLookup := strings.TrimSpace(source.Dir) == ""
	if err := validateTodoRunCardinality(len(refs)); err != nil {
		return nil, source, nil, opts, http.StatusBadRequest, err
	}
	origin := requestOrigin(r)
	if globalLookup {
		provider, owningSource, todo, err := s.resolveTodoReference(r.Context(), source, refs[0])
		if err != nil {
			return provider, owningSource, nil, opts, http.StatusNotFound, err
		}
		todo.MarkdownBody = absolutizeAttachmentURLs(todo.MarkdownBody, origin)
		return provider, owningSource, []*types.TODO{todo}, opts, http.StatusOK, nil
	}

	provider, source, err := s.todoProviderContext(r.Context(), source)
	if err != nil {
		return nil, source, nil, opts, http.StatusBadRequest, err
	}
	todoList := make([]*types.TODO, 0, len(refs))
	for _, ref := range refs {
		todo, err := provider.Get(r.Context(), ref)
		if err != nil {
			return provider, source, nil, opts, http.StatusNotFound, err
		}
		todo.MarkdownBody = absolutizeAttachmentURLs(todo.MarkdownBody, origin)
		todoList = append(todoList, todo)
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
		Registry: &s.todoRuns,
		Todos:    todoList,
		Source:   source,
		Backend:  backend,
		Options:  opts,
	}
	if _, _, err := newTodoRunExecutorContext(r.Context(), req); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	resp := todoRunResponse{
		Status:    "started",
		Ref:       todos.TODOReference(todoList[0]),
		Refs:      todoRunRefs(todoList),
		Count:     len(todoList),
		Dir:       source.Dir,
		Agent:     opts.Agent,
		Mode:      opts.Mode,
		Driver:    opts.Driver,
		Backend:   string(opts.Spec.Backend),
		Model:     opts.Spec.Name,
		Effort:    string(opts.Spec.Effort),
		RunMode:   string(opts.RunMode),
		Plan:      opts.RunMode == types.ModePlan,
		Resume:    opts.Resume,
		SessionID: opts.Spec.SessionID,
		Timeout:   opts.timeout().String(),
		MaxBudget: opts.Budget.Cost,
		MaxTurns:  opts.Budget.MaxTurns,
		Commit:    specCommit(opts.Spec) && !specDryRun(opts.Spec),
		Message:   todoRunStartedMessage(len(todoList)),
	}
	// A dry run (Workflow.PostRun.DryRun) still executes the agent — it only skips
	// the post-run commit (see maybeCommitAfterRun). "Show what would run without
	// running" is covered by the live prompt preview (handleTodoRunPreview).
	if err := startTodoRun(req); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
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
	previewPrompt, err := buildTodoRunPromptPreview(r.Context(), provider, source.Dir, todoList, opts)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	resp := todoRunPreviewResponse{
		Prompt:  previewPrompt,
		Mode:    opts.Mode,
		Agent:   opts.Agent,
		Backend: string(opts.Spec.Backend),
		Effort:  string(opts.Spec.Effort),
		RunMode: string(opts.RunMode),
		Plan:    opts.RunMode == types.ModePlan,
		Count:   len(todoList),
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// buildTodoRunPromptPreview renders the exact prompt a run would dispatch:
// framing, TODO sections, effort directive, and the mode's structured-output
// schema instruction. The editable prompt override replaces the body but the
// schema instruction is always re-appended, so the preview shows the full
// contract the agent receives.
func buildTodoRunPromptPreview(ctx context.Context, provider todos.Provider, dir string, todoList []*types.TODO, opts todoRunOptions) (string, error) {
	mode := opts.RunMode
	if mode == "" {
		mode = types.ModeRun
	}
	tmpl, err := todoprompt.ResolveTemplate(dir, mode)
	if err != nil {
		return "", err
	}
	existingPlan, err := todoPlanMarkdown(ctx, provider, todoList, mode)
	if err != nil {
		return "", err
	}
	req, _, err := todoprompt.Render(todoList, todoprompt.Options{
		WorkDir: dir, Mode: mode, Effort: string(opts.Spec.Effort), Template: tmpl, ExistingPlan: existingPlan,
	})
	if err != nil {
		return "", err
	}
	return req.Prompt.User, nil
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
	return openTodoProvider(ctx, p.ResolvedDir())
}

// openTodoProvider is the single API/UI native runtime seam. It is a variable so
// package tests can inject an in-memory implementation without opening PostgreSQL.
var openTodoProvider = func(ctx context.Context, dir string) (todos.Provider, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return todoruntime.Open(ctx, dir)
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
	if !validTodoStatus(status) {
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
	}
	if todo.LLM != nil {
		out.SessionID = todo.LLM.SessionId
	}
	if out.Ref == "" {
		out.Ref = todo.FilePath
	}
	out.HasPlan = todos.HasPlan(todo)
	out.HasVerification = todos.ExtractVerificationFixture(todo.MarkdownBody) != ""
	if detail {
		out.Body = strings.TrimSpace(todo.MarkdownBody)
		out.Implementation = strings.TrimSpace(todo.Implementation)
		out.Events = todo.ProviderEvents
		out.Criteria = todo.AcceptanceCriteria
		out.VerificationMarkdown = todos.ExtractVerificationFixture(todo.MarkdownBody)
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
		counts.Total++
		switch item.Status {
		case types.StatusCompleted:
			counts.Completed++
		case types.StatusDraft:
			counts.Open++
			counts.Draft++
		case types.StatusInProgress:
			counts.Open++
			counts.InProgress++
		case types.StatusReview:
			counts.Open++
			counts.Review++
		case types.StatusAsk:
			counts.Open++
			counts.Ask++
		case types.StatusFailed:
			counts.Open++
			counts.Failed++
		case types.StatusVerified:
			counts.Open++
			counts.Verified++
		case types.StatusSkipped:
			counts.Open++
			counts.Skipped++
		default:
			counts.Open++
			counts.Pending++
		}
	}
	return counts
}

func validTodoStatus(status types.Status) bool {
	return types.IsKnownStatus(status)
}

func validTodoPriority(priority types.Priority) bool {
	switch priority {
	case types.PriorityHigh, types.PriorityMedium, types.PriorityLow:
		return true
	default:
		return false
	}
}

func writeTodoError(w http.ResponseWriter, status int, err error) {
	switch {
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

func todoRunRefs(todoList []*types.TODO) []string {
	refs := make([]string, len(todoList))
	for i, todo := range todoList {
		refs[i] = todos.TODOReference(todo)
	}
	return refs
}

func todoRunStartedMessage(count int) string {
	if count > 1 {
		return fmt.Sprintf("Started run for %d todos", count)
	}
	return "Todo run started"
}

func todoRunLabel(todoList []*types.TODO) string {
	if len(todoList) == 1 {
		return todos.TODOReference(todoList[0])
	}
	return fmt.Sprintf("%d todos", len(todoList))
}

func normalizeTodoRunOptions(payload todoRunPayload) (todoRunOptions, error) {
	kind, err := resolveDriverFromPayload(payload)
	if err != nil {
		return todoRunOptions{}, err
	}
	backend, model, err := resolveTodoRunBackendModel(kind, string(payload.Backend), payload.Name)
	if err != nil {
		return todoRunOptions{}, err
	}
	// The coding agent is derived from the resolved model, not the driver.
	agent, _ := claude.ResolveAgent(model)

	// Mode is the legacy mechanism label (cmux vs inline) some response/preview
	// paths still read; the authoritative selection is Driver.
	mode := "inline"
	if kind == drivers.Cmux {
		mode = "cmux"
	}

	// RunMode selects the built-in prompt; the legacy plan bool maps to plan
	// until the dashboard sends runMode. Plan works on every driver now (the
	// plan template's frontmatter carries the plan permission posture).
	runModeValue := payload.RunMode
	if runModeValue == "" && payload.Plan {
		runModeValue = string(types.ModePlan)
	}
	runMode, err := types.ParseRunMode(runModeValue)
	if err != nil {
		return todoRunOptions{}, err
	}
	effort := api.Effort(strings.ToLower(strings.TrimSpace(string(payload.Effort))))
	if effort == "" {
		effort = "medium"
	}
	if !validTodoRunEffort(string(effort)) {
		return todoRunOptions{}, fmt.Errorf("invalid effort %q", payload.Effort)
	}

	timeout := 30 * time.Minute
	if raw := strings.TrimSpace(payload.Budget.Timeout); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return todoRunOptions{}, fmt.Errorf("invalid timeout %q", payload.Budget.Timeout)
		}
		if parsed <= 0 {
			return todoRunOptions{}, fmt.Errorf("timeout must be greater than zero")
		}
		timeout = parsed
	}

	// The rendered .prompt template's request is not built here — normalize only
	// resolves and validates the run's Spec overlay (model/backend/effort/budget
	// plus any prompt/permissions overrides the dashboard sent); newTodoRunExecutor
	// and buildRequest overlay it onto the template frontmatter's defaults.
	spec := payload.Spec
	spec.Name = model
	spec.Backend = api.Backend(backend)
	spec.Effort = effort
	spec.Budget.Timeout = timeout.String()
	if err := spec.Budget.Validate(); err != nil {
		return todoRunOptions{}, err
	}
	if err := spec.Permissions.Validate(); err != nil {
		return todoRunOptions{}, err
	}
	// Permissions.Validate reaches Tools.Modes only via Tools.Policies(), whose
	// enabled/ask/disabled switch has no default case — an unrecognized mode is
	// silently dropped from the map rather than surfaced as invalid. Check the
	// raw modes directly so a mistyped value fails loud at this wire boundary
	// instead of being dropped.
	for tool, mode := range spec.Permissions.Tools.Modes {
		if !mode.Valid() {
			return todoRunOptions{}, fmt.Errorf("invalid tool mode %q for tool %q", mode, tool)
		}
	}
	// Validate the run's Workflow (verify scope/maxIterations); nil is allowed.
	if err := spec.Workflow.Validate(); err != nil {
		return todoRunOptions{}, fmt.Errorf("workflow: %w", err)
	}

	return todoRunOptions{
		Spec:    spec,
		Agent:   agent,
		Mode:    mode,
		Driver:  string(kind),
		RunMode: runMode,
		Resume:  payload.Resume,
	}, nil
}

func defaultStartTodoRun(req todoRunRequest) error {
	if req.Registry == nil {
		return errors.New("todo run registry is required")
	}
	executor, sessionID, err := newTodoRunExecutor(req)
	if err != nil {
		return err
	}
	ctx, timeoutCancel := context.WithTimeout(context.Background(), req.Options.timeout())
	runCtx, stop := context.WithCancelCause(ctx)
	cleanup, err := req.Registry.register(todoRunIssueIDs(req.Todos), strings.HasPrefix(executor.Name(), "headless-"), stop)
	if err != nil {
		timeoutCancel()
		return err
	}
	go func() {
		defer timeoutCancel()
		defer cleanup()

		execCtx := todos.NewExecutorContext(runCtx, logger.StandardLogger(), nil)
		runner := todos.NewTODOExecutor(req.Source.Dir, executor, sessionID, req.Provider)
		runner.SetMode(req.Options.RunMode)
		runner.SetResume(req.Options.Resume)
		var runErr error
		var result *todos.ExecutionResult
		// A single selection runs through Execute; a multi-select runs every todo
		// in one combined agent session via ExecuteGroup.
		if len(req.Todos) == 1 {
			result, runErr = runner.Execute(execCtx, req.Todos[0])
		} else {
			var results []*todos.ExecutionResult
			results, runErr = runner.ExecuteGroup(execCtx, req.Todos)
			if len(results) > 0 {
				result = results[0]
			}
		}
		if runErr != nil && (result == nil || !result.Cancelled) {
			logger.Warnf("todo run %s failed: %v", todoRunLabel(req.Todos), runErr)
		}
		maybeCommitAfterRun(req, result)
	}()
	return nil
}

// maybeCommitAfterRun runs the `gavel commit` pipeline over the agent's changes
// once a dashboard run finishes, mirroring the CLI's `todos run --commit`, then
// runs issue verification over the resulting commits when enabled. Auto-commit
// is skipped for plan-only runs (which make no changes) and whenever the
// executor already committed its own changes (see todos.ShouldCommitAfter).
func maybeCommitAfterRun(req todoRunRequest, result *todos.ExecutionResult) {
	if len(req.Todos) == 0 || req.Todos[0] == nil {
		return
	}
	todo := req.Todos[0]

	// An ask outcome has half-done work by design: committing it is a decision
	// for the answer turn, not the run tail.
	askEnded := result != nil && result.EndStatus == types.EndAsk
	// A dry run executes but suppresses the auto-commit (api.PostRun.DryRun).
	commitEnabled := specCommit(req.Options.Spec) && !specDryRun(req.Options.Spec)
	if enabled := commitEnabled && req.Options.RunMode != types.ModePlan && !askEnded; todos.ShouldCommitAfter(result, enabled) {
		meta := commit.AgentRunMetadata{IssueID: todo.ID}
		if todo.LLM != nil {
			meta.SessionID = todo.LLM.SessionId
		}
		if _, err := commit.RunAfterAgent(context.Background(), req.Source.Dir, todo.CWD, meta); err != nil {
			logger.Errorf("commit after todo run failed: %v", err)
		}
	}
	// Verification is the run loop's definition of done (the fixture/CEL verdict),
	// not a separate post-commit scoring pass. The Verification tab and
	// `gavel todos check` re-run it manually.
}

// resolveDriverFromPayload selects the driver mechanism: the explicit Driver
// field when set, otherwise the legacy agent+mode pair (mode cmux → cmux, mode
// inline → sdk; codex was never an inline agent). The driver is mechanism-only;
// the coding agent is derived from the model downstream.
func resolveDriverFromPayload(p todoRunPayload) (drivers.Kind, error) {
	if s := strings.TrimSpace(p.Driver); s != "" {
		return drivers.Parse(s)
	}
	agent := strings.ToLower(strings.TrimSpace(p.Agent))
	if agent == "" {
		agent, _ = claude.ResolveAgent(strings.TrimSpace(p.Name))
	}
	mode := strings.ToLower(strings.TrimSpace(p.Mode))
	if mode == "" {
		mode = "cmux"
	}
	switch mode {
	case "cmux":
		return drivers.Cmux, nil
	case "inline":
		if agent == "codex" {
			return "", fmt.Errorf("codex runs require a cmux driver")
		}
		return drivers.Sdk, nil
	default:
		return "", fmt.Errorf("invalid mode %q", p.Mode)
	}
}

func newTodoRunExecutor(req todoRunRequest) (todos.Executor, string, error) {
	return newTodoRunExecutorContext(context.Background(), req)
}

func newTodoRunExecutorContext(ctx context.Context, req todoRunRequest) (todos.Executor, string, error) {
	kind, err := drivers.Parse(req.Options.Driver)
	if err != nil {
		return nil, "", err
	}
	// cmux returns "" as the orchestrator session id (it manages its own
	// --session-id, passed via SessionID) so TODOExecutor does not overwrite the
	// todo's recorded prior session.
	mode := req.Options.RunMode
	if mode == "" {
		mode = types.ModeRun
	}
	// Post-run checks run inside the agent loop as fixture-backed verify
	// plugins; a failing round's feedback re-runs the same session.
	var verifiers []agent.Verify
	var maxIterations int
	if mode == types.ModeRun {
		var err error
		verifiers, maxIterations, err = todos.BuildCheckVerifiers(req.Source.Dir, req.Todos, specVerify(req.Options.Spec))
		if err != nil {
			return nil, "", err
		}
	}
	// The todo's recorded plan feeds both flows: a plan re-run reports
	// updated/unchanged, and an implement run follows the approved/edited plan.
	// Single-todo only — a group run has no single plan to attribute.
	existingPlan, err := todoPlanMarkdown(ctx, req.Provider, req.Todos, mode)
	if err != nil {
		return nil, "", err
	}
	return drivers.New(kind, drivers.Config{
		WorkDir:        req.Source.Dir,
		Model:          req.Options.Spec.Name,
		Backend:        string(req.Options.Spec.Backend),
		Fallbacks:      req.Options.Spec.Model.Fallbacks,
		Effort:         string(req.Options.Spec.Effort),
		Mode:           mode,
		ExistingPlan:   existingPlan,
		Verifiers:      verifiers,
		MaxIterations:  maxIterations,
		Resume:         req.Options.Resume,
		SessionID:      req.Options.Spec.SessionID,
		Timeout:        req.Options.timeout(),
		MaxBudgetUsd:   req.Options.Budget.Cost,
		MaxTurns:       req.Options.Budget.MaxTurns,
		Dirty:          specDirty(req.Options.Spec),
		PromptOverride: req.Options.Prompt.User,
		ToolModes:      toolModesFromPermissions(req.Options.Permissions.Tools),
		PermissionMode: string(req.Options.Permissions.Mode),
		// The dashboard is the approval resolver (handleTodoSessionApprove), so
		// brokered tool permissions can be surfaced and answered here.
		Approvals: true,
	})
}

func todoPlanMarkdown(ctx context.Context, provider todos.Provider, todoList []*types.TODO, mode types.RunMode) (string, error) {
	if len(todoList) != 1 || (mode != types.ModePlan && mode != types.ModeRun) {
		return "", nil
	}
	content, ok := provider.(todos.PlanContentProvider)
	if !ok {
		return "", fmt.Errorf("PostgreSQL TODO runtime does not support durable plan content")
	}
	return content.PlanMarkdown(ctx, todoList[0], mode)
}

// toolModesFromPermissions flattens api.Tools (the Spec's allow/deny/modes
// policy) into the legacy tool-name→mode map (enabled/ask/disabled) that
// headless.Config/drivers.Config key their allow/deny/broker logic on.
func toolModesFromPermissions(tools api.Tools) map[string]string {
	if len(tools.Allow) == 0 && len(tools.Deny) == 0 && len(tools.Modes) == 0 {
		return nil
	}
	out := make(map[string]string, len(tools.Allow)+len(tools.Deny)+len(tools.Modes))
	for _, tool := range tools.Allow {
		out[tool] = "enabled"
	}
	for _, tool := range tools.Deny {
		out[tool] = "disabled"
	}
	for tool, mode := range tools.Modes {
		switch mode {
		case api.ToolModeOn:
			out[tool] = "enabled"
		case api.ToolModeAsk:
			out[tool] = "ask"
		case api.ToolModeOff:
			out[tool] = "disabled"
		case api.ToolModeAuto:
			delete(out, tool)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// resolveRunSessionID determines the claude session id a run will use, so the
// caller knows it up front. A resume run reuses the todo's prior session; a
// fresh cmux run mints a new id (claude is launched with it, so the dashboard
// can follow the log immediately); inline resumes a single todo's session if it
// has one and otherwise lets claude manage its own id.
func resolveRunSessionID(opts todoRunOptions, todoList []*types.TODO) string {
	if opts.Resume {
		if sid := firstTodoSessionID(todoList); sid != "" {
			return sid
		}
	}
	switch opts.Mode {
	case "cmux":
		return uuid.NewString()
	case "inline":
		if len(todoList) == 1 {
			return firstTodoSessionID(todoList)
		}
	}
	return ""
}

func firstTodoSessionID(todoList []*types.TODO) string {
	for _, todo := range todoList {
		if todo != nil && todo.LLM != nil && todo.LLM.SessionId != "" {
			return todo.LLM.SessionId
		}
	}
	return ""
}

func countProjectTodos(ctx context.Context, dir string) (todoCounts, error) {
	if dir == "" {
		return todoCounts{}, nil
	}
	nativeProvider, err := openTodoProvider(ctx, dir)
	if err != nil {
		return todoCounts{}, err
	}
	items, err := nativeProvider.List(ctx, todos.DiscoveryFilters{})
	if err != nil {
		return todoCounts{}, err
	}
	return summarizeTodos(items), nil
}
