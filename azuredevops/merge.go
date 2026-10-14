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

func (c *Client) MergePR(ctx context.Context, number int, input model.MergeInput) error {
	strategies := map[string]string{"rebase": "rebase", "squash": "squash", "merge": "noFastForward"}
	strategy, ok := strategies[strings.ToLower(strings.TrimSpace(input.Method))]
	if number <= 0 || !ok || strings.TrimSpace(input.HeadSHA) == "" {
		return fmt.Errorf("azure merge requires a positive PR number, rebase/squash/merge method and reviewed head SHA")
	}
	if input.Auto {
		return fmt.Errorf("azure auto-complete is not supported by this action")
	}
	if err := c.Preflight(ctx); err != nil {
		return err
	}
	payload := map[string]any{
		"status":                "completed",
		"lastMergeSourceCommit": commitRef{ID: input.HeadSHA},
		"completionOptions":     map[string]any{"mergeStrategy": strategy, "bypassPolicy": false, "deleteSourceBranch": false},
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
