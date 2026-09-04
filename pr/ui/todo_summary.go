package ui

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
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
	// Lifecycle is the todo's step state as the lifecycle definition sees it:
	// which steps apply, which one comes next, and why. It is populated on detail
	// responses only, and computed server-side — the definition is data the
	// browser has no copy of, so a client re-deriving it would be guessing.
	Lifecycle *todoLifecycleState `json:"lifecycle,omitempty"`
	// Envelope-driven fields from the last agent run: the native plan file the
	// agent reported (rendered by the Plan tab / review mode), its status, the
	// final summary, and the questions blocking an ask todo. The step an answer
	// resumes is not among them: Phases marks it, and `runMode` is a key the run
	// endpoint rejects on input, so it is not one a response may hand out.
	PlanPath       string                `json:"planPath,omitempty"`
	PlanStatus     types.PlanStatus      `json:"planStatus,omitempty"`
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
