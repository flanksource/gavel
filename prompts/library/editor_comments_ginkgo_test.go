package library_test

import (
	"io/fs"
	"strings"

	promptlib "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts/library"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("editor-comments.prompt", func() {
	It("runs in the project and declares no tools of its own, since the editor attaches the thread tools to every session", func() {
		tpl, err := promptlib.LoadFS(library.FS, "editor-comments.prompt")
		Expect(err).NotTo(HaveOccurred())
		spec, _, err := tpl.Render(promptlib.RenderOptions{Declared: true, Data: map[string]any{
			"gavelExecutable": "/bin/gavel",
			"dir":             "/src/acme's repo",
			"actor":           "eyJraW5kIjoiYWdlbnQifQ",
			"scope":           "gavel://project/acme",
			"target":          "gavel://project/acme/files/docs/a.md",
			"message":         "Rename {{x}} to y",
		}})
		Expect(err).NotTo(HaveOccurred())
		Expect(spec.Setup.Cwd).To(Equal("/src/acme's repo"))
		Expect(spec.Model.Mode).To(Equal(api.ModeAgent))
		Expect(spec.Prompt.User).To(Equal("Rename {{x}} to y"))
		Expect(spec.Tools).To(BeEmpty())
	})

	DescribeTable("tells a session started before its first thread that the thread tools come with it",
		func(waiting bool, matcher OmegaMatcher) {
			tpl, err := promptlib.LoadFS(library.FS, "editor-comments.prompt")
			Expect(err).NotTo(HaveOccurred())
			spec, _, err := tpl.Render(promptlib.RenderOptions{Declared: true, Data: map[string]any{
				"dir": "/src/acme", "message": "ready?", "waiting": waiting,
			}})
			Expect(err).NotTo(HaveOccurred())
			Expect(spec.Prompt.System).To(matcher)
		},
		Entry("waiting for threads", true, ContainSubstring("No thread has arrived yet")),
		Entry("started with threads", false, Not(ContainSubstring("No thread has arrived yet"))),
	)

	It("is discoverable with valid unrendered frontmatter", func() {
		raw, err := fs.ReadFile(library.FS, "editor-comments.prompt")
		Expect(err).NotTo(HaveOccurred())
		var frontmatter map[string]any
		Expect(yaml.Unmarshal([]byte(strings.SplitN(string(raw), "---\n", 3)[1]), &frontmatter)).To(Succeed())
		Expect(frontmatter).To(HaveKeyWithValue("name", "editor-comments"))
	})
})
