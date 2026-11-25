package blame

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

type Request struct {
	Context                     context.Context
	Dir                         string
	Number, Limit, HistoryLimit int
	Exclude                     []string
	Since                       time.Time
}

type Recommendations struct {
	*Result
	Number         int    `json:"number"`
	URL            string `json:"url"`
	Title          string `json:"title"`
	Author         string `json:"author"`
	TotalReviewers int    `json:"totalReviewers"`
	Provider       string `json:"provider"`
}

func (r Recommendations) Pretty() api.Text {
	text := clicky.Text(fmt.Sprintf("Recommended reviewers for PR #%d: %s", r.Number, r.Title), "font-bold")
	text = text.NewLine().Append(fmt.Sprintf("%d changed files · showing %d of %d candidates", len(r.Files), len(r.Reviewers), r.TotalReviewers), "muted")
	if r.Since != nil {
		text = text.NewLine().Append("History since: "+r.Since.Format(time.RFC3339), "muted")
	}
	if r.CodeownersFile != "" {
		text = text.NewLine().Append("Ownership: "+r.CodeownersFile+" at the base branch tip", "muted")
	}
	if len(r.Reviewers) == 0 {
		return text.NewLine().Append("No eligible reviewers found in ownership rules or file history.", "muted")
	}
	for i, reviewer := range r.Reviewers {
		name := reviewer.Login
		if name == "" {
			name = reviewer.Name
			if reviewer.Email != "" {
				name += " <" + reviewer.Email + ">"
			}
		}
		text = text.NewLine().Append(fmt.Sprintf("%d. %s", i+1, name), "font-bold")
		for _, policy := range reviewer.Policies {
			requirement := "optional"
			if policy.Required {
				requirement = "required"
			}
			text = text.NewLine().Append(fmt.Sprintf("  %s (policy %d, %s)", policy.Name, policy.ID, requirement), "")
			if policy.Status != "" {
				text = text.Append(" ["+policy.Status+"]", "muted")
			}
			if len(policy.Patterns) > 0 {
				text = text.Append(": "+strings.Join(policy.Patterns, "; "), "muted")
			}
		}
		for _, evidence := range reviewer.Files {
			reasons := []string{}
			if evidence.CodeOwner {
				reasons = append(reasons, fmt.Sprintf("CODEOWNERS %s:%d (%s)", r.CodeownersFile, evidence.RuleLine, evidence.Rule))
			}
			if evidence.Lines > 0 {
				reasons = append(reasons, fmt.Sprintf("%d changed lines", evidence.Lines))
			}
			if evidence.AdjacentLines > 0 {
				reasons = append(reasons, fmt.Sprintf("%d adjacent lines", evidence.AdjacentLines))
			}
			if evidence.Commits > 0 {
				reasons = append(reasons, fmt.Sprintf("%d file commits", evidence.Commits))
			}
			text = text.NewLine().Append("  "+evidence.Path+": "+strings.Join(reasons, ", "), "")
		}
	}
	return text
}
