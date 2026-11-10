// A PR of a run without GitHub credentials, in a sandbox job or a substrate environment: its PR's
// status, thread resolution and source branch pushes go to the run's run API, serve's or its
// caller's, which acts for that run alone.
package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/sandbox/runclient"
)

// RunAPI is the run API a run without GitHub credentials reaches its PR through.
type RunAPI interface {
	Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error)
}

// installed is the run API UseRunAPI installed; nil leaves it to the sandbox job's environment.
var installed RunAPI

// UseRunAPI makes api the run API this process's PR client goes through: a run's program in a
// substrate environment reaches its caller's.
func UseRunAPI(api RunAPI) { installed = api }

// inJob is the run API this process reaches its PR through: the one installed, or the run API of
// the sandbox job this process runs in; nil outside both.
var inJob = func() RunAPI {
	if installed != nil {
		return installed
	}
	client, err := runclient.FromEnv(os.Getenv)
	if err != nil {
		return nil
	}
	return client
}

// InJob reports whether this process reaches its PR through a run API rather than credentials of
// its own.
func InJob() bool { return inJob() != nil }

// jobClient is a PR client through a run API.
type jobClient struct{ api RunAPI }

func (c jobClient) Kind() string { return "github" }

func (c jobClient) Status(ctx context.Context, number int, opts model.StatusOptions) (*model.StatusSnapshot, error) {
	query := url.Values{
		"number":   {strconv.Itoa(number)},
		"threads":  {strconv.FormatBool(opts.Threads)},
		"logs":     {strconv.FormatBool(opts.Logs)},
		"tailLogs": {strconv.Itoa(opts.TailLogs)},
	}
	resp, err := c.api.Do(ctx, http.MethodGet, "pr/status?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("read the status of PR #%d through its run API: %w", number, err)
	}
	defer resp.Body.Close()
	var snapshot model.StatusSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		return nil, fmt.Errorf("decode the status of PR #%d: %w", number, err)
	}
	return &snapshot, nil
}

func (c jobClient) ResolveThread(ctx context.Context, number int, threadID string) error {
	query := url.Values{"number": {strconv.Itoa(number)}}
	resp, err := c.api.Do(ctx, http.MethodPost, "pr/threads/"+url.PathEscape(threadID)+"/resolve?"+query.Encode(), nil)
	if err != nil {
		return fmt.Errorf("resolve review thread %s of PR #%d through its run API: %w", threadID, number, err)
	}
	return resp.Body.Close()
}

func (c jobClient) Preflight(context.Context) error { return nil }

func (c jobClient) DefaultBranch(context.Context) (string, error) {
	return "", errInJob("reading a repository's default branch")
}

func (c jobClient) OpenPRs(context.Context) ([]model.PRInfo, error) {
	return nil, errInJob("listing open PRs")
}

func (c jobClient) CreatePR(context.Context, model.CreatePRInput) (*model.CreatePRResult, error) {
	return nil, errInJob("creating a PR")
}

func (c jobClient) MergePR(context.Context, int, model.MergeInput) error {
	return errInJob("merging a PR")
}

func errInJob(what string) error {
	return fmt.Errorf("%s isn't available inside a sandbox job", what)
}

// fetchHeadInJob checks the PR's source branch is still at t.HeadSHA, which the job's checkout
// holds, as FetchHead does with the branch itself.
func (t *RunTarget) fetchHeadInJob(ctx context.Context, api RunAPI, dir string) error {
	resp, err := api.Do(ctx, http.MethodGet, "pr/head", nil)
	if err != nil {
		return fmt.Errorf("read the PR's source branch through its run API: %w", err)
	}
	defer resp.Body.Close()
	var head struct{ SHA string }
	if err := json.NewDecoder(resp.Body).Decode(&head); err != nil {
		return fmt.Errorf("decode the PR's source branch: %w", err)
	}
	if head.SHA != t.HeadSHA {
		return fmt.Errorf("PR source branch advanced from %s to %s; refresh the PR before continuing", t.HeadSHA, head.SHA)
	}
	if _, err := targetGit(ctx, dir, "cat-file", "-e", t.HeadSHA+"^{commit}"); err != nil {
		return fmt.Errorf("the job's checkout doesn't have the PR's head %s: %w", t.HeadSHA, err)
	}
	return nil
}

// pushHeadInJob uploads the commits since t.HeadSHA for the run API to fast-forward the PR's source
// branch with, as PushHead pushes them itself.
func (t *RunTarget) pushHeadInJob(ctx context.Context, api RunAPI, dir, head string) error {
	bundle := filepath.Join(os.TempDir(), "gavel-pr-push-"+head+".bundle")
	defer os.Remove(bundle)
	if _, err := targetGit(ctx, dir, "bundle", "create", "--quiet", bundle, "HEAD", "^"+t.HeadSHA); err != nil {
		return fmt.Errorf("bundle the commits for the PR: %w", err)
	}
	data, err := os.ReadFile(bundle)
	if err != nil {
		return fmt.Errorf("read the bundle of the commits for the PR: %w", err)
	}
	resp, err := api.Do(ctx, http.MethodPost, "pr/push?"+url.Values{"base": {t.HeadSHA}}.Encode(), bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("push the PR's source branch through its run API: %w", err)
	}
	defer resp.Body.Close()
	var pushed struct{ SHA string }
	if err := json.NewDecoder(resp.Body).Decode(&pushed); err != nil {
		return fmt.Errorf("decode the PR push: %w", err)
	}
	if pushed.SHA != head {
		return fmt.Errorf("the run API pushed %s to the PR's source branch, not the run's head %s", pushed.SHA, head)
	}
	return nil
}
