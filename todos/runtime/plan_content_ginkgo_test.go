package runtime

import (
	"context"
	"os"
	"path/filepath"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("plan protocol normalization", func() {
	DescribeTable("persists Markdown without assistant protocol wrappers",
		func(input, expected string) {
			content, path, err := (&Provider{}).planResultContent(context.Background(), &todos.ExecutionResult{
				Plan: &types.PlanResult{Content: input},
			}, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(BeEmpty())
			Expect(content).To(Equal(expected))
		},
		Entry("Codex plan", "<proposed_plan>\n# Plan\n\n1. Fix the editor.\n</proposed_plan>", "# Plan\n\n1. Fix the editor."),
		Entry("plan with trailing commentary", "<proposed_plan># Plan</proposed_plan>\nReady for review.", "# Plan"),
		Entry("ordinary Markdown", "  # Plan\n\n1. Fix the editor.  ", "# Plan\n\n1. Fix the editor."),
		Entry("literal protocol example", "```xml\n<proposed_plan>example</proposed_plan>\n```", "```xml\n<proposed_plan>example</proposed_plan>\n```"),
		Entry("embedded protocol example", "Document <proposed_plan>example</proposed_plan>", "Document <proposed_plan>example</proposed_plan>"),
	)

	It("normalizes a plan read from the reported file", func() {
		path := filepath.Join(GinkgoT().TempDir(), "plan.md")
		Expect(os.WriteFile(path, []byte("<proposed_plan># File plan</proposed_plan>"), 0o600)).To(Succeed())
		content, resolvedPath, err := (&Provider{}).planResultContent(context.Background(), &todos.ExecutionResult{
			Plan: &types.PlanResult{Path: path, Content: "Short summary"},
		}, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(resolvedPath).To(Equal(path))
		Expect(content).To(Equal("# File plan"))
	})

	It("normalizes a plan recovered from Captain", func() {
		original := resolveSessionPlan
		DeferCleanup(func() { resolveSessionPlan = original })
		resolveSessionPlan = func(_ context.Context, _ *captaindb.DB, sessionID string) (string, string, error) {
			Expect(sessionID).To(Equal("plan-session"))
			return "", "<proposed_plan># Recovered plan</proposed_plan>", nil
		}
		content, path, err := (&Provider{}).planResultContent(context.Background(), nil, "plan-session")
		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(BeEmpty())
		Expect(content).To(Equal("# Recovered plan"))
	})
})
