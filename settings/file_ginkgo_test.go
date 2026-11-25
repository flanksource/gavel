package settings_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/settings"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// projectTree is a project at <tmp>/acme with a slides/arb sub-folder, its own
// HOME so the user's real ~/.gavel.yaml never joins the chain.
func projectTree() (home, root, page string) {
	GinkgoHelper()
	base := GinkgoT().TempDir()
	base, err := filepath.EvalSymlinks(base)
	Expect(err).NotTo(HaveOccurred())
	home = filepath.Join(base, "home")
	root = filepath.Join(home, "acme")
	Expect(os.MkdirAll(filepath.Join(root, "slides", "arb"), 0o755)).To(Succeed())
	GinkgoT().Setenv("HOME", home)
	page = filepath.Join(root, "slides", "arb", "16-security.mdx")
	Expect(os.WriteFile(page, []byte("# Security\n"), 0o644)).To(Succeed())
	// A git checkout, so files synced in DB mode record a HEAD.
	git(root, "init", "--quiet")
	git(root, "-c", "user.email=dev@example.com", "-c", "user.name=dev", "commit", "--quiet", "--allow-empty", "-m", "init")
	return home, root, page
}

func git(dir string, args ...string) string {
	GinkgoHelper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), string(out))
	return strings.TrimSpace(string(out))
}

func writeConfig(dir, name, body string) {
	GinkgoHelper()
	Expect(os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644)).To(Succeed())
}

var _ = Describe("file-only settings", func() {
	ctx := context.Background()

	It("layers home, project root, sub-folder and its snippets in order, tracing each field to its file", func() {
		home, root, page := projectTree()
		slides := filepath.Join(root, "slides")
		writeConfig(home, verify.GavelConfigFileName, "pr:\n  base: origin/home\n  draft: true\n")
		writeConfig(root, verify.GavelConfigFileName, "pr:\n  base: origin/main\n")
		writeConfig(slides, ".gavel.agents.yaml", "comments:\n  agents:\n    default:\n      enabled: true\n      prompt: slides-review\n")
		target := settings.Target{Path: page, Project: &settings.Project{ID: uuid.New(), Dir: root}}

		resolved, err := settings.File{}.Resolve(ctx, target)
		Expect(err).NotTo(HaveOccurred())
		refs := []settings.Ref{}
		for _, layer := range resolved.Layers {
			refs = append(refs, layer.Ref)
		}
		Expect([]any{
			refs,
			resolved.Merged.PR.Base, resolved.Merged.PR.Draft, resolved.Merged.Comments.Agents.Default.Prompt,
			resolved.Provenance["/pr/base"], resolved.Provenance["/pr/draft"], resolved.Provenance["/comments/agents/default/prompt"],
		}).To(Equal([]any{
			[]settings.Ref{{Kind: settings.KindUser, File: ".gavel.yaml"}, {Kind: settings.KindProject, File: ".gavel.yaml"}, {Kind: settings.KindPath, Path: "slides", File: ".gavel.agents.yaml"}},
			"origin/main", true, "slides-review",
			filepath.Join(root, ".gavel.yaml"), filepath.Join(home, ".gavel.yaml"), filepath.Join(slides, ".gavel.agents.yaml"),
		}))
	})

	It("lists every position of the chain, with the files at each", func() {
		_, root, page := projectTree()
		writeConfig(filepath.Join(root, "slides"), ".gavel.zeta.yaml", "{}\n")
		writeConfig(filepath.Join(root, "slides"), verify.GavelConfigFileName, "{}\n")

		resolved, err := settings.File{}.Resolve(ctx, settings.Target{Path: page, Project: &settings.Project{ID: uuid.New(), Dir: root}})
		Expect(err).NotTo(HaveOccurred())
		positions := map[string][]string{}
		for _, position := range resolved.Positions {
			positions[string(position.Kind)+":"+position.Path] = position.Files
		}
		Expect(positions).To(Equal(map[string][]string{
			"user:":           {},
			"project:":        {},
			"path:slides":     {".gavel.yaml", ".gavel.zeta.yaml"},
			"path:slides/arb": {},
		}))
	})

	It("reaches every folder of a project that is not a git checkout", func() {
		base, err := filepath.EvalSymlinks(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		GinkgoT().Setenv("HOME", filepath.Join(base, "home"))
		root := filepath.Join(base, "plain")
		Expect(os.MkdirAll(filepath.Join(root, "slides", "arb"), 0o755)).To(Succeed())
		page := filepath.Join(root, "slides", "arb", "16-security.mdx")
		Expect(os.WriteFile(page, []byte("# Security\n"), 0o644)).To(Succeed())
		writeConfig(filepath.Join(root, "slides"), verify.GavelConfigFileName, "pr:\n  base: origin/slides\n")

		resolved, err := settings.File{}.Resolve(ctx, settings.Target{Path: page, Project: &settings.Project{ID: uuid.New(), Dir: root}})
		Expect(err).NotTo(HaveOccurred())
		kinds := []string{}
		for _, position := range resolved.Positions {
			kinds = append(kinds, string(position.Kind)+":"+position.Path)
		}
		Expect([]any{kinds, resolved.Merged.PR.Base}).To(Equal([]any{[]string{"user:", "project:", "path:slides", "path:slides/arb"}, "origin/slides"}))
	})

	It("saves to a new snippet at a chosen position and refuses a position outside the chain or a database layer", func() {
		_, root, page := projectTree()
		target := settings.Target{Path: page, Project: &settings.Project{ID: uuid.New(), Dir: root}}
		draft := func(cfg *verify.GavelConfig) error { cfg.PR.Draft = true; return nil }

		Expect(settings.File{}.Save(ctx, target, settings.Ref{Kind: settings.KindPath, Path: "slides/arb", File: ".gavel.pr.yaml"}, draft)).To(Succeed())
		saved, err := verify.LoadSingleGavelConfig(filepath.Join(root, "slides", "arb", ".gavel.pr.yaml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(saved.PR.Draft).To(BeTrue())

		Expect(settings.File{}.Save(ctx, target, settings.Ref{Kind: settings.KindPath, Path: "other", File: ".gavel.yaml"}, draft)).
			To(MatchError(ContainSubstring("not in the target's configuration chain")))
		Expect(settings.File{}.Save(ctx, target, settings.Ref{Kind: settings.KindProject}, draft)).
			To(MatchError(ContainSubstring("no database layer")))
		Expect(settings.File{}.Save(ctx, target, settings.Ref{Kind: settings.KindProject, File: "notes.yaml"}, draft)).
			To(MatchError(ContainSubstring("is not a .gavel.yaml")))
	})
})
