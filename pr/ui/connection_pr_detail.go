// Reads the dashboard detail of a pull request a git connection's repository holds through that
// connection's provider: the PR, the CI its pipelines ran with failed steps' log tails, comments.
package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/prwatch"
	"github.com/flanksource/gavel/todos/types"
)

// connectionLogTail is how many of a failed step's last log lines its detail carries.
const connectionLogTail = 100

// connectionDetailTimeout bounds reading one pull request's detail through its connection.
const connectionDetailTimeout = 2 * time.Minute

// PRRemote reaches the repository of a listed pull request a git connection brings, with the
// provider its detail is read through.
type PRRemote func(ctx context.Context, pr github.PRListItem) (*git.Repository, error)

// SetPRRemote reads the detail of listed pull requests from git connections through remote.
func (s *Server) SetPRRemote(remote PRRemote) { s.prRemote = remote }

// connectionPR is the listed pull request repo#number when a git connection brings it.
func (s *Server) connectionPR(repo string, number int) (github.PRListItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, pr := range s.prs {
		if pr.Provider != "" && pr.Repo == repo && pr.Number == number {
			return pr, true
		}
	}
	return github.PRListItem{}, false
}

// connectionPRAt is the listed pull request at url when a git connection brings it.
func (s *Server) connectionPRAt(url string) (github.PRListItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, pr := range s.prs {
		if pr.Provider != "" && pr.URL == url {
			return pr, true
		}
	}
	return github.PRListItem{}, false
}

// autoMergeRefusals say why the dashboard offers no auto-merge for a connection provider's pull
// requests.
var autoMergeRefusals = map[string]string{
	types.PRProviderGitLab: "auto-merge isn't offered for GitLab pull requests from the dashboard",
}

// mergeConnectionPR completes pr, the head that was reviewed, through its remote with the method
// asked; auto-merge has the provider complete it once the base's policies pass (Azure DevOps
// auto-complete).
func (s *Server) mergeConnectionPR(w http.ResponseWriter, r *http.Request, pr github.PRListItem, input model.MergeInput) {
	if refusal, refused := autoMergeRefusals[pr.Provider]; input.Auto && refused {
		writeJSONError(w, http.StatusBadRequest, errors.New(refusal))
		return
	}
	if s.prRemote == nil {
		writeJSONError(w, http.StatusBadGateway, fmt.Errorf("this dashboard has no connection to merge %s pull request %s#%d through", pr.Provider, pr.Repo, pr.Number))
		return
	}
	remote, err := s.prRemote(r.Context(), pr)
	if err == nil {
		options := git.MergePullRequestOptions{Method: git.MergeMethod(input.Method), HeadSHA: input.HeadSHA}
		merger, canMerge := remote.Provider.(git.PullRequestMerger)
		autoMerger, canAutoMerge := remote.Provider.(git.PullRequestAutoMerger)
		switch {
		case input.Auto && !canAutoMerge:
			err = fmt.Errorf("auto-merge isn't offered for %s pull requests", pr.Provider)
		case input.Auto:
			err = autoMerger.AutoMergePullRequest(r.Context(), remote.URL, pr.Number, options)
		case !canMerge:
			err = fmt.Errorf("%s pull requests can't be merged from the dashboard", pr.Provider)
		default:
			err = merger.MergePullRequest(r.Context(), remote.URL, pr.Number, options)
		}
	}
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, fmt.Errorf("merge pull request %s#%d: %w", pr.Repo, pr.Number, err))
		return
	}
	s.afterPRAction(pr.Repo, pr.Number)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"ok"}`)
}

// approveConnectionPR approves pr through its remote as the connection's identity, leaving comment
// when there is one.
func (s *Server) approveConnectionPR(w http.ResponseWriter, r *http.Request, pr github.PRListItem, comment string) {
	if s.prRemote == nil {
		writeJSONError(w, http.StatusBadGateway, fmt.Errorf("this dashboard has no connection to approve %s pull request %s#%d through", pr.Provider, pr.Repo, pr.Number))
		return
	}
	remote, err := s.prRemote(r.Context(), pr)
	if err == nil {
		approver, ok := remote.Provider.(git.PullRequestApprover)
		if !ok {
			err = fmt.Errorf("%s pull requests can't be approved from the dashboard", pr.Provider)
		} else {
			err = approver.ApprovePullRequest(r.Context(), remote.URL, pr.Number, comment)
		}
	}
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, fmt.Errorf("approve pull request %s#%d: %w", pr.Repo, pr.Number, err))
		return
	}
	s.afterPRAction(pr.Repo, pr.Number)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"ok"}`)
}

// fetchConnectionPRDetail reads pr's detail through its remote, with Error set on failure as
// fetchPRDetail's callers expect.
func (s *Server) fetchConnectionPRDetail(ctx context.Context, pr github.PRListItem) prDetail {
	if s.prRemote == nil {
		return prDetail{Error: fmt.Sprintf("this dashboard has no connection to read %s pull request %s#%d through", pr.Provider, pr.Repo, pr.Number)}
	}
	ctx, cancel := context.WithTimeout(ctx, connectionDetailTimeout)
	defer cancel()
	remote, err := s.prRemote(ctx, pr)
	if err == nil {
		var detail prDetail
		if detail, err = connectionPRDetail(ctx, remote, pr.Number, s.gitCache); err == nil {
			return detail
		}
	}
	logger.Warnf("failed to fetch PR %s#%d: %v", pr.Repo, pr.Number, err)
	return prDetail{Error: fmt.Sprintf("pull request %s#%d: %v", pr.Repo, pr.Number, err)}
}

// connectionPRDetail reads pull request number of remote through its provider, with its merge
// readiness, threads and CI (the last lines of failed steps' logs), and with cache, its commits
// and changed files from git.
func connectionPRDetail(ctx context.Context, remote *git.Repository, number int, cache *git.Cache) (prDetail, error) {
	status, err := provider.NewConnectionClient(remote, "").Status(ctx, number, model.StatusOptions{Threads: true, Logs: true, TailLogs: connectionLogTail})
	if err != nil {
		return prDetail{}, err
	}
	info := status.PR
	// Every comment on such a pull request is in a thread, which carries its own state.
	info.Comments = info.ReviewThreads
	pr, err := remote.Provider.GetPullRequest(ctx, remote.URL, number)
	if err != nil {
		return prDetail{}, fmt.Errorf("read pull request %d: %w", number, err)
	}
	if cache != nil && !fromFork(pr, remote.URL) {
		if err := connectionChanges(ctx, cache, remote, info); err != nil {
			return prDetail{}, fmt.Errorf("read pull request %d's changes: %w", number, err)
		}
	}
	comments := prwatch.MergeAndFilter(append([]model.PRComment(nil), info.Comments...), info.ReviewThreads)
	return prDetail{PR: info, Runs: status.Runs, Comments: comments}, nil
}

// fromFork reports whether pr's head is in another repository than repo, whose commits a cache of
// repo doesn't reach.
func fromFork(pr git.PullRequest, repo git.RepoURL) bool {
	return pr.HeadRepository != "" && !strings.EqualFold(pr.HeadRepository, repo.Project+"/"+repo.Name)
}
