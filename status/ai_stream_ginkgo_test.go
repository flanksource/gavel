package status

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/clicky"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AI summary cancellation", func() {
	It("drains every active summary before closing the update stream", func() {
		files := []FileStatus{
			{Path: "first.go"},
			{Path: "second.go"},
			{Path: "third.go"},
			{Path: "fourth.go"},
		}
		started := make(chan string, len(files))
		stopped := make(chan string, len(files))
		release := make(chan struct{})

		previous := summarizeFileChangeWithAIFunc
		summarizeFileChangeWithAIFunc = func(ctx context.Context, _ SummaryOptions, file FileStatus) (string, error) {
			started <- file.Path
			select {
			case <-ctx.Done():
				stopped <- file.Path
				return "", ctx.Err()
			case <-release:
				return "released", nil
			}
		}
		DeferCleanup(func() {
			summarizeFileChangeWithAIFunc = previous
			close(release)
			clicky.CancelAllGlobalTasks()
			clicky.WaitForGlobalCompletionSilent()
			clicky.ClearGlobalTasks()
		})

		updates := StreamAISummaries(context.Background(), SummaryOptions{Files: files, MaxWorkers: 2})
		collected := make(chan []AISummaryUpdate, 1)
		go func() {
			var result []AISummaryUpdate
			for update := range updates {
				result = append(result, update)
			}
			collected <- result
		}()

		Eventually(func() int { return len(started) }).Should(Equal(2))
		clicky.CancelAllGlobalTasks()

		var result []AISummaryUpdate
		Eventually(collected, time.Second).Should(Receive(&result))
		Expect(stopped).To(HaveLen(2))
		Expect(started).To(HaveLen(2))

		running := 0
		failures := make(map[int]string, len(files))
		for _, update := range result {
			switch update.Status {
			case AISummaryStatusRunning:
				running++
			case AISummaryStatusFailed:
				failures[update.Index] = update.Error
			}
		}
		Expect(running).To(Equal(2))
		Expect(failures).To(HaveLen(len(files)))
		for index := range files {
			Expect(failures).To(HaveKeyWithValue(index, context.Canceled.Error()), fmt.Sprintf("file index %d", index))
		}
	})
})
