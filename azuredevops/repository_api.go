package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/flanksource/gavel/pr/model"
)

func (c *Client) Preflight(ctx context.Context) error {
	if c.info != nil {
		return nil
	}
	if _, err := validateRepository(c.repo); err != nil {
		return err
	}
	data, _, err := c.request(ctx, http.MethodGet, c.repositoryPath(), nil, nil)
	if err != nil {
		return err
	}
	var info repositoryInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return fmt.Errorf("decode Azure repository: %w", err)
	}
	if info.ID == "" || info.Name == "" || info.Project.ID == "" || info.Project.Name == "" {
		return fmt.Errorf("azure repository response is missing repository or project identity")
	}
	c.info = &info
	c.repo.Project, c.repo.Name = info.Project.Name, info.Name
	return nil
}

func (c *Client) Repository(ctx context.Context) (Repository, error) {
	if err := c.Preflight(ctx); err != nil {
		return Repository{}, err
	}
	return c.repo, nil
}

func (c *Client) DefaultBranch(ctx context.Context) (string, error) {
	if err := c.Preflight(ctx); err != nil {
		return "", err
	}
	branch := strings.TrimPrefix(c.info.DefaultBranch, "refs/heads/")
	if branch == "" || branch == c.info.DefaultBranch {
		return "", fmt.Errorf("azure repository default branch is missing or invalid")
	}
	return branch, nil
}

func (c *Client) OpenPRs(ctx context.Context) ([]model.PRInfo, error) {
	if err := c.Preflight(ctx); err != nil {
		return nil, err
	}
	prs, err := collection[pullRequest](ctx, c, c.repositoryPath()+"/pullrequests", url.Values{"searchCriteria.status": {"active"}})
	if err != nil {
		return nil, err
	}
	result := make([]model.PRInfo, 0, len(prs))
	for _, pr := range prs {
		info, err := c.prInfo(pr)
		if err != nil {
			return nil, err
		}
		result = append(result, *info)
	}
	return result, nil
}

func (c *Client) CreatePR(ctx context.Context, in model.CreatePRInput) (*model.CreatePRResult, error) {
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Head) == "" {
		return nil, fmt.Errorf("azure PR title and source branch are required")
	}
	if err := c.Preflight(ctx); err != nil {
		return nil, err
	}
	if in.Base == "" {
		base, err := c.DefaultBranch(ctx)
		if err != nil {
			return nil, err
		}
		in.Base = base
	}
	payload := struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Source      string `json:"sourceRefName"`
		Target      string `json:"targetRefName"`
		Draft       bool   `json:"isDraft"`
	}{Title: in.Title, Description: in.Body, Source: fullRef(in.Head), Target: fullRef(in.Base), Draft: in.Draft}
	if payload.Source == payload.Target {
		return nil, fmt.Errorf("azure PR source and target branches must differ")
	}
	data, _, err := c.request(ctx, http.MethodPost, c.repositoryPath()+"/pullrequests", nil, payload)
	if err != nil {
		return nil, err
	}
	var pr pullRequest
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, fmt.Errorf("decode Azure created PR: %w", err)
	}
	if pr.ID <= 0 {
		return nil, fmt.Errorf("azure created PR response is missing pullRequestId")
	}
	return &model.CreatePRResult{Number: pr.ID, Title: pr.Title, State: pr.Status, URL: fmt.Sprintf("%s/pullrequest/%d", c.repo.URL(), pr.ID), Base: strings.TrimPrefix(in.Base, "refs/heads/")}, nil
}

func fullRef(branch string) string { return "refs/heads/" + strings.TrimPrefix(branch, "refs/heads/") }

func (c *Client) prInfo(pr pullRequest) (*model.PRInfo, error) {
	if pr.ID <= 0 {
		return nil, fmt.Errorf("azure PR response is missing pullRequestId")
	}
	states := map[string]string{"active": "OPEN", "completed": "MERGED", "abandoned": "CLOSED"}
	state, ok := states[pr.Status]
	if !ok {
		return nil, fmt.Errorf("unexpected Azure PR state %q", pr.Status)
	}
	return &model.PRInfo{Number: pr.ID, Provider: "azuredevops", Title: pr.Title, Body: pr.Description, State: state, IsDraft: pr.Draft,
		HeadRefName: strings.TrimPrefix(pr.Source, "refs/heads/"), BaseRefName: strings.TrimPrefix(pr.Target, "refs/heads/"), HeadRefOID: pr.SourceCommit.ID, BaseRefOID: pr.TargetCommit.ID,
		Author: model.PRAuthor{Login: pr.CreatedBy.Login, Name: pr.CreatedBy.Name, AvatarURL: pr.CreatedBy.Avatar}, URL: fmt.Sprintf("%s/pullrequest/%d", c.repo.URL(), pr.ID)}, nil
}
