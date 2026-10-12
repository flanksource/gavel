package native_test

import (
	"encoding/json"

	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("issue comments", Ordered, func() {
	var fixture *coordinatorFixture

	BeforeAll(func(ctx SpecContext) {
		fixture = openCoordinatorFixture(ctx, "gavel_native_comments")
	})

	anchor := types.LineCommentAnchor{
		Path: "pkg/widget.go", Side: types.DiffSideNew, Line: 42, LineText: "return nil",
		Commit: "bbbbbbb", Branch: "shell/0123abcd",
	}

	It("stores a line comment's anchor payload and folds its resolution from the history", func(ctx SpecContext) {
		issue := fixture.createIssue(ctx, "Review the widget")

		comment, err := fixture.repository.AddComment(ctx, issue.ID, issue.Version, native.CommentInput{
			Actor: "reviewer", Body: "handle the error", Payload: types.LineCommentPayload{Anchor: &anchor},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(comment.Kind).To(Equal(types.EventKindComment))

		resolved, err := fixture.repository.ResolveComment(ctx, issue.ID, issue.Version+1, "reviewer", comment.ID, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Kind).To(Equal(types.EventKindCommentResolved))
		var payload types.CommentResolution
		Expect(json.Unmarshal(resolved.Payload, &payload)).To(Succeed())
		Expect(payload).To(Equal(types.CommentResolution{CommentID: comment.ID.String(), Resolved: true}))

		events, err := fixture.repository.ListEvents(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		var history []types.ProviderEvent
		for _, event := range events {
			history = append(history, types.ProviderEvent{ID: event.ID.String(), Kind: event.Kind, Body: event.Body, Payload: event.Payload})
		}
		comments, err := types.LineComments(history)
		Expect(err).NotTo(HaveOccurred())
		Expect(comments).To(HaveLen(1))
		Expect(comments[0].Anchor).To(Equal(anchor))
		Expect(comments[0].Resolved).To(BeTrue())
	})

	It("refuses an empty comment body", func(ctx SpecContext) {
		issue := fixture.createIssue(ctx, "Empty comment")

		_, err := fixture.repository.AddComment(ctx, issue.ID, issue.Version, native.CommentInput{Actor: "reviewer", Body: "  "})

		Expect(err).To(MatchError(native.ErrInvalidInput))
	})

	It("refuses to resolve a comment that does not exist", func(ctx SpecContext) {
		issue := fixture.createIssue(ctx, "Missing comment")
		missing := uuid.New()

		_, err := fixture.repository.ResolveComment(ctx, issue.ID, issue.Version, "reviewer", missing, true)

		Expect(err).To(MatchError(native.ErrNotFound))
		Expect(err).To(MatchError(ContainSubstring(missing.String())))
	})

	It("refuses to resolve another issue's comment", func(ctx SpecContext) {
		owner := fixture.createIssue(ctx, "Owner")
		other := fixture.createIssue(ctx, "Other")
		comment, err := fixture.repository.AddComment(ctx, owner.ID, owner.Version, native.CommentInput{Actor: "reviewer", Body: "mine"})
		Expect(err).NotTo(HaveOccurred())

		_, err = fixture.repository.ResolveComment(ctx, other.ID, other.Version, "reviewer", comment.ID, true)

		Expect(err).To(MatchError(native.ErrNotFound))
	})

	It("refuses to resolve an event that is not a comment", func(ctx SpecContext) {
		issue := fixture.createIssue(ctx, "Not a comment")
		events, err := fixture.repository.ListEvents(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(events).NotTo(BeEmpty())

		_, err = fixture.repository.ResolveComment(ctx, issue.ID, issue.Version, "reviewer", events[0].ID, true)

		Expect(err).To(MatchError(native.ErrInvalidInput))
		Expect(err).To(MatchError(ContainSubstring(events[0].Kind)))
	})
})
