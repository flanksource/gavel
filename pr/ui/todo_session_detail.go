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
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
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
	// Landing is how the attempt's worktree commits reached a branch; absent
	// until `todos land` records one.
	Landing *native.RunLanding `json:"landing,omitempty"`
	// OutcomeError is why gavel could not persist the outcome of a run Captain
	// had finished. Such an attempt failed: its status reads failed and its
	// error carries this, while State stays the run state Captain filed.
	OutcomeError string `json:"outcomeError,omitempty"`
	CanStop      bool   `json:"canStop,omitempty"`
	Stopping     bool   `json:"stopping,omitempty"`
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
	ids := make([]uuid.UUID, len(links))
	for index := range links {
		ids[index] = links[index].PromptRunID
	}
	overviews, err := provider.Captain().ListPromptRunOverviews(ctx, captaindb.PromptRunOverviewFilter{IDs: ids})
	if err != nil {
		return nil, err
	}
	for index := range overviews {
		if err := promptrun.ReconcileWorkspace(ctx, provider.Captain(), &overviews[index].PromptRun); err != nil {
			return nil, err
		}
	}
	verifications, err := provider.Captain().LatestPromptRunVerifications(ctx, ids)
	if err != nil {
		return nil, err
	}
	landings, err := provider.Repository().ListLandings(ctx, issueID)
	if err != nil {
		return nil, fmt.Errorf("list landings of TODO %s: %w", issueID, err)
	}
	events, err := provider.Repository().ListEvents(ctx, issueID)
	if err != nil {
		return nil, fmt.Errorf("list events of TODO %s: %w", issueID, err)
	}
	unpersisted, err := unpersistedOutcomes(events)
	if err != nil {
		return nil, fmt.Errorf("read lifecycle outcomes of TODO %s: %w", issueID, err)
	}
	return assembleTodoAttempts(todoAttemptRecords{
		Links: links, Overviews: overviews, Verifications: verifications, Landings: landings, Unpersisted: unpersisted,
	})
}

// todoAttemptRecords is everything recorded about a TODO's prompt runs.
type todoAttemptRecords struct {
	Links         []native.PromptRunLink
	Overviews     []captaindb.PromptRunOverview
	Verifications map[uuid.UUID]captaindb.PromptRunVerification
	Landings      []native.RunLanding
	// Unpersisted maps a run to why its outcome could not be persisted.
	Unpersisted map[uuid.UUID]string
}

// unpersistedOutcomes reads, from a TODO's history, the runs whose latest
// lifecycle outcome records that it could not be persisted, and why.
func unpersistedOutcomes(events []native.Event) (map[uuid.UUID]string, error) {
	type outcomePayload struct {
		PromptRunID string `json:"promptRunId"`
		Status      string `json:"status"`
		Error       string `json:"error"`
	}
	unpersisted := map[uuid.UUID]string{}
	sort.SliceStable(events, func(i, j int) bool { return events[i].Sequence < events[j].Sequence })
	for _, event := range events {
		if event.Kind != lifecycle.EventLifecycleOutcome {
			continue
		}
		var payload outcomePayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return nil, fmt.Errorf("decode lifecycle outcome %d: %w", event.Sequence, err)
		}
		if payload.PromptRunID == "" {
			continue
		}
		runID, err := uuid.Parse(payload.PromptRunID)
		if err != nil {
			return nil, fmt.Errorf("lifecycle outcome %d names invalid prompt run %q: %w", event.Sequence, payload.PromptRunID, err)
		}
		delete(unpersisted, runID)
		if payload.Status == string(types.StatusFailed) && payload.Error != "" {
			unpersisted[runID] = payload.Error
		}
	}
	return unpersisted, nil
}

// assembleTodoAttempts joins a TODO's prompt-run links with captain's overview
// of each run — which carries the run's recorded workspace — its latest
// verification, its landing and an outcome gavel could not persist, newest
// first. A link whose run captain does not know, or a landing of a run that is
// not linked, is an error, never a silently shorter history.
func assembleTodoAttempts(records todoAttemptRecords) ([]todoAttemptDetail, error) {
	links := records.Links
	sort.SliceStable(links, func(i, j int) bool { return links[i].CreatedAt.Before(links[j].CreatedAt) })
	byID := make(map[uuid.UUID]captaindb.PromptRunOverview, len(records.Overviews))
	for index := range records.Overviews {
		byID[records.Overviews[index].ID] = records.Overviews[index]
	}
	landingByRun := make(map[uuid.UUID]*native.RunLanding, len(records.Landings))
	for index := range records.Landings {
		landingByRun[records.Landings[index].PromptRunID] = &records.Landings[index]
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
		if verification, ok := records.Verifications[link.PromptRunID]; ok {
			attempt.Verification = verification.Report
		}
		if reason, ok := records.Unpersisted[link.PromptRunID]; ok {
			attempt.OutcomeError, attempt.Status = reason, captaindb.PromptRunStatusFailed
			attempt.Error = strings.TrimSpace(strings.Join([]string{attempt.Error, reason}, "\n"))
		}
		attempt.Landing = landingByRun[link.PromptRunID]
		delete(landingByRun, link.PromptRunID)
		attempts[len(links)-1-index] = attempt
	}
	for runID := range landingByRun {
		return nil, fmt.Errorf("%w: landing of prompt run %s, which is not linked to the TODO", native.ErrLinkConflict, runID)
	}
	return attempts, nil
}
