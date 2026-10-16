// The run API's PR endpoints: an agent run on a linked PR has no GitHub credentials, so serve
// reads its PR's status, resolves its threads and pushes its source branch for it, for that PR
// alone, with the organization's GitHub App installation.
package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/sandbox"
)

// runPR is the PR a run works on: its client and number.
type runPR struct {
	client provider.Client
	number int
}

// withPR resolves the run's PR, refusing a run that works on none and a request naming another.
func (a *RunsAPI) withPR(next func(http.ResponseWriter, *http.Request, sandbox.RunRecord, runPR)) func(http.ResponseWriter, *http.Request, sandbox.RunRecord) {
	return func(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord) {
		if run.Spec.Kind != "agent" || run.Spec.PR == "" {
			writeCIError(w, http.StatusBadRequest, fmt.Sprintf("run %s works on no PR", run.ID))
			return
		}
		if a.PRs == nil {
			writeCIError(w, http.StatusServiceUnavailable, "this serve has no GitHub App to act on PRs with")
			return
		}
		client, number, err := a.PRs(r.Context(), run.Org, run.Spec.PR)
		if err != nil {
			logger.Warnf("open the PR of sandbox run %s: %v", run.ID, err)
			writeCIError(w, http.StatusBadGateway, fmt.Sprintf("could not reach PR %s", run.Spec.PR))
			return
		}
		if asked := r.URL.Query().Get("number"); asked != "" && asked != strconv.Itoa(number) {
			writeCIError(w, http.StatusForbidden, fmt.Sprintf("run %s works on PR #%d, not #%s", run.ID, number, asked))
			return
		}
		next(w, r, run, runPR{client: client, number: number})
	}
}

func (a *RunsAPI) handlePRStatus(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord, pr runPR) {
	query := r.URL.Query()
	tail, _ := strconv.Atoi(query.Get("tailLogs"))
	snapshot, err := pr.client.Status(r.Context(), pr.number, model.StatusOptions{
		Threads: query.Get("threads") == "true", Logs: query.Get("logs") == "true", TailLogs: tail,
	})
	if err != nil {
		writeCIError(w, http.StatusBadGateway, fmt.Sprintf("read the status of PR #%d: %v", pr.number, err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snapshot)
}

func (a *RunsAPI) handlePRHead(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord, pr runPR) {
	head, ok := prHead(r.Context(), w, pr)
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"sha": head.HeadRefOID, "branch": head.HeadRefName})
}

func (a *RunsAPI) handlePRResolve(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord, pr runPR) {
	thread := r.PathValue("thread")
	if err := pr.client.ResolveThread(r.Context(), pr.number, thread); err != nil {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("resolve review thread %s of PR #%d: %v", thread, pr.number, err))
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// handlePRPush fast-forwards the PR's source branch, which GitHub names, from base, the head the
// run's turn started at, to the commits the run uploaded.
func (a *RunsAPI) handlePRPush(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord, pr runPR) {
	head, ok := prHead(r.Context(), w, pr)
	if !ok {
		return
	}
	base := r.URL.Query().Get("base")
	if base != head.HeadRefOID {
		writeCIError(w, http.StatusConflict, fmt.Sprintf("the PR's source branch is at %s, not %s", head.HeadRefOID, base))
		return
	}
	repo, ok := a.runRepo(w, r, run)
	if !ok {
		return
	}
	sha, err := a.Sources.PushPRHead(r.Context(), repo, head.HeadRepositoryURL, head.HeadRefName, base, http.MaxBytesReader(w, r.Body, maxRunCommitsBytes))
	if err != nil {
		writeCIError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"sha": sha})
}

// prHead is the PR's open source branch, answering the request itself when it can't read one.
func prHead(ctx context.Context, w http.ResponseWriter, pr runPR) (*model.PRInfo, bool) {
	snapshot, err := pr.client.Status(ctx, pr.number, model.StatusOptions{})
	if err != nil {
		writeCIError(w, http.StatusBadGateway, fmt.Sprintf("read the status of PR #%d: %v", pr.number, err))
		return nil, false
	}
	if snapshot == nil || snapshot.PR == nil || snapshot.PR.State != "OPEN" {
		writeCIError(w, http.StatusConflict, fmt.Sprintf("PR #%d is no longer open", pr.number))
		return nil, false
	}
	return snapshot.PR, true
}
