package prwatch

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/github"
)

type PRWatchResult struct {
	PR   *github.PRInfo                `json:"pr"`
	Runs map[int64]*github.WorkflowRun `json:"runs,omitempty"`
}

func (r PRWatchResult) Pretty() api.Text {
	text := r.PR.Pretty()
	text = text.NewLine().NewLine().Add(r.prettyWorkflows())
	return text
}

func (r PRWatchResult) prettyWorkflows() api.Text {
	if len(r.Runs) == 0 && len(r.PR.StatusCheckRollup) == 0 {
		return clicky.Text("  No checks found", "text-gray-500")
	}

	text := clicky.Text("Workflows:", "font-bold")

	rendered := make(map[int64]bool)
	for _, run := range r.Runs {
		if rendered[run.DatabaseID] {
			continue
		}
		rendered[run.DatabaseID] = true
		text = text.NewLine().Add(run.Pretty())
	}

	for _, check := range r.PR.StatusCheckRollup {
		runID, err := github.ExtractRunID(check.DetailsURL)
		if err == nil && rendered[runID] {
			continue
		}
		text = text.NewLine().Append("  ", "").
			Add(github.StatusIcon(check.Status, check.Conclusion)).
			Append(" "+check.Name, "")
	}

	return text
}
