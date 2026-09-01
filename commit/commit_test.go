package commit

import (
	"errors"
	"os"
	"testing"

	"github.com/flanksource/captain/pkg/aiflags"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api "github.com/flanksource/captain/pkg/api"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/verify"
)

func TestMain(m *testing.M) {
	_ = os.Unsetenv(EnvSessionID)
	for _, marker := range sessionEnvironmentMarkers {
		_ = os.Unsetenv(marker)
	}
	os.Exit(m.Run())
}

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

var _ = Describe("verify.Merge of CommitConfig", func() {
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
		merged := verify.Merge(base, override)
		Expect(merged.Hooks).To(HaveLen(3))
		Expect(merged.Hooks[0].Name).To(Equal("home-a"))
		Expect(merged.Hooks[1].Name).To(Equal("repo-a"))
		Expect(merged.Hooks[2].Name).To(Equal("repo-b"))
	})

	It("overrides the Message spec model when set, preserves otherwise", func() {
		base := msgSpecCfg("claude-haiku-4-5")
		Expect(verify.Merge(base, verify.CommitConfig{}).Message.Spec.Model.Name).To(Equal("claude-haiku-4-5"))
		Expect(verify.Merge(base, msgSpecCfg("gpt-4o")).Message.Spec.Model.Name).To(Equal("gpt-4o"))
	})

	It("overrides the Grouping spec model when set, preserves otherwise", func() {
		base := verify.CommitConfig{Grouping: verify.PromptSpec{Spec: api.Spec{Model: api.Model{Name: "claude-sonnet-4-5"}}}}
		Expect(verify.Merge(base, verify.CommitConfig{}).Grouping.Spec.Model.Name).To(Equal("claude-sonnet-4-5"))
		override := verify.CommitConfig{Grouping: verify.PromptSpec{Spec: api.Spec{Model: api.Model{Name: "claude-opus-4-1"}}}}
		Expect(verify.Merge(base, override).Grouping.Spec.Model.Name).To(Equal("claude-opus-4-1"))
	})

	It("overrides MaxCommits when non-zero, preserves otherwise", func() {
		base := verify.CommitConfig{MaxCommits: 7}
		Expect(verify.Merge(base, verify.CommitConfig{}).MaxCommits).To(Equal(7))
		Expect(verify.Merge(base, verify.CommitConfig{MaxCommits: 3}).MaxCommits).To(Equal(3))
	})
})

// modelName resolves a ladder and returns just the name, for the precedence
// tables below. Mode/effort fidelity is asserted separately, in
// "carries a compact model's backend and effort through to the agent".
func modelName(m api.Model, err error) string {
	Expect(err).ToNot(HaveOccurred())
	return m.Name
}

// modelFlags builds the flag surface with just a model set.
func modelFlags(model string) aiflags.ModelFlags { return aiflags.ModelFlags{Model: model} }

var _ = Describe("Options model resolution", func() {
	DescribeTable("messageModel prefers --model, then commit.message.model, then ai.model, else the built-in default",
		func(opts Options, expected string) {
			Expect(modelName(opts.messageModel())).To(Equal(expected))
		},
		Entry("flag wins", Options{Flags: modelFlags("haiku-flag"), Config: msgSpecCfg("haiku-cfg"), AI: aiBaseSpec("base")}, "haiku-flag"),
		Entry("message spec used when no flag", Options{Config: msgSpecCfg("haiku-cfg"), AI: aiBaseSpec("base")}, "haiku-cfg"),
		Entry("ai base used when no flag or op spec", Options{AI: aiBaseSpec("base")}, "base"),
		Entry("defaults to haiku-class when nothing set", Options{}, defaultMessageModel),
	)

	DescribeTable("groupModel cascades --group-model -> --model -> commit.grouping.model -> ai.model -> default",
		func(opts Options, expected string) {
			Expect(modelName(opts.groupModel())).To(Equal(expected))
		},
		Entry("--group-model wins over everything",
			Options{GroupModel: "opus-flag", Flags: modelFlags("haiku-flag"), Config: msgGroupSpecCfg("haiku-cfg", "sonnet-cfg"), AI: aiBaseSpec("base")}, "opus-flag"),
		// The inversion this refactor removes: --model used to be read LAST for
		// grouping, so any `ai: model:` in ~/.gavel.yaml silently disabled it.
		Entry("--model beats the ai base spec",
			Options{Flags: modelFlags("haiku-flag"), AI: aiBaseSpec("base")}, "haiku-flag"),
		Entry("--model beats commit.grouping.model",
			Options{Flags: modelFlags("haiku-flag"), Config: msgGroupSpecCfg("haiku-cfg", "sonnet-cfg")}, "haiku-flag"),
		Entry("commit.grouping.model used when no flag",
			Options{Config: msgGroupSpecCfg("haiku-cfg", "sonnet-cfg")}, "sonnet-cfg"),
		Entry("ai base used when no flag or grouping spec",
			Options{Config: msgSpecCfg("haiku-cfg"), AI: aiBaseSpec("base")}, "base"),
		// Grouping deliberately no longer reads commit.message.model: that is the
		// message operation's model, and borrowing it across operations meant
		// configuring one silently redirected the other.
		Entry("ignores commit.message.model and takes the default",
			Options{Config: msgSpecCfg("haiku-cfg")}, defaultGroupModel),
		Entry("defaults to sonnet-class when nothing set",
			Options{}, defaultGroupModel),
	)

	// The reported bug, end to end: `gavel commit --model agent:gpt-5.6-luna:medium`
	// ran as api:gpt-5.6-luna. It failed three independent ways — the config beat
	// the flag, the string return dropped Mode/Effort, and BuildAgent's
	// DefaultConfig() re-inferred the backend from the bare name.
	It("carries a compact model's backend and effort through to the agent", func() {
		var got clickyai.AgentConfig
		previousAgent := newAgentFunc
		newAgentFunc = func(cfg clickyai.AgentConfig) (clickyai.Agent, error) {
			got = cfg
			return nil, errors.New("stop before provider construction")
		}
		DeferCleanup(func() { newAgentFunc = previousAgent })

		opts := Options{
			Flags: modelFlags("agent:gpt-5.6-luna:medium"),
			// The ~/.gavel.yaml `ai:` block that used to win over --model.
			AI: api.Spec{Model: api.Model{Name: "sonnet", Mode: api.ModeAPI}},
		}
		model, err := opts.groupModel()
		Expect(err).ToNot(HaveOccurred())
		_, _ = BuildAgent(opts, model)

		Expect(got.Model.Name).To(Equal("gpt-5.6-luna"))
		Expect(got.Model.Mode).To(Equal(api.ModeAgent), "the agent: mode must survive; it used to become the api mode")
		Expect(got.Model.Effort).To(Equal(api.EffortMedium), "the :medium effort must survive; it used to be dropped")
	})

	It("keeps the configured budget instead of resetting to the agent default", func() {
		var got clickyai.AgentConfig
		previousAgent := newAgentFunc
		newAgentFunc = func(cfg clickyai.AgentConfig) (clickyai.Agent, error) {
			got = cfg
			return nil, errors.New("stop before provider construction")
		}
		DeferCleanup(func() { newAgentFunc = previousAgent })

		opts := Options{AI: api.Spec{Budget: api.Budget{Cost: 5, MaxTokens: 2000}}}
		_, _ = BuildAgent(opts, api.Model{Name: "claude-sonnet-5"})
		Expect(got.Budget.Cost).To(Equal(5.0))
		Expect(got.Budget.MaxTokens).To(Equal(2000))
	})
})

var _ = Describe("BuildAgent errors", func() {
	It("preserves provider-specific credential guidance without appending every provider key", func() {
		previousAgent := newAgentFunc
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) {
			return nil, errors.New("API key not found for backend openai; similar environment variable found: OPEN_AI_API_KEY (did you mean OPENAI_API_KEY?)")
		}
		DeferCleanup(func() { newAgentFunc = previousAgent })

		_, err := BuildAgent(Options{}, api.Model{Name: "gpt-5.6-terra", Mode: api.ModeAPI})
		Expect(err).To(MatchError("LLM agent unavailable: API key not found for backend openai; similar environment variable found: OPEN_AI_API_KEY (did you mean OPENAI_API_KEY?)"))
		Expect(err.Error()).ToNot(ContainSubstring("ANTHROPIC_API_KEY"))
		Expect(errors.Is(err, ErrLLMUnavailable)).To(BeTrue())
	})
})
