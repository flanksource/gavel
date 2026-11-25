package agents

import (
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("a managed session's file write tools", func() {
	DescribeTable("over editor-comments' frontmatter allows",
		func(ctx SpecContext, edits string, options RunOptions, want api.Tools) {
			dir := GinkgoT().TempDir()
			thread := comments.Comment{ID: "c-edits", Body: "Reword the intro", Author: comments.Actor{Kind: "user", ID: "user:reviewer"},
				Target: comments.Target{URI: "gavel://project/acme/files/README.md", ScopeURI: "gavel://project/acme", Kind: "file"}}
			invocation, err := delivery{}.invocation(sessionRequest{Threads: []comments.Comment{thread}, Dir: dir, Agent: previewAgent, Prompt: "editor-comments", Options: options, Edits: edits})
			Expect(err).NotTo(HaveOccurred())
			promptCtx, err := promptRunContext(ctx, dir)
			Expect(err).NotTo(HaveOccurred())
			rendered, err := captaincli.RenderPrompt(promptCtx, invocation)
			Expect(err).NotTo(HaveOccurred())
			Expect(rendered.Input.Permissions.Tools).To(Equal(want))
		},
		Entry("direct mode keeps them allowed", comments.EditsDirect, RunOptions{},
			api.Tools{"Read": api.ToolPolicyAllow, "Glob": api.ToolPolicyAllow, "Grep": api.ToolPolicyAllow, "Edit": api.ToolPolicyAllow, "Write": api.ToolPolicyAllow, "Bash": api.ToolPolicyAllow}),
		Entry("propose mode denies them, keeping a run spec's own rules", comments.EditsPropose,
			RunOptions{Spec: &api.Spec{Permissions: api.Permissions{Tools: api.Tools{"WebFetch": api.ToolPolicyDeny, "Write": api.ToolPolicyAllow}}}},
			api.Tools{"Read": api.ToolPolicyAllow, "Glob": api.ToolPolicyAllow, "Grep": api.ToolPolicyAllow, "Bash": api.ToolPolicyAllow, "WebFetch": api.ToolPolicyDeny,
				"Edit": api.ToolPolicyDeny, "Write": api.ToolPolicyDeny, "MultiEdit": api.ToolPolicyDeny, "NotebookEdit": api.ToolPolicyDeny}),
	)
})
