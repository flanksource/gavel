package github

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	nethttp "net/http"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons/http"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/service"
)

type Options struct {
	Context context.Context `json:"-"`
	WorkDir string
	Repo    string // owner/repo
	Token   string // optional; falls back to GITHUB_TOKEN then GH_TOKEN env
	// TokenSource, when set, is the only credential: no fallback to Token, the
	// environment or auth.json, so a hosted organization never uses the host's token.
	TokenSource func() (string, error)
	// Cache is the GitHub response cache of the organization this client acts for;
	// nil is the shared local-org cache.
	Cache *cache.Store
}

func (o Options) context() context.Context {
	if o.Context != nil {
		return o.Context
	}
	return context.Background()
}

func (o Options) cacheStore() *cache.Store {
	if o.Cache != nil {
		return o.Cache
	}
	return cache.Shared()
}

// ErrNoToken is wrapped by every error Options.token() returns when no token
// was ever configured, so callers (ProbeToken, the PR dashboard) can tell
// that state (AuthStateNoToken) from other failures with errors.Is.
var ErrNoToken = errors.New("no GitHub token")

func (o Options) token() (string, error) {
	if o.TokenSource != nil {
		token, err := o.TokenSource()
		if err == nil && token == "" {
			err = fmt.Errorf("%w: the token source returned an empty token", ErrNoToken)
		}
		return token, err
	}
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
	return "", fmt.Errorf("%w: set Options.Token, GITHUB_TOKEN, GH_TOKEN, or run `gavel system install` to persist one", ErrNoToken)
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
	Path       string `json:"path"`
}

func (r restRun) toWorkflowRun(jobs []model.Job) *model.WorkflowRun {
	return &model.WorkflowRun{
		DatabaseID:   r.ID,
		Name:         r.Name,
		Status:       r.Status,
		Conclusion:   r.Conclusion,
		URL:          r.HTMLURL,
		HeadSHA:      r.HeadSHA,
		WorkflowID:   r.WorkflowID,
		WorkflowPath: r.Path,
		Jobs:         jobs,
	}
}

type restJobsResponse struct {
	Jobs       []restJob `json:"jobs"`
	TotalCount int       `json:"total_count"`
}

type restJob struct {
	ID          int64      `json:"id"`
	RunID       int64      `json:"run_id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Conclusion  string     `json:"conclusion"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt time.Time  `json:"completed_at"`
	HTMLURL     string     `json:"html_url"`
	Steps       []restStep `json:"steps"`
}

func (j restJob) toJob() model.Job {
	steps := make([]model.Step, len(j.Steps))
	for i, s := range j.Steps {
		steps[i] = s.toStep()
	}
	return model.Job{
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

func (s restStep) toStep() model.Step {
	return model.Step{
		Name:       s.Name,
		Status:     s.Status,
		Conclusion: s.Conclusion,
		Number:     s.Number,
	}
}

// RunLogOptions selects an exact job and controls which log tail to attach.
type RunLogOptions struct {
	JobID     int64
	FetchLogs bool
	TailLines int
	// Checks is the PR's current check rollup. Re-running a workflow keeps its
	// run id, so a cached completed run these checks contradict is fetched again.
	Checks model.StatusChecks
	// Revalidate asks GitHub for the run even when it is cached as completed.
	Revalidate bool
}

func FetchRunJobs(opts Options, runID int64, logOpts RunLogOptions) (*model.WorkflowRun, error) {
	token, err := opts.token()
	if err != nil {
		return nil, err
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return nil, err
	}
	var cached *model.WorkflowRun
	if logOpts.JobID == 0 {
		cached = loadCompletedRun(opts.cacheStore(), runID)
	}
	if cached != nil && !logOpts.Revalidate && !contradictedByChecks(cached, logOpts.Checks) {
		logger.Tracef("github cache hit for completed run %d", runID)
		if logOpts.FetchLogs {
			FetchAndAttachLogs(opts, cached, logOpts.TailLines)
		}
		return cached, nil
	}

	runPath := fmt.Sprintf("/repos/%s/actions/runs/%d", repo, runID)
	runResp, err := cachedGet(opts.context(), opts.cacheStore(), token, runPath, nil)
	if err != nil {
		return nil, err
	}
	var run restRun
	if err := json.Unmarshal(runResp.Body, &run); err != nil {
		return nil, fmt.Errorf("parse run response: %w", err)
	}

	jobs, staleJobs, err := fetchRunJobs(opts, runID, logOpts.JobID)
	if err != nil {
		return nil, err
	}

	result := run.toWorkflowRun(jobs)
	result.Stale = cmp.Or(runResp.Stale, staleJobs)
	logger.Debugf("fetched run %d %q: status=%s conclusion=%s jobs=%d stale=%q",
		result.DatabaseID, result.Name, result.Status, result.Conclusion, len(result.Jobs), result.Stale)

	// Cache the run metadata before attaching request-specific tails from raw job logs.
	if logOpts.JobID == 0 {
		switch {
		case runSettled(result):
			persistCompletedRun(opts.cacheStore(), repo, runID, result)
		case cached != nil && result.Stale == "":
			opts.cacheStore().DeleteCompletedRun(runID)
		}
	}
	if logOpts.FetchLogs {
		FetchAndAttachLogs(opts, result, logOpts.TailLines)
	}

	return result, nil
}

// loadCompletedRun returns the cached completed run, or nil. A payload that is
// not a settled run (a job still in progress) is not trusted.
func loadCompletedRun(store *cache.Store, runID int64) *model.WorkflowRun {
	payload := store.GetCompletedRunPayload(runID)
	if payload == nil {
		return nil
	}
	var run model.WorkflowRun
	if err := json.Unmarshal(payload, &run); err != nil {
		logger.Warnf("github cache: corrupt completed run %d, refetching: %v", runID, err)
		return nil
	}
	if !runSettled(&run) {
		logger.Warnf("github cache: completed run %d has unfinished jobs, refetching", runID)
		return nil
	}
	// Older cached payloads contain tails from a previous request.
	for i := range run.Jobs {
		job := &run.Jobs[i]
		job.Logs, job.LogHeadings = "", nil
		for j := range job.Steps {
			job.Steps[j].Logs, job.Steps[j].LogHeadings = "", nil
		}
	}
	return &run
}

// runSettled reports whether GitHub confirmed the run and all its jobs finished.
func runSettled(run *model.WorkflowRun) bool {
	return run.Stale == "" && run.Status == "completed" &&
		!slices.ContainsFunc(run.Jobs, func(job model.Job) bool { return job.Status != "completed" })
}

// contradictedByChecks reports whether the PR's checks show the run was re-run:
// one of its checks is unfinished, or names a job the run does not have.
func contradictedByChecks(run *model.WorkflowRun, checks model.StatusChecks) bool {
	for _, check := range checks {
		if runID, err := ExtractRunID(check.DetailsURL); err != nil || runID != run.DatabaseID {
			continue
		}
		if check.Status != "COMPLETED" {
			return true
		}
		jobID, err := ExtractJobID(check.DetailsURL)
		if err != nil {
			continue
		}
		if !slices.ContainsFunc(run.Jobs, func(job model.Job) bool { return job.DatabaseID == jobID }) {
			return true
		}
	}
	return false
}

// fetchRunJobs returns the run's latest jobs, or only jobID's when it is set,
// and why they are stale when GitHub could not confirm them.
func fetchRunJobs(opts Options, runID, jobID int64) (jobs []model.Job, stale string, err error) {
	token, err := opts.token()
	if err != nil {
		return nil, "", err
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return nil, "", err
	}
	if jobID != 0 {
		response, err := cachedGet(opts.context(), opts.cacheStore(), token, fmt.Sprintf("/repos/%s/actions/jobs/%d", repo, jobID), nil)
		if err != nil {
			return nil, "", err
		}
		var job restJob
		if err := json.Unmarshal(response.Body, &job); err != nil {
			return nil, "", fmt.Errorf("parse job response: %w", err)
		}
		if job.ID != jobID || job.RunID != runID {
			return nil, "", fmt.Errorf("job %d belongs to run %d, expected job %d in run %d", job.ID, job.RunID, jobID, runID)
		}
		return []model.Job{job.toJob()}, response.Stale, nil
	}
	for page := 1; ; page++ {
		path := fmt.Sprintf("/repos/%s/actions/runs/%d/jobs?per_page=100&page=%d", repo, runID, page)
		response, err := cachedGet(opts.context(), opts.cacheStore(), token, path, nil)
		if err != nil {
			return nil, "", err
		}
		stale = cmp.Or(stale, response.Stale)
		var payload restJobsResponse
		if err := json.Unmarshal(response.Body, &payload); err != nil {
			return nil, "", fmt.Errorf("parse jobs response: %w", err)
		}
		for _, job := range payload.Jobs {
			jobs = append(jobs, job.toJob())
		}
		if len(jobs) >= payload.TotalCount {
			return jobs, stale, nil
		}
		if len(payload.Jobs) == 0 {
			return nil, "", fmt.Errorf("workflow run %d returned an empty jobs page before total_count %d", runID, payload.TotalCount)
		}
	}
}

func persistCompletedRun(store *cache.Store, repo string, runID int64, run *model.WorkflowRun) {
	payload, err := cache.MarshalJSON(run)
	if err != nil {
		logger.Warnf("github cache: marshal run %d failed: %v", runID, err)
		return
	}
	store.PutCompletedRun(repo, runID, run.Status, run.Conclusion, payload)
}

func FetchAndAttachLogs(opts Options, run *model.WorkflowRun, tailLines int) {
	for i := range run.Jobs {
		job := &run.Jobs[i]
		if !model.NeedsJobLogs(job.Conclusion) {
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
func FetchJobLogs(opts Options, job *model.Job, tailLines int) error {
	token, err := opts.token()
	if err != nil {
		return fmt.Errorf("cannot fetch logs: %w", err)
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return fmt.Errorf("cannot resolve repo for logs: %w", err)
	}
	if logs, ok := opts.cacheStore().GetJobLogs(job.DatabaseID); ok {
		logger.Tracef("github cache hit for job logs %d", job.DatabaseID)
		attachLogsToSteps(job, logs, tailLines)
		return nil
	}

	path := fmt.Sprintf("/repos/%s/actions/jobs/%d/logs", repo, job.DatabaseID)
	result, err := cachedGet(opts.context(), opts.cacheStore(), token, path, nil)
	if err != nil {
		return err
	}
	logs := string(result.Body)
	attachLogsToSteps(job, logs, tailLines)

	// Only persist logs GitHub served after the job reached a terminal state —
	// otherwise the cache would lock in a half-finished log stream.
	if job.Conclusion != "" && result.Stale == "" {
		opts.cacheStore().PutJobLogs(job.DatabaseID, repo, logs)
	}
	return nil
}

type restWorkflow struct {
	Path string `json:"path"`
}

func FetchWorkflowDefinition(opts Options, run *model.WorkflowRun) (string, error) {
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
		if yaml, path, ok := opts.cacheStore().GetWorkflowDef(repo, run.WorkflowID, run.HeadSHA); ok {
			logger.Tracef("github cache hit for workflow def %s/%d@%s", repo, run.WorkflowID, run.HeadSHA)
			run.WorkflowPath = path
			run.WorkflowYAML = yaml
			return yaml, nil
		}
	}

	ctx := context.Background()

	wfPath := fmt.Sprintf("/repos/%s/actions/workflows/%d", repo, run.WorkflowID)
	wfResp, err := cachedGet(ctx, opts.cacheStore(), token, wfPath, nil)
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
	contentResp, err := cachedGet(ctx, opts.cacheStore(), token, contentPath, map[string]string{
		"Accept": "application/vnd.github.raw+json",
	})
	if err != nil {
		return "", err
	}
	yaml := string(contentResp.Body)
	run.WorkflowYAML = yaml

	if run.HeadSHA != "" {
		opts.cacheStore().PutWorkflowDef(repo, run.WorkflowID, run.HeadSHA, wf.Path, yaml)
	}
	return yaml, nil
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

func ParsePRJSON(data []byte) (*model.PRInfo, error) {
	var pr model.PRInfo
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, fmt.Errorf("parse PR JSON: %w", err)
	}
	return &pr, nil
}

func ParseRunJSON(data []byte) (*model.WorkflowRun, error) {
	var run model.WorkflowRun
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

var jobIDRegexp = regexp.MustCompile(`/actions/runs/\d+/(?:attempts/\d+/)?job/(\d+)`)

func ExtractJobID(detailsURL string) (int64, error) {
	matches := jobIDRegexp.FindStringSubmatch(detailsURL)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no job ID found in URL: %s", detailsURL)
	}
	return strconv.ParseInt(matches[1], 10, 64)
}

var runURLPattern = regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+/[^/]+)/actions/runs/(\d+)(?:/attempts/\d+)?(?:/job/(\d+))?/?(?:[?#].*)?$`)

// ParseRunURL parses a GitHub Actions run or job URL, e.g.
// https://github.com/owner/repo/actions/runs/123/job/456. jobID is 0 for a
// run URL without a job.
func ParseRunURL(url string) (repo string, runID, jobID int64, err error) {
	m := runURLPattern.FindStringSubmatch(strings.TrimSpace(url))
	if len(m) < 4 {
		return "", 0, 0, fmt.Errorf("cannot parse Actions run URL: %q", url)
	}
	runID, err = strconv.ParseInt(m[2], 10, 64)
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid run ID in %q: %w", url, err)
	}
	if runID <= 0 {
		return "", 0, 0, fmt.Errorf("invalid run ID in %q: expected a positive integer", url)
	}
	if m[3] != "" {
		if jobID, err = strconv.ParseInt(m[3], 10, 64); err != nil {
			return "", 0, 0, fmt.Errorf("invalid job ID in %q: %w", url, err)
		}
		if jobID <= 0 {
			return "", 0, 0, fmt.Errorf("invalid job ID in %q: expected a positive integer", url)
		}
	}
	return m[1], runID, jobID, nil
}
