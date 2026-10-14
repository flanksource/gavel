package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// prIndexStore is the persisted PR index the poller writes and the PR link and
// search endpoints read. It is a variable so package tests can point it at a
// provisioned database instead of the process-wide store.
var prIndexStore = func() *cache.Store { return cache.Shared() }

// defaultPRSearchLimit is the page a PR picker asks for when it names no limit.
const defaultPRSearchLimit = 20

// todoPRLinkItem is one PR link as the TODO pages render it: the link plus the
// title, state and author the poller last indexed for that PR. The indexed
// fields are absent for a PR the poller has not synced.
type todoPRLinkItem struct {
	todos.PRLink
	Title   string `json:"title,omitempty"`
	State   string `json:"state,omitempty"`
	Author  string `json:"author,omitempty"`
	IsDraft bool   `json:"isDraft,omitempty"`
}

type todoPRLinkPayload struct {
	Dir      string `json:"dir"`
	ID       string `json:"id"`
	URL      string `json:"url"`
	Relation string `json:"relation"`
}

// prIndexRow is the poller's view of a PR as the index stores it. The poller
// only searches GitHub.
func prIndexRow(pr github.PRListItem) cache.PullRequest {
	return cache.PullRequest{
		Provider: types.PRProviderGitHub, Repo: pr.Repo, Number: pr.Number, URL: pr.URL,
		Title: pr.Title, Author: pr.Author, State: pr.State, IsDraft: pr.IsDraft,
		Head: pr.Source, Base: pr.Target, UpdatedAt: pr.UpdatedAt,
	}
}

func (s *Server) handleTodoPRLinks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		query := r.URL.Query()
		linker, todo, ok := s.todoPRLinkTarget(w, r.Context(), query.Get("dir"), query.Get("id"))
		if !ok {
			return
		}
		links, err := linker.PRLinks(r.Context(), todo)
		if err != nil {
			writeTodoError(w, http.StatusInternalServerError, err)
			return
		}
		items, err := withIndexedPRs(links)
		if err != nil {
			writePRIndexError(w, err)
			return
		}
		writeTodoJSON(w, http.StatusOK, items)
	case http.MethodPost:
		s.handleTodoPRLinkCreate(w, r)
	case http.MethodDelete:
		query := r.URL.Query()
		if _, err := types.PRRefFromURL(query.Get("url")); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
		linker, todo, ok := s.todoPRLinkTarget(w, r.Context(), query.Get("dir"), query.Get("id"))
		if !ok {
			return
		}
		if err := linker.UnlinkPR(r.Context(), todo, query.Get("url")); err != nil {
			writeTodoError(w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeTodoError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s not allowed", r.Method))
	}
}

func (s *Server) handleTodoPRLinkCreate(w http.ResponseWriter, r *http.Request) {
	var payload todoPRLinkPayload
	if err := decodeTodoRequest(r, &payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	if _, err := types.PRRefFromURL(payload.URL); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	relation, err := types.ParsePRRelation(payload.Relation)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	dir := payload.Dir
	if dir == "" {
		dir = r.URL.Query().Get("dir")
	}
	linker, todo, ok := s.todoPRLinkTarget(w, r.Context(), dir, payload.ID)
	if !ok {
		return
	}
	link, err := linker.LinkPR(r.Context(), todo, todos.PRLinkRequest{URL: payload.URL, Relation: relation, Source: types.PRLinkSourceManual})
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	items, err := withIndexedPRs([]todos.PRLink{*link})
	if err != nil {
		writePRIndexError(w, err)
		return
	}
	writeTodoJSON(w, http.StatusCreated, items[0])
}

// todoPRLinkTarget resolves the TODO named by id in dir's workspace. It writes
// the error response itself; ok is false when the caller must stop.
func (s *Server) todoPRLinkTarget(w http.ResponseWriter, ctx context.Context, dir, id string) (todos.PRLinkProvider, *types.TODO, bool) {
	if id = strings.TrimSpace(id); id == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("id is required"))
		return nil, nil, false
	}
	provider, linker, err := s.todoPRLinkProvider(ctx, dir)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return nil, nil, false
	}
	todo, err := provider.Get(ctx, id)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return nil, nil, false
	}
	return linker, todo, true
}

func (s *Server) todoPRLinkProvider(ctx context.Context, dir string) (todos.Provider, todos.PRLinkProvider, error) {
	provider, _, err := s.todoProviderContext(ctx, todoSource{Dir: strings.TrimSpace(dir)})
	if err != nil {
		return nil, nil, err
	}
	linker, ok := provider.(todos.PRLinkProvider)
	if !ok {
		return nil, nil, fmt.Errorf("TODO provider does not support PR links; native PostgreSQL storage is required")
	}
	return provider, linker, nil
}

// withIndexedPRs joins each link to the PR index for its title and state.
func withIndexedPRs(links []todos.PRLink) ([]todoPRLinkItem, error) {
	items := make([]todoPRLinkItem, len(links))
	if len(links) == 0 {
		return items, nil
	}
	keys := make([]cache.PRKey, len(links))
	for i, link := range links {
		keys[i] = cache.PRKey{Provider: link.Provider, Repo: link.Repo, Number: link.Number}
	}
	indexed, err := prIndexStore().GetPRs(keys)
	if err != nil {
		return nil, err
	}
	for i, link := range links {
		items[i] = todoPRLinkItem{PRLink: link}
		if pr, ok := indexed[keys[i]]; ok {
			items[i].Title, items[i].State, items[i].Author, items[i].IsDraft = pr.Title, pr.State, pr.Author, pr.IsDraft
		}
	}
	return items, nil
}

// handlePRTodos lists the TODOs in dir's workspace (the server's own when dir is
// empty) that link the PR at url.
func (s *Server) handlePRTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	if _, err := types.PRRefFromURL(query.Get("url")); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	_, linker, err := s.todoPRLinkProvider(r.Context(), query.Get("dir"))
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	linked, err := linker.TodosForPR(r.Context(), query.Get("url"))
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	writeTodoJSON(w, http.StatusOK, linked)
}

// handlePRSearch searches the persisted PR index. An empty q lists the most
// recently updated PRs, which is what a picker shows before anything is typed.
func (s *Server) handlePRSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	limit := defaultPRSearchLimit
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			writeTodoError(w, http.StatusBadRequest, fmt.Errorf("limit %q is not a number", raw))
			return
		}
		limit = parsed
	}
	rows, err := prIndexStore().SearchPRs(cache.PRQuery{
		Text: query.Get("q"), Repo: query.Get("repo"), State: query.Get("state"), Limit: limit,
	})
	if err != nil {
		writePRIndexError(w, err)
		return
	}
	writeTodoJSON(w, http.StatusOK, rows)
}

func writePRIndexError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, cache.ErrPRIndexUnavailable):
		writeTodoErrorStatus(w, http.StatusServiceUnavailable, err)
	case errors.Is(err, cache.ErrInvalidPRQuery):
		writeTodoErrorStatus(w, http.StatusBadRequest, err)
	default:
		writeTodoErrorStatus(w, http.StatusInternalServerError, err)
	}
}
