package agents_test

import (
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("managed comment run completion", func() {
	It("shares a scope's default session across pages and starts another only when fresh is explicit", func(ctx SpecContext) {
		f := newFixture(ctx, "comment_default_shared", nil)
		target := comments.DeliverTarget{Kind: "run", Prompt: "editor-comments"}
		var sessions []string
		for _, page := range []string{"first", "second", "fresh", "after-fresh"} {
			thread := f.create(ctx, page, "Fix "+page)
			target.Fresh = page == "fresh"
			_, err := f.svc.Deliver(ctx, comments.DeliverRequest{CommentID: thread.ID, ExpectedVersion: thread.Version, RequestID: page, Target: target})
			Expect(err).NotTo(HaveOccurred())
			sessions = append(sessions, f.runOf(ctx, thread.ID)().SessionID)
		}
		_, launched := f.bridge.Snapshot()
		Expect(launched).To(HaveLen(2))
		Expect(sessions).To(Equal([]string{"managed-1", "managed-1", "managed-2", "managed-2"}))
		Expect(launched[0].SessionKey).To(Equal(launched[1].SessionKey))
		Expect(launched[1].Fresh).To(BeTrue())
		Expect(f.bridge.End("managed-1")).To(Succeed())
		thread := f.create(ctx, "explicit", "Continue the original session")
		_, err := f.svc.Deliver(ctx, comments.DeliverRequest{CommentID: thread.ID, ExpectedVersion: thread.Version, RequestID: "explicit", Target: comments.DeliverTarget{Kind: "session", SessionID: "managed-1"}})
		Expect(err).NotTo(HaveOccurred())
		_, launched = f.bridge.Snapshot()
		Expect(launched[len(launched)-1].SessionKey).To(BeEmpty())
		Expect(launched[len(launched)-1].Options.Resume).To(Equal("managed-1"))
	})

	DescribeTable("records asynchronous failures on every thread in a batch",
		func(ctx SpecContext, started bool) {
			f := newFixture(ctx, "comment_completion", nil)
			f.bridge.DeferEvents = !started
			first := f.create(ctx, "first", "Fix the first heading")
			second := f.create(ctx, "second", "Fix the second heading")
			_, err := f.svc.DeliverBatch(ctx, comments.DeliverBatchRequest{
				RequestID: "completion-batch", Target: comments.DeliverTarget{Kind: "run", Prompt: "editor-comments", Fresh: true},
				Threads: []comments.DeliverThread{{ID: first.ID, ExpectedVersion: first.Version}, {ID: second.ID, ExpectedVersion: second.Version}},
			})
			Expect(err).NotTo(HaveOccurred())
			_, launched := f.bridge.Snapshot()
			Expect(launched[0].OnComplete).NotTo(BeNil())
			launched[0].OnComplete(captaincli.PromptRunSummary{Error: "session admission refused"})
			for _, id := range []string{first.ID, second.ID} {
				run := f.runOf(ctx, id)()
				Expect([]string{run.State, run.Error}).To(Equal([]string{"error", "session admission refused"}))
			}
		},
		Entry("before the first provider event", false),
		Entry("after the session started", true),
	)

	It("does not overwrite a newer hand-off when an older run completes", func(ctx SpecContext) {
		f := newFixture(ctx, "comment_completion_retry", nil)
		thread := f.create(ctx, "retry", "Fix the heading")
		target := comments.DeliverTarget{Kind: "run", Prompt: "editor-comments", Fresh: true}
		for _, request := range []string{"first", "retry"} {
			current := f.get(ctx, thread.ID)
			_, err := f.svc.Deliver(ctx, comments.DeliverRequest{CommentID: thread.ID, ExpectedVersion: current.Version, RequestID: request, Target: target})
			Expect(err).NotTo(HaveOccurred())
		}
		_, launched := f.bridge.Snapshot()
		before := f.runOf(ctx, thread.ID)()
		Expect(launched[0].OnComplete).NotTo(BeNil())
		launched[0].OnComplete(captaincli.PromptRunSummary{Error: "old failure"})
		launched[0].OnEvent(0, api.Event{Kind: api.EventSystem, SessionID: "old-session"})
		Expect(f.runOf(ctx, thread.ID)()).To(Equal(before))
	})

	It("records a terminal failure without a provider error string", func(ctx SpecContext) {
		f := newFixture(ctx, "comment_completion_verdict", nil)
		thread := f.create(ctx, "verdict", "Fix the heading")
		_, err := f.svc.Deliver(ctx, comments.DeliverRequest{CommentID: thread.ID, ExpectedVersion: thread.Version, RequestID: "verdict", Target: comments.DeliverTarget{Kind: "run", Prompt: "editor-comments", Fresh: true}})
		Expect(err).NotTo(HaveOccurred())
		_, launched := f.bridge.Snapshot()
		Expect(launched[0].OnComplete).NotTo(BeNil())
		launched[0].OnComplete(captaincli.PromptRunSummary{Success: false, Duration: time.Second.String()})
		Expect(f.runOf(ctx, thread.ID)().State).To(Equal("error"))
		Expect(f.runOf(ctx, thread.ID)().Error).NotTo(BeEmpty())
	})
})
