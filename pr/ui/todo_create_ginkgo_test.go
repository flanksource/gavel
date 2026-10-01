package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("create TODO from PR", func() {
	It("keeps selected comment markdown and generated sections in the created body", func() {
		workDir := GinkgoT().TempDir()
		server := &Server{ghOpts: github.Options{WorkDir: workDir}}
		comment := "```markdown\n## Acceptance Criteria\n## Verification\n```"
		payload := todoNewPayload{todoCreatePayload: todoCreatePayload{
			Title: "Address PR feedback",
			Body:  "## Review comments\n\n### Review summary\n\n#### Details\n\n" + comment,
			Criteria: []types.AcceptanceCriterion{
				{Text: "Tests pass"},
			},
			PRVerification: &todoPRVerificationPayload{PRNumber: 17, Repo: "acme/widget", CommentIDs: []int64{901}},
		}}
		raw, err := json.Marshal(payload)
		Expect(err).NotTo(HaveOccurred())
		req := httptest.NewRequest(http.MethodPost, "/api/todos/new?dir="+url.QueryEscape(workDir), bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		server.Handler().ServeHTTP(rec, req)

		Expect(rec.Code).To(Equal(http.StatusCreated), rec.Body.String())
		var response todoNewResponse
		Expect(json.Unmarshal(rec.Body.Bytes(), &response)).To(Succeed())
		Expect(response.Todo.Body).To(ContainSubstring(comment))
		Expect(response.Todo.Criteria).To(ConsistOf(types.AcceptanceCriterion{Text: "Tests pass"}))
		Expect(response.Todo.VerificationMarkdown).To(ContainSubstring("gavel pr status 17 --repo acme/widget --comments 901"))
	})
})
