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
	"github.com/flanksource/gavel/todos/remote"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RunsAPI serves /api/runs/{id}/... to sandbox jobs. It is mounted outside the Clerk middleware.
type RunsAPI struct {
	DB       *gorm.DB
	TokenKey []byte
	Cache    *cache.Store
	Sources  *sandbox.Sources
	// Relay hands an agent run's messages to the host waiting for the run.
	Relay *remote.Relay
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
	mux.HandleFunc("POST /api/runs/{id}/messages", a.withRun(a.handleMessage))
	mux.HandleFunc("POST /api/runs/{id}/calls", a.withRun(a.handleCall))
	mux.HandleFunc("POST /api/runs/{id}/commits", a.withRun(a.handleCommits))
	mux.HandleFunc("POST /api/runs/{id}/transcripts", a.withRun(a.handleTranscript))
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

// handleMessage hands one message of an agent run to the host waiting for the run, once it has
// taken it, so a job's messages arrive in the order it sends them.
func (a *RunsAPI) handleMessage(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord) {
	message, ok := a.runMessage(w, r, run)
	if !ok {
		return
	}
	if err := a.Relay.Deliver(r.Context(), run.ID, message); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, remote.ErrNoListener) {
			status = http.StatusConflict
		}
		writeCIError(w, status, err.Error())
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// handleCall hands a call of an agent run, a tool call or an approval request, to the host
// waiting for the run, and answers with its reply once there is one.
func (a *RunsAPI) handleCall(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord) {
	message, ok := a.runMessage(w, r, run)
	if !ok {
		return
	}
	if message.Kind != remote.KindTool && message.Kind != remote.KindApproval {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("a %q message isn't a call; calls are tool calls and approval requests", message.Kind))
		return
	}
	reply, err := a.Relay.Call(r.Context(), run.ID, message)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, remote.ErrNoListener) {
			status = http.StatusConflict
		}
		writeCIError(w, status, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(reply)
}

// runMessage decodes the message of an agent run, answering the request itself when it can't.
func (a *RunsAPI) runMessage(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord) (remote.Message, bool) {
	if run.Spec.Kind != "agent" {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("run %s is a %s run; only agent runs send messages", run.ID, run.Spec.Kind))
		return remote.Message{}, false
	}
	var message remote.Message
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRunMessageBytes)).Decode(&message); err != nil {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("decode the message of run %s: %v", run.ID, err))
		return remote.Message{}, false
	}
	if a.Relay == nil {
		writeCIError(w, http.StatusConflict, fmt.Sprintf("%v: run %s", remote.ErrNoListener, run.ID))
		return remote.Message{}, false
	}
	return message, true
}

// maxRunMessageBytes bounds one message: a run's result carries its whole loop.
const maxRunMessageBytes = 64 << 20

func (a *RunsAPI) handleSpec(w http.ResponseWriter, _ *http.Request, run sandbox.RunRecord) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(run.Spec)
}

func (a *RunsAPI) handleResults(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord) {
	if run.Spec.RepoID == uuid.Nil {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("run %s has no repository to store results under", run.ID))
		return
	}
	storeResults(w, r, a.Cache, run.Org, run.Spec.RepoID, run.Spec.SHA, "sandbox run "+run.ID.String())
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

// handleCommits puts the commits an agent run made, a git bundle, on the run's own branch:
// gavel/<todo>-<run>, from the commit the run started at.
func (a *RunsAPI) handleCommits(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord) {
	if run.Spec.Kind != "agent" || run.Spec.TodoID == nil {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("run %s is a %s run; only agent runs push commits", run.ID, run.Spec.Kind))
		return
	}
	repo, ok := a.runRepo(w, r, run)
	if !ok {
		return
	}
	branch := sandbox.RunBranchPrefix + run.Spec.TodoID.String()[:8] + "-" + run.ID.String()[:8]
	sha, err := a.Sources.PushRunCommits(r.Context(), repo, run.Spec.SHA, branch, http.MaxBytesReader(w, r.Body, maxRunCommitsBytes))
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, sandbox.ErrBranchExists) {
			status = http.StatusConflict
		}
		writeCIError(w, status, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"branch": branch, "sha": sha})
}

// maxRunCommitsBytes bounds the bundle of a run's commits.
const maxRunCommitsBytes = 256 << 20

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
