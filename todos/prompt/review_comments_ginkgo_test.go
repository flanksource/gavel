package prompt

import (
	"encoding/json"
	"strings"

	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = Describe("review comments in the run prompt", func() {
	lineComment := func(id, actor, body string, anchor types.LineCommentAnchor) types.ProviderEvent {
		GinkgoHelper()
		payload, err := json.Marshal(types.LineCommentPayload{Anchor: &anchor})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return types.ProviderEvent{ID: id, Kind: types.EventKindComment, Actor: actor, Body: body, Payload: payload}
	}
	resolved := func(id string) types.ProviderEvent {
		GinkgoHelper()
		payload, err := json.Marshal(types.CommentResolution{CommentID: id, Resolved: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return types.ProviderEvent{Kind: types.EventKindCommentResolved, Payload: payload}
	}
	render := func(todo *types.TODO, opts Options) string {
		GinkgoHelper()
		if opts.Mode == "" {
			opts.Mode = types.ModeRun
		}
		req, _, err := renderResolvedForTest([]*types.TODO{todo}, opts)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return req.Prompt.User
	}

	It("renders unresolved line comments grouped by file and leaves them out of the comments section", func() {
		todo := newTestTODO("fix-widget", "Fix the widget")
		todo.ProviderEvents = []types.ProviderEvent{
			{ID: "c0", Kind: types.EventKindComment, Actor: "maintainer", Body: "Keep the API stable"},
			lineComment("c1", "reviewer", "Handle the error here", types.LineCommentAnchor{
				Path: "pkg/widget.go", Side: types.DiffSideNew, Line: 42, LineText: "return nil", Commit: "abc1234def5678",
			}),
			lineComment("c2", "reviewer", "Already fixed", types.LineCommentAnchor{
				Path: "pkg/widget.go", Side: types.DiffSideNew, Line: 7, LineText: "x := 1", Commit: "abc1234def5678",
			}),
			lineComment("c3", "", "Why was this removed?", types.LineCommentAnchor{
				Path: "README.md", Side: types.DiffSideOld, Line: 3, LineText: "# Widgets", Commit: "abc1234def5678",
			}),
			lineComment("c4", "reviewer", "Also log it", types.LineCommentAnchor{
				Path: "pkg/widget.go", Side: types.DiffSideNew, Line: 40, LineText: "if err != nil {", Commit: "abc1234def5678",
			}),
			resolved("c2"),
		}

		prompt := render(todo, Options{})

		gomega.Expect(prompt).To(gomega.ContainSubstring("## Comments\n\n**maintainer:**\n\nKeep the API stable\n\n## Review comments\n\n" +
			"### `README.md:3` (old side, at abc1234)\n\n```\n# Widgets\n```\n\n**unknown:** Why was this removed?\n\n" +
			"### `pkg/widget.go:40` (new side, at abc1234)\n\n```go\nif err != nil {\n```\n\n**reviewer:** Also log it\n\n" +
			"### `pkg/widget.go:42` (new side, at abc1234)\n\n```go\nreturn nil\n```\n\n**reviewer:** Handle the error here\n\n"))
		gomega.Expect(prompt).NotTo(gomega.ContainSubstring("Already fixed"))
		gomega.Expect(strings.Count(prompt, "Handle the error here")).To(gomega.Equal(1), "a line comment is not repeated as a plain comment")
	})

	It("omits the section when every line comment is resolved", func() {
		todo := newTestTODO("fix-widget", "Fix the widget")
		todo.ProviderEvents = []types.ProviderEvent{
			lineComment("c1", "reviewer", "Handle the error here", types.LineCommentAnchor{
				Path: "pkg/widget.go", Side: types.DiffSideNew, Line: 42, Commit: "abc1234",
			}),
			resolved("c1"),
		}

		gomega.Expect(render(todo, Options{})).NotTo(gomega.ContainSubstring("## Review comments"))
	})

	It("refuses to render a todo whose line comment payload does not decode", func() {
		todo := newTestTODO("fix-widget", "Fix the widget")
		todo.ProviderEvents = []types.ProviderEvent{
			{ID: "c-bad", Kind: types.EventKindComment, Body: "broken", Payload: json.RawMessage(`{"anchor":{"line":"x"}}`)},
		}

		_, _, err := Render([]*types.TODO{todo}, Options{Mode: types.ModeRun})

		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("c-bad")))
	})
})
