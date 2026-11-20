// The PR client of a repository a git connection reaches, through commons-db's provider: its pull
// requests' status (with CI, threads and merge readiness), and opening, merging and resolving.
package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/azuredevops"
	"github.com/flanksource/gavel/pr/model"
)

var prStates = map[git.PullRequestState]string{git.PullRequestOpen: "OPEN", git.PullRequestMerged: "MERGED", git.PullRequestClosed: "CLOSED"}

var prMergeable = map[string]string{"mergeable": "MERGEABLE", "conflicts": "CONFLICTING"}

// PRState is a pull request state as the dashboard spells it: OPEN, MERGED or CLOSED.
func PRState(state git.PullRequestState) string { return prStates[state] }

// PRMergeable is a pull request's mergeability as the dashboard spells it: MERGEABLE, CONFLICTING
// or UNKNOWN.
func PRMergeable(mergeability string) string {
	if spelled, ok := prMergeable[mergeability]; ok {
		return spelled
	}
	return "UNKNOWN"
}

// kinds names each connection type's PR client kind, as callers of Client.Kind compare it.
var kinds = map[string]string{models.ConnectionTypeAzureDevops: "azuredevops", models.ConnectionTypeGithub: "github", models.ConnectionTypeGitlab: "gitlab"}

// NewConnectionClient is the PR client of remote, a repository a git connection reaches;
// defaultBranch is its default branch as the caller knows it, "" when it doesn't.
func NewConnectionClient(remote *git.Repository, defaultBranch string) Client {
	return connectionClient{remote: remote, defaultBranch: defaultBranch}
}

type connectionClient struct {
	remote        *git.Repository
	defaultBranch string
}

func (c connectionClient) Kind() string { return kinds[c.remote.URL.Provider] }

func (c connectionClient) Preflight(ctx context.Context) error { return c.remote.Provider.Verify(ctx) }

// DefaultBranch is the default branch the client was given, else the one its provider names.
func (c connectionClient) DefaultBranch(ctx context.Context) (string, error) {
	if c.defaultBranch != "" {
		return c.defaultBranch, nil
	}
	info, err := c.repository(ctx)
	if err != nil {
		return "", err
	}
	if info.DefaultBranch == "" {
		return "", fmt.Errorf("%s names no default branch for %s", c.Kind(), c.remote.URL.Redacted())
	}
	return info.DefaultBranch, nil
}

// repository is the client's repository as its provider describes it.
func (c connectionClient) repository(ctx context.Context) (git.RepositoryInfo, error) {
	reader, ok := c.remote.Provider.(git.RepositoryReader)
	if !ok {
		return git.RepositoryInfo{}, fmt.Errorf("%s doesn't describe its repositories to gavel", c.Kind())
	}
	info, err := reader.Repository(ctx, c.remote.URL)
	if err != nil {
		return git.RepositoryInfo{}, fmt.Errorf("read repository %s: %w", c.remote.URL.Redacted(), err)
	}
	return info, nil
}

func (c connectionClient) OpenPRs(ctx context.Context) ([]model.PRInfo, error) {
	pulls, err := c.remote.Provider.ListPullRequests(ctx, c.remote.URL, git.ListPullRequestsOptions{State: git.PullRequestOpen})
	if err != nil {
		return nil, err
	}
	infos := make([]model.PRInfo, 0, len(pulls))
	for _, pull := range pulls {
		infos = append(infos, c.info(pull))
	}
	return infos, nil
}

func (c connectionClient) CreatePR(ctx context.Context, in model.CreatePRInput) (*model.CreatePRResult, error) {
	if in.Base == "" {
		base, err := c.DefaultBranch(ctx)
		if err != nil {
			return nil, err
		}
		in.Base = base
	}
	pull, err := c.remote.Provider.OpenPullRequest(ctx, c.remote.URL, git.PullRequestRequest{Title: in.Title, Body: in.Body, Head: in.Head, Base: in.Base, Draft: in.Draft})
	if err != nil {
		return nil, err
	}
	return &model.CreatePRResult{Number: pull.Number, URL: pull.URL, Title: pull.Title, State: PRState(pull.State), Base: pull.Base}, nil
}

func (c connectionClient) MergePR(ctx context.Context, number int, in model.MergeInput) error {
	if in.Auto {
		return fmt.Errorf("auto-merge isn't offered for %s pull requests: completing one waits on its policies already", c.Kind())
	}
	merger, ok := c.remote.Provider.(git.PullRequestMerger)
	if !ok {
		return fmt.Errorf("%s pull requests can't be merged through gavel", c.Kind())
	}
	return merger.MergePullRequest(ctx, c.remote.URL, number, git.MergePullRequestOptions{Method: git.MergeMethod(in.Method), HeadSHA: in.HeadSHA})
}

func (c connectionClient) ResolveThread(ctx context.Context, number int, thread string) error {
	writer, ok := c.remote.Provider.(git.PullRequestThreadWriter)
	if !ok {
		return fmt.Errorf("%s pull request threads can't be resolved through gavel", c.Kind())
	}
	current, err := writer.PullRequestThread(ctx, c.remote.URL, number, thread)
	if err != nil {
		return err
	}
	if current.Resolved {
		return nil
	}
	_, err = writer.ResolvePullRequestThread(ctx, c.remote.URL, number, thread, true)
	return err
}

// info is pull as the dashboard's PR model holds it, before its CI, threads and readiness.
func (c connectionClient) info(pull git.PullRequest) model.PRInfo {
	info := model.PRInfo{Provider: c.Kind(), Number: pull.Number, Title: pull.Title, Body: pull.Body, State: PRState(pull.State),
		Mergeable: "UNKNOWN", HeadRefName: pull.Head, BaseRefName: pull.Base, HeadRefOID: pull.HeadSHA, Author: model.PRAuthor{Name: pull.Author}, URL: pull.URL}
	if metadata := pull.Metadata; metadata != nil {
		info.IsDraft, info.Author.AvatarURL, info.BaseRefOID = metadata.Draft, metadata.AuthorAvatarURL, metadata.BaseSHA
		info.Mergeable = PRMergeable(metadata.Mergeability)
	}
	return info
}

// Status is pull request number with its merge readiness, its threads when asked, and the CI its
// pipelines ran: a check per job, and when asked the log tails of failed steps.
func (c connectionClient) Status(ctx context.Context, number int, opts model.StatusOptions) (*model.StatusSnapshot, error) {
	if opts.TailLogs < 0 {
		return nil, fmt.Errorf("--tail-logs must not be negative")
	}
	provider, repo := c.remote.Provider, c.remote.URL
	pull, err := provider.GetPullRequest(ctx, repo, number)
	if err != nil {
		return nil, fmt.Errorf("read pull request %d: %w", number, err)
	}
	info := c.info(pull)
	if reader, ok := provider.(git.PullRequestReadinessReader); ok {
		readiness, err := reader.PullRequestReadiness(ctx, repo, number)
		if err != nil {
			return nil, fmt.Errorf("read pull request %d's merge readiness: %w", number, err)
		}
		if readiness.State != "" {
			info.MergeReadiness = &model.MergeReadiness{State: readiness.State, Reasons: readiness.Reasons,
				WaitingForBuilds: readiness.WaitingForBuilds, WaitingForMerge: readiness.WaitingForMerge}
		}
	}
	if reader, ok := provider.(git.PullRequestThreadReader); ok && opts.Threads {
		if info.ReviewThreads, err = threadComments(ctx, reader, repo, number); err != nil {
			return nil, err
		}
	}
	runs := map[int64]*model.WorkflowRun{}
	if reader, ok := provider.(git.PullRequestPipelineReader); ok {
		if runs, err = pipelineRuns(ctx, reader, repo, number, opts, &info); err != nil {
			return nil, err
		}
	}
	return &model.StatusSnapshot{PR: &info, Runs: runs}, nil
}

// threadComments reads pull request number's threads as review thread comments.
func threadComments(ctx context.Context, reader git.PullRequestThreadReader, repo git.RepoURL, number int) ([]model.PRComment, error) {
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
				IsResolved: thread.Resolved, IsReviewThread: true, CanResolve: true})
		}
	}
	return comments, nil
}

// pipelineRuns reads the CI pull request number ran as workflow runs, adding a check for each
// job, or for a run without jobs, to info's rollup.
func pipelineRuns(ctx context.Context, reader git.PullRequestPipelineReader, repo git.RepoURL, number int, opts model.StatusOptions, info *model.PRInfo) (map[int64]*model.WorkflowRun, error) {
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
			converted, err := pipelineJob(ctx, reader, execution, job, opts)
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

// pipelineJob is job as the PR model holds it, with, when opts asks for logs, the log tails of its
// failed steps, or its own when it failed without steps.
func pipelineJob(ctx context.Context, reader git.PullRequestPipelineReader, execution git.PullRequestExecution, job git.PipelineJob, opts model.StatusOptions) (model.Job, error) {
	status, conclusion, err := azuredevops.ExecutionState(job.Status, job.Result)
	if err != nil {
		return model.Job{}, fmt.Errorf("job %s: %w", job.ID, err)
	}
	converted := model.Job{NativeID: job.ID, Name: job.Name, Status: status, Conclusion: conclusion, StartedAt: job.StartedAt,
		CompletedAt: job.FinishedAt, URL: job.URL}
	tail := func(log string) (string, error) {
		logs, err := reader.PipelineLog(ctx, execution.Repository, execution.ID, log, opts.TailLogs)
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
		if opts.Logs && model.NeedsJobLogs(conclusion) && step.LogID != "" {
			if converted.Steps[len(converted.Steps)-1].Logs, err = tail(step.LogID); err != nil {
				return model.Job{}, err
			}
		}
	}
	if opts.Logs && model.NeedsJobLogs(conclusion) && len(job.Steps) == 0 && job.LogID != "" {
		if converted.Logs, err = tail(job.LogID); err != nil {
			return model.Job{}, err
		}
	}
	return converted, nil
}
