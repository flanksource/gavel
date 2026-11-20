// Reads the dashboard detail of a pull request a git connection's repository holds through that
// connection's provider: the PR, the CI its pipelines ran with failed steps' log tails, comments.
package ui

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/azuredevops"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/prwatch"
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

// mergeConnectionPR completes pr, the head that was reviewed, through its remote with the method
// asked. Its provider queues completion behind the base's policies itself, so there is no
// separate auto-merge.
func (s *Server) mergeConnectionPR(w http.ResponseWriter, r *http.Request, pr github.PRListItem, input model.MergeInput) {
	if input.Auto {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("auto-merge isn't offered for Azure DevOps pull requests: completing one waits on its policies already"))
		return
	}
	if s.prRemote == nil {
		writeJSONError(w, http.StatusBadGateway, fmt.Errorf("this dashboard has no connection to merge %s pull request %s#%d through", pr.Provider, pr.Repo, pr.Number))
		return
	}
	remote, err := s.prRemote(r.Context(), pr)
	if err == nil {
		merger, ok := remote.Provider.(git.PullRequestMerger)
		if !ok {
			err = fmt.Errorf("%s pull requests can't be merged from the dashboard", pr.Provider)
		} else {
			err = merger.MergePullRequest(r.Context(), remote.URL, pr.Number,
				git.MergePullRequestOptions{Method: git.MergeMethod(input.Method), HeadSHA: input.HeadSHA})
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

// connectionPRDetail reads pull request number of remote through its provider, and with cache, its
// commits and changed files from git. Azure DevOps is the provider whose CI it reads; its
// executions' states take the dashboard's GitHub spelling.
func connectionPRDetail(ctx context.Context, remote *git.Repository, number int, cache *git.Cache) (prDetail, error) {
	provider, repo := remote.Provider, remote.URL
	pr, err := provider.GetPullRequest(ctx, repo, number)
	if err != nil {
		return prDetail{}, fmt.Errorf("read pull request %d: %w", number, err)
	}
	metadata, err := git.ReadPullRequestMetadata(ctx, provider, repo, git.PullRequestMetadataOptions{Number: number, Listed: &pr, SkipChecks: true})
	if err != nil {
		return prDetail{}, fmt.Errorf("read pull request %d's metadata: %w", number, err)
	}
	info := &model.PRInfo{Provider: "azuredevops", Number: pr.Number, Title: pr.Title, Body: pr.Body, State: prStates[pr.State],
		IsDraft: metadata.Draft, Mergeable: "UNKNOWN", HeadRefName: pr.Head, BaseRefName: pr.Base, HeadRefOID: pr.HeadSHA,
		Author: model.PRAuthor{Name: pr.Author, AvatarURL: metadata.AuthorAvatarURL}, URL: pr.URL}
	if mergeable, ok := prMergeable[metadata.Mergeability]; ok {
		info.Mergeable = mergeable
	}
	if reader, ok := provider.(git.PullRequestThreadReader); ok {
		if info.ReviewThreads, err = connectionThreads(ctx, reader, repo, number); err != nil {
			return prDetail{}, err
		}
		// Every comment on such a pull request is in a thread, which carries its own state.
		info.Comments = info.ReviewThreads
	}
	runs := map[int64]*model.WorkflowRun{}
	if reader, ok := provider.(git.PullRequestPipelineReader); ok {
		if runs, err = connectionRuns(ctx, reader, repo, number, info); err != nil {
			return prDetail{}, err
		}
	}
	if cache != nil && !fromFork(pr, repo) {
		if err := connectionChanges(ctx, cache, remote, info); err != nil {
			return prDetail{}, fmt.Errorf("read pull request %d's changes: %w", number, err)
		}
	}
	comments := prwatch.MergeAndFilter(append([]model.PRComment(nil), info.Comments...), info.ReviewThreads)
	return prDetail{PR: info, Runs: runs, Comments: comments}, nil
}

// fromFork reports whether pr's head is in another repository than repo, whose commits a cache of
// repo doesn't reach.
func fromFork(pr git.PullRequest, repo git.RepoURL) bool {
	return pr.HeadRepository != "" && !strings.EqualFold(pr.HeadRepository, repo.Project+"/"+repo.Name)
}

// connectionThreads reads pull request number's threads as review thread comments.
func connectionThreads(ctx context.Context, reader git.PullRequestThreadReader, repo git.RepoURL, number int) ([]model.PRComment, error) {
	threads, err := reader.PullRequestThreads(ctx, repo, number)
	if err != nil {
		return nil, fmt.Errorf("read pull request %d's threads: %w", number, err)
	}
	var comments []model.PRComment
	for _, thread := range threads {
		for _, comment := range thread.Comments {
			id, err := strconv.ParseInt(comment.ID, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("thread %s comment ID %q: %w", thread.ID, comment.ID, err)
			}
			comments = append(comments, model.PRComment{ThreadID: thread.ID, ID: id, Body: comment.Body, Author: comment.Author,
				AvatarURL: comment.AuthorAvatarURL, URL: thread.URL, CreatedAt: comment.CreatedAt, Path: thread.Path, Line: thread.Line,
				IsResolved: thread.Resolved, IsReviewThread: true})
		}
	}
	return comments, nil
}

// connectionRuns reads the CI pull request number ran as workflow runs, adding a check for each
// job, or for a run without jobs, to info's rollup.
func connectionRuns(ctx context.Context, reader git.PullRequestPipelineReader, repo git.RepoURL, number int, info *model.PRInfo) (map[int64]*model.WorkflowRun, error) {
	executions, err := reader.PullRequestExecutions(ctx, repo, number)
	if err != nil {
		return nil, fmt.Errorf("read pull request %d's CI: %w", number, err)
	}
	runs := make(map[int64]*model.WorkflowRun, len(executions))
	for _, execution := range executions {
		id, err := strconv.ParseInt(execution.ID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("pipeline execution ID %q: %w", execution.ID, err)
		}
		pipeline, _ := strconv.ParseInt(execution.Pipeline, 10, 64)
		status, conclusion, err := azuredevops.ExecutionState(execution.Status, execution.Result)
		if err != nil {
			return nil, fmt.Errorf("pipeline execution %s: %w", execution.ID, err)
		}
		run := &model.WorkflowRun{DatabaseID: id, WorkflowID: pipeline, Name: execution.PipelineName, Status: status, Conclusion: conclusion,
			URL: execution.URL, HeadSHA: execution.SHA}
		for _, job := range execution.Jobs {
			converted, err := connectionJob(ctx, reader, execution, job)
			if err != nil {
				return nil, err
			}
			run.Jobs = append(run.Jobs, converted)
			info.StatusCheckRollup = append(info.StatusCheckRollup, model.StatusCheck{RunID: id, JobID: job.ID, Name: job.Name,
				WorkflowName: run.Name, Status: converted.Status, Conclusion: converted.Conclusion, DetailsURL: job.URL})
		}
		if len(execution.Jobs) == 0 {
			info.StatusCheckRollup = append(info.StatusCheckRollup, model.StatusCheck{RunID: id, Name: run.Name, WorkflowName: run.Name,
				Status: status, Conclusion: conclusion, DetailsURL: run.URL})
		}
		runs[id] = run
	}
	return runs, nil
}

// connectionJob is job as the dashboard shows it, with the log tails of its failed steps, or its
// own when it failed without steps.
func connectionJob(ctx context.Context, reader git.PullRequestPipelineReader, execution git.PullRequestExecution, job git.PipelineJob) (model.Job, error) {
	status, conclusion, err := azuredevops.ExecutionState(job.Status, job.Result)
	if err != nil {
		return model.Job{}, fmt.Errorf("job %s: %w", job.ID, err)
	}
	converted := model.Job{NativeID: job.ID, Name: job.Name, Status: status, Conclusion: conclusion, StartedAt: job.StartedAt,
		CompletedAt: job.FinishedAt, URL: job.URL}
	tail := func(log string) (string, error) {
		logs, err := reader.PipelineLog(ctx, execution.Repository, execution.ID, log, connectionLogTail)
		if err != nil {
			return "", fmt.Errorf("read the log of job %s: %w", job.Name, err)
		}
		return logs, nil
	}
	for _, step := range job.Steps {
		status, conclusion, err := azuredevops.ExecutionState(step.Status, step.Result)
		if err != nil {
			return model.Job{}, fmt.Errorf("step %s: %w", step.Name, err)
		}
		converted.Steps = append(converted.Steps, model.Step{Name: step.Name, Number: step.Number, Status: status, Conclusion: conclusion})
		if model.NeedsJobLogs(conclusion) && step.LogID != "" {
			if converted.Steps[len(converted.Steps)-1].Logs, err = tail(step.LogID); err != nil {
				return model.Job{}, err
			}
		}
	}
	if model.NeedsJobLogs(conclusion) && len(job.Steps) == 0 && job.LogID != "" {
		if converted.Logs, err = tail(job.LogID); err != nil {
			return model.Job{}, err
		}
	}
	return converted, nil
}
