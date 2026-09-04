package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo run admission contract", func() {
	It("previews the exact rendered Captain spec as YAML", func(ctx SpecContext) {
		workDir := GinkgoT().TempDir()
		server := &Server{ghOpts: github.Options{WorkDir: workDir}}
		created, err := uiTestProviderFor(workDir).Create(ctx, todos.CreateRequest{
			Title:    "Render the Captain request",
			Body:     "Keep every advanced option.",
			Priority: types.PriorityMedium,
			Status:   types.StatusPending,
		})
		Expect(err).NotTo(HaveOccurred())

		body, err := json.Marshal(todoRunPayload{
			Ref:  todos.TODOReference(created),
			Step: "run",
			Spec: api.Spec{
				Model:  api.Model{Mode: api.ModeAgent, Name: "gpt-5.5", Effort: api.EffortHigh},
				Budget: api.Budget{Timeout: "45m", MaxTurns: 12},
				Prompt: api.Prompt{AppendSystem: "Keep the contract visible."},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		recorder := httptest.NewRecorder()
		server.handleTodoRunPreview(recorder, httptest.NewRequest(http.MethodPost, "/api/todos/run/preview", strings.NewReader(string(body))))
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())

		var response todoRunPreviewResponse
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response.SpecYAML).NotTo(BeEmpty())
		var rendered api.Spec
		Expect(yaml.Unmarshal([]byte(response.SpecYAML), &rendered)).To(Succeed())
		Expect(rendered.Prompt.User).To(Equal(response.Prompt))
		Expect(rendered.Prompt.AppendSystem).To(Equal("Keep the contract visible."))
		Expect(rendered.Budget).To(Equal(api.Budget{Timeout: "45m0s", MaxTurns: 12}))
		Expect(rendered.Mode).To(Equal(api.ModeAgent))
		Expect(rendered.Name).To(Equal("gpt-5.5"))
	})

	It("returns the admitted Captain session id before reporting started", func(ctx SpecContext) {
		workDir := GinkgoT().TempDir()
		server := &Server{ghOpts: github.Options{WorkDir: workDir}}
		created, err := uiTestProviderFor(workDir).Create(ctx, todos.CreateRequest{
			Title:  "Tail immediately",
			Status: types.StatusPending,
		})
		Expect(err).NotTo(HaveOccurred())

		previousStart := run.Start
		DeferCleanup(func() { run.Start = previousStart })
		run.Start = func(todoRunRequest) (todoRunStartResult, error) {
			return todoRunStartResult{Status: "started", SessionID: "11111111-1111-4111-8111-111111111111"}, nil
		}

		body, err := json.Marshal(todoRunPayload{
			Ref:  todos.TODOReference(created),
			Spec: specPayload("gpt-5.5", "high"),
		})
		Expect(err).NotTo(HaveOccurred())
		recorder := httptest.NewRecorder()
		server.handleTodoRun(recorder, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body))))
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())

		var response todoRunResponse
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response.Status).To(Equal("started"))
		Expect(response.SessionID).To(Equal("11111111-1111-4111-8111-111111111111"))
	})
})
