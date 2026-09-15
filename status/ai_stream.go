package status

import (
	"context"
	"time"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/clicky"
	clickytask "github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
	clickyai "github.com/flanksource/gavel/ai"
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

type SummaryOptions struct {
	WorkDir    string
	Files      []FileStatus
	MaxWorkers int
	Prompt     *SummaryPrompt
	NewAgent   func(captaincli.AIRuntimeResolved) (clickyai.Agent, error)
}

func StreamAISummaries(ctx context.Context, options SummaryOptions) <-chan AISummaryUpdate {
	files, maxWorkers := options.Files, options.MaxWorkers
	updates := make(chan AISummaryUpdate, len(files)*2+1)
	if len(files) == 0 {
		close(updates)
		return updates
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if maxWorkers <= 0 {
		maxWorkers = defaultAISummaryWorkers
	}
	if options.NewAgent == nil {
		options.NewAgent = func(runtime captaincli.AIRuntimeResolved) (clickyai.Agent, error) {
			return clickyai.NewAgent(runtime.Config)
		}
	}

	group := clicky.StartGroup[string]("status ai summaries", clickytask.WithConcurrency(maxWorkers))
	handles := make([]clickytask.TypedTask[string], 0, len(files))
	for i := range files {
		index := i
		file := files[i]
		handles = append(handles, group.Add(
			"summarize "+file.Path,
			func(taskCtx flanksourceContext.Context, _ *clickytask.Task) (string, error) {
				select {
				case updates <- AISummaryUpdate{Index: index, Status: AISummaryStatusRunning}:
				case <-taskCtx.Done():
					return "", taskCtx.Err()
				}
				return summarizeFileChangeWithAIFunc(taskCtx, options, file)
			},
			clickytask.WithContext(ctx),
			clickytask.WithTaskTimeout(defaultAISummaryItemTimeout),
			clickytask.WithCancellationDrain(),
		))
	}

	completed := make(chan AISummaryUpdate, len(handles))
	for index, handle := range handles {
		go func(index int, handle clickytask.TypedTask[string]) {
			summary, err := handle.GetResult()
			if err != nil {
				completed <- AISummaryUpdate{Index: index, Status: AISummaryStatusFailed, Error: err.Error()}
				return
			}
			completed <- AISummaryUpdate{Index: index, Status: AISummaryStatusDone, Summary: summary}
		}(index, handle)
	}

	go func() {
		defer close(updates)
		for range handles {
			updates <- <-completed
		}
	}()

	return updates
}
