package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo run defaults provenance", func() {
	var dir string
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		dir = GinkgoT().TempDir()
		previous := runCaptainWhoami
		DeferCleanup(func() { runCaptainWhoami = previous })
		runCaptainWhoami = func(captaincli.WhoamiOptions) (any, error) {
			return captaincli.WhoamiResult{
				Adapters:         []captaincli.AdapterStatus{{Provider: "openai", Mode: "agent", Models: []string{"gpt-5.6-sol"}}},
				DefaultProvider:  "openai",
				ProviderDefaults: map[string]captaincli.ProviderDefaultView{"openai": {Mode: "agent", Model: "gpt-5.6-sol"}},
			}, nil
		}
	})

	readContext := func() map[string]any {
		recorder := httptest.NewRecorder()
		(&Server{}).handleTodoRunContext(recorder, httptest.NewRequest(http.MethodGet, "/api/todos/run/context?dir="+url.QueryEscape(dir), nil))
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var result map[string]any
		Expect(json.Unmarshal(recorder.Body.Bytes(), &result)).To(Succeed())
		return result
	}

	It("does not make catalog display defaults executable defaults for a model-less step", func() {
		result := readContext()
		Expect(result["defaultProvider"]).NotTo(Equal("openai"))
		for _, value := range result["modes"].([]any) {
			Expect(value).To(HaveKeyWithValue("defaultModel", ""))
		}
		defaults := result["promptDefaults"].(map[string]any)["verify"].(map[string]any)
		Expect(defaults).NotTo(HaveKey("model"))
		Expect(defaults).To(HaveKey("spec"))
	})

	It("projects a full authored step snapshot and provenance without resolving its model", func() {
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("todos:\n  run:\n    model: agent:sonnet\n    budget:\n      maxTurns: 7\n    permissions:\n      mode: plan\n    prompt:\n      appendSystem: Keep the shared snapshot.\n"), 0o600)).To(Succeed())
		defaults := readContext()["promptDefaults"].(map[string]any)["run"].(map[string]any)
		Expect(defaults).To(HaveKeyWithValue("model", "sonnet"))
		Expect(defaults["spec"]).To(SatisfyAll(
			HaveKeyWithValue("budget", HaveKeyWithValue("maxTurns", float64(7))),
			HaveKeyWithValue("permissions", HaveKeyWithValue("mode", "plan")),
			HaveKeyWithValue("prompt", HaveKeyWithValue("appendSystem", "Keep the shared snapshot.")),
		))
		Expect(defaults["trace"]).NotTo(BeEmpty())
		Expect(defaults["provenance"]).To(HaveKeyWithValue("/budget/maxTurns", HaveKeyWithValue("source", HaveKeyWithValue("kind", "layer"))))
	})

	It("returns the final full Spec and field provenance from the same preview resolution", func() {
		created := specTodo(dir, types.StatusPending)
		server := &Server{ghOpts: github.Options{WorkDir: dir}}
		recorder := postJSON(server.handleTodoRunPreview, "/api/todos/run/preview", map[string]any{
			"ref": created.ID, "step": "plan", "spec": map[string]any{"model": "claude-sonnet-5", "mode": "agent", "budget": map[string]any{"maxTurns": 7}},
		})
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var result map[string]any
		Expect(json.Unmarshal(recorder.Body.Bytes(), &result)).To(Succeed())
		Expect(result["spec"]).To(HaveKeyWithValue("budget", HaveKeyWithValue("maxTurns", float64(7))))
		Expect(result["provenance"]).To(HaveKeyWithValue("/budget/maxTurns", HaveKeyWithValue("source", HaveKeyWithValue("kind", "layer"))))
	})

	It("retains saved native fields and their provenance in the shared step snapshot", func() {
		Expect(os.WriteFile(filepath.Join(os.Getenv("HOME"), ".captain.yaml"), []byte("ai:\n  maxTokens: 1536\n  noCache: true\n"), 0o600)).To(Succeed())
		defaults := readContext()["promptDefaults"].(map[string]any)["verify"].(map[string]any)
		Expect(defaults["spec"]).To(SatisfyAll(
			HaveKeyWithValue("budget", HaveKeyWithValue("maxTokens", float64(1536))),
			HaveKeyWithValue("noCache", true),
		))
		Expect(defaults["provenance"]).To(HaveKeyWithValue("/budget/maxTokens", HaveKeyWithValue("source", SatisfyAll(
			HaveKeyWithValue("kind", "saved"), HaveKeyWithValue("key", "ai.maxTokens"),
		))))
	})
})
