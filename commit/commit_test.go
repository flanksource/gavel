package commit

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api "github.com/flanksource/captain/pkg/api"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/verify"
)

// msgSpecCfg builds a CommitConfig whose commit.message spec pins model.
func msgSpecCfg(model string) verify.CommitConfig {
	return verify.CommitConfig{Message: verify.PromptSpec{Spec: api.Spec{Model: api.Model{Name: model}}}}
}

// msgGroupSpecCfg builds a CommitConfig pinning both the commit.message and
// commit.grouping spec models.
func msgGroupSpecCfg(message, group string) verify.CommitConfig {
	cfg := msgSpecCfg(message)
	cfg.Grouping = verify.PromptSpec{Spec: api.Spec{Model: api.Model{Name: group}}}
	return cfg
}

// aiBaseSpec builds a base ai: spec pinning model.
func aiBaseSpec(model string) api.Spec {
	return api.Spec{Model: api.Model{Name: model}}
}

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

	It("overrides the Message spec model when set, preserves otherwise", func() {
		base := msgSpecCfg("claude-haiku-4-5")
		Expect(verify.MergeCommitConfig(base, verify.CommitConfig{}).Message.Model.Name).To(Equal("claude-haiku-4-5"))
		Expect(verify.MergeCommitConfig(base, msgSpecCfg("gpt-4o")).Message.Model.Name).To(Equal("gpt-4o"))
	})

	It("overrides the Grouping spec model when set, preserves otherwise", func() {
		base := verify.CommitConfig{Grouping: verify.PromptSpec{Spec: api.Spec{Model: api.Model{Name: "claude-sonnet-4-5"}}}}
		Expect(verify.MergeCommitConfig(base, verify.CommitConfig{}).Grouping.Model.Name).To(Equal("claude-sonnet-4-5"))
		override := verify.CommitConfig{Grouping: verify.PromptSpec{Spec: api.Spec{Model: api.Model{Name: "claude-opus-4-1"}}}}
		Expect(verify.MergeCommitConfig(base, override).Grouping.Model.Name).To(Equal("claude-opus-4-1"))
	})

	It("overrides MaxCommits when non-zero, preserves otherwise", func() {
		base := verify.CommitConfig{MaxCommits: 7}
		Expect(verify.MergeCommitConfig(base, verify.CommitConfig{}).MaxCommits).To(Equal(7))
		Expect(verify.MergeCommitConfig(base, verify.CommitConfig{MaxCommits: 3}).MaxCommits).To(Equal(3))
	})
})

var _ = Describe("Options model resolution", func() {
	DescribeTable("messageModel prefers --model, then commit.message.model, then ai.model, else empty",
		func(opts Options, expected string) {
			Expect(opts.messageModel()).To(Equal(expected))
		},
		Entry("flag wins", Options{Model: "haiku-flag", Config: msgSpecCfg("haiku-cfg"), AI: aiBaseSpec("base")}, "haiku-flag"),
		Entry("message spec used when no flag", Options{Config: msgSpecCfg("haiku-cfg"), AI: aiBaseSpec("base")}, "haiku-cfg"),
		Entry("ai base used when no flag or op spec", Options{AI: aiBaseSpec("base")}, "base"),
		Entry("empty when nothing set", Options{}, ""),
	)

	DescribeTable("groupModel cascades group flag -> commit.grouping.model -> ai.model -> message model -> default",
		func(opts Options, expected string) {
			Expect(opts.groupModel()).To(Equal(expected))
		},
		Entry("--group-model wins over everything",
			Options{GroupModel: "opus-flag", Config: msgGroupSpecCfg("haiku-cfg", "sonnet-cfg"), AI: aiBaseSpec("base")}, "opus-flag"),
		Entry("commit.grouping.model used when no group flag",
			Options{Config: msgGroupSpecCfg("haiku-cfg", "sonnet-cfg")}, "sonnet-cfg"),
		Entry("ai base used when no group flag or grouping spec",
			Options{Config: msgSpecCfg("haiku-cfg"), AI: aiBaseSpec("base")}, "base"),
		Entry("falls back to --model when no group/ai model",
			Options{Model: "haiku-flag"}, "haiku-flag"),
		Entry("falls back to commit.message.model when no group/ai model",
			Options{Config: msgSpecCfg("haiku-cfg")}, "haiku-cfg"),
		Entry("defaults to sonnet-class when nothing set",
			Options{}, defaultGroupModel),
	)
})

var _ = Describe("BuildAgent errors", func() {
	It("preserves provider-specific credential guidance without appending every provider key", func() {
		previousAgent := newAgentFunc
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) {
			return nil, errors.New("API key not found for backend openai; similar environment variable found: OPEN_AI_API_KEY (did you mean OPENAI_API_KEY?)")
		}
		DeferCleanup(func() { newAgentFunc = previousAgent })

		_, err := BuildAgent(Options{}, "api:terra")
		Expect(err).To(MatchError("LLM agent unavailable: API key not found for backend openai; similar environment variable found: OPEN_AI_API_KEY (did you mean OPENAI_API_KEY?)"))
		Expect(err.Error()).ToNot(ContainSubstring("ANTHROPIC_API_KEY"))
		Expect(errors.Is(err, ErrLLMUnavailable)).To(BeTrue())
	})
})
