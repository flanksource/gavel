// The run API sandbox jobs call back to: their spec and their results, authenticated only by the
// run token the dispatcher gave the job, for that run alone.
package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/sandbox"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RunsAPI serves /api/runs/{id}/... to sandbox jobs. It is mounted outside the Clerk middleware.
type RunsAPI struct {
	DB       *gorm.DB
	TokenKey []byte
	Cache    *cache.Store
	Sources  *sandbox.Sources
	// PRs is the client an organization's PR at url is read and resolved through, and its
	// number: the organization's GitHub App installation.
	PRs func(ctx context.Context, org uuid.UUID, url string) (provider.Client, int, error)
}

// Handler routes the run API.
func (a *RunsAPI) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/runs/{id}/spec", a.withRun(a.handleSpec))
	mux.HandleFunc("GET /api/runs/{id}/source", a.withRun(a.handleSource))
	mux.HandleFunc("POST /api/runs/{id}/results", a.withRun(a.handleResults))
	mux.HandleFunc("GET /api/runs/{id}/pr/status", a.withRun(a.withPR(a.handlePRStatus)))
	mux.HandleFunc("GET /api/runs/{id}/pr/head", a.withRun(a.withPR(a.handlePRHead)))
	mux.HandleFunc("POST /api/runs/{id}/pr/threads/{thread}/resolve", a.withRun(a.withPR(a.handlePRResolve)))
	mux.HandleFunc("POST /api/runs/{id}/pr/push", a.withRun(a.withPR(a.handlePRPush)))
	return mux
}

// withRun authenticates the run token, which must be for the run in the path, and loads the run.
func (a *RunsAPI) withRun(next func(http.ResponseWriter, *http.Request, sandbox.RunRecord)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeCIError(w, http.StatusBadRequest, fmt.Sprintf("%q is not a run id", r.PathValue("id")))
			return
		}
		raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || raw == "" {
			writeCIError(w, http.StatusUnauthorized, "send the job's run token as Authorization: Bearer")
			return
		}
		tokenRun, tokenOrg, err := sandbox.VerifyRunToken(a.TokenKey, raw, time.Now())
		if err != nil {
			writeCIError(w, http.StatusUnauthorized, err.Error())
			return
		}
		if tokenRun != id {
			writeCIError(w, http.StatusForbidden, fmt.Sprintf("this run token is for run %s, not %s", tokenRun, id))
			return
		}
		run, err := sandbox.LoadRun(r.Context(), a.DB, id)
		if errors.Is(err, sandbox.ErrRunNotFound) {
			writeCIError(w, http.StatusNotFound, err.Error())
			return
		}
		if err != nil {
			logger.Warnf("load sandbox run %s: %v", id, err)
			writeCIError(w, http.StatusInternalServerError, "could not read the run")
			return
		}
		if run.Org != tokenOrg {
			writeCIError(w, http.StatusForbidden, fmt.Sprintf("run %s belongs to another organization", id))
			return
		}
		next(w, r, run)
	}
}

func (a *RunsAPI) handleSpec(w http.ResponseWriter, _ *http.Request, run sandbox.RunRecord) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(run.Spec)
}

func (a *RunsAPI) handleResults(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord) {
	if run.Spec.RepoID == uuid.Nil {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("run %s has no repository to store results under", run.ID))
		return
	}
	storeResults(w, r, a.Cache, run.Org, run.Spec.RepoID, run.Spec.SHA, "sandbox run "+run.ID.String(), "")
}

// runRepo reads run's repository, answering the request itself when it can't.
func (a *RunsAPI) runRepo(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord) (sandbox.Repo, bool) {
	var repo sandbox.Repo
	result := a.DB.WithContext(r.Context()).Raw(`SELECT id, org_id AS org, source, name, COALESCE(github_repo_id, 0) AS git_hub_repo_id, COALESCE(installation_id, 0) AS installation_id
		FROM repos WHERE id = ? AND org_id = ?`, run.Spec.RepoID, run.Org).Scan(&repo)
	if result.Error != nil {
		logger.Warnf("read the repository of sandbox run %s: %v", run.ID, result.Error)
		writeCIError(w, http.StatusInternalServerError, "could not read the run's repository")
		return sandbox.Repo{}, false
	}
	if result.RowsAffected == 0 {
		writeCIError(w, http.StatusNotFound, fmt.Sprintf("run %s has no repository", run.ID))
		return sandbox.Repo{}, false
	}
	return repo, true
}

// handleSource streams a git bundle of the run's ref.
func (a *RunsAPI) handleSource(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord) {
	repo, ok := a.runRepo(w, r, run)
	if !ok {
		return
	}
	body := &startedWriter{ResponseWriter: w, contentType: "application/x-git-bundle"}
	if _, err := a.Sources.Bundle(r.Context(), repo, run.Spec.Ref, body); err != nil {
		logger.Warnf("bundle the source of sandbox run %s: %v", run.ID, err)
		if !body.started {
			writeCIError(w, http.StatusInternalServerError, err.Error())
		}
	}
}

// startedWriter sets the content type on the first write and records that the response began,
// so an error before any bundle bytes can still be a proper error response.
type startedWriter struct {
	http.ResponseWriter
	contentType string
	started     bool
}

func (w *startedWriter) Write(p []byte) (int, error) {
	if !w.started {
		w.started = true
		w.Header().Set("Content-Type", w.contentType)
	}
	return w.ResponseWriter.Write(p)
}
