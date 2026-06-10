package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/gavel/github/activity"
)

// graphqlEndpoint returns the GraphQL endpoint, honoring GITHUB_API_URL the
// same way githubAPIBase() does so tests and GitHub Enterprise can redirect it.
func graphqlEndpoint() string {
	return githubAPIBase() + "/graphql"
}

// postGraphQL runs a single GraphQL query against endpoint, records the call in
// the shared activity recorder, and returns the raw response body for the
// caller to unmarshal into its own typed shape. It surfaces HTTP-level failures
// and top-level GraphQL `errors[]` as Go errors — callers never have to repeat
// that boilerplate. kind tags the call for activity stats (e.g. KindSearch,
// KindGraphQL).
func postGraphQL(token, endpoint, kind, query string, variables map[string]any) ([]byte, *RateLimit, error) {
	body := map[string]any{"query": query, "variables": variables}

	ctx := context.Background()
	client := newClient(token)

	start := time.Now()
	resp, err := client.R(ctx).
		Header("Content-Type", "application/json").
		Post(endpoint, body)
	if err != nil {
		activity.Shared().Record(activity.Entry{
			Method: "POST", URL: "/graphql", Kind: kind,
			Duration: time.Since(start), Error: err.Error(),
		})
		return nil, nil, fmt.Errorf("GraphQL request: %w", err)
	}
	rl := ParseRateLimit(resp.Header)
	if !resp.IsOK() {
		respBody, _ := resp.AsString()
		activity.Shared().Record(activity.Entry{
			Method: "POST", URL: "/graphql", Kind: kind,
			StatusCode: resp.StatusCode, Duration: time.Since(start),
			SizeBytes: len(respBody),
			Error:     fmt.Sprintf("status %d", resp.StatusCode),
		})
		return nil, rl, fmt.Errorf("GraphQL request: status %d: %s", resp.StatusCode, respBody)
	}

	respBody, _ := resp.AsString()
	activity.Shared().Record(activity.Entry{
		Method: "POST", URL: "/graphql", Kind: kind,
		StatusCode: resp.StatusCode, Duration: time.Since(start),
		SizeBytes: len(respBody),
	})

	if errs := graphQLErrors([]byte(respBody)); errs != nil {
		return nil, rl, errs
	}
	return []byte(respBody), rl, nil
}

// graphQLErrors decodes just the top-level `errors` array and returns a joined
// error if any are present, nil otherwise. Kept separate from response decoding
// so each caller can unmarshal its own `data` shape while sharing error checks.
func graphQLErrors(body []byte) error {
	var envelope struct {
		Errors []graphQLError `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("parse GraphQL response: %w", err)
	}
	if len(envelope.Errors) == 0 {
		return nil
	}
	msgs := make([]string, len(envelope.Errors))
	for i, e := range envelope.Errors {
		msgs[i] = e.Message
	}
	return fmt.Errorf("GraphQL errors: %s", strings.Join(msgs, "; "))
}

// summarizeRollup classifies a status-check rollup's contexts into pass/fail/
// running/pending counts. Shared by the PR search path and the org default-
// branch status path so the classification logic lives in exactly one place.
func summarizeRollup(rollup *graphQLStatusCheckRollup) *CheckSummary {
	if rollup == nil {
		return nil
	}
	var cs CheckSummary
	for _, check := range rollup.Contexts.Nodes {
		sc := check.toStatusCheck()
		switch {
		case sc.Status == "COMPLETED" && (sc.Conclusion == "SUCCESS" || sc.Conclusion == "NEUTRAL" || sc.Conclusion == "SKIPPED"):
			cs.Passed++
		case sc.Status == "COMPLETED" && (sc.Conclusion == "FAILURE" || sc.Conclusion == "TIMED_OUT" || sc.Conclusion == "STARTUP_FAILURE"):
			cs.Failed++
			cs.Failures = append(cs.Failures, FailedCheck{
				Name:       sc.Name,
				DetailsURL: sc.DetailsURL,
			})
		case sc.Status == "IN_PROGRESS":
			cs.Running++
		default:
			cs.Pending++
		}
	}
	return &cs
}
