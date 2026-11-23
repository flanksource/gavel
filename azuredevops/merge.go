package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/pr/model"
)

var mergeStrategies = map[string]string{"rebase": "rebase", "squash": "squash", "merge": "noFastForward"}

func mergeCompletionOptions(strategy string) map[string]any {
	return map[string]any{"mergeStrategy": strategy, "bypassPolicy": false, "deleteSourceBranch": false}
}

func (c *Client) MergePR(ctx context.Context, number int, input model.MergeInput) error {
	strategy, ok := mergeStrategies[strings.ToLower(strings.TrimSpace(input.Method))]
	if number <= 0 || !ok || strings.TrimSpace(input.HeadSHA) == "" {
		return fmt.Errorf("azure merge requires a positive PR number, rebase/squash/merge method and reviewed head SHA")
	}
	if input.Auto {
		return fmt.Errorf("azure auto-complete is not supported by this action")
	}
	if err := c.Preflight(ctx); err != nil {
		return err
	}
	completion := mergeCompletionOptions(strategy)
	if message := strings.TrimSpace(strings.TrimSpace(input.CommitTitle) + "\n\n" + strings.TrimSpace(input.CommitBody)); message != "" && strategy != "rebase" {
		completion["mergeCommitMessage"] = message
	}
	payload := map[string]any{
		"status":                "completed",
		"lastMergeSourceCommit": commitRef{ID: input.HeadSHA},
		"completionOptions":     completion,
	}
	data, _, err := c.request(ctx, http.MethodPatch, c.repositoryPath()+"/pullrequests/"+strconv.Itoa(number), nil, payload)
	if err != nil {
		return err
	}
	var result pullRequest
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("decode Azure PR completion: %w", err)
	}
	queued := result.Status == "active" && !result.CompletionQueueTime.IsZero() && result.MergeFailure == ""
	if result.ID != number || (result.Status != "completed" && !queued) {
		return fmt.Errorf("azure PR completion returned unexpected identity/status: PR %d, status %q, merge status %q", result.ID, result.Status, result.MergeStatus)
	}
	return nil
}
