package provider

import (
	"context"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
)

type gitHubClient struct{ opts github.Options }

func (c gitHubClient) Kind() string { return "github" }
func (c gitHubClient) Preflight(ctx context.Context) error {
	_, err := github.DefaultBranch(c.opts)
	return err
}
func (c gitHubClient) DefaultBranch(ctx context.Context) (string, error) {
	return github.DefaultBranch(c.opts)
}
func (c gitHubClient) CreatePR(ctx context.Context, in model.CreatePRInput) (*model.CreatePRResult, error) {
	return github.CreatePR(c.opts, in)
}
func (c gitHubClient) OpenPRs(ctx context.Context) ([]model.PRInfo, error) {
	prs, _, err := github.SearchPRs(c.opts, github.PRSearchOptions{State: "open", Author: "@me"})
	if err != nil {
		return nil, err
	}
	out := make([]model.PRInfo, 0, len(prs))
	for _, pr := range prs {
		out = append(out, model.PRInfo{Number: pr.Number, Title: pr.Title, URL: pr.URL, HeadRefName: pr.Source, BaseRefName: pr.Target, State: pr.State})
	}
	return out, nil
}
func (c gitHubClient) Status(ctx context.Context, number int, opts model.StatusOptions) (*model.StatusSnapshot, error) {
	pr, err := github.FetchPR(c.opts, number)
	if err != nil {
		return nil, err
	}
	runs := map[int64]*model.WorkflowRun{}
	for _, check := range pr.StatusCheckRollup {
		id, err := github.ExtractRunID(check.DetailsURL)
		if err != nil {
			continue
		}
		if _, found := runs[id]; found {
			continue
		}
		run, err := github.FetchRunJobs(c.opts, id, github.RunLogOptions{FetchLogs: opts.Logs, TailLines: opts.TailLogs})
		if err != nil {
			return nil, err
		}
		runs[id] = run
	}
	return &model.StatusSnapshot{PR: pr, Runs: runs}, nil
}
