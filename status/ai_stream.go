package status

import (
	"context"
	"time"

	clickyai "github.com/flanksource/clicky/ai"
	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
)

type AISummaryStatus string

const (
	AISummaryStatusIdle    AISummaryStatus = ""
	AISummaryStatusPending AISummaryStatus = "pending"
	AISummaryStatusRunning AISummaryStatus = "running"
	AISummaryStatusDone    AISummaryStatus = "done"
	AISummaryStatusFailed  AISummaryStatus = "failed"
)

type AISummaryUpdate struct {
	Index   int
	Status  AISummaryStatus
	Summary string
	Error   string
}

type aiSummaryResult struct {
	Index    int
	HasIndex bool
	Summary  string
}

const (
	defaultAISummaryWorkers     = 4
	defaultAISummaryItemTimeout = 5 * time.Minute
)

func (r *Result) PrepareAISummaries() {
	if r == nil {
		return
	}
	for i := range r.Files {
		r.Files[i].AISummary = ""
		r.Files[i].AIError = ""
		r.Files[i].AIStatus = AISummaryStatusPending
	}
}

func (r *Result) ApplyAISummaryUpdate(update AISummaryUpdate) {
	if r == nil || update.Index < 0 || update.Index >= len(r.Files) {
		return
	}

	file := &r.Files[update.Index]
	file.AIStatus = update.Status

	switch update.Status {
	case AISummaryStatusDone:
		file.AISummary = update.Summary
		file.AIError = ""
	case AISummaryStatusFailed:
		file.AISummary = ""
		file.AIError = update.Error
	default:
		if update.Error != "" {
			file.AIError = update.Error
		}
	}
}

func StreamAISummaries(ctx context.Context, workDir string, agent clickyai.Agent, files []FileStatus, maxWorkers int) <-chan AISummaryUpdate {
	updates := make(chan AISummaryUpdate, len(files)*2+1)
	if agent == nil || len(files) == 0 {
		close(updates)
		return updates
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if maxWorkers <= 0 {
		maxWorkers = defaultAISummaryWorkers
	}

	items := make([]func(logger.Logger) (aiSummaryResult, error), 0, len(files))
	for i := range files {
		index := i
		file := files[i]
		items = append(items, func(logger.Logger) (aiSummaryResult, error) {
			select {
			case updates <- AISummaryUpdate{Index: index, Status: AISummaryStatusRunning}:
			case <-ctx.Done():
				return aiSummaryResult{Index: index, HasIndex: true}, ctx.Err()
			}

			summary, err := summarizeFileChangeWithAIFunc(ctx, workDir, agent, file)
			if err != nil {
				return aiSummaryResult{Index: index, HasIndex: true}, err
			}

			return aiSummaryResult{
				Index:    index,
				HasIndex: true,
				Summary:  summary,
			}, nil
		})
	}

	go func() {
		defer close(updates)

		previousNoRender := clickytask.IsNoRender()
		clickytask.SetNoRender(true)
		defer clickytask.SetNoRender(previousNoRender)

		batch := clickytask.Batch[aiSummaryResult]{
			Name:        "status ai summaries",
			Items:       items,
			MaxWorkers:  maxWorkers,
			ItemTimeout: defaultAISummaryItemTimeout,
			Timeout:     time.Duration(len(items)+1) * defaultAISummaryItemTimeout,
		}

		for result := range batch.Run() {
			if result.Error != nil {
				if !result.Value.HasIndex {
					continue
				}
				updates <- AISummaryUpdate{
					Index:  result.Value.Index,
					Status: AISummaryStatusFailed,
					Error:  result.Error.Error(),
				}
				continue
			}

			updates <- AISummaryUpdate{
				Index:   result.Value.Index,
				Status:  AISummaryStatusDone,
				Summary: result.Value.Summary,
			}
		}
	}()

	return updates
}
