package ui

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/bulk"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func todoRuntimePresetWorkspace() string {
	dir := GinkgoT().TempDir()
	presets := filepath.Join(dir, ".captain", "presets")
	Expect(os.MkdirAll(presets, 0o700)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(presets, "reviewer.yaml"), []byte("name: reviewer\ndescription: Review changes\nscope: context\nspec:\n  model: agent:sonnet\npresets: []\n"), 0o600)).To(Succeed())
	return dir
}

var _ = Describe("todo runtime presets", func() {
	It("carries ordered presets through JSON decoding and run options", func() {
		var payload todoRunPayload
		request := httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(`{"ref":"example","step":"plan","presets":["organization","reviewer"]}`))
		Expect(decodeTodoRequest(request, &payload)).To(Succeed())
		encoded, err := json.Marshal(payload)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).To(ContainSubstring(`"presets":["organization","reviewer"]`))
		options, err := buildTodoRunOptions(payload, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(options.Presets).To(Equal([]string{"organization", "reviewer"}))
		Expect(options.PresetsSet).To(BeTrue())
	})

	It("carries presets through the dashboard bulk resolver", func() {
		options, err := resolveBulkRunOptions(context.Background(), bulk.RunRequest{
			Step: "plan", Todo: &types.TODO{ID: "example"}, Flags: bulk.RunFlags{Presets: []string{"reviewer"}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(options.Presets).To(Equal([]string{"reviewer"}))
		Expect(options.PresetsSet).To(BeTrue())
	})

	It("accepts preset through the generated entity HTTP action", func() {
		dir := todoRuntimePresetWorkspace()
		todo, err := uiTestProviderFor(dir).Create(context.Background(), todos.CreateRequest{Title: "Bulk review", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		original := run.Start
		DeferCleanup(func() { run.Start = original })
		var dispatched *run.Request
		run.Start = func(request run.Request) (run.StartResult, error) {
			dispatched = &request
			return run.StartResult{Status: "started", SessionID: "example-session"}, nil
		}
		recorder, result := postTodoAction("plan", []string{todo.ID}, map[string]string{"preset": "reviewer"})
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		Expect(result.Applied).To(Equal(1), "%+v", result.Results)
		Expect(dispatched).NotTo(BeNil())
		Expect(dispatched.Options.Presets).To(Equal([]string{"reviewer"}))
		Expect(dispatched.Options.PresetsSet).To(BeTrue())
	})

	It("lists catalog metadata and reuses the discovered catalog", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "reviewer.yaml"), []byte("name: reviewer\ndescription: Review changes\nscope: context\nspec:\n  model: agent:sonnet\npresets: []\n"), 0o600)).To(Succeed())
		source, err := runtimeprofiles.NewFileSource(runtimeprofiles.FileSourceOptions{Kind: runtimeprofiles.KindPreset, Dir: dir, Label: "project"})
		Expect(err).NotTo(HaveOccurred())
		catalog, err := runtimeprofiles.NewCatalog(source)
		Expect(err).NotTo(HaveOccurred())
		calls := 0
		host := &lifecycle.Host{Catalog: func(context.Context) (*runtimeprofiles.Catalog, error) { calls++; return catalog, nil }}
		presets, err := todoRunPresetOptions(context.Background(), host)
		Expect(err).NotTo(HaveOccurred())
		Expect(presets).To(HaveExactElements(SatisfyAll(
			HaveField("Name", "reviewer"), HaveField("Description", "Review changes"),
			HaveField("Spec.Name", "agent:sonnet"), HaveField("Source", source.Info()),
		)))
		Expect(presets[0].ID).NotTo(BeEmpty())
		_, err = host.RuntimeCatalog(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(Equal(1))
	})

	It("preserves catalog discovery errors", func() {
		host := &lifecycle.Host{Catalog: func(context.Context) (*runtimeprofiles.Catalog, error) { return nil, errors.New("catalog unavailable") }}
		_, err := todoRunPresetOptions(context.Background(), host)
		Expect(err).To(MatchError(ContainSubstring("catalog unavailable")))
	})

	It("exposes the catalog and canonical per-step preset defaults together", func() {
		dir := todoRuntimePresetWorkspace()
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("todos:\n  presets: [reviewer]\n"), 0o600)).To(Succeed())
		previous := runCaptainWhoami
		DeferCleanup(func() { runCaptainWhoami = previous })
		runCaptainWhoami = func(captaincli.WhoamiOptions) (any, error) {
			return captaincli.WhoamiResult{Adapters: []captaincli.AdapterStatus{{Provider: "anthropic", Mode: "agent", Models: []string{"claude-sonnet-5"}}}}, nil
		}
		result, err := todoRunContext(context.Background(), dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RuntimePresets).To(HaveExactElements(
			HaveField("Name", "reviewer"),
			HaveField("Name", "Edit"), HaveField("Name", "Plan"), HaveField("Name", "Read-only"),
		))
		Expect(result.RuntimeProfiles).To(BeEmpty())
		Expect(result.PromptDefaults["run"].Presets).To(Equal([]string{result.RuntimePresets[0].ID}))
		Expect(result.PromptDefaults["verify"].Presets).To(Equal([]string{result.RuntimePresets[0].ID}))
	})

	It("returns selected presets and their effective trace in a preview", func() {
		dir := todoRuntimePresetWorkspace()
		todo, err := uiTestProviderFor(dir).Create(context.Background(), todos.CreateRequest{Title: "Review change", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		body, err := json.Marshal(todoRunPayload{
			Ref: todo.ID, Step: "plan", Presets: []string{"reviewer"},
			Spec: api.Spec{Model: api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent}},
		})
		Expect(err).NotTo(HaveOccurred())
		recorder := httptest.NewRecorder()
		server := &Server{ghOpts: github.Options{WorkDir: dir}}
		server.handleTodoRunPreview(recorder, httptest.NewRequest(http.MethodPost, "/api/todos/run/preview", strings.NewReader(string(body))))
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var preview todoRunPreviewResponse
		Expect(json.Unmarshal(recorder.Body.Bytes(), &preview)).To(Succeed())
		Expect(preview.RuntimePresets).To(HaveExactElements(HaveField("Name", "reviewer")))
		Expect(preview.RuntimeProfile).To(BeNil())
		Expect(preview.Model).To(Equal("claude-sonnet-5"))
		Expect(preview.RuntimeMode).To(Equal("agent"))
		Expect(preview.Trace).To(ContainElement(HaveField("Source", api.SpecLayerSourcePreset)))
		Expect(preview.SpecYAML).NotTo(BeEmpty())
	})

	It("retains the deprecated profile wire field as a warning-only no-op", func() {
		dir := todoRuntimePresetWorkspace()
		todo, err := uiTestProviderFor(dir).Create(context.Background(), todos.CreateRequest{Title: "Review change", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		recorder := postJSON((&Server{ghOpts: github.Options{WorkDir: dir}}).handleTodoRunPreview, "/api/todos/run/preview", todoRunPayload{
			Ref: todo.ID, Step: "plan", RuntimeProfile: "reviewer",
		})
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var preview todoRunPreviewResponse
		Expect(json.Unmarshal(recorder.Body.Bytes(), &preview)).To(Succeed())
		Expect(preview.RuntimeProfile).To(BeNil())
		Expect(preview.Warnings).To(ContainElement(api.RuntimeProfileDeprecationWarning))
	})
})
