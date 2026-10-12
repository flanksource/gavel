package azuredevops

import (
	"fmt"
	"github.com/flanksource/gavel/pr/model"
	"sort"
	"strings"
)

const buildPolicyType = "0609b952-1397-4640-95ec-e00a01b2c241"

func readiness(pr pullRequest, policies []policyEvaluation) *model.MergeReadiness {
	if pr.Status != "active" {
		return nil
	}
	r := &model.MergeReadiness{State: "ready"}
	switch pr.MergeStatus {
	case "succeeded":
	case "conflicts":
		r.State = "conflicting"
		r.Reasons = append(r.Reasons, "conflicts with target branch")
	case "queued":
		r.State = "pending"
		r.WaitingForMerge = true
		r.Reasons = append(r.Reasons, "merge computation queued")
	case "failure", "rejectedByPolicy":
		r.State = "blocked"
		r.Reasons = append(r.Reasons, "merge "+pr.MergeStatus)
	default:
		r.State = "unknown"
		r.Reasons = append(r.Reasons, "merge status: "+pr.MergeStatus)
	}
	if pr.MergeFailure != "" {
		r.Reasons = append(r.Reasons, pr.MergeFailure)
	}
	if pr.Draft {
		if r.State == "ready" || r.State == "pending" {
			r.State = "blocked"
		}
		r.Reasons = append(r.Reasons, "PR is a draft")
	}
	applyPolicies(r, policies)
	return r
}

func applyPolicies(r *model.MergeReadiness, policies []policyEvaluation) {
	for _, p := range policies {
		if !p.Configuration.Enabled || !p.Configuration.Blocking || p.Configuration.Deleted {
			continue
		}
		switch p.Status {
		case "approved", "notApplicable":
			continue
		case "queued", "running":
			if p.Configuration.Type.ID == buildPolicyType {
				r.WaitingForBuilds = true
				if r.State == "ready" {
					r.State = "pending"
				}
			} else if r.State == "ready" || r.State == "pending" {
				r.State = "blocked"
			}
		case "rejected", "broken":
			if r.State == "ready" || r.State == "pending" {
				r.State = "blocked"
			}
		default:
			if r.State != "conflicting" {
				r.State = "unknown"
			}
		}
		r.Reasons = append(r.Reasons, fmt.Sprintf("%s (policy %d): %s", p.Configuration.Type.Name, p.Configuration.ID, p.Status))
	}
}

func currentBuilds(pr pullRequest, builds []build) []build {
	latest := map[int64]build{}
	for _, b := range builds {
		expected := pr.SourceCommit.ID
		if b.SourceBranch == fmt.Sprintf("refs/pull/%d/merge", pr.ID) {
			expected = pr.MergeCommit.ID
		} else if b.SourceBranch != pr.Source {
			continue
		}
		if expected == "" || b.SourceVersion != expected {
			continue
		}
		old, found := latest[b.Definition.ID]
		if !found || b.QueueTime.After(old.QueueTime) || (b.QueueTime.Equal(old.QueueTime) && b.ID > old.ID) {
			latest[b.Definition.ID] = b
		}
	}
	out := make([]build, 0, len(latest))
	for _, b := range latest {
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Definition.ID < out[j].Definition.ID })
	return out
}

func executionState(state, result string) (string, string, error) {
	switch state {
	case "pending", "notStarted", "postponed":
		return "QUEUED", "", nil
	case "inProgress", "cancelling":
		return "IN_PROGRESS", "", nil
	case "completed":
		conclusions := map[string]string{"succeeded": "SUCCESS", "failed": "FAILURE", "canceled": "CANCELLED", "skipped": "SKIPPED", "abandoned": "CANCELLED", "partiallySucceeded": "PARTIAL_SUCCESS", "succeededWithIssues": "PARTIAL_SUCCESS"}
		if conclusion, ok := conclusions[result]; ok {
			return "COMPLETED", conclusion, nil
		}
	}
	return "", "", fmt.Errorf("unexpected Azure execution state %q/result %q", strings.TrimSpace(state), strings.TrimSpace(result))
}
