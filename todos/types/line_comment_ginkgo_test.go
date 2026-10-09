package types

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LineComments", func() {
	const (
		commentID = "11111111-0000-0000-0000-000000000001"
		otherID   = "11111111-0000-0000-0000-000000000002"
		plainID   = "11111111-0000-0000-0000-000000000003"
	)
	anchor := LineCommentAnchor{
		Path: "pkg/widget.go", Side: DiffSideNew, Line: 42, LineText: "return nil",
		Base: "aaaaaaa", Commit: "bbbbbbb", Branch: "shell/0123abcd", AttemptID: "attempt-1",
	}
	otherAnchor := LineCommentAnchor{Path: "pkg/old.go", Side: DiffSideOld, Line: 7, LineText: "x := 1", Commit: "bbbbbbb"}

	payload := func(value any) json.RawMessage {
		GinkgoHelper()
		raw, err := json.Marshal(value)
		Expect(err).NotTo(HaveOccurred())
		return raw
	}
	lineComment := func(id string, anchor LineCommentAnchor, body string) ProviderEvent {
		return ProviderEvent{ID: id, Kind: EventKindComment, Actor: "reviewer", Body: body,
			Payload: payload(LineCommentPayload{Anchor: &anchor})}
	}
	resolution := func(id string, resolved bool) ProviderEvent {
		return ProviderEvent{Kind: EventKindCommentResolved, Payload: payload(CommentResolution{CommentID: id, Resolved: resolved})}
	}

	It("folds anchored comments in event order and applies the latest resolution of each", func() {
		first := lineComment(commentID, anchor, "handle the error")
		second := lineComment(otherID, otherAnchor, "why was this removed?")
		events := []ProviderEvent{
			{ID: plainID, Kind: EventKindComment, Body: "a plain comment", Payload: json.RawMessage(`{}`)},
			first,
			{Kind: "status_changed", Payload: json.RawMessage(`{}`)},
			second,
			resolution(commentID, true),
			resolution(otherID, true),
			resolution(otherID, false),
		}

		comments, err := LineComments(events)

		Expect(err).NotTo(HaveOccurred())
		Expect(comments).To(Equal([]LineComment{
			{Event: first, Anchor: anchor, Resolved: true},
			{Event: second, Anchor: otherAnchor, Resolved: false},
		}))
	})

	It("ignores a resolution of a comment it did not fold, such as a plain one", func() {
		events := []ProviderEvent{
			{ID: plainID, Kind: EventKindComment, Body: "plain", Payload: json.RawMessage(`{}`)},
			resolution(plainID, true),
		}

		comments, err := LineComments(events)

		Expect(err).NotTo(HaveOccurred())
		Expect(comments).To(BeEmpty())
	})

	It("refuses an anchored comment whose payload does not decode", func() {
		events := []ProviderEvent{{ID: commentID, Kind: EventKindComment, Payload: json.RawMessage(`{"anchor":{"line":"forty-two"}}`)}}

		_, err := LineComments(events)

		Expect(err).To(MatchError(ContainSubstring(commentID)))
	})

	It("refuses a resolution whose payload does not decode", func() {
		events := []ProviderEvent{{ID: otherID, Kind: EventKindCommentResolved, Payload: json.RawMessage(`[]`)}}

		_, err := LineComments(events)

		Expect(err).To(MatchError(ContainSubstring(otherID)))
	})
})

var _ = Describe("LineCommentAnchor.Validate", func() {
	valid := LineCommentAnchor{Path: "a.go", Side: DiffSideOld, Line: 1, Commit: "abc1234"}

	It("accepts an anchor with every required field and an empty line text", func() {
		Expect(valid.Validate()).To(Succeed())
	})

	DescribeTable("refuses a missing or invalid required field",
		func(mutate func(*LineCommentAnchor), field string) {
			anchor := valid
			mutate(&anchor)
			Expect(anchor.Validate()).To(MatchError(ContainSubstring(field)))
		},
		Entry("path", func(a *LineCommentAnchor) { a.Path = "  " }, "path"),
		Entry("side", func(a *LineCommentAnchor) { a.Side = "left" }, "side"),
		Entry("line zero", func(a *LineCommentAnchor) { a.Line = 0 }, "line"),
		Entry("negative line", func(a *LineCommentAnchor) { a.Line = -3 }, "line"),
		Entry("commit", func(a *LineCommentAnchor) { a.Commit = "" }, "commit"),
	)
})
