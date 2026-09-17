package ui

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo configuration error boundaries", func() {
	var dir string
	var server *Server
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		dir = GinkgoT().TempDir()
		server = &Server{ghOpts: github.Options{WorkDir: dir}}
	})

	DescribeTable("keeps invalid project layers fatal despite a valid request override",
		func(config, field string) {
			created := specTodo(dir, types.StatusPending)
			called := stubSpecRunStart(nil)
			Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte(config), 0o600)).To(Succeed())
			temperature := 0.5
			payload := todoRunPayload{Ref: created.ID, Step: "plan", Spec: api.Spec{
				Model:       api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent, Temperature: &temperature},
				Permissions: api.Permissions{Mode: api.PermissionPlan},
			}}
			for _, handler := range []http.HandlerFunc{server.handleTodoRunPreview, server.handleTodoRun} {
				recorder := postJSON(handler, "/api/todos/run", payload)
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError), recorder.Body.String())
				Expect(recorder.Body.String()).To(And(ContainSubstring(".gavel.yaml"), ContainSubstring(field)))
			}
			Expect(*called).To(BeFalse())
		},
		Entry("AI temperature", "ai:\n  temperature: 3\n", "temperature"),
		Entry("AI permission mode", "ai:\n  permissions:\n    mode: invalid-policy\n", "permission"),
		Entry("loaded step permission mode", "todos:\n  plan:\n    permissions:\n      mode: invalid-policy\n", "permission"),
	)

	DescribeTable("refuses invalid project configuration before changing plan review state",
		func(action string) {
			created := specTodo(dir, types.StatusReview)
			called := stubSpecRunStart(nil)
			Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("todos:\n  plan:\n    permissions:\n      mode: invalid-policy\n"), 0o600)).To(Succeed())
			handler := server.handleTodoPlanApprove
			body := map[string]any{"ref": created.ID, "run": true}
			if action == "revise" {
				handler = server.handleTodoPlanRevise
				body = map[string]any{"ref": created.ID, "feedback": "Keep the smaller change."}
			}
			recorder := postJSON(handler, "/api/todos/plan/"+action, body)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError), recorder.Body.String())
			Expect(recorder.Body.String()).To(And(ContainSubstring(".gavel.yaml"), ContainSubstring("permission")))
			Expect(created.Status).To(Equal(types.StatusReview))
			Expect(*called).To(BeFalse())
		}, Entry("approve", "approve"), Entry("revise", "revise"),
	)

	It("returns a configuration error before recording an answer or dispatching", func() {
		created := specAskTodo(dir, "example-asking-session", types.PlanPhase)
		stubApprovalStore(errNoCaptainDatabase)
		called := stubSpecRunStart(nil)
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("ai:\n  permissions:\n    mode: invalid-policy\n"), 0o600)).To(Succeed())
		recorder := postJSON(server.handleTodoAnswer, "/api/todos/answer", todoAnswerPayload{
			Ref: created.ID, Answer: "Continue with the smaller change.",
			Options: &todoRunPayload{Spec: api.Spec{
				Model:       api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent},
				Permissions: api.Permissions{Mode: api.PermissionPlan},
			}},
		})
		Expect(recorder.Code).To(Equal(http.StatusInternalServerError), recorder.Body.String())
		Expect(recorder.Body.String()).To(And(ContainSubstring(".gavel.yaml"), ContainSubstring("permission")))
		Expect(uiTestProviderFor(dir).comments).To(BeEmpty())
		Expect(created.Status).To(Equal(types.StatusAsk))
		Expect(*called).To(BeFalse())
	})

	invalidTemperature := 3.0
	DescribeTable("keeps invalid request layers as client errors",
		func(spec api.Spec, field string) {
			created := specTodo(dir, types.StatusPending)
			called := stubSpecRunStart(nil)
			for _, handler := range []http.HandlerFunc{server.handleTodoRunPreview, server.handleTodoRun} {
				recorder := postJSON(handler, "/api/todos/run", todoRunPayload{Ref: created.ID, Step: "plan", Spec: spec})
				Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
				Expect(recorder.Body.String()).To(ContainSubstring(field))
			}
			Expect(*called).To(BeFalse())
		},
		Entry("permission mode", api.Spec{Permissions: api.Permissions{Mode: "invalid-policy"}}, "permission mode"),
		Entry("timeout", api.Spec{Budget: api.Budget{Timeout: "-1s"}}, "timeout"),
		Entry("temperature validated by Captain", api.Spec{Model: api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent, Temperature: &invalidTemperature}}, "temperature"),
	)

	DescribeTable("rejects invalid selected preset definitions",
		func(preset, detail string) {
			created := specTodo(dir, types.StatusPending)
			called := stubSpecRunStart(nil)
			presets := filepath.Join(dir, ".captain", "presets")
			Expect(os.MkdirAll(presets, 0o700)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(presets, "reviewer.yaml"), []byte(preset), 0o600)).To(Succeed())
			temperature := 0.5
			payload := todoRunPayload{Ref: created.ID, Step: "plan", Presets: []string{"reviewer"}, Spec: api.Spec{
				Model: api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent, Temperature: &temperature},
			}}
			for _, handler := range []http.HandlerFunc{server.handleTodoRunPreview, server.handleTodoRun} {
				recorder := postJSON(handler, "/api/todos/run", payload)
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError), recorder.Body.String())
				Expect(recorder.Body.String()).To(And(ContainSubstring("reviewer"), ContainSubstring(detail)))
			}
			Expect(*called).To(BeFalse())
		},
		Entry("invalid preset field", "name: reviewer\nscope: context\nspec:\n  temperature: 3\n", "temperature"),
		Entry("nested preset", "name: reviewer\nscope: context\npresets: [base]\n", "nested runtime presets are not supported"),
	)

	It("keeps an unknown requested preset as a client error", func() {
		created := specTodo(dir, types.StatusPending)
		called := stubSpecRunStart(nil)
		for _, handler := range []http.HandlerFunc{server.handleTodoRunPreview, server.handleTodoRun} {
			recorder := postJSON(handler, "/api/todos/run", todoRunPayload{Ref: created.ID, Step: "plan", Presets: []string{"missing-preset"}})
			Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
			Expect(recorder.Body.String()).To(ContainSubstring("missing-preset"))
		}
		Expect(*called).To(BeFalse())
	})

	It("evaluates preset runtime policy before the explicit request and reports each warning once", func() {
		created := specTodo(dir, types.StatusPending)
		presets := filepath.Join(dir, ".captain", "presets")
		Expect(os.MkdirAll(presets, 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(presets, "reviewer.yaml"), []byte(`name: reviewer
scope: context
spec:
  model: gpt-5.6-sol
  mode: agent
  permissions:
    tools:
      Read: allow
    plugins:
      review-tools: enabled
`), 0o600)).To(Succeed())
		recorder := postJSON(server.handleTodoRunPreview, "/api/todos/run/preview", todoRunPayload{
			Ref: created.ID, Step: "plan", Presets: []string{"reviewer"},
			Spec: api.Spec{Model: api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent}},
		})
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var preview todoRunPreviewResponse
		Expect(json.Unmarshal(recorder.Body.Bytes(), &preview)).To(Succeed())
		Expect(preview.Model).To(Equal("claude-sonnet-5"))
		Expect(preview.RuntimeMode).To(Equal("agent"))
		Expect(preview.RuntimePresets).To(HaveExactElements(HaveField("Name", "reviewer")))
		Expect(preview.RuntimeProfile).To(BeNil())
		Expect(preview.Trace).To(ContainElement(HaveField("Source", api.SpecLayerSourcePreset)))
		Expect(preview.Warnings).To(HaveExactElements(ContainSubstring("plugins")))
	})
})
