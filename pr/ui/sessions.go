package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/git/gitstate"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

const (
	defaultSessionLimit = 60
	maxSessionLimit     = 200
	// sessionSummaryScan is how many of the most recently active non-subagent
	// session rows are read: Captain and gavel admission rows interleave with
	// the agent sessions, so the scan is wider than the page.
	sessionSummaryScan = 400
)

// sessionsResponse is GET /api/sessions. Errors lists the projects whose git
// state or todo links could not be read; their sessions are still listed,
// without the parts that failed.
type sessionsResponse struct {
	Sessions []sessionRow          `json:"sessions"`
	Errors   []sessionProjectError `json:"errors"`
}

type sessionProjectError struct {
	Project string `json:"project"`
	Error   string `json:"error"`
}

// sessionLinkSource is the slice of a native todo provider the Sessions tab
// reads: Captain's sessions and runs, and the todo issues runs are linked to.
type sessionLinkSource interface {
	Captain() *captaindb.DB
	Repository() *native.Repository
	Workspace() *native.Workspace
}

type sessionProjectSources struct {
	project sessionProject
	links   sessionLinkSource
	err     error
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	limit, err := sessionLimit(r.URL.Query().Get("limit"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	projects, err := LoadProjects()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	tracker, ok := s.requestGitTracker(w)
	if !ok {
		return
	}
	sources := loadSessionProjectSources(r.Context(), tracker, projects)
	captain := sessionCaptain(sources)
	if captain == nil {
		respondError(w, http.StatusServiceUnavailable, "captain session store unavailable: no configured project has a native todo provider")
		return
	}
	page, err := captain.ListSessionSummaries(r.Context(), captaindb.SessionListFilter{ExcludeSubagents: true, Limit: sessionSummaryScan})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	overviews := page.Rows
	runs, err := sessionRuns(r.Context(), captain, overviews, sources, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.mu.RLock()
	prs := append([]github.PRListItem(nil), s.prs...)
	s.mu.RUnlock()

	response := sessionsResponse{Errors: []sessionProjectError{}}
	sessionProjects := make([]sessionProject, len(sources))
	for i, source := range sources {
		sessionProjects[i] = source.project
		if source.err != nil {
			response.Errors = append(response.Errors, sessionProjectError{Project: source.project.Project.Name, Error: source.err.Error()})
		}
	}
	response.Sessions = buildSessionRows(overviews, sessionProjects, runs, prs, strings.TrimSpace(r.URL.Query().Get("project")), limit)
	respondJSON(w, http.StatusOK, response)
}

func sessionLimit(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultSessionLimit, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 || limit > maxSessionLimit {
		return 0, fmt.Errorf("limit %q: expected an integer between 1 and %d", raw, maxSessionLimit)
	}
	return limit, nil
}

// loadSessionProjectSources reads every project's stored git state in one
// batch and each project's native todo provider concurrently. A project's
// failure is recorded on its own entry and never fails the batch.
func loadSessionProjectSources(ctx context.Context, tracker *gitstate.Tracker, projects []Project) []sessionProjectSources {
	dirs := make([]string, len(projects))
	for i, project := range projects {
		dirs[i] = project.ResolvedDir()
	}
	states, gitErrs := tracker.States(ctx, dirs)
	sources := make([]sessionProjectSources, len(projects))
	var group errgroup.Group
	group.SetLimit(todoBatchConcurrency)
	for i, project := range projects {
		group.Go(func() error {
			source := sessionProjectSources{project: sessionProject{Project: project}}
			var errs []error
			if err := gitErrs[dirs[i]]; err != nil {
				errs = append(errs, fmt.Errorf("git state: %w", err))
			} else {
				state := states[dirs[i]]
				source.project.Git = &state
			}
			if links, err := sessionLinksFor(ctx, project.ResolvedDir()); err != nil {
				errs = append(errs, fmt.Errorf("todo links: %w", err))
			} else {
				source.links = links
			}
			source.err = errors.Join(errs...)
			sources[i] = source
			return nil
		})
	}
	_ = group.Wait()
	return sources
}

func sessionLinksFor(ctx context.Context, dir string) (sessionLinkSource, error) {
	provider, err := openTodoProvider(ctx, dir)
	if err != nil {
		return nil, err
	}
	links, ok := provider.(sessionLinkSource)
	if !ok || links.Captain() == nil || links.Repository() == nil || links.Workspace() == nil {
		return nil, fmt.Errorf("todo provider for %s is not backed by native storage", dir)
	}
	return links, nil
}

// sessionCaptain is the Captain store every native provider shares (they all
// read through the one process pool), taken from the first that has one.
func sessionCaptain(sources []sessionProjectSources) *captaindb.DB {
	for _, source := range sources {
		if source.links != nil {
			return source.links.Captain()
		}
	}
	return nil
}

// sessionRuns resolves the gavel run each transcript session executed: the
// run whose execution session it is, the todo the run is linked to, and the
// worktree it recorded. Only transcript children of a run can have one, so
// sessions a person drives directly cost no lookup; and only those that can
// reach the page — live or waiting on input, or among the `limit` most recent
// top-level sessions — are looked up at all.
func sessionRuns(ctx context.Context, captain *captaindb.DB, overviews []captaindb.SessionListSummary, sources []sessionProjectSources, limit int) (map[uuid.UUID]sessionRun, error) {
	issueByRun, err := sessionRunIssues(ctx, sources)
	if err != nil {
		return nil, err
	}
	runs := map[uuid.UUID]sessionRun{}
	recent := 0
	for _, overview := range overviews {
		if !isTopLevelAgentSession(overview) {
			continue
		}
		recent++
		urgent := overview.ProcessActive || overview.ActivityState == string(captaindb.SessionActivityAsk) || overview.ActivityState == string(captaindb.SessionActivityApproval)
		if overview.ParentRelation != captaindb.SessionParentRelationTranscript || (recent > limit && !urgent) {
			continue
		}
		id := overview.ID
		promptRuns, err := captain.ListPromptRuns(ctx, captaindb.PromptRunFilter{ExecutionSessionID: &id})
		if err != nil {
			return nil, fmt.Errorf("prompt runs of session %s: %w", id, err)
		}
		run, ok := newestPromptRun(promptRuns)
		if !ok {
			continue
		}
		entry := sessionRun{}
		if run.Workspace != nil {
			entry.Worktree = run.Workspace.Worktree
		}
		if link, linked := issueByRun[run.ID]; linked {
			issue, err := link.source.Repository().GetIssue(ctx, link.issueID)
			if err != nil {
				return nil, fmt.Errorf("todo %s of session %s: %w", link.issueID, id, err)
			}
			entry.Project = link.project
			entry.Todo = &sessionTodo{ID: issue.ID.String(), Title: issue.Title, Status: string(issue.Status)}
		}
		runs[id] = entry
	}
	return runs, nil
}

type sessionRunIssue struct {
	issueID uuid.UUID
	project string
	source  sessionLinkSource
}

func sessionRunIssues(ctx context.Context, sources []sessionProjectSources) (map[uuid.UUID]sessionRunIssue, error) {
	issues := map[uuid.UUID]sessionRunIssue{}
	for _, source := range sources {
		if source.links == nil {
			continue
		}
		links, err := source.links.Repository().ListPromptRunLinks(ctx, source.links.Workspace().ID)
		if err != nil {
			return nil, fmt.Errorf("prompt run links of project %s: %w", source.project.Project.Name, err)
		}
		for issueID, issueLinks := range links {
			for _, link := range issueLinks {
				issues[link.PromptRunID] = sessionRunIssue{issueID: issueID, project: source.project.Project.Name, source: source.links}
			}
		}
	}
	return issues, nil
}

func newestPromptRun(runs []captaindb.PromptRun) (captaindb.PromptRun, bool) {
	if len(runs) == 0 {
		return captaindb.PromptRun{}, false
	}
	newest := runs[0]
	for _, run := range runs[1:] {
		if run.CreatedAt.After(newest.CreatedAt) {
			newest = run
		}
	}
	return newest, true
}
