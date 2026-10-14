package todos

import (
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("linked PR definition of done", func() {
	It("requires the whole current PR without rewriting the stored fixture", func() {
		stored := "### command: Local test\n\n```bash\necho checked\n```"
		todo := &types.TODO{VerificationMarkdown: stored}
		todo.PR = &types.PR{Number: 27, URL: "https://github.com/acme/service/pull/27"}
		dod, err := BuildDefinitionOfDone(DefinitionOfDoneOptions{WorkDir: GinkgoT().TempDir(), Todos: []*types.TODO{todo}})
		Expect(err).NotTo(HaveOccurred())
		Expect(dod.Declared()).To(BeTrue())
		Expect(dod.Fixture).To(ContainSubstring("--whole-pr --head-sha HEAD --follow --fail-fast --logs"))
		Expect(dod.Fixture).To(ContainSubstring("echo checked"))
		Expect(todo.VerificationMarkdown).To(Equal(stored))
		updated, err := PRVerificationFixture(todo.PR, dod.Fixture)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Count(updated, "Entire PR at the verified revision")).To(Equal(1))
	})
	It("preserves fixture front matter when adding the mandatory gate again", func() {
		pr := &types.PR{Number: 27, URL: "https://github.com/acme/service/pull/27"}
		stored := "---\nenv:\n  FIXTURE_MODE: offline\nskip: 'true'\n---\n\n| Name | Command | Exit Code |\n|---|---|---|\n| Local | true | 0 |"
		fixture, err := PRVerificationFixture(pr, stored)
		Expect(err).NotTo(HaveOccurred())
		front, _, err := fixtures.SplitFrontMatter(fixture)
		Expect(err).NotTo(HaveOccurred())
		Expect(front).NotTo(BeNil())
		Expect(fixture).To(ContainSubstring("FIXTURE_MODE: offline"))
		root, err := fixtures.ParseMarkdownDocument("linked-pr", fixture, GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		var gate *fixtures.FixtureTest
		root.Walk(func(node *fixtures.FixtureNode) {
			if node.Test.Name == "Entire PR at the verified revision" {
				gate = node.Test
			}
		})
		Expect(gate).NotTo(BeNil())
		reason, err := gate.ShouldSkip(fixtures.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(reason).To(BeEmpty())
	})
})
