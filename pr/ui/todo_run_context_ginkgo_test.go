package ui

import (
	"errors"
	"net/http"
	"net/http/httptest"

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
					Backend:       "codex-agent",
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
					"openai": {Agent: "codex-agent", Model: "gpt-captain-default", Effort: string(api.EffortHigh)},
				},
			}, nil
		}

		context, err := todoRunContext()

		Expect(err).NotTo(HaveOccurred())
		Expect(context.DefaultBackend).To(Equal("codex-agent"))
		Expect(context.Backends).To(HaveLen(1))
		Expect(context.Backends[0].ID).To(Equal("codex-agent"))
		Expect(context.Backends[0].DefaultModel).To(Equal("gpt-captain-default"))
		Expect(context.Backends[0].Models).To(HaveExactElements(
			HaveField("ID", "gpt-captain-default"),
			HaveField("ID", "gpt-captain-other"),
		))
	})

	It("prefers the agent backend when Captain defaults the provider to CLI", func() {
		runCaptainWhoami = func(captaincli.WhoamiOptions) (any, error) {
			return captaincli.WhoamiResult{
				Adapters: []captaincli.AdapterStatus{
					{Backend: "claude-cli", Models: []string{"claude-opus-4-8"}},
					{Backend: "claude-agent", Models: []string{"claude-opus-4-8"}},
				},
				DefaultProvider: "anthropic",
				ProviderDefaults: map[string]captaincli.ProviderDefaultView{
					"anthropic": {Agent: "claude-cli", Model: "claude-opus-4-8"},
				},
			}, nil
		}

		context, err := todoRunContext()

		Expect(err).NotTo(HaveOccurred())
		Expect(context.DefaultBackend).To(Equal("claude-agent"))
	})

	It("does not synthesize a model when Captain returns an empty adapter catalog", func() {
		runCaptainWhoami = func(captaincli.WhoamiOptions) (any, error) {
			return captaincli.WhoamiResult{
				Adapters: []captaincli.AdapterStatus{{
					Backend:    "claude-agent",
					Type:       "cli",
					ModelError: "Captain returned no models",
				}},
			}, nil
		}

		context, err := todoRunContext()

		Expect(err).NotTo(HaveOccurred())
		Expect(context.Backends).To(HaveLen(1))
		Expect(context.Backends[0].DefaultModel).To(BeEmpty())
		Expect(context.Backends[0].Models).To(BeEmpty())
		Expect(context.Backends[0].ModelError).To(Equal("Captain returned no models"))
	})
})
