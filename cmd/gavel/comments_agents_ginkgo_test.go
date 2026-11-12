package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/projects"
	"github.com/flanksource/gavel/settings"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("comments agents", func() {
	DescribeTable("--layer names a settings layer position",
		func(layer string, expected settings.Ref) {
			Expect(parseLayerRef(layer)).To(Equal(expected))
		},
		Entry("the user's home", "user", settings.Ref{Kind: settings.KindUser}),
		Entry("the project root", "project", settings.Ref{Kind: settings.KindProject}),
		Entry("an absolute parent directory", "/srv/workspace/", settings.Ref{Kind: settings.KindParent, Path: "/srv/workspace"}),
		Entry("a project folder", "slides/./arb/", settings.Ref{Kind: settings.KindPath, Path: "slides/arb"}),
	)

	DescribeTable("--layer rejects what names no position",
		func(layer string) {
			_, err := parseLayerRef(layer)
			Expect(err).To(MatchError(ContainSubstring("--layer")))
		},
		Entry("empty", ""),
		Entry("the project root as a folder", "."),
		Entry("a folder outside the project", "../elsewhere"),
	)

	It("saves a page entry into a .gavel.<name>.yaml and reads it back traced to that entry", func() {
		base, err := filepath.EvalSymlinks(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		home := filepath.Join(base, "home")
		root := filepath.Join(home, "acme")
		slides := filepath.Join(root, "slides")
		Expect(os.MkdirAll(slides, 0o755)).To(Succeed())
		GinkgoT().Setenv("HOME", home)
		page := filepath.Join(slides, "16-security.mdx")
		Expect(os.WriteFile(page, []byte("# Security\n"), 0o644)).To(Succeed())
		out, err := exec.Command("git", "-C", root, "init", "--quiet").CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(out))
		catalogPath := filepath.Join(base, "projects.json")
		DeferCleanup(projects.UseFileForTest(catalogPath))
		_, err = projects.NewFile(catalogPath).Create(context.Background(), projects.Project{Name: "acme", Dir: root})
		Expect(err).NotTo(HaveOccurred())

		run := func(args ...string) agentSettings {
			GinkgoHelper()
			var stdout bytes.Buffer
			command := commentsAgentsCommand()
			command.SetArgs(args)
			command.SetOut(&stdout)
			Expect(command.ExecuteContext(context.Background())).To(Succeed())
			var shown agentSettings
			Expect(json.Unmarshal(stdout.Bytes(), &shown)).To(Succeed())
			return shown
		}
		run("set", page, "--layer", "slides", "--file", ".gavel.agents.yaml", "--page", "--prompt", "security-review", "--enabled")
		shown := run("get", page)

		saved, err := verify.LoadSingleGavelConfig(filepath.Join(slides, ".gavel.agents.yaml"))
		Expect(err).NotTo(HaveOccurred())
		layer := filepath.Join(slides, ".gavel.agents.yaml")
		source := verify.CommentAgentSource{Layer: layer, Page: "16-security.mdx"}
		Expect([]any{saved.Comments.Agents, shown.Agent}).To(Equal([]any{
			verify.CommentAgentsConfig{Pages: map[string]verify.CommentAgent{"16-security.mdx": {Enabled: ptrTo(true), Prompt: "security-review"}}},
			verify.ResolvedCommentAgent{
				Agent:      verify.CommentAgent{Enabled: ptrTo(true), Prompt: "security-review"},
				Provenance: map[string]verify.CommentAgentSource{"enabled": source, "prompt": source},
				Found:      true,
			},
		}))

		Expect(run("delete", page, "--layer", "slides", "--file", ".gavel.agents.yaml", "--page").Agent.Found).To(BeFalse())
	})
})

func ptrTo[T any](value T) *T { return &value }
