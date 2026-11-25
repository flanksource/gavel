package verify

import (
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The tree is root/.gavel.yaml → root/org/.gavel.yaml → root/org/repo (git
// root), with $HOME elsewhere: org is the git root's parent and root sits two
// levels up, outside the chain.
var _ = Describe("git-root-parent config layer", func() {
	const parentRule, twoUpRule = "acme-parent-rule", "acme-two-up-rule"
	var root, org, repo string

	writeConfig := func(dir, content string) {
		Expect(os.WriteFile(filepath.Join(dir, GavelConfigFileName), []byte(content), 0o600)).To(Succeed())
	}

	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		root = GinkgoT().TempDir()
		org = filepath.Join(root, "org")
		repo = filepath.Join(org, "repo")
		Expect(os.MkdirAll(filepath.Join(repo, ".git"), 0o755)).To(Succeed())
		writeConfig(root, "lint: {ignore: [{rule: "+twoUpRule+"}]}\n")
		writeConfig(repo, "procfile: {profile: repo}\n")
	})

	It("merges the parent beneath the git root and keeps its lint.ignore, ignoring two levels up", func() {
		writeConfig(org, "procfile: {profile: parent, mem: 1g}\nlint: {ignore: [{rule: "+parentRule+", source: betterleaks}]}\n")

		cfg, err := LoadGavelConfig(repo)

		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Procfile).To(Equal(ProcfileConfig{Profile: "repo", Mem: "1g"}))
		Expect(cfg.Lint.Ignore).To(Equal([]LintIgnoreRule{{Rule: parentRule, Source: "betterleaks"}}))
	})

	It("lists the parent as a git-root-parent source in the trace", func() {
		writeConfig(org, "procfile: {mem: 1g}\n")

		trace, err := LoadGavelConfigTrace(repo)

		Expect(err).NotTo(HaveOccurred())
		Expect(trace.Sources).To(HaveLen(2))
		Expect([]string{trace.Sources[0].Origin, trace.Sources[0].Path}).To(Equal([]string{"git-root-parent", filepath.Join(org, GavelConfigFileName)}))
		Expect([]string{trace.Sources[1].Origin, trace.Sources[1].Path}).To(Equal([]string{"git-root", filepath.Join(repo, GavelConfigFileName)}))
	})

	It("does not read $HOME twice when the repo sits directly under it", func() {
		GinkgoT().Setenv("HOME", org)
		writeConfig(org, "lint: {ignore: [{rule: "+parentRule+"}]}\n")

		Expect(GavelConfigDirs(repo)).To(Equal([]GavelConfigDir{
			{Origin: "user-home", Dir: org},
			{Origin: "git-root", Dir: repo},
		}))
		cfg, err := LoadGavelConfig(repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Lint.Ignore).To(Equal([]LintIgnoreRule{{Rule: parentRule}}))
	})

	It("has no parent layer outside a git repository", func() {
		Expect(os.Remove(filepath.Join(repo, ".git"))).To(Succeed())
		home, err := os.UserHomeDir()
		Expect(err).NotTo(HaveOccurred())

		Expect(GavelConfigDirs(repo)).To(Equal([]GavelConfigDir{
			{Origin: "user-home", Dir: home},
			{Origin: "target-directory", Dir: repo},
		}))
	})

	It("invalidates the cached config when the parent file is created", func() {
		first, err := LoadGavelConfig(repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(first.Procfile.Mem).To(BeEmpty())

		writeConfig(org, "procfile: {mem: 2g}\n")

		second, err := LoadGavelConfig(repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(second.Procfile).To(Equal(ProcfileConfig{Profile: "repo", Mem: "2g"}))
	})
})

var _ = Describe("raw configuration spec layers", func() {
	DescribeTable("rejects invalid home configuration before repository overrides",
		func(homeConfig, repositoryConfig, field string) {
			home, repository := GinkgoT().TempDir(), GinkgoT().TempDir()
			GinkgoT().Setenv("HOME", home)
			path := filepath.Join(home, ".gavel.yaml")
			Expect(os.WriteFile(path, []byte(homeConfig), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(repository, ".gavel.yaml"), []byte(repositoryConfig), 0600)).To(Succeed())
			_, err := LoadGavelConfig(repository)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(path))
			Expect(err.Error()).To(ContainSubstring(field))
			_, err = LoadGavelConfigTrace(repository)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(path))
			Expect(err.Error()).To(ContainSubstring(field))
		},
		Entry("model temperature", "ai: {temperature: 3}\n", "ai: {temperature: 1}\n", "temperature"),
		Entry("step budget", "todos: {run: {budget: {maxTurns: -1}}}\n", "todos: {run: {budget: {maxTurns: 5}}}\n", "maxTurns"),
		Entry("custom step budget", "todos: {steps: {handoff: {budget: {cost: -1}}}}\n", "todos: {steps: {handoff: {budget: {cost: 5}}}}\n", "cost"),
		Entry("timeout constraint", "todos: {timeout: invalid}\n", "todos: {timeout: 5m}\n", "timeout"),
	)

	It("keeps structurally valid partial runtimes available for later composition", func() {
		home, repository := GinkgoT().TempDir(), GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		Expect(os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("ai: {mode: agent, permissions: {tools: {Read: allow}}}\n"), 0600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repository, ".gavel.yaml"), []byte("ai: {model: gpt-5.6-sol}\n"), 0600)).To(Succeed())
		config, err := LoadGavelConfig(repository)
		Expect(err).NotTo(HaveOccurred())
		Expect(config.AI.Name).To(Equal("gpt-5.6-sol"))
		Expect(string(config.AI.Permissions.Tools["Read"])).To(Equal("allow"))
	})

	// A step block is a later layer than the `ai:` base, so it supplies the
	// value: the base only decides what a step does not name.
	It("lets a step block replace what the ai: base configured", func() {
		home, repository := GinkgoT().TempDir(), GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		config := "ai: {permissions: {tools: {Bash: deny}}}\ntodos: {run: {permissions: {tools: {Bash: allow}}}}\n"
		Expect(os.WriteFile(filepath.Join(repository, ".gavel.yaml"), []byte(config), 0600)).To(Succeed())

		loaded, err := LoadGavelConfig(repository)

		Expect(err).NotTo(HaveOccurred())
		Expect(loaded.AI.Permissions.Tools).To(Equal(api.Tools{"Bash": api.ToolPolicyDeny}))
		Expect(loaded.Todos.Run.Spec.Permissions.Tools).To(Equal(api.Tools{"Bash": api.ToolPolicyAllow}))
	})
})
