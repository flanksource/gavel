package ui

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	prlifecycle "github.com/flanksource/gavel/pr/lifecycle"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/workflows"
	"github.com/google/uuid"
)

type prLifecycleRequest struct {
	prlifecycle.Selection
	Operation string    `json:"operation"`
	SpecHash  string    `json:"specHash"`
	RunID     uuid.UUID `json:"runId"`
}

type prLifecyclePrepared struct {
	Spec     api.ResolvedSpec `json:"spec"`
	SpecHash string           `json:"specHash"`
}

type prLifecycleControl struct {
	cancel context.CancelFunc
	done   chan struct{}
}

var prLifecycleControls sync.Map

func (s *Server) prLifecycleHost(r *http.Request) (prlifecycle.Host, error) {
	project, err := GetProject(r.URL.Query().Get("project"))
	if err != nil {
		return prlifecycle.Host{}, err
	}
	repo := r.URL.Query().Get("repo")
	owned := false
	for _, name := range project.Repos {
		if name == repo {
			owned = true
		}
	}
	if !owned {
		return prlifecycle.Host{}, fmt.Errorf("repository %q is not registered for project %s", repo, project.Name)
	}
	number, err := strconv.Atoi(r.URL.Query().Get("number"))
	if err != nil || number <= 0 {
		return prlifecycle.Host{}, fmt.Errorf("positive PR number is required")
	}
	client, err := github.GitClient(r.Context(), s.ghOpts)
	if err != nil {
		return prlifecycle.Host{}, err
	}
	repository, err := client.Repository("https://github.com/" + repo)
	if err != nil {
		return prlifecycle.Host{}, err
	}
	store, err := openTodoProvider(r.Context(), project.ResolvedDir())
	if err != nil {
		return prlifecycle.Host{}, err
	}
	backed, ok := store.(interface{ Captain() *captaindb.DB })
	if !ok || backed.Captain() == nil {
		return prlifecycle.Host{}, fmt.Errorf("PR lifecycles require a Captain database")
	}
	view, err := workflows.Inspect("pr", project.ResolvedDir())
	if err != nil {
		return prlifecycle.Host{}, err
	}
	return prlifecycle.Host{DB: backed.Captain(), Dir: project.ResolvedDir(), Provider: client.Provider, Repo: repository.URL, Number: number, Definition: view.Effective}, nil
}

func (s *Server) handlePRLifecycle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	host, err := s.prLifecycleHost(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	snapshot, err := host.Snapshot(r.Context())
	if err != nil {
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}
	if r.Method == http.MethodGet {
		snapshot.Failures, err = s.prLifecycleFailures(host, snapshot)
		if err != nil {
			respondError(w, http.StatusBadGateway, err.Error())
			return
		}
		view, err := host.Preview(r.Context(), snapshot)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, view)
		return
	}
	var request prLifecycleRequest
	if err := decodeRuntimePresetRequest(r, &request); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if request.Operation == "cancel" {
		cancelPRLifecycle(w, r, host, snapshot, request.RunID)
		return
	}
	if request.Operation != "preview" && request.Operation != "start" {
		respondError(w, http.StatusBadRequest, "operation must be preview, start or cancel")
		return
	}
	snapshot.Failures, err = s.prLifecycleFailures(host, snapshot)
	if err != nil {
		respondError(w, http.StatusBadGateway, err.Error())
		return
	}
	input, err := host.Prepare(r.Context(), snapshot, request.Selection)
	if err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}
	encoded, err := json.Marshal(input.Resolved.Spec)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	fingerprint := fmt.Sprintf("%x", sha256.Sum256(encoded))
	if request.Operation == "preview" {
		respondJSON(w, http.StatusOK, prLifecyclePrepared{Spec: input.Resolved, SpecHash: fingerprint})
		return
	}
	if request.SpecHash != fingerprint {
		respondError(w, http.StatusConflict, "resolved prompt changed; preview again before starting")
		return
	}
	source, err := host.Repo.SourceRepository(snapshot.PR.HeadRepository)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	target := provider.RunTarget{RepositoryURL: source.Clone, Branch: snapshot.PR.Head, HeadSHA: snapshot.PR.HeadSHA}
	if err := target.FetchHead(r.Context(), host.Dir); err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}
	run, err := promptrun.Admit(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}
	runCtx, cancel := context.WithCancel(s.context())
	control := &prLifecycleControl{cancel: cancel, done: make(chan struct{})}
	prLifecycleControls.Store(run.ID, control)
	go func() {
		defer cancel()
		defer close(control.done)
		defer prLifecycleControls.Delete(run.ID)
		if _, err := promptrun.Run(runCtx, input); err != nil {
			logger.Errorf("PR lifecycle run %s: %v", run.ID, err)
			outcome := promptrun.Outcome{State: captaindb.PromptRunStateFailed, Error: err.Error()}
			if runCtx.Err() != nil {
				outcome.State = captaindb.PromptRunStateCancelled
			}
			if _, settleErr := promptrun.Settle(context.WithoutCancel(runCtx), host.DB, run.ID, outcome); settleErr != nil && !errors.Is(settleErr, promptrun.ErrRunFinished) {
				logger.Errorf("settle PR lifecycle run %s: %v", run.ID, settleErr)
			}
		}
	}()
	respondJSON(w, http.StatusAccepted, run)
}

func cancelPRLifecycle(w http.ResponseWriter, r *http.Request, host prlifecycle.Host, snapshot prlifecycle.Snapshot, id uuid.UUID) {
	run, err := host.DB.GetPromptRun(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if run.SessionID != prlifecycle.SessionID(snapshot.PR.URL) {
		respondError(w, http.StatusBadRequest, "run does not belong to this PR")
		return
	}
	if value, ok := prLifecycleControls.Load(id); ok {
		control := value.(*prLifecycleControl)
		control.cancel()
		select {
		case <-control.done:
		case <-r.Context().Done():
			return
		}
		run, err = host.DB.GetPromptRun(r.Context(), id)
	} else {
		if run.State == captaindb.PromptRunStateRunning {
			respondError(w, http.StatusConflict, "run is executing outside this dashboard; stop it through its Captain session")
			return
		}
		run, err = promptrun.Cancel(r.Context(), host.DB, id, "Cancelled from PR details")
	}
	if err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, run)
}
