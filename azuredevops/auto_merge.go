package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type autoCompleteIdentity struct {
	ID uuid.UUID `json:"id"`
}

func (c *Client) EnableAutoMerge(ctx context.Context, number int, method string) error {
	strategy, ok := mergeStrategies[strings.ToLower(strings.TrimSpace(method))]
	if number <= 0 || !ok {
		return fmt.Errorf("azure auto-complete requires a positive PR number and rebase/squash/merge method (got %d and %q)", number, method)
	}
	if err := c.Preflight(ctx); err != nil {
		return err
	}
	actor, err := c.authenticatedUser(ctx)
	if err != nil {
		return err
	}
	payload := map[string]any{"autoCompleteSetBy": actor, "completionOptions": mergeCompletionOptions(strategy)}
	data, _, err := c.request(ctx, http.MethodPatch, c.repositoryPath()+"/pullrequests/"+strconv.Itoa(number), nil, payload)
	if err != nil {
		return err
	}
	var result struct {
		ID                int                  `json:"pullRequestId"`
		Status            string               `json:"status"`
		AutoCompleteSetBy autoCompleteIdentity `json:"autoCompleteSetBy"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("decode Azure PR auto-complete: %w", err)
	}
	enabled := result.Status == "active" && result.AutoCompleteSetBy == actor
	if result.ID != number || (!enabled && result.Status != "completed") {
		return fmt.Errorf("azure did not confirm auto-complete for PR #%d: returned PR #%d with status %q", number, result.ID, result.Status)
	}
	return nil
}

func (c *Client) authenticatedUser(ctx context.Context) (autoCompleteIdentity, error) {
	query := url.Values{"api-version": {"7.1-preview.1"}}
	data, _, err := c.request(ctx, http.MethodGet, c.projectPath("", "connectionData"), query, nil)
	if err != nil {
		return autoCompleteIdentity{}, err
	}
	var connection struct {
		AuthenticatedUser autoCompleteIdentity `json:"authenticatedUser"`
	}
	if err := json.Unmarshal(data, &connection); err != nil {
		return autoCompleteIdentity{}, fmt.Errorf("decode Azure connection data: %w", err)
	}
	if connection.AuthenticatedUser.ID == uuid.Nil {
		return autoCompleteIdentity{}, fmt.Errorf("azure connection data is missing authenticated user ID")
	}
	return connection.AuthenticatedUser, nil
}
