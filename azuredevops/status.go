package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/pr/model"
)

func (c *Client) Status(ctx context.Context, number int, opts model.StatusOptions) (*model.StatusSnapshot, error) {
	if number <= 0 {
		return nil, fmt.Errorf("azure PR number must be positive")
	}
	if opts.TailLogs < 0 {
		return nil, fmt.Errorf("--tail-logs must not be negative")
	}
	if err := c.Preflight(ctx); err != nil {
		return nil, err
	}
	data, _, err := c.request(ctx, http.MethodGet, c.repositoryPath()+"/pullrequests/"+strconv.Itoa(number), nil, nil)
	if err != nil {
		return nil, err
	}
	var pr pullRequest
	if err = json.Unmarshal(data, &pr); err != nil {
		return nil, fmt.Errorf("decode Azure PR: %w", err)
	}
	if pr.ID != number || !strings.HasPrefix(pr.Source, "refs/heads/") || !strings.HasPrefix(pr.Target, "refs/heads/") {
		return nil, fmt.Errorf("azure PR response has invalid identity or branch refs")
	}
	info, err := c.prInfo(pr)
	if err != nil {
		return nil, err
	}
	if pr.Status == "active" {
		policies, err := collection[policyEvaluation](ctx, c, c.apiPath("policy/evaluations"), url.Values{"artifactId": {fmt.Sprintf("vstfs:///CodeReview/CodeReviewId/%s/%d", c.info.Project.ID, number)}, "api-version": {"7.1-preview.1"}})
		if err != nil {
			return nil, err
		}
		info.MergeReadiness = readiness(pr, policies)
		if err := c.refreshHead(ctx, &pr, info); err != nil {
			return nil, err
		}
	}
	runs, err := c.pipelineRuns(ctx, pr, opts)
	if err != nil {
		return nil, err
	}
	for _, b := range runs {
		if len(b.Jobs) == 0 {
			info.StatusCheckRollup = append(info.StatusCheckRollup, model.StatusCheck{Name: b.Name, WorkflowName: b.Name, RunID: b.DatabaseID, Status: b.Status, Conclusion: b.Conclusion, DetailsURL: b.URL})
		}
		for _, j := range b.Jobs {
			info.StatusCheckRollup = append(info.StatusCheckRollup, model.StatusCheck{Name: j.Name, WorkflowName: b.Name, RunID: b.DatabaseID, JobID: j.NativeID, Status: j.Status, Conclusion: j.Conclusion, DetailsURL: j.URL})
		}
	}
	sort.Slice(info.StatusCheckRollup, func(i, j int) bool {
		a, b := info.StatusCheckRollup[i], info.StatusCheckRollup[j]
		if a.WorkflowName != b.WorkflowName {
			return a.WorkflowName < b.WorkflowName
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.JobID < b.JobID
	})
	return &model.StatusSnapshot{PR: info, Runs: runs}, nil
}

func (c *Client) refreshHead(ctx context.Context, pr *pullRequest, info *model.PRInfo) error {
	source, err := c.sourceRepository(*pr)
	if err != nil {
		return err
	}
	head, err := c.refOID(ctx, c.projectPath(source.Project.Name, "git/repositories/"+url.PathEscape(source.ID)), pr.Source)
	if err != nil {
		return err
	}
	stale := head != pr.SourceCommit.ID
	pr.SourceCommit.ID, info.HeadRefOID = head, head
	if pr.TargetCommit.ID != "" {
		base, err := c.refOID(ctx, c.repositoryPath(), pr.Target)
		if err != nil {
			return err
		}
		stale = stale || base != pr.TargetCommit.ID
		pr.TargetCommit.ID, info.BaseRefOID = base, base
	}
	if stale {
		pr.MergeCommit.ID = ""
		info.MergeReadiness.State = "pending"
		info.MergeReadiness.WaitingForMerge = true
		info.MergeReadiness.Reasons = append(info.MergeReadiness.Reasons, "merge computation has not caught up to the source or target branch")
	}
	return nil
}

func (c *Client) sourceRepository(pr pullRequest) (repositoryInfo, error) {
	if pr.ForkSource == nil {
		return *c.info, nil
	}
	source := pr.ForkSource.Repository
	if source.ID == "" || source.Project.Name == "" {
		return repositoryInfo{}, fmt.Errorf("azure fork source repository is missing identity or project")
	}
	return source, nil
}

func (c *Client) refOID(ctx context.Context, path, branch string) (string, error) {
	refs, err := collection[struct {
		Name string `json:"name"`
		OID  string `json:"objectId"`
	}](ctx, c, path+"/refs", url.Values{"filter": {strings.TrimPrefix(branch, "refs/")}})
	if err != nil {
		return "", err
	}
	for _, ref := range refs {
		if ref.Name != branch {
			continue
		}
		if ref.OID == "" {
			return "", fmt.Errorf("azure branch %q has no objectId", branch)
		}
		return ref.OID, nil
	}
	return "", fmt.Errorf("azure branch %q was not found", branch)
}

func (c *Client) pipelineRuns(ctx context.Context, pr pullRequest, opts model.StatusOptions) (map[int64]*model.WorkflowRun, error) {
	var builds []build
	source, err := c.sourceRepository(pr)
	if err != nil {
		return nil, err
	}
	projects := map[int64]string{}
	queries := []struct {
		repository repositoryInfo
		branch     string
	}{{*c.info, fmt.Sprintf("refs/pull/%d/merge", pr.ID)}, {source, pr.Source}}
	for _, query := range queries {
		items, err := collection[build](ctx, c, c.projectPath(query.repository.Project.Name, "build/builds"), url.Values{"repositoryId": {query.repository.ID}, "repositoryType": {"TfsGit"}, "branchName": {query.branch}, "queryOrder": {"queueTimeDescending"}})
		if err != nil {
			return nil, err
		}
		for _, b := range items {
			if b.SourceBranch != query.branch {
				continue
			}
			builds = append(builds, b)
			projects[b.ID] = query.repository.Project.Name
		}
	}
	runs := make(map[int64]*model.WorkflowRun)
	for _, b := range currentBuilds(pr, builds) {
		if b.ID <= 0 || b.Definition.ID <= 0 || b.Definition.Name == "" {
			return nil, fmt.Errorf("azure build has invalid identity or pipeline definition")
		}
		status, conclusion, err := executionState(b.Status, b.Result)
		if err != nil {
			return nil, fmt.Errorf("build %d: %w", b.ID, err)
		}
		run := &model.WorkflowRun{DatabaseID: b.ID, WorkflowID: b.Definition.ID, Name: b.Definition.Name, Status: status, Conclusion: conclusion, HeadSHA: b.SourceVersion, URL: fmt.Sprintf("https://dev.azure.com/%s/%s/_build/results?buildId=%d", url.PathEscape(c.repo.Organization), url.PathEscape(projects[b.ID]), b.ID)}
		if b.Status != "notStarted" && b.Status != "postponed" {
			if err := c.loadTimeline(ctx, run, timelineOptions{StatusOptions: opts, Project: projects[b.ID]}); err != nil {
				return nil, err
			}
		}
		runs[b.ID] = run
	}
	return runs, nil
}
