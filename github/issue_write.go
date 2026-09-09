package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// IssueInput describes one issue write. A zero Number opens a new issue; a
// non-zero Number edits that issue in place, replacing every field that is set.
type IssueInput struct {
	Number    int
	Title     string
	Body      string
	Labels    []string
	Assignees []string
}

type IssueResult struct {
	Number int    `json:"number"`
	URL    string `json:"html_url"`
	Title  string `json:"title"`
	State  string `json:"state"`
	NodeID string `json:"node_id"` // GraphQL global node ID
	// Repo is the resolved owner/repo the issue lives on, so callers can build
	// an external reference without resolving the remote a second time.
	Repo string `json:"-"`
	// Updated reports that an existing issue was edited rather than opened.
	Updated bool `json:"-"`
}

// SaveIssue opens a new issue on the resolved repo, or edits the one named by
// in.Number.
func SaveIssue(opts Options, in IssueInput) (*IssueResult, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, fmt.Errorf("SaveIssue: Title is required")
	}

	token, err := opts.token()
	if err != nil {
		return nil, err
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"title": in.Title,
		"body":  in.Body,
	}
	// GitHub rejects an explicit null for these, so only send them when set.
	// On an edit that also leaves the issue's existing labels/assignees alone.
	if len(in.Labels) > 0 {
		payload["labels"] = in.Labels
	}
	if len(in.Assignees) > 0 {
		payload["assignees"] = in.Assignees
	}

	request := newClient(token).Header("Content-Type", "application/json").R(context.Background())
	url := fmt.Sprintf("%s/repos/%s/issues", githubAPIBase(), repo)
	write, action := request.Post, "create issue on"
	if in.Number > 0 {
		url = fmt.Sprintf("%s/%d", url, in.Number)
		write, action = request.Patch, fmt.Sprintf("update issue #%d on", in.Number)
	}

	resp, err := write(url, payload)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", action, repo, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read issue-write response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s: HTTP %d: %s%s",
			action, repo, resp.StatusCode, string(body), saveIssueHint(resp.StatusCode, repo, in.Number))
	}

	var out IssueResult
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("parse issue-write response: %w", err)
	}
	out.Repo = repo
	out.Updated = in.Number > 0
	return &out, nil
}

// saveIssueHint explains the two failures whose GitHub message alone does not
// say what to change.
func saveIssueHint(status int, repo string, number int) string {
	switch status {
	case http.StatusGone:
		return fmt.Sprintf(" (issues are disabled on %s)", repo)
	case http.StatusNotFound:
		if number > 0 {
			return fmt.Sprintf(" (issue %s#%d not found, or the token lacks issues: write)", repo, number)
		}
		return fmt.Sprintf(" (repo %s not found, or the token lacks issues: write)", repo)
	default:
		return ""
	}
}
