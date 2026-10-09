package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/flanksource/gavel/pr/model"
)

type repoInfo struct {
	DefaultBranch string `json:"default_branch"`
}

// DefaultBranch returns the repo's default branch (e.g. "main") via REST.
func DefaultBranch(opts Options) (string, error) {
	token, err := opts.token()
	if err != nil {
		return "", err
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return "", err
	}

	resp, err := cachedGet(context.Background(), token, fmt.Sprintf("/repos/%s", repo), nil)
	if err != nil {
		return "", fmt.Errorf("fetch repo %s: %w", repo, err)
	}
	var info repoInfo
	if err := json.Unmarshal(resp.Body, &info); err != nil {
		return "", fmt.Errorf("parse repo response: %w", err)
	}
	if info.DefaultBranch == "" {
		return "", fmt.Errorf("repo %s: empty default_branch in response", repo)
	}
	return info.DefaultBranch, nil
}

// CreatePR opens a pull request against the resolved repo. Base defaults to
// the repo's default branch when empty.
func CreatePR(opts Options, in model.CreatePRInput) (*model.CreatePRResult, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, fmt.Errorf("CreatePR: Title is required")
	}
	if strings.TrimSpace(in.Head) == "" {
		return nil, fmt.Errorf("CreatePR: Head branch is required")
	}

	token, err := opts.token()
	if err != nil {
		return nil, err
	}
	repo, err := opts.resolveRepo()
	if err != nil {
		return nil, err
	}

	base := in.Base
	if base == "" {
		base, err = DefaultBranch(opts)
		if err != nil {
			return nil, fmt.Errorf("resolve default branch: %w", err)
		}
	}

	payload := map[string]any{
		"title": in.Title,
		"head":  in.Head,
		"base":  base,
		"body":  in.Body,
		"draft": in.Draft,
	}

	client := newClient(token).Header("Content-Type", "application/json")
	resp, err := client.R(context.Background()).Post(
		fmt.Sprintf("%s/repos/%s/pulls", githubAPIBase(), repo), payload,
	)
	if err != nil {
		return nil, fmt.Errorf("create PR on %s: %w", repo, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read create-PR response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("create PR on %s: HTTP %d: %s", repo, resp.StatusCode, string(body))
	}

	var out model.CreatePRResult
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("parse create-PR response: %w", err)
	}
	out.Base = base
	return &out, nil
}
