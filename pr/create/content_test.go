package prcreate

import (
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/aiflags"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	commitpkg "github.com/flanksource/gavel/commit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ContentInput", func() {
	It("describes the given commits in order with their messages and files", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		dir := GinkgoT().TempDir()
		run := func(args ...string) string {
			GinkgoHelper()
			out, err := captureGit(dir, args...)
			Expect(err).NotTo(HaveOccurred())
			return out
		}
		commit := func(name, message string) string {
			GinkgoHelper()
			Expect(os.WriteFile(filepath.Join(dir, name), []byte(name+"\n"), 0o600)).To(Succeed())
			run("add", name)
			run("-c", "user.email=test@example.com", "-c", "user.name=test", "commit", "-q", "-m", message)
			return run("rev-parse", "HEAD")
		}
		run("init", "-q", "-b", "main")
		first := commit("a.txt", "feat: add a")
		second := commit("b.txt", "feat: add b\n\nWith a body.")
		saved := &captainconfig.Config{}

		got, err := ContentInput(dir, []string{first, second}, aiflags.ModelFlags{}, saved)

		Expect(err).NotTo(HaveOccurred())
		Expect(got.Commits).To(Equal([]commitpkg.PRCommitInput{
			{Message: "feat: add a", Files: []string{"a.txt"}},
			{Message: "feat: add b\n\nWith a body.", Files: []string{"b.txt"}},
		}))
		Expect(got.Options.WorkDir).To(Equal(dir))
		Expect(got.Options.Saved).To(BeIdenticalTo(saved))
	})
})

var _ = Describe("PR content configuration snapshot", func() {
	It("carries the worktree full specification and captured saved settings to generation", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("ai:\n  model: api:sonnet\n  memory: {skipHooks: true}\npr:\n  content:\n    file: review.prompt\n    budget: {cost: 2, maxTurns: 3}\n    setup: {envVars: [{name: REVIEW_SCOPE, value: pull-request}]}\n"), 0o600)).To(Succeed())
		saved := &captainconfig.Config{AI: captainconfig.AIDefaults{MaxTokens: 1700}}
		flags := aiflags.ModelFlags{Model: "api:gpt-5", Temperature: "0"}
		got, err := loadContentOptions(dir, Input{Flags: flags, Saved: saved})
		Expect(err).NotTo(HaveOccurred())
		Expect(got.WorkDir).To(Equal(dir))
		Expect(got.Saved).To(BeIdenticalTo(saved))
		Expect(got.Flags).To(Equal(flags))
		Expect(got.AI.Memory.SkipHooks).To(BeTrue())
		Expect(got.PR.Content.File).To(Equal("review.prompt"))
		Expect(got.PR.Content.Spec.Budget).To(Equal(api.Budget{Cost: 2, MaxTurns: 3}))
		Expect(got.PR.Content.Spec.Setup.EnvVars).To(HaveLen(1))
		Expect(got.PR.Content.Spec.Setup.EnvVars[0].Name).To(Equal("REVIEW_SCOPE"))
		Expect(got.PR.Content.Spec.Setup.EnvVars[0].ValueStatic).To(Equal("pull-request"))
	})
})
