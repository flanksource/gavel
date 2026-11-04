package prwatch

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
)

func fetchActionsSnapshot(_ context.Context, opts WatchOptions) (*PRWatchResult, error) {
	run, err := github.FetchRunJobs(opts.Options, opts.RunID, github.RunLogOptions{JobID: opts.JobID})
	if err != nil {
		return nil, err
	}
	if opts.ExpectedHead != "" && run.HeadSHA != opts.ExpectedHead {
		return nil, fmt.Errorf("workflow %d reports head %s, expected %s", opts.RunID, run.HeadSHA, opts.ExpectedHead)
	}
	if opts.JobID != 0 {
		run.Status, run.Conclusion = scopedPipelineState(run.Jobs)
	}
	result := &PRWatchResult{Runs: map[int64]*model.WorkflowRun{opts.RunID: run}}
	newResultFilters(nil, opts.Actions).apply(result)
	if len(result.Runs) == 0 {
		return nil, fmt.Errorf("--actions %s matched no jobs or workflows in run %d", strings.Join(opts.Actions, ","), opts.RunID)
	}
	if opts.Logs {
		for _, selected := range result.Runs {
			for i := range selected.Jobs {
				job := &selected.Jobs[i]
				if model.NeedsJobLogs(job.Conclusion) {
					if err := github.FetchJobLogs(opts.Options, job, opts.TailLogs); err != nil {
						return nil, fmt.Errorf("fetch job %d logs: %w", job.DatabaseID, err)
					}
				}
			}
		}
	}
	return result, nil
}

func actionsComplete(result *PRWatchResult) bool {
	if len(result.Runs) == 0 {
		return false
	}
	for _, run := range result.Runs {
		if run == nil || !strings.EqualFold(run.Status, "completed") {
			return false
		}
	}
	return true
}

func actionsConclusionFailed(conclusion string) bool {
	return model.IsFailureConclusion(conclusion) || strings.EqualFold(conclusion, "ACTION_REQUIRED") || strings.EqualFold(conclusion, "STALE")
}
