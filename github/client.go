package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons/http"
	"github.com/flanksource/commons/logger"
)

type Options struct {
	WorkDir string
	Repo    string // owner/repo
	Token   string // optional; falls back to GITHUB_TOKEN then GH_TOKEN env
}

func (o Options) token() (string, error) {
	if o.Token != "" {
		return o.Token, nil
	}
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t, nil
	}
	if t := os.Getenv("GH_TOKEN"); t != "" {
		return t, nil
	}
	return "", fmt.Errorf("no GitHub token: set Options.Token, GITHUB_TOKEN, or GH_TOKEN")
}

func (o Options) resolveRepo() (string, error) {
	if o.Repo != "" {
		return o.Repo, nil
	}
	cmd := exec.Command("git", "remote", "get-url", "origin")
	if o.WorkDir != "" {
		cmd.Dir = o.WorkDir
	}
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git remote get-url origin: %w", err)
	}
	return parseGitHubRepo(strings.TrimSpace(string(out)))
}

func (o Options) currentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	if o.WorkDir != "" {
		cmd.Dir = o.WorkDir
	}
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch --show-current: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

var repoPatterns = []*regexp.Regexp{
	regexp.MustCompile(`github\.com[:/]([^/]+/[^/.]+?)(?:\.git)?$`),
}

func parseGitHubRepo(remoteURL string) (string, error) {
	for _, re := range repoPatterns {
		if m := re.FindStringSubmatch(remoteURL); len(m) >= 2 {
			return m[1], nil
		}
	}
	return "", fmt.Errorf("cannot parse GitHub owner/repo from %q", remoteURL)
}

func newClient(token string) *http.Client {
	return http.NewClient().
		BaseURL("https://api.github.com").
		Header("Authorization", "Bearer "+token).
		Header("Accept", "application/vnd.github+json")
}

// REST response types for /actions endpoints (snake_case JSON)

type restRun struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HTMLURL    string `json:"html_url"`
}

func (r restRun) toWorkflowRun(jobs []Job) *WorkflowRun {
	return &WorkflowRun{
		DatabaseID: r.ID,
		Name:       r.Name,
		Status:     r.Status,
		Conclusion: r.Conclusion,
		URL:        r.HTMLURL,
		Jobs:       jobs,
	}
}

type restJobsResponse struct {
	Jobs []restJob `json:"jobs"`
}

type restJob struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Conclusion  string     `json:"conclusion"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt time.Time  `json:"completed_at"`
	HTMLURL     string     `json:"html_url"`
	Steps       []restStep `json:"steps"`
}

func (j restJob) toJob() Job {
	steps := make([]Step, len(j.Steps))
	for i, s := range j.Steps {
		steps[i] = s.toStep()
	}
	return Job{
		DatabaseID:  j.ID,
		Name:        j.Name,
		Status:      j.Status,
		Conclusion:  j.Conclusion,
		StartedAt:   j.StartedAt,
		CompletedAt: j.CompletedAt,
		URL:         j.HTMLURL,
		Steps:       steps,
	}
}

type restStep struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Number     int    `json:"number"`
}

func (s restStep) toStep() Step {
	return Step{
		Name:       s.Name,
		Status:     s.Status,
		Conclusion: s.Conclusion,
		Number:     s.Number,
	}
}

func FetchRunJobs(opts Options, runID int64) (*WorkflowRun, error) {
	token, err := opts.token()
	if err != nil {
		return nil, err
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	client := newClient(token)

	runURL := fmt.Sprintf("/repos/%s/actions/runs/%d", repo, runID)
	logger.Tracef("fetching run: GET %s", runURL)
	resp, err := client.R(ctx).Get(runURL)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", runURL, err)
	}
	if !resp.IsOK() {
		body, _ := resp.AsString()
		return nil, fmt.Errorf("GET %s: status %d: %s", runURL, resp.StatusCode, body)
	}
	var run restRun
	if err := resp.Into(&run); err != nil {
		return nil, fmt.Errorf("parse run response: %w", err)
	}

	jobsURL := fmt.Sprintf("/repos/%s/actions/runs/%d/jobs", repo, runID)
	logger.Tracef("fetching jobs: GET %s", jobsURL)
	resp, err = client.R(ctx).Get(jobsURL)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", jobsURL, err)
	}
	if !resp.IsOK() {
		body, _ := resp.AsString()
		return nil, fmt.Errorf("GET %s: status %d: %s", jobsURL, resp.StatusCode, body)
	}
	var jobsResp restJobsResponse
	if err := resp.Into(&jobsResp); err != nil {
		return nil, fmt.Errorf("parse jobs response: %w", err)
	}

	jobs := make([]Job, len(jobsResp.Jobs))
	for i, j := range jobsResp.Jobs {
		jobs[i] = j.toJob()
	}

	result := run.toWorkflowRun(jobs)
	logger.Debugf("fetched run %d %q: status=%s conclusion=%s jobs=%d", result.DatabaseID, result.Name, result.Status, result.Conclusion, len(result.Jobs))
	return result, nil
}

func FetchAndAttachLogs(opts Options, run *WorkflowRun, tailLines int) {
	token, err := opts.token()
	if err != nil {
		logger.Warnf("cannot fetch logs: %v", err)
		return
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		logger.Warnf("cannot resolve repo for logs: %v", err)
		return
	}

	ctx := context.Background()
	client := newClient(token)

	for i := range run.Jobs {
		job := &run.Jobs[i]
		if !strings.EqualFold(job.Conclusion, "failure") {
			continue
		}
		endpoint := fmt.Sprintf("/repos/%s/actions/jobs/%d/logs", repo, job.DatabaseID)
		logger.Tracef("fetching job logs: GET %s", endpoint)
		resp, err := client.R(ctx).Get(endpoint)
		if err != nil {
			logger.Warnf("failed to fetch logs for job %d (%s): %v", job.DatabaseID, job.Name, err)
			continue
		}
		if !resp.IsOK() {
			logger.Warnf("failed to fetch logs for job %d (%s): status %d", job.DatabaseID, job.Name, resp.StatusCode)
			continue
		}
		body, err := resp.AsString()
		if err != nil {
			logger.Warnf("failed to read logs for job %d (%s): %v", job.DatabaseID, job.Name, err)
			continue
		}
		attachLogsToSteps(job, body, tailLines)
	}
}

func attachLogsToSteps(job *Job, rawLog string, tailLines int) {
	job.Logs = tailString(rawLog, tailLines)
	sections := parseLogSections(rawLog)
	for i := range job.Steps {
		step := &job.Steps[i]
		if !strings.EqualFold(step.Conclusion, "failure") {
			continue
		}
		if logs, ok := sections[step.Name]; ok {
			step.Logs = tailString(logs, tailLines)
		}
	}
}

var logTimestampPrefix = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T[\d:.]+Z\s*`)

func parseLogSections(raw string) map[string]string {
	sections := make(map[string]string)
	var currentStep string
	var buf strings.Builder

	for _, line := range strings.Split(raw, "\n") {
		cleaned := logTimestampPrefix.ReplaceAllString(line, "")
		if strings.HasPrefix(cleaned, "##[group]") {
			if currentStep != "" {
				sections[currentStep] = buf.String()
			}
			currentStep = strings.TrimPrefix(cleaned, "##[group]")
			buf.Reset()
			continue
		}
		if strings.HasPrefix(cleaned, "##[endgroup]") {
			if currentStep != "" {
				sections[currentStep] = buf.String()
				currentStep = ""
				buf.Reset()
			}
			continue
		}
		if currentStep != "" {
			if buf.Len() > 0 {
				buf.WriteByte('\n')
			}
			buf.WriteString(cleaned)
		}
	}
	if currentStep != "" {
		sections[currentStep] = buf.String()
	}
	return sections
}

func tailString(s string, maxLines int) string {
	if maxLines <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[len(lines)-maxLines:], "\n")
}

func ParsePRJSON(data []byte) (*PRInfo, error) {
	var pr PRInfo
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, fmt.Errorf("parse PR JSON: %w", err)
	}
	return &pr, nil
}

func ParseRunJSON(data []byte) (*WorkflowRun, error) {
	var run WorkflowRun
	if err := json.Unmarshal(data, &run); err != nil {
		return nil, fmt.Errorf("parse run JSON: %w", err)
	}
	return &run, nil
}

var runIDRegexp = regexp.MustCompile(`/actions/runs/(\d+)`)

func ExtractRunID(detailsURL string) (int64, error) {
	matches := runIDRegexp.FindStringSubmatch(detailsURL)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no run ID found in URL: %s", detailsURL)
	}
	return strconv.ParseInt(matches[1], 10, 64)
}
