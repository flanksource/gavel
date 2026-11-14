package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/pr/blame"
)

const requiredReviewersPolicyType = "fd2167ab-b0be-447a-8ec8-39368250530e"

type policyReviewer struct {
	ID    string `json:"id"`
	Name  string `json:"displayName"`
	Login string `json:"uniqueName"`
	Group bool   `json:"isContainer"`
}

func (c *Client) reviewerPolicies(ctx context.Context, number int) ([]blame.PolicyReviewer, error) {
	policies, err := collection[policyEvaluation](ctx, c, c.apiPath("policy/evaluations"), url.Values{
		"artifactId":  {fmt.Sprintf("vstfs:///CodeReview/CodeReviewId/%s/%d", c.info.Project.ID, number)},
		"api-version": {"7.1-preview.1"}, "includeNotApplicable": {"true"}})
	if err != nil {
		return nil, err
	}
	var result []blame.PolicyReviewer
	var identities map[string]policyReviewer
	for _, evaluation := range policies {
		policy := evaluation.Configuration
		if !policy.Enabled || policy.Deleted || !strings.EqualFold(policy.Type.ID, requiredReviewersPolicyType) || evaluation.Status == "notApplicable" {
			continue
		}
		switch evaluation.Status {
		case "approved", "rejected", "queued", "running":
		default:
			return nil, fmt.Errorf("Azure reviewer policy %d has unusable evaluation status %q", policy.ID, evaluation.Status)
		}
		var settings struct {
			Reviewers []string `json:"requiredReviewerIds"`
			Patterns  []string `json:"filenamePatterns"`
		}
		if err := json.Unmarshal(policy.Settings, &settings); err != nil {
			return nil, fmt.Errorf("decode Azure reviewer policy %d: %w", policy.ID, err)
		}
		if policy.ID <= 0 || len(settings.Reviewers) == 0 {
			return nil, fmt.Errorf("Azure reviewer policy %d requires requiredReviewerIds", policy.ID)
		}
		if identities == nil {
			reviewers, err := collection[policyReviewer](ctx, c, c.repositoryPath()+"/pullrequests/"+strconv.Itoa(number)+"/reviewers", nil)
			if err != nil {
				return nil, err
			}
			identities = make(map[string]policyReviewer, len(reviewers))
			for _, reviewer := range reviewers {
				identities[strings.ToLower(reviewer.ID)] = reviewer
			}
		}
		for _, id := range settings.Reviewers {
			reviewer, found := identities[strings.ToLower(id)]
			if !found || reviewer.ID == "" || reviewer.Name == "" {
				return nil, fmt.Errorf("Azure reviewer policy %d identity %q is missing from PR reviewers; refresh its policy evaluation", policy.ID, id)
			}
			identity := blame.Identity{ID: reviewer.ID, Name: reviewer.Name}
			if !reviewer.Group {
				if reviewer.Login == "" {
					return nil, fmt.Errorf("Azure reviewer %s has no uniqueName", id)
				}
				identity.Email = reviewer.Login
			}
			result = append(result, blame.PolicyReviewer{Identity: identity, Policy: blame.Policy{ID: policy.ID, Name: policy.Type.Name, Required: policy.Blocking, Status: evaluation.Status, Patterns: settings.Patterns}})
		}
	}
	return result, nil
}
