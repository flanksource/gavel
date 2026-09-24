package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
)

type sessionDetailProvider interface {
	Captain() *captaindb.DB
	Repository() *native.Repository
}

// todoAttemptDetail is one prompt run of a TODO. Its session ids name the
// transcript the dashboard loads from Captain's own session handler
// (/api/captain/sessions/{id}); this endpoint never projects the transcript.
type todoAttemptDetail struct {
	captaindb.PromptRunOverview
	PromptRunID      uuid.UUID `json:"promptRunId"`
	AdmissionSession uuid.UUID `json:"admissionSessionId"`
	Ordinal          int       `json:"ordinal"`
	Step             string    `json:"step"`
	// Verification is the newest definition-of-done report the attempt's
	// verifiers produced, read from captain's per-iteration record. It is
	// always present — null for an attempt that was never verified — so a
	// reader can tell "the server knows there is none" from "not sent".
	Verification *api.VerifyReport `json:"verification"`
	CanStop      bool              `json:"canStop,omitempty"`
	Stopping     bool              `json:"stopping,omitempty"`
}

type todoSessionDetailResponse struct {
	// Attempts is newest first.
	Attempts []todoAttemptDetail `json:"attempts"`
}

func (s *Server) handleTodoSessionDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if ref == "" {
		writeTodoError(w, http.StatusBadRequest, errors.New("ref is required"))
		return
	}
	dir, err := s.resolveTodoDir(r.URL.Query().Get("dir"))
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	provider, err := openTodoProvider(r.Context(), dir)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	detailProvider, ok := provider.(sessionDetailProvider)
	if !ok || detailProvider.Captain() == nil || detailProvider.Repository() == nil {
		writeTodoError(w, http.StatusNotImplemented, errors.New("session details require native TODO storage"))
		return
	}
	todo, err := provider.Get(r.Context(), ref)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	issueID, err := uuid.Parse(todo.ID)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, fmt.Errorf("native TODO has invalid ID %q: %w", todo.ID, err))
		return
	}
	attempts, err := listTodoAttempts(r.Context(), detailProvider, issueID)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	issue, err := detailProvider.Repository().GetIssue(r.Context(), issueID)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	if issue.ActivePromptRunID != nil {
		status := todoRuns().Status(issueID)
		for index := range attempts {
			if attempts[index].PromptRunID == *issue.ActivePromptRunID {
				attempts[index].CanStop = status.CanStop
				attempts[index].Stopping = status.Stopping
			}
		}
	}
	if err := json.NewEncoder(w).Encode(todoSessionDetailResponse{Attempts: attempts}); err != nil {
		panic(fmt.Errorf("encode TODO session detail: %w", err))
	}
}

// listTodoAttempts returns the TODO's prompt runs, newest first, each with its
// ordinal (1 for the first run) and latest verification report.
func listTodoAttempts(ctx context.Context, provider sessionDetailProvider, issueID uuid.UUID) ([]todoAttemptDetail, error) {
	links, err := provider.Repository().ListPromptRuns(ctx, issueID)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(links, func(i, j int) bool { return links[i].CreatedAt.Before(links[j].CreatedAt) })
	ids := make([]uuid.UUID, len(links))
	for index := range links {
		ids[index] = links[index].PromptRunID
	}
	overviews, err := provider.Captain().ListPromptRunOverviews(ctx, captaindb.PromptRunOverviewFilter{IDs: ids})
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]captaindb.PromptRunOverview, len(overviews))
	for index := range overviews {
		byID[overviews[index].ID] = overviews[index]
	}
	verifications, err := provider.Captain().LatestPromptRunVerifications(ctx, ids)
	if err != nil {
		return nil, err
	}
	attempts := make([]todoAttemptDetail, len(links))
	for index, link := range links {
		overview, ok := byID[link.PromptRunID]
		if !ok {
			return nil, fmt.Errorf("%w: linked prompt run %s", captaindb.ErrPromptRunNotFound, link.PromptRunID)
		}
		attempt := todoAttemptDetail{
			PromptRunOverview: overview, PromptRunID: overview.ID, AdmissionSession: overview.SessionID,
			Ordinal: index + 1, Step: string(link.StepKind),
		}
		if verification, ok := verifications[link.PromptRunID]; ok {
			attempt.Verification = verification.Report
		}
		attempts[len(links)-1-index] = attempt
	}
	return attempts, nil
}
