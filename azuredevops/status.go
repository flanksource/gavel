package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/flanksource/gavel/pr/model"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

func (c *Client) Status(ctx context.Context, number int, opts model.StatusOptions) (*model.StatusSnapshot, error) {
	if number <= 0 {
		return nil, fmt.Errorf("Azure PR number must be positive")
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
		return nil, fmt.Errorf("Azure PR response has invalid identity or branch refs")
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
	refs, err := collection[struct {
		Name string `json:"name"`
		OID  string `json:"objectId"`
	}](ctx, c, c.repositoryPath()+"/refs", url.Values{"filter": {strings.TrimPrefix(pr.Source, "refs/")}})
	if err != nil {
		return err
	}
	for _, ref := range refs {
		if ref.Name != pr.Source {
			continue
		}
		if ref.OID == "" {
			return fmt.Errorf("Azure source ref has no objectId")
		}
		if ref.OID != pr.SourceCommit.ID {
			pr.SourceCommit.ID = ref.OID
			pr.MergeCommit.ID = ""
			info.HeadRefOID = ref.OID
			info.MergeReadiness.State = "pending"
			info.MergeReadiness.WaitingForMerge = true
			info.MergeReadiness.Reasons = append(info.MergeReadiness.Reasons, "merge computation has not caught up to the source branch")
		}
		return nil
	}
	return fmt.Errorf("Azure source branch %q was not found", pr.Source)
}

func (c *Client) pipelineRuns(ctx context.Context, pr pullRequest, opts model.StatusOptions) (map[int64]*model.WorkflowRun, error) {
	var builds []build
	for _, branch := range []string{fmt.Sprintf("refs/pull/%d/merge", pr.ID), pr.Source} {
		items, err := collection[build](ctx, c, c.apiPath("build/builds"), url.Values{"repositoryId": {c.info.ID}, "repositoryType": {"TfsGit"}, "branchName": {branch}, "queryOrder": {"queueTimeDescending"}})
		if err != nil {
			return nil, err
		}
		builds = append(builds, items...)
	}
	runs := make(map[int64]*model.WorkflowRun)
	for _, b := range currentBuilds(pr, builds) {
		if b.ID <= 0 || b.Definition.ID <= 0 || b.Definition.Name == "" {
			return nil, fmt.Errorf("Azure build has invalid identity or pipeline definition")
		}
		status, conclusion, err := executionState(b.Status, b.Result)
		if err != nil {
			return nil, fmt.Errorf("build %d: %w", b.ID, err)
		}
		run := &model.WorkflowRun{DatabaseID: b.ID, WorkflowID: b.Definition.ID, Name: b.Definition.Name, Status: status, Conclusion: conclusion, HeadSHA: b.SourceVersion, URL: fmt.Sprintf("https://dev.azure.com/%s/%s/_build/results?buildId=%d", url.PathEscape(c.repo.Organization), url.PathEscape(c.repo.Project), b.ID)}
		if b.Status != "notStarted" && b.Status != "postponed" {
			if err := c.loadTimeline(ctx, run, opts); err != nil {
				return nil, err
			}
		}
		runs[b.ID] = run
	}
	return runs, nil
}
