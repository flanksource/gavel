package azuredevops

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/pr/model"
)

const minimumReviewersPolicyType = "fa4e907d-c16b-4a4c-9dfa-4906e5d171dd"

func normalizeReviewers(reviewers []reviewer) ([]model.PRReviewer, error) {
	states := map[int]string{
		10:  model.ReviewApproved,
		5:   model.ReviewApprovedWithSuggestions,
		0:   model.ReviewNoVote,
		-5:  model.ReviewWaitingForAuthor,
		-10: model.ReviewRejected,
	}
	result := make([]model.PRReviewer, 0, len(reviewers))
	for _, reviewer := range reviewers {
		state, ok := states[reviewer.Vote]
		if !ok {
			return nil, fmt.Errorf("unexpected Azure PR reviewer vote %d for %q", reviewer.Vote, reviewer.ID)
		}
		result = append(result, model.PRReviewer{
			ID:       reviewer.ID,
			PRAuthor: model.PRAuthor{Login: reviewer.Login, Name: reviewer.Name, AvatarURL: reviewer.Avatar},
			State:    state,
			Required: reviewer.Required,
		})
	}
	return result, nil
}

func reviewDecision(reviewers []model.PRReviewer, policies []policyEvaluation) string {
	for _, reviewer := range reviewers {
		if reviewer.Required && reviewer.State != model.ReviewApproved && reviewer.State != model.ReviewApprovedWithSuggestions {
			return model.ReviewRequired
		}
	}

	approvalPolicy := false
	for _, evaluation := range policies {
		policy := evaluation.Configuration
		if !policy.Enabled || !policy.Blocking || policy.Deleted || evaluation.Status == "notApplicable" {
			continue
		}
		if !strings.EqualFold(policy.Type.ID, minimumReviewersPolicyType) && !strings.EqualFold(policy.Type.ID, requiredReviewersPolicyType) {
			continue
		}
		approvalPolicy = true
		if evaluation.Status != "approved" {
			return model.ReviewRequired
		}
	}
	if approvalPolicy {
		return model.ReviewApproved
	}

	var approved, waiting bool
	for _, reviewer := range reviewers {
		switch reviewer.State {
		case model.ReviewRejected:
			return "CHANGES_REQUESTED"
		case model.ReviewWaitingForAuthor:
			waiting = true
		case model.ReviewApproved, model.ReviewApprovedWithSuggestions:
			approved = true
		}
	}
	switch {
	case waiting:
		return model.ReviewWaitingForAuthor
	case approved:
		return model.ReviewApproved
	case len(reviewers) > 0:
		return model.ReviewNoVote
	default:
		return ""
	}
}
