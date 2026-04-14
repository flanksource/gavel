package commit

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/verify"
)

func TestCommit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commit Suite")
}

var _ = Describe("RunHooks", func() {
	var workDir string

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
	})

	It("passes when all hooks succeed", func() {
		hooks := []verify.CommitHook{
			{Name: "one", Run: "true"},
			{Name: "two", Run: "exit 0"},
		}
		results, err := RunHooks(workDir, hooks, []string{"main.go"})
		Expect(err).ToNot(HaveOccurred())
		Expect(results).To(HaveLen(2))
		Expect(results[0]).To(Equal(HookResult{Name: "one"}))
		Expect(results[1]).To(Equal(HookResult{Name: "two"}))
	})

	It("returns ErrHookFailed on first failing hook and short-circuits", func() {
		hooks := []verify.CommitHook{
			{Name: "passing", Run: "true"},
			{Name: "failing", Run: "exit 7"},
			{Name: "never-runs", Run: "exit 99"},
		}
		results, err := RunHooks(workDir, hooks, []string{"main.go"})
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrHookFailed)).To(BeTrue(), "expected wrapped ErrHookFailed, got %v", err)
		Expect(results).To(HaveLen(2), "third hook must not run after second fails")
		Expect(results[0].Name).To(Equal("passing"))
		Expect(results[1].Name).To(Equal("failing"))
		Expect(results[1].ExitCode).To(Equal(7))
	})

	It("skips file-filtered hooks when no staged files match", func() {
		hooks := []verify.CommitHook{
			{Name: "py-only", Run: "exit 99", Files: []string{"**/*.py"}},
		}
		results, err := RunHooks(workDir, hooks, []string{"main.go", "README.md"})
		Expect(err).ToNot(HaveOccurred(), "hook should skip entirely, never run 'exit 99'")
		Expect(results).To(HaveLen(1))
		Expect(results[0].Skipped).To(BeTrue())
		Expect(results[0].Name).To(Equal("py-only"))
	})

	It("runs file-filtered hooks when at least one staged file matches", func() {
		hooks := []verify.CommitHook{
			{Name: "go-only", Run: "true", Files: []string{"**/*.go"}},
		}
		results, err := RunHooks(workDir, hooks, []string{"README.md", "pkg/foo.go"})
		Expect(err).ToNot(HaveOccurred())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Skipped).To(BeFalse())
	})

	It("treats an empty hook list as a no-op", func() {
		results, err := RunHooks(workDir, nil, []string{"main.go"})
		Expect(err).ToNot(HaveOccurred())
		Expect(results).To(BeEmpty())
	})
})

var _ = Describe("verify.MergeCommitConfig", func() {
	It("appends hooks from override in order", func() {
		base := verify.CommitConfig{
			Hooks: []verify.CommitHook{{Name: "home-a", Run: "true"}},
		}
		override := verify.CommitConfig{
			Hooks: []verify.CommitHook{
				{Name: "repo-a", Run: "true"},
				{Name: "repo-b", Run: "true"},
			},
		}
		merged := verify.MergeCommitConfig(base, override)
		Expect(merged.Hooks).To(HaveLen(3))
		Expect(merged.Hooks[0].Name).To(Equal("home-a"))
		Expect(merged.Hooks[1].Name).To(Equal("repo-a"))
		Expect(merged.Hooks[2].Name).To(Equal("repo-b"))
	})

	It("overrides Model when non-empty, preserves otherwise", func() {
		base := verify.CommitConfig{Model: "claude-haiku-4.5"}
		Expect(verify.MergeCommitConfig(base, verify.CommitConfig{}).Model).To(Equal("claude-haiku-4.5"))
		Expect(verify.MergeCommitConfig(base, verify.CommitConfig{Model: "gpt-4o"}).Model).To(Equal("gpt-4o"))
	})
})
