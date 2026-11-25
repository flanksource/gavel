package agents

import (
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("resuming a managed session", func() {
	It("renders the recorded prompt as a run of the session it resumes", func(ctx SpecContext) {
		dir := GinkgoT().TempDir()
		thread := comments.Comment{ID: "c-resume", Body: "Mention the fix", Author: comments.Actor{Kind: "user", ID: "user:reviewer"},
			Target: comments.Target{URI: "gavel://project/acme/files/CHANGELOG.md", ScopeURI: "gavel://project/acme", Kind: "file"}}
		invocation, err := delivery{}.invocation(sessionRequest{Threads: []comments.Comment{thread}, Dir: dir, Agent: previewAgent, Prompt: "editor-comments"})
		Expect(err).NotTo(HaveOccurred())
		invocation.Options.Resume = "managed-session-1"
		promptCtx, err := promptRunContext(ctx, dir)
		Expect(err).NotTo(HaveOccurred())
		rendered, err := captaincli.RenderPrompt(promptCtx, invocation)
		Expect(err).NotTo(HaveOccurred())
		Expect(rendered.Input.SessionID).To(Equal("managed-session-1"))
	})
})
