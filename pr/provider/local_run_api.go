// A local caller's run API for a run on substrate: the run's program has no credentials, so the
// caller answers its PR calls itself, with its own PR client and git, as serve answers a hosted run's.
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/pr/model"
	"github.com/google/uuid"
)

// maxPushBytes bounds the bundle of the commits a run pushes to its PR.
const maxPushBytes = 256 << 20

// LocalRunAPI answers a run's PR calls for PR number through client, the caller's own: the PR's
// status and head, its review threads resolved, and pushes to target's source branch, taken in
// through repo, the caller's checkout, and pushed with the caller's git credentials. A push must
// continue the head the run saw, which must still be the PR's, and is leased on it.
func LocalRunAPI(client Client, number int, target RunTarget, repo string) http.Handler {
	api := localRunAPI{client: client, number: number, target: target, repo: repo}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/runs/{id}/pr/status", api.status)
	mux.HandleFunc("GET /api/runs/{id}/pr/head", api.head)
	mux.HandleFunc("POST /api/runs/{id}/pr/threads/{thread}/resolve", api.resolve)
	mux.HandleFunc("POST /api/runs/{id}/pr/push", api.push)
	return mux
}

type localRunAPI struct {
	client Client
	number int
	target RunTarget
	repo   string
}

// refuse answers a request the run API can't fulfil, with why.
func refuse(w http.ResponseWriter, status int, reason string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": reason})
}

// samePR refuses a request naming a PR other than the run's.
func (a localRunAPI) samePR(w http.ResponseWriter, r *http.Request) bool {
	if asked := r.URL.Query().Get("number"); asked != "" && asked != strconv.Itoa(a.number) {
		refuse(w, http.StatusForbidden, fmt.Sprintf("the run works on PR #%d, not #%s", a.number, asked))
		return false
	}
	return true
}

func (a localRunAPI) status(w http.ResponseWriter, r *http.Request) {
	if !a.samePR(w, r) {
		return
	}
	query := r.URL.Query()
	tail, _ := strconv.Atoi(query.Get("tailLogs"))
	snapshot, err := a.client.Status(r.Context(), a.number, model.StatusOptions{
		Threads: query.Get("threads") == "true", Logs: query.Get("logs") == "true", TailLogs: tail,
	})
	if err != nil {
		refuse(w, http.StatusBadGateway, fmt.Sprintf("read the status of PR #%d: %v", a.number, err))
		return
	}
	_ = json.NewEncoder(w).Encode(snapshot)
}

// openHead is the PR's source branch head, answering the request itself when the PR has none.
func (a localRunAPI) openHead(ctx context.Context, w http.ResponseWriter) (*model.PRInfo, bool) {
	snapshot, err := a.client.Status(ctx, a.number, model.StatusOptions{})
	if err != nil {
		refuse(w, http.StatusBadGateway, fmt.Sprintf("read the status of PR #%d: %v", a.number, err))
		return nil, false
	}
	if snapshot == nil || snapshot.PR == nil || snapshot.PR.State != "OPEN" {
		refuse(w, http.StatusConflict, fmt.Sprintf("PR #%d is no longer open", a.number))
		return nil, false
	}
	return snapshot.PR, true
}

func (a localRunAPI) head(w http.ResponseWriter, r *http.Request) {
	pr, ok := a.openHead(r.Context(), w)
	if !ok {
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"sha": pr.HeadRefOID, "branch": pr.HeadRefName})
}

func (a localRunAPI) resolve(w http.ResponseWriter, r *http.Request) {
	if !a.samePR(w, r) {
		return
	}
	thread := r.PathValue("thread")
	if err := a.client.ResolveThread(r.Context(), a.number, thread); err != nil {
		refuse(w, http.StatusBadRequest, fmt.Sprintf("resolve review thread %s of PR #%d: %v", thread, a.number, err))
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// push fast-forwards the PR's source branch from base, the head the run's turn started at, to
// the commits the run sent.
func (a localRunAPI) push(w http.ResponseWriter, r *http.Request) {
	pr, ok := a.openHead(r.Context(), w)
	if !ok {
		return
	}
	base := r.URL.Query().Get("base")
	if base != pr.HeadRefOID {
		refuse(w, http.StatusConflict, fmt.Sprintf("the PR's source branch is at %s, not %s", pr.HeadRefOID, base))
		return
	}
	sha, done, err := a.receive(r.Context(), base, http.MaxBytesReader(w, r.Body, maxPushBytes))
	if err != nil {
		refuse(w, http.StatusBadRequest, err.Error())
		return
	}
	defer done()
	lease := fmt.Sprintf("--force-with-lease=refs/heads/%s:%s", a.target.Branch, base)
	if _, err := targetGit(r.Context(), a.repo, "push", "--porcelain", lease, "--", a.target.RepositoryURL, sha+":refs/heads/"+a.target.Branch); err != nil {
		refuse(w, http.StatusBadRequest, fmt.Sprintf("push the PR's source branch %s: %v", a.target.Branch, err))
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"sha": sha})
}

// receive takes the one branch bundle holds into the caller's checkout under a temporary ref,
// after checking it continues from base, and returns its commit and a func that drops the ref.
func (a localRunAPI) receive(ctx context.Context, base string, bundle io.Reader) (string, func(), error) {
	file, err := os.CreateTemp("", "gavel-pr-push-*.bundle")
	if err != nil {
		return "", nil, fmt.Errorf("save the run's commits: %w", err)
	}
	defer os.Remove(file.Name())
	_, err = io.Copy(file, bundle)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", nil, fmt.Errorf("save the run's commits: %w", err)
	}
	heads, err := targetGit(ctx, a.repo, "bundle", "list-heads", file.Name())
	if err != nil {
		return "", nil, fmt.Errorf("read the run's bundle: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(heads), "\n")
	if len(lines) != 1 || lines[0] == "" {
		return "", nil, fmt.Errorf("the run's bundle has %d branches; a run's commits are on one", len(lines))
	}
	sha, ref, _ := strings.Cut(lines[0], " ")
	incoming := "refs/gavel/incoming/" + uuid.NewString()
	if _, err := targetGit(ctx, a.repo, "fetch", "--quiet", file.Name(), ref+":"+incoming); err != nil {
		return "", nil, fmt.Errorf("read the run's commits: %w", err)
	}
	done := func() { _, _ = targetGit(context.WithoutCancel(ctx), a.repo, "update-ref", "-d", incoming) }
	if _, err := targetGit(ctx, a.repo, "merge-base", "--is-ancestor", base, sha); err != nil {
		done()
		return "", nil, fmt.Errorf("the run's commits don't continue the PR's head %s", base)
	}
	return sha, done, nil
}
