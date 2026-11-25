package model

import (
	"cmp"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

type PRReviewer struct {
	PRAuthor
	ID       string `json:"id"`
	State    string `json:"state"`
	Required bool   `json:"required"`
}

const (
	ReviewApproved                = "APPROVED"
	ReviewApprovedWithSuggestions = "APPROVED_WITH_SUGGESTIONS"
	ReviewNoVote                  = "NO_VOTE"
	ReviewWaitingForAuthor        = "WAITING_FOR_AUTHOR"
	ReviewRejected                = "REJECTED"
	ReviewRequired                = "REVIEW_REQUIRED"
)

func (r PRReviewer) Pretty() api.Text {
	name := cmp.Or(r.Name, r.Login, r.ID)
	if r.Required {
		name += " (required)"
	}
	label := map[string]string{
		ReviewApproved:                "Approved",
		ReviewApprovedWithSuggestions: "Approved with suggestions",
		ReviewNoVote:                  "No vote",
		ReviewWaitingForAuthor:        "Waiting for author",
		ReviewRejected:                "Rejected",
	}[r.State]
	return clicky.Text(name+": ", "").Append(cmp.Or(label, r.State), ReviewStyle(r.State))
}

type MergeReadiness struct {
	State            string   `json:"state"`
	Reasons          []string `json:"reasons,omitempty"`
	WaitingForBuilds bool     `json:"-"`
	WaitingForMerge  bool     `json:"-"`
}

type StatusSnapshot struct {
	PR   *PRInfo
	Runs map[int64]*WorkflowRun
}

type StatusOptions struct {
	Threads  bool
	Logs     bool
	TailLogs int
}
