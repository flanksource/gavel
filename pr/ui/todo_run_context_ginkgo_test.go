package ui

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo run context catalog", func() {
	var previous func(captaincli.WhoamiOptions) (any, error)

	BeforeEach(func() {
		previous = runCaptainWhoami
	})

	AfterEach(func() {
		runCaptainWhoami = previous
	})

	It("returns a service error when Captain catalog discovery fails", func() {
		runCaptainWhoami = func(captaincli.WhoamiOptions) (any, error) {
			return nil, errors.New("captain catalog unavailable")
		}

		recorder := httptest.NewRecorder()
		(&Server{}).handleTodoRunContext(recorder, httptest.NewRequest(http.MethodGet, "/api/todos/run/context", nil))

		Expect(recorder.Code).To(Equal(http.StatusServiceUnavailable))
		Expect(recorder.Body.String()).To(ContainSubstring("load run providers from Captain: captain catalog unavailable"))
	})

	It("projects only Captain adapters and uses Captain provider defaults", func() {
		runCaptainWhoami = func(captaincli.WhoamiOptions) (any, error) {
			return captaincli.WhoamiResult{
				Adapters: []captaincli.AdapterStatus{{
					Provider:      "openai",
					Mode:          "agent",
					Type:          "cli",
					Authenticated: true,
					Binary:        "/usr/local/bin/codex",
					ModelDetails: []captainai.ModelDef{
						{ID: "gpt-captain-default", Name: "Captain Default", CapabilitiesKnown: true, Reasoning: true},
						{ID: "gpt-captain-other", Name: "Captain Other", CapabilitiesKnown: true, Reasoning: true},
					},
				}},
				DefaultProvider: "openai",
				ProviderDefaults: map[string]captaincli.ProviderDefaultView{
					"openai": {Mode: "agent", Model: "gpt-captain-default", Effort: string(api.EffortHigh)},
				},
			}, nil
		}

		context, err := todoRunContext("")

		Expect(err).NotTo(HaveOccurred())
		Expect(context.DefaultMode).To(Equal("agent"))
		Expect(context.DefaultProvider).To(Equal("openai"))
		Expect(context.Modes).To(HaveLen(1))
		Expect(context.Modes[0].ID).To(Equal("agent"))
		Expect(context.Modes[0].Driver).To(Equal("agent"))
		Expect(context.Modes[0].DefaultModel).To(Equal("gpt-captain-default"))
		Expect(context.Modes[0].Models).To(HaveExactElements(
			HaveField("ID", "gpt-captain-default"),
			HaveField("ID", "gpt-captain-other"),
		))
		Expect(context.Models[0]).To(SatisfyAll(
			HaveField("Provider", "openai"),
			HaveField("Modes", []string{"agent"}),
			HaveField("Runtime", api.Model{Name: "gpt-captain-default"}),
		))
	})

	It("prefers the agent mode when Captain defaults the provider to CLI", func() {
		runCaptainWhoami = func(captaincli.WhoamiOptions) (any, error) {
			return captaincli.WhoamiResult{
				Adapters: []captaincli.AdapterStatus{
					{Provider: "anthropic", Mode: "cli", Models: []string{"claude-opus-4-8"}},
					{Provider: "anthropic", Mode: "agent", Models: []string{"claude-opus-4-8"}},
				},
				DefaultProvider: "anthropic",
				ProviderDefaults: map[string]captaincli.ProviderDefaultView{
					"anthropic": {Mode: "cli", Model: "claude-opus-4-8"},
				},
			}, nil
		}

		context, err := todoRunContext("")

		Expect(err).NotTo(HaveOccurred())
		Expect(context.DefaultMode).To(Equal("agent"))
		Expect(context.DefaultProvider).To(Equal("anthropic"))
	})

	It("does not synthesize a model when Captain returns an empty adapter catalog", func() {
		runCaptainWhoami = func(captaincli.WhoamiOptions) (any, error) {
			return captaincli.WhoamiResult{
				Adapters: []captaincli.AdapterStatus{{
					Provider:   "anthropic",
					Mode:       "agent",
					Type:       "cli",
					ModelError: "Captain returned no models",
				}},
			}, nil
		}

		context, err := todoRunContext("")

		Expect(err).NotTo(HaveOccurred())
		Expect(context.Modes).To(HaveLen(1))
		Expect(context.Modes[0].DefaultModel).To(BeEmpty())
		Expect(context.Modes[0].Models).To(BeEmpty())
		Expect(context.Modes[0].ModelError).To(Equal("Captain returned no models"))
	})

	// The dialog used to seed every action from DefaultMode, which is the
	// account-wide default and knows nothing about a prompt's frontmatter. Under
	// an openai agent default it therefore sent that runtime as if the operator
	// had chosen it, outranking the `model: claude` that todos-triage.prompt and
	// todos-plan.prompt pin — and because both prompts also declare a per-tool
	// policy only the Claude transports carry, Captain refused the run with
	// "the openai agent cannot enforce a per-tool policy (Glob, Grep, Read)".
	It("resolves each prompt's own runtime rather than the account default", func() {
		runCaptainWhoami = func(captaincli.WhoamiOptions) (any, error) {
			return captaincli.WhoamiResult{
				Adapters: []captaincli.AdapterStatus{
					{
						Provider: "openai", Mode: "agent", Type: "cli", Authenticated: true,
						ModelDetails: []captainai.ModelDef{{ID: "gpt-5.6-sol", CapabilitiesKnown: true}},
					},
					{
						Provider: "anthropic", Mode: "agent", Type: "cli", Authenticated: true,
						ModelDetails: []captainai.ModelDef{{ID: "claude-opus-4-8", CapabilitiesKnown: true}},
					},
				},
				DefaultProvider: "openai",
				ProviderDefaults: map[string]captaincli.ProviderDefaultView{
					"openai": {Mode: "agent", Model: "gpt-5.6-sol"},
				},
			}, nil
		}
		// The reported configuration: a codex account default, with todos.run
		// pinned to codex above the prompt frontmatter, and todos.triage silent.
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(
			filepath.Join(dir, ".gavel.yaml"),
			[]byte("ai:\n  mode: agent\n  model: gpt-5.6-luna\n"+
				"todos:\n  run:\n    mode: agent\n    model: gpt-5.6-sol\n  triage: {}\n"),
			0o600,
		)).To(Succeed())

		context, err := todoRunContext(dir)

		Expect(err).NotTo(HaveOccurred())
		Expect(context.DefaultMode).To(Equal("agent"), "the account default mechanism is unchanged")
		Expect(context.DefaultProvider).To(Equal("openai"))
		// The map must not be uniform, or it would prove nothing: todos-run.prompt
		// pins no provider, so `run` keeps the configured codex mode, while plan
		// and triage pin `model: claude` and must move off it.
		Expect(context.PromptDefaults).To(HaveKeyWithValue("run",
			HaveField("Mode", "agent")))
		Expect(context.PromptDefaults).To(HaveKeyWithValue("plan",
			HaveField("Mode", "agent")))
		Expect(context.PromptDefaults).To(HaveKeyWithValue("triage",
			HaveField("Mode", "agent")))
	})
})
