package provider

import (
	"context"
	"fmt"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
)

func (c gitHubClient) MergePR(ctx context.Context, number int, input model.MergeInput) error {
	if _, err := github.MergeMethodFor(input.Method); err != nil {
		return err
	}
	pr, err := github.FetchPR(c.opts, number)
	if err != nil {
		return err
	}
	if pr.Number != number || pr.NodeID == "" || pr.HeadRefOID == "" || pr.State != "OPEN" || pr.IsDraft {
		return fmt.Errorf("PR #%d is not an open, ready pull request", number)
	}
	if input.HeadSHA != "" && pr.HeadRefOID != input.HeadSHA {
		return fmt.Errorf("PR head changed from %s to %s; refresh before merging", input.HeadSHA, pr.HeadRefOID)
	}
	if input.Auto {
		return github.EnableAutoMerge(c.opts, pr.NodeID, input.Method)
	}
	return github.MergePullRequest(c.opts, github.MergeOptions{NodeID: pr.NodeID, Method: input.Method, HeadSHA: pr.HeadRefOID})
}
