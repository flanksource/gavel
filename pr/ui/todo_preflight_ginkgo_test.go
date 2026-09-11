package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo preview preflight", func() {
	var server *Server
	var todo *types.TODO
	BeforeEach(func() {
		dir := GinkgoT().TempDir()
		var err error
		todo, err = uiTestProviderFor(dir).Create(context.Background(), todos.CreateRequest{Title: "Review preflight", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		server = &Server{ghOpts: github.Options{WorkDir: dir}}
		original := run.Start
		DeferCleanup(func() { run.Start = original })
		run.Start = func(run.Request) (run.StartResult, error) {
			Fail("preflight must not start a run")
			return run.StartResult{}, nil
		}
	})

	It("returns capability warnings beside the resolved layer trace", func() {
		body, err := json.Marshal(todoRunPayload{Ref: todo.ID, Step: "plan", Spec: api.Spec{
			Model:           api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent},
			Permissions:     api.Permissions{Plugins: api.ResourcePolicies{"review-tools": api.ResourceEnabled}},
			ToolPreferences: api.ToolPreferences{"review": api.ToolPolicyAsk},
		}})
		Expect(err).NotTo(HaveOccurred())
		recorder := httptest.NewRecorder()
		server.handleTodoRunPreview(recorder, httptest.NewRequest(http.MethodPost, "/api/todos/run/preview", strings.NewReader(string(body))))
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var response map[string]json.RawMessage
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(string(response["warnings"])).To(ContainSubstring("plugins"))
		Expect(string(response["warnings"])).NotTo(ContainSubstring("requires Config.CanUseTool"), "preview carries the dashboard's actual deferred broker")
		Expect(string(response["trace"])).To(ContainSubstring(`"source":"request"`))
	})

	DescribeTable("refuses unsupported built-in planning policy before preview or run",
		func(mode api.RuntimeMode) {
			body, err := json.Marshal(todoRunPayload{Ref: todo.ID, Step: "plan", Spec: api.Spec{
				Model:       api.Model{Name: "gpt-5.6-sol", Mode: mode},
				Permissions: api.Permissions{Tools: api.Tools{"shell": api.ToolPolicyAllow}},
			}})
			Expect(err).NotTo(HaveOccurred())
			for _, handler := range []http.HandlerFunc{server.handleTodoRunPreview, server.handleTodoRun} {
				recorder := httptest.NewRecorder()
				handler(recorder, httptest.NewRequest(http.MethodPost, "/api/todos/run/preview", strings.NewReader(string(body))))
				Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
				Expect(recorder.Body.String()).To(ContainSubstring("cannot enforce a per-tool policy"))
			}
		}, Entry("agent", api.ModeAgent), Entry("cmux", api.ModeCmux),
	)

	It("rejects a missing judge in both preview and run before dispatch", func() {
		body, err := json.Marshal(todoRunPayload{Ref: todo.ID, Step: "plan", Spec: api.Spec{
			Workflow: &api.Workflow{Verify: &api.Verify{Prompts: []string{"missing-review.prompt"}}},
		}})
		Expect(err).NotTo(HaveOccurred())
		for _, handler := range []http.HandlerFunc{server.handleTodoRunPreview, server.handleTodoRun} {
			recorder := httptest.NewRecorder()
			handler(recorder, httptest.NewRequest(http.MethodPost, "/api/todos/run/preview", strings.NewReader(string(body))))
			Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
			Expect(recorder.Body.String()).To(ContainSubstring("missing-review.prompt"))
		}
	})
})
