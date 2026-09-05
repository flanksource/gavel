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

func todoRuntimeProfileWorkspace() string {
	dir := GinkgoT().TempDir()
	profiles := filepath.Join(dir, ".captain", "profiles")
	Expect(os.MkdirAll(profiles, 0o700)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(profiles, "reviewer.yaml"), []byte("name: reviewer\ndescription: Review changes\nspec:\n  model: agent:sonnet\npresets: []\n"), 0o600)).To(Succeed())
	return dir
}

var _ = Describe("todo runtime profiles", func() {
	It("carries an explicit profile through JSON decoding and run options", func() {
		var payload todoRunPayload
		request := httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(`{"ref":"example","step":"plan","runtimeProfile":"reviewer"}`))
		Expect(decodeTodoRequest(request, &payload)).To(Succeed())
		encoded, err := json.Marshal(payload)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).To(ContainSubstring(`"runtimeProfile":"reviewer"`))
		options, err := buildTodoRunOptions(payload, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(options.RuntimeProfile).To(Equal("reviewer"))
	})

	It("carries a profile through the dashboard bulk resolver", func() {
		options, err := resolveBulkRunOptions(context.Background(), bulk.RunRequest{
			Step: "plan", Todo: &types.TODO{ID: "example"}, Flags: bulk.RunFlags{RuntimeProfile: "reviewer"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(options.RuntimeProfile).To(Equal("reviewer"))
	})

	It("accepts runtime-profile through the generated entity HTTP action", func() {
		dir := todoRuntimeProfileWorkspace()
		todo, err := uiTestProviderFor(dir).Create(context.Background(), todos.CreateRequest{Title: "Bulk review", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		original := run.Start
		DeferCleanup(func() { run.Start = original })
		var dispatched *run.Request
		run.Start = func(request run.Request) (run.StartResult, error) {
			dispatched = &request
			return run.StartResult{Status: "started", SessionID: "example-session"}, nil
		}
		recorder, result := postTodoAction("plan", []string{todo.ID}, map[string]string{"runtime-profile": "reviewer"})
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		Expect(result.Applied).To(Equal(1), "%+v", result.Results)
		Expect(dispatched).NotTo(BeNil())
		Expect(dispatched.Options.RuntimeProfile).To(Equal("reviewer"))
	})

	It("lists catalog metadata and reuses the discovered catalog", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "reviewer.yaml"), []byte("name: reviewer\ndescription: Review changes\nspec:\n  model: agent:sonnet\npresets: []\n"), 0o600)).To(Succeed())
		source, err := runtimeprofiles.NewFileSource(runtimeprofiles.FileSourceOptions{Kind: runtimeprofiles.KindProfile, Dir: dir, Label: "project"})
		Expect(err).NotTo(HaveOccurred())
		catalog, err := runtimeprofiles.NewCatalog(source)
		Expect(err).NotTo(HaveOccurred())
		calls := 0
		host := &lifecycle.Host{Catalog: func(context.Context) (*runtimeprofiles.Catalog, error) { calls++; return catalog, nil }}
		profiles, err := todoRunProfileOptions(context.Background(), host)
		Expect(err).NotTo(HaveOccurred())
		Expect(profiles).To(HaveExactElements(SatisfyAll(
			HaveField("Name", "reviewer"), HaveField("Description", "Review changes"),
			HaveField("Model", "agent:sonnet"), HaveField("Source", source.Info()),
		)))
		Expect(profiles[0].ID).NotTo(BeEmpty())
		_, err = host.RuntimeCatalog(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(Equal(1))
	})

	It("preserves catalog discovery errors", func() {
		host := &lifecycle.Host{Catalog: func(context.Context) (*runtimeprofiles.Catalog, error) { return nil, errors.New("catalog unavailable") }}
		_, err := todoRunProfileOptions(context.Background(), host)
		Expect(err).To(MatchError(ContainSubstring("catalog unavailable")))
	})

	It("exposes the catalog and canonical per-step profile defaults together", func() {
		dir := todoRuntimeProfileWorkspace()
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("todos:\n  runtimeProfile: reviewer\n"), 0o600)).To(Succeed())
		previous := runCaptainWhoami
		DeferCleanup(func() { runCaptainWhoami = previous })
		runCaptainWhoami = func(captaincli.WhoamiOptions) (any, error) {
			return captaincli.WhoamiResult{Adapters: []captaincli.AdapterStatus{{Provider: "anthropic", Mode: "agent", Models: []string{"claude-sonnet-5"}}}}, nil
		}
		result, err := todoRunContext(context.Background(), dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RuntimeProfiles).To(HaveExactElements(HaveField("Name", "reviewer")))
		Expect(result.PromptDefaults["run"].RuntimeProfile).To(Equal(result.RuntimeProfiles[0].ID))
		Expect(result.PromptDefaults["verify"].RuntimeProfile).To(Equal(result.RuntimeProfiles[0].ID))
	})

	It("returns selected profile metadata and the effective trace in a preview", func() {
		dir := todoRuntimeProfileWorkspace()
		todo, err := uiTestProviderFor(dir).Create(context.Background(), todos.CreateRequest{Title: "Review change", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		body, err := json.Marshal(todoRunPayload{
			Ref: todo.ID, Step: "plan", RuntimeProfile: "reviewer",
			Spec: api.Spec{Model: api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent}},
		})
		Expect(err).NotTo(HaveOccurred())
		recorder := httptest.NewRecorder()
		server := &Server{ghOpts: github.Options{WorkDir: dir}}
		server.handleTodoRunPreview(recorder, httptest.NewRequest(http.MethodPost, "/api/todos/run/preview", strings.NewReader(string(body))))
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var preview todoRunPreviewResponse
		Expect(json.Unmarshal(recorder.Body.Bytes(), &preview)).To(Succeed())
		Expect(preview.RuntimeProfile).NotTo(BeNil())
		Expect(preview.RuntimeProfile.Profile.Name).To(Equal("reviewer"))
		Expect(preview.Model).To(Equal("claude-sonnet-5"))
		Expect(preview.RuntimeMode).To(Equal("agent"))
		Expect(preview.Trace).To(ContainElement(HaveField("Source", api.SpecLayerSourceProfile)))
		Expect(preview.SpecYAML).NotTo(BeEmpty())
		var wire map[string]json.RawMessage
		Expect(json.Unmarshal(recorder.Body.Bytes(), &wire)).To(Succeed())
		Expect(string(wire["runtimeProfile"])).NotTo(ContainSubstring(`"resolved"`))
	})
})
