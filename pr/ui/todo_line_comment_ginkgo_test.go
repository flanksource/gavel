package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo line comments API", func() {
	const commentID = "11111111-0000-0000-0000-00000000c0de"
	var (
		dir      string
		server   *Server
		provider *uiTestTODOProvider
		todo     *types.TODO
	)
	anchor := types.LineCommentAnchor{
		Path: "pkg/widget.go", Side: types.DiffSideNew, Line: 42, LineText: "return nil",
		Base: "aaaaaaa", Commit: "bbbbbbb", Branch: "shell/0123abcd", AttemptID: "attempt-1",
	}

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		server = &Server{ghOpts: github.Options{WorkDir: dir}}
		provider = uiTestProviderFor(dir)
		var err error
		todo, err = provider.Create(GinkgoT().Context(), todos.CreateRequest{Title: "Review the widget", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
	})

	patch := func(payload map[string]any) *httptest.ResponseRecorder {
		GinkgoHelper()
		payload["ref"] = todo.ID
		body, err := json.Marshal(payload)
		Expect(err).NotTo(HaveOccurred())
		recorder := httptest.NewRecorder()
		server.handleTodoItem(recorder, httptest.NewRequest(http.MethodPatch,
			"/api/todos/item?dir="+url.QueryEscape(dir), strings.NewReader(string(body))))
		return recorder
	}

	It("records a line comment with its anchor", func() {
		recorder := patch(map[string]any{"lineComment": map[string]any{"anchor": anchor, "body": "handle the error"}})

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		Expect(provider.lineComments).To(Equal([]todos.CommentRequest{{Body: "handle the error", Anchor: &anchor}}))
	})

	DescribeTable("refuses a line comment that cannot be anchored",
		func(mutate func(map[string]any), want string) {
			raw, err := json.Marshal(anchor)
			Expect(err).NotTo(HaveOccurred())
			fields := map[string]any{}
			Expect(json.Unmarshal(raw, &fields)).To(Succeed())
			lineComment := map[string]any{"anchor": fields, "body": "handle the error"}
			mutate(lineComment)

			recorder := patch(map[string]any{"lineComment": lineComment})

			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(ContainSubstring(want))
			Expect(provider.lineComments).To(BeEmpty())
		},
		Entry("empty body", func(c map[string]any) { c["body"] = "  " }, "body"),
		Entry("no anchor", func(c map[string]any) { delete(c, "anchor") }, "anchor"),
		Entry("empty path", func(c map[string]any) { c["anchor"].(map[string]any)["path"] = "" }, "path"),
		Entry("unknown side", func(c map[string]any) { c["anchor"].(map[string]any)["side"] = "left" }, "side"),
		Entry("line zero", func(c map[string]any) { c["anchor"].(map[string]any)["line"] = 0 }, "line"),
		Entry("no commit", func(c map[string]any) { c["anchor"].(map[string]any)["commit"] = "" }, "commit"),
	)

	It("resolves a recorded comment of the todo", func() {
		todo.ProviderEvents = []types.ProviderEvent{{ID: commentID, Kind: types.EventKindComment, Body: "handle the error"}}

		recorder := patch(map[string]any{"resolveComment": map[string]any{"id": commentID, "resolved": true}})

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		Expect(provider.resolutions).To(Equal([]types.CommentResolution{{CommentID: commentID, Resolved: true}}))
	})

	It("answers 404 for a comment id the todo does not carry", func() {
		recorder := patch(map[string]any{"resolveComment": map[string]any{"id": commentID, "resolved": true}})

		Expect(recorder.Code).To(Equal(http.StatusNotFound))
		Expect(recorder.Body.String()).To(ContainSubstring(commentID))
		Expect(provider.resolutions).To(BeEmpty())
	})

	It("refuses a resolution without a comment id", func() {
		recorder := patch(map[string]any{"resolveComment": map[string]any{"id": " ", "resolved": true}})

		Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		Expect(recorder.Body.String()).To(ContainSubstring("resolveComment.id"))
	})

	It("answers 501 when storage cannot resolve comments", func() {
		changes, comments, err := todoUpdateChanges(todoUpdatePayload{
			ResolveComment: &todoResolveCommentPayload{ID: commentID, Resolved: true},
		}, nil)
		Expect(err).NotTo(HaveOccurred())

		_, err = applyTodoPatch(GinkgoT().Context(), parentlessProvider{provider}, todo, changes, comments)

		Expect(err).To(MatchError(ContainSubstring("resolving comments requires native TODO storage")))
	})
})
