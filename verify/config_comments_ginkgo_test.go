package verify

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func ptr[T any](value T) *T { return &value }

var _ = Describe("comment agents config", func() {
	DescribeTable("accepts only a clean relative file path beneath the directory as a page key",
		func(key string, valid bool) {
			Expect(IsCommentPageKey(key)).To(Equal(valid))
		},
		Entry("a file", "16-security.mdx", true),
		Entry("a nested file", "arb/16-security.mdx", true),
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

	It("sets a page entry and a default, then removes the page entry without leaving an empty map", func() {
		reviewer := CommentAgent{Enabled: ptr(true), Assign: ptr(true), Prompt: "slides-review", Presets: []string{"fast"}, Instructions: "Keep the tables"}
		var configured CommentAgentsConfig

		Expect(configured.Set("16-security.mdx", &reviewer)).To(Succeed())
		Expect(configured.Set("", &CommentAgent{Enabled: ptr(false)})).To(Succeed())
		Expect(configured).To(Equal(CommentAgentsConfig{Default: &CommentAgent{Enabled: ptr(false)}, Pages: map[string]CommentAgent{"16-security.mdx": reviewer}}))

		Expect(configured.Set("16-security.mdx", nil)).To(Succeed())
		Expect(configured).To(Equal(CommentAgentsConfig{Default: &CommentAgent{Enabled: ptr(false)}}))
	})

	It("refuses a page key that escapes the layer's directory", func() {
		var configured CommentAgentsConfig
		Expect(configured.Set("../x.md", &CommentAgent{})).To(MatchError(ContainSubstring(`"../x.md" must be a relative file path`)))
	})

	Describe("patching an entry with a JSON merge patch", func() {
		const page = "16-security.mdx"
		entry := CommentAgent{Enabled: ptr(true), Prompt: "slides-review", Instructions: "Keep the tables", Variables: map[string]any{"tone": "terse", "length": "short"}}
		patched := func(patch string) (CommentAgentsConfig, error) {
			configured := CommentAgentsConfig{Default: &CommentAgent{Enabled: ptr(false)}, Pages: map[string]CommentAgent{page: entry}}
			edit, err := PatchCommentAgent([]byte(patch))
			if err != nil {
				return configured, err
			}
			return configured, configured.Edit(page, edit)
		}
		It("replaces present fields, removes null ones so they inherit, and merges nested objects", func() {
			configured, err := patched(`{"edits":"propose","instructions":null,"variables":{"length":null,"audience":"ops"}}`)
			Expect(err).NotTo(HaveOccurred())
			Expect(configured.Pages[page]).To(Equal(CommentAgent{Enabled: ptr(true), Prompt: "slides-review", Edits: CommentEditsPropose, Variables: map[string]any{"tone": "terse", "audience": "ops"}}))
			Expect(configured.Default).To(Equal(&CommentAgent{Enabled: ptr(false)}), "only the named entry changes")
		})
		It("deletes the entry for a null patch and creates one a patch names for the first time", func() {
			removed, err := patched(`null`)
			Expect(err).NotTo(HaveOccurred())
			Expect(removed.Pages).To(BeNil())
			var fresh CommentAgentsConfig
			edit, err := PatchCommentAgent([]byte(`{"assign":true}`))
			Expect(err).NotTo(HaveOccurred())
			Expect(fresh.Edit("", edit)).To(Succeed())
			Expect(fresh.Default).To(Equal(&CommentAgent{Assign: ptr(true)}))
		})
		DescribeTable("rejects a patch before editing anything",
			func(patch string) {
				_, err := PatchCommentAgent([]byte(patch))
				Expect(err).To(HaveOccurred())
			},
			Entry("an unknown field", `{"model":"opus"}`),
			Entry("a field of the wrong type", `{"enabled":"yes"}`),
			Entry("an edits policy outside the enum", `{"edits":"rewrite"}`),
			Entry("a patch that is not an object", `["enabled"]`),
		)
	})

	Describe("resolving a file's agent across layers", func() {
		project := CommentAgentLayer{Label: "/repo/.gavel.yaml", Dir: ".", Agents: CommentAgentsConfig{Default: &CommentAgent{Enabled: ptr(true), Prompt: "editor-comments", Presets: []string{"fast"}}}}
		slides := CommentAgentLayer{Label: "db:path:slides", Dir: "slides", Agents: CommentAgentsConfig{
			Default: &CommentAgent{Assign: ptr(true)},
			Pages:   map[string]CommentAgent{"16-security.mdx": {Prompt: "slides-review"}, "draft.mdx": {Enabled: ptr(false)}},
		}}
		layers := []CommentAgentLayer{project, slides}

		DescribeTable("merges ancestors first, then the directory default, then the page entry, field by field",
			func(file string, want CommentAgent, found bool) {
				resolved, err := ResolveCommentAgent(layers, file)
				Expect(err).NotTo(HaveOccurred())
				Expect([]any{resolved.Agent, resolved.Found}).To(Equal([]any{want, found}))
			},
			Entry("a page entry overriding only the prompt", "slides/16-security.mdx",
				CommentAgent{Enabled: ptr(true), Assign: ptr(true), Prompt: "slides-review", Presets: []string{"fast"}}, true),
			Entry("a page opting out", "slides/draft.mdx",
				CommentAgent{Enabled: ptr(false), Assign: ptr(true), Prompt: "editor-comments", Presets: []string{"fast"}}, true),
			Entry("a folder page with no entry", "slides/other.mdx",
				CommentAgent{Enabled: ptr(true), Assign: ptr(true), Prompt: "editor-comments", Presets: []string{"fast"}}, true),
			Entry("a file outside the folder", "README.md",
				CommentAgent{Enabled: ptr(true), Prompt: "editor-comments", Presets: []string{"fast"}}, true),
			Entry("a sibling sharing the folder's leading characters", "slides-old/x.mdx",
				CommentAgent{Enabled: ptr(true), Prompt: "editor-comments", Presets: []string{"fast"}}, true),
		)

		It("traces each field to the layer default or page entry that last set it", func() {
			resolved, err := ResolveCommentAgent(layers, "slides/16-security.mdx")
			Expect(err).NotTo(HaveOccurred())
			Expect(resolved.Provenance).To(Equal(map[string]CommentAgentSource{
				"enabled": {Layer: "/repo/.gavel.yaml"},
				"presets": {Layer: "/repo/.gavel.yaml"},
				"assign":  {Layer: "db:path:slides"},
				"prompt":  {Layer: "db:path:slides", Page: "16-security.mdx"},
			}))
		})

		It("finds nothing when no layer configures the file", func() {
			resolved, err := ResolveCommentAgent([]CommentAgentLayer{slides}, "README.md")
			Expect(err).NotTo(HaveOccurred())
			Expect([]any{resolved.Found, resolved.Provenance}).To(Equal([]any{false, map[string]CommentAgentSource{}}))
		})

		It("treats an unset enabled as off and an unset prompt as the editor prompt", func() {
			Expect([]any{CommentAgent{}.IsEnabled(), CommentAgent{}.Assigns(), CommentAgent{}.PromptOrDefault(), CommentAgent{Prompt: "slides-review"}.PromptOrDefault()}).
				To(Equal([]any{false, false, DefaultCommentAgentPrompt, "slides-review"}))
		})
	})
})
