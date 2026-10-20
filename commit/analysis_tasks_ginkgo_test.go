package commit

import (
	"context"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/task"
	flanksourcecontext "github.com/flanksource/commons/context"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("commit analysis inside an active workflow", func() {
	It("generates a message without waiting for unrelated foreground tasks", func() {
		started, release := make(chan struct{}), make(chan struct{})
		finished := make(chan struct{})
		previous := analyzeCommitMessageWithAIFunc
		active := clicky.StartTask("active workflow", func(_ flanksourcecontext.Context, _ *task.Task) (struct{}, error) {
			close(started)
			<-release
			return struct{}{}, nil
		})
		DeferCleanup(func() {
			close(release)
			_, _ = active.GetResult()
			<-finished
			analyzeCommitMessageWithAIFunc = previous
		})
		Eventually(started).Should(BeClosed())
		analyzeCommitMessageWithAIFunc = func(_ context.Context, analysis models.CommitAnalysis, _ clickyai.Agent, _ git.AnalyzeOptions) (models.CommitAnalysis, error) {
			analysis.CommitType, analysis.Subject = models.CommitTypeFix, "complete the turn"
			return analysis, nil
		}
		type result struct {
			analysis commitAIAnalysis
			err      error
		}
		done := make(chan result, 1)
		go func() {
			defer close(finished)
			analysis, err := generateCommitAnalysisWithAgent(context.Background(), "diff", &agentLifecycleProbe{}, git.AnalyzeOptions{})
			done <- result{analysis, err}
		}()
		var got result
		Eventually(done, time.Second).Should(Receive(&got))
		Expect(got.err).NotTo(HaveOccurred())
		Expect(got.analysis.Message).To(Equal("fix: complete the turn"))
	})
})
