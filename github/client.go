package github

import (
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons/http"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/service"
)

type Options struct {
	WorkDir string
	Repo    string // owner/repo
	Token   string // optional; falls back to GITHUB_TOKEN then GH_TOKEN env
}

// ErrNoTokenMarker is a substring of the "no GitHub token" error returned
// by Options.token(). ProbeToken uses it to distinguish "token was never
// configured" (AuthStateNoToken) from other token() failures.
const ErrNoTokenMarker = "no GitHub token"

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
	// Last resort: ~/.config/gavel/auth.json written by `gavel system
	// install`. This is the path used by the launchd/systemd daemon, which
	// doesn't inherit the user's shell env.
	cfg, err := service.LoadAuthConfig()
	if err != nil {
		return "", fmt.Errorf("load auth config: %w", err)
	}
	if cfg.Token != "" {
		return cfg.Token, nil
	}
	return "", fmt.Errorf("no GitHub token: set Options.Token, GITHUB_TOKEN, GH_TOKEN, or run `gavel system install` to persist one")
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

func ResolveRepoFromDir(dir string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git remote get-url origin in %s: %w", dir, err)
	}
	return parseGitHubRepo(strings.TrimSpace(string(out)))
}

func parseGitHubRepo(remoteURL string) (string, error) {
	for _, re := range repoPatterns {
		if m := re.FindStringSubmatch(remoteURL); len(m) >= 2 {
			return m[1], nil
		}
	}
	return "", fmt.Errorf("cannot parse GitHub owner/repo from %q", remoteURL)
}

var prURLPattern = regexp.MustCompile(`(?:https?://)?github\.com/([^/]+/[^/]+)/pull/(\d+)`)
var repoURLPattern = regexp.MustCompile(`(?:https?://)?github\.com/([^/]+/[^/]+?)(?:\.git)?(?:/|$)`)

func ParsePRURL(url string) (repo string, prNumber int, err error) {
	if m := prURLPattern.FindStringSubmatch(url); len(m) >= 3 {
		n, _ := strconv.Atoi(m[2])
		return m[1], n, nil
	}
	return "", 0, fmt.Errorf("cannot parse PR URL: %q", url)
}

func ParseRepoURL(url string) (string, error) {
	if m := repoURLPattern.FindStringSubmatch(url); len(m) >= 2 {
		return m[1], nil
	}
	return "", fmt.Errorf("cannot parse repo URL: %q", url)
}

func newClient(token string) *http.Client {
	return http.NewClient().
		BaseURL("https://api.github.com").
		Header("Authorization", "Bearer "+token).
		Header("Accept", "application/vnd.github+json")
}

type RateLimit struct {
	Limit     int    `json:"limit"`
	Remaining int    `json:"remaining"`
	Used      int    `json:"used"`
	Reset     int64  `json:"reset"`
	Resource  string `json:"resource"`
}

func ParseRateLimit(header nethttp.Header) *RateLimit {
	remaining := header.Get("X-RateLimit-Remaining")
	if remaining == "" {
		return nil
	}
	rl := &RateLimit{Resource: header.Get("X-RateLimit-Resource")}
	rl.Remaining, _ = strconv.Atoi(remaining)
	rl.Limit, _ = strconv.Atoi(header.Get("X-RateLimit-Limit"))
	rl.Used, _ = strconv.Atoi(header.Get("X-RateLimit-Used"))
	rl.Reset, _ = strconv.ParseInt(header.Get("X-RateLimit-Reset"), 10, 64)
	return rl
}

// REST response types for /actions endpoints (snake_case JSON)

type restRun struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HTMLURL    string `json:"html_url"`
	HeadSHA    string `json:"head_sha"`
	WorkflowID int64  `json:"workflow_id"`
}

func (r restRun) toWorkflowRun(jobs []Job) *WorkflowRun {
	return &WorkflowRun{
		DatabaseID: r.ID,
		Name:       r.Name,
		Status:     r.Status,
		Conclusion: r.Conclusion,
		URL:        r.HTMLURL,
		HeadSHA:    r.HeadSHA,
		WorkflowID: r.WorkflowID,
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

	// Phase 3 short-circuit: completed runs are immutable, so once we've
	// stored one we never have to talk to GitHub again.
	if payload := cache.Shared().GetCompletedRunPayload(runID); payload != nil {
		var cached WorkflowRun
		if err := json.Unmarshal(payload, &cached); err == nil {
			logger.Tracef("github cache hit for completed run %d", runID)
			return &cached, nil
		}
		logger.Warnf("github cache: corrupt completed run %d, refetching", runID)
	}

	ctx := context.Background()

	runPath := fmt.Sprintf("/repos/%s/actions/runs/%d", repo, runID)
	runResp, err := cachedGet(ctx, token, runPath, nil)
	if err != nil {
		return nil, err
	}
	var run restRun
	if err := json.Unmarshal(runResp.Body, &run); err != nil {
		return nil, fmt.Errorf("parse run response: %w", err)
	}

	jobsPath := fmt.Sprintf("/repos/%s/actions/runs/%d/jobs", repo, runID)
	jobsResp, err := cachedGet(ctx, token, jobsPath, nil)
	if err != nil {
		return nil, err
	}
	var jobsPayload restJobsResponse
	if err := json.Unmarshal(jobsResp.Body, &jobsPayload); err != nil {
		return nil, fmt.Errorf("parse jobs response: %w", err)
	}

	jobs := make([]Job, len(jobsPayload.Jobs))
	for i, j := range jobsPayload.Jobs {
		jobs[i] = j.toJob()
	}

	result := run.toWorkflowRun(jobs)
	logger.Debugf("fetched run %d %q: status=%s conclusion=%s jobs=%d (cached=%t)",
		result.DatabaseID, result.Name, result.Status, result.Conclusion, len(result.Jobs),
		runResp.FromCache && jobsResp.FromCache)

	// Persist completed runs into the immutable cache. Subsequent calls
	// (including from new processes) will short-circuit at the top of this
	// function.
	if result.Status == "completed" {
		if payload, err := cache.MarshalJSON(result); err == nil {
			cache.Shared().PutCompletedRun(repo, runID, result.Status, result.Conclusion, payload)
		}
	}

	return result, nil
}

func FetchAndAttachLogs(opts Options, run *WorkflowRun, tailLines int) {
	for i := range run.Jobs {
		job := &run.Jobs[i]
		if !strings.EqualFold(job.Conclusion, "failure") {
			continue
		}
		if err := FetchJobLogs(opts, job, tailLines); err != nil {
			logger.Warnf("failed to fetch logs for job %d (%s): %v", job.DatabaseID, job.Name, err)
		}
	}
}

// FetchJobLogs fetches logs for a single job from the GitHub API and attaches them
// to the job and its steps (via attachLogsToSteps). The job must have DatabaseID set.
//
// Logs for jobs with a terminal Conclusion are persisted to the immutable
// cache so a subsequent fetch of the same job is a zero-RTT lookup. In-flight
// jobs (Conclusion == "") fall through to the ETag-aware HTTP cache only.
func FetchJobLogs(opts Options, job *Job, tailLines int) error {
	token, err := opts.token()
	if err != nil {
		return fmt.Errorf("cannot fetch logs: %w", err)
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return fmt.Errorf("cannot resolve repo for logs: %w", err)
	}

	if logs, ok := cache.Shared().GetJobLogs(job.DatabaseID); ok {
		logger.Tracef("github cache hit for job logs %d", job.DatabaseID)
		attachLogsToSteps(job, logs, tailLines)
		return nil
	}

	ctx := context.Background()
	path := fmt.Sprintf("/repos/%s/actions/jobs/%d/logs", repo, job.DatabaseID)
	result, err := cachedGet(ctx, token, path, nil)
	if err != nil {
		return err
	}
	logs := string(result.Body)
	attachLogsToSteps(job, logs, tailLines)

	// Only persist logs once the job has reached a terminal state — otherwise
	// the cache would lock in a half-finished log stream.
	if job.Conclusion != "" {
		cache.Shared().PutJobLogs(job.DatabaseID, repo, logs)
	}
	return nil
}

type restWorkflow struct {
	Path string `json:"path"`
}

func FetchWorkflowDefinition(opts Options, run *WorkflowRun) (string, error) {
	if run.WorkflowID == 0 {
		return "", fmt.Errorf("workflow ID not set on run %d", run.DatabaseID)
	}
	token, err := opts.token()
	if err != nil {
		return "", err
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return "", err
	}

	// Workflow YAML at a specific SHA is immutable; check the immutable
	// cache before any HTTP. We need the path so the cache key is the
	// (repo, workflowID, sha) tuple — the path is recovered from the cache
	// when present.
	if run.HeadSHA != "" {
		if yaml, path, ok := cache.Shared().GetWorkflowDef(repo, run.WorkflowID, run.HeadSHA); ok {
			logger.Tracef("github cache hit for workflow def %s/%d@%s", repo, run.WorkflowID, run.HeadSHA)
			run.WorkflowPath = path
			run.WorkflowYAML = yaml
			return yaml, nil
		}
	}

	ctx := context.Background()

	wfPath := fmt.Sprintf("/repos/%s/actions/workflows/%d", repo, run.WorkflowID)
	wfResp, err := cachedGet(ctx, token, wfPath, nil)
	if err != nil {
		return "", err
	}
	var wf restWorkflow
	if err := json.Unmarshal(wfResp.Body, &wf); err != nil {
		return "", fmt.Errorf("parse workflow response: %w", err)
	}
	run.WorkflowPath = wf.Path

	ref := run.HeadSHA
	if ref == "" {
		ref = "HEAD"
	}
	contentPath := fmt.Sprintf("/repos/%s/contents/%s?ref=%s", repo, wf.Path, ref)
	contentResp, err := cachedGet(ctx, token, contentPath, map[string]string{
		"Accept": "application/vnd.github.raw+json",
	})
	if err != nil {
		return "", err
	}
	yaml := string(contentResp.Body)
	run.WorkflowYAML = yaml

	if run.HeadSHA != "" {
		cache.Shared().PutWorkflowDef(repo, run.WorkflowID, run.HeadSHA, wf.Path, yaml)
	}
	return yaml, nil
}

func cleanRawLog(s string) string {
	var sb strings.Builder
	for i, line := range strings.Split(s, "\n") {
		cleaned := logTimestampPrefix.ReplaceAllString(line, "")
		cleaned = ansiEscape.ReplaceAllString(cleaned, "")
		if strings.HasPrefix(cleaned, "##[group]") || strings.HasPrefix(cleaned, "##[endgroup]") {
			continue
		}
		cleaned = actionsAnnotation.ReplaceAllString(cleaned, "")
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(cleaned)
	}
	return sb.String()
}

func attachLogsToSteps(job *Job, rawLog string, tailLines int) {
	job.Logs = tailString(cleanRawLog(rawLog), tailLines)
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
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
var actionsAnnotation = regexp.MustCompile(`##\[(error|warning|debug|notice|command)\]`)

func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

func parseLogSections(raw string) map[string]string {
	sections := make(map[string]string)
	var currentStep string
	var buf strings.Builder

	for _, line := range strings.Split(raw, "\n") {
		cleaned := logTimestampPrefix.ReplaceAllString(line, "")
		cleaned = stripANSI(cleaned)
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
			cleaned = actionsAnnotation.ReplaceAllString(cleaned, "")
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
