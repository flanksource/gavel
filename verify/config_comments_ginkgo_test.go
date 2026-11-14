package verify

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("comment agents config", func() {
	DescribeTable("accepts only a clean relative file path beneath the directory as a page key",
		func(key string, valid bool) {
			Expect(IsCommentPageKey(key)).To(Equal(valid))
		},
		Entry("a file", "19-security.mdx", true),
		Entry("a nested file", "arb/19-security.mdx", true),
		Entry("the directory itself", ".", false),
		Entry("empty", "", false),
		Entry("a parent", "../other.md", false),
		Entry("an absolute path", "/etc/passwd", false),
		Entry("an unclean path", "arb/../x.md", false),
		Entry("a windows separator", `arb\x.md`, false),
	)

	It("rejects a .gavel.yaml whose page key escapes its directory", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, GavelConfigFileName), []byte("comments:\n  agents:\n    pages:\n      ../x.md:\n        enabled: true\n"), 0o644)).To(Succeed())

		_, err := LoadSingleGavelConfig(filepath.Join(dir, GavelConfigFileName))
		Expect(err).To(MatchError(ContainSubstring(`comments.agents.pages: "../x.md"`)))
	})

	It("saves a page agent and a directory opt-out beside the file's other settings, then removes the page", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, GavelConfigFileName)
		Expect(os.WriteFile(path, []byte("commit:\n  types:\n  - feat\n"), 0o644)).To(Succeed())
		reviewer := CommentAgent{Enabled: true, Assign: true, Prompt: "slides-review", Presets: []string{"fast"}, Instructions: "Keep the tables"}

		Expect(SaveCommentAgent(dir, "19-security.mdx", &reviewer)).To(Succeed())
		Expect(SaveCommentAgent(dir, "", &CommentAgent{Enabled: false})).To(Succeed())
		saved, err := LoadSingleGavelConfig(path)
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{saved.Commit.Types, saved.Comments.Agents}).To(Equal([]any{
			[]string{"feat"},
			CommentAgentsConfig{Default: &CommentAgent{Enabled: false}, Pages: map[string]CommentAgent{"19-security.mdx": reviewer}},
		}))

		Expect(SaveCommentAgent(dir, "19-security.mdx", nil)).To(Succeed())
		saved, err = LoadSingleGavelConfig(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(saved.Comments.Agents).To(Equal(CommentAgentsConfig{Default: &CommentAgent{Enabled: false}}))
	})

	It("defaults an agent with no prompt to the editor's comment prompt", func() {
		Expect([]string{CommentAgent{}.PromptOrDefault(), CommentAgent{Prompt: "slides-review"}.PromptOrDefault()}).
			To(Equal([]string{DefaultCommentAgentPrompt, "slides-review"}))
	})
})
