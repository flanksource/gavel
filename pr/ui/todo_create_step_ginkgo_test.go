package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create TODO steps", func() {
	var workDir string
	var server *Server
	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		server = &Server{ghOpts: github.Options{WorkDir: workDir}}
	})

	getSteps := func(query string) (*httptest.ResponseRecorder, todoCreateStepsResponse) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/todos/new/steps?dir="+url.QueryEscape(workDir)+query, nil)
		server.Handler().ServeHTTP(recorder, request)
		var response todoCreateStepsResponse
		if recorder.Code == http.StatusOK {
			Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		}
		return recorder, response
	}
	stepNames := func(response todoCreateStepsResponse) []string {
		names := make([]string, 0, len(response.Steps))
		for _, step := range response.Steps {
			names = append(names, step.Name)
		}
		return names
	}

	DescribeTable("lists the steps the builtin lifecycle applies to a new todo, in definition order",
		func(query string, want []string) {
			recorder, response := getSteps(query)
			Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
			Expect(stepNames(response)).To(Equal(want))
		},
		Entry("omitted status is draft", "", []string{"triage", "plan", "triage.new"}),
		Entry("draft", "&status=draft", []string{"triage", "plan", "triage.new"}),
		Entry("verified has no plan or run", "&status=verified", []string{"triage", "triage.new"}),
	)

	It("projects each step with its label, prompt and auxiliary flag", func() {
		_, response := getSteps("&status=draft")
		Expect(response.Steps).To(ContainElements(
			todoCreateStepOption{Name: "plan", Label: "Plan", Prompt: "todos.plan", Auxiliary: false},
			todoCreateStepOption{Name: "triage.new", Label: "Triage new", Prompt: "todos.triage.new", Auxiliary: true},
		))
	})

	DescribeTable("decodes the create step with body precedence over query", func(contentType, body, query, want string) {
		request := httptest.NewRequest(http.MethodPost, "/api/todos/new"+query, strings.NewReader(body))
		request.Header.Set("Content-Type", contentType)
		payload, _, err := parseTodoNewPayload(request)
		Expect(err).NotTo(HaveOccurred())
		Expect(payload.Step).To(Equal(want))
	},
		Entry("JSON over query", "application/json", `{"step":"plan"}`, "?step=run", "plan"),
		Entry("form over query", "application/x-www-form-urlencoded", "step=plan", "?step=run", "plan"),
		Entry("query alone", "", "", "?step=run", "run"),
	)

	It("rejects a status a caller cannot assign", func() {
		recorder, _ := getSteps("&status=review")
		Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
	})

	Context("POST /api/todos/new with a step", func() {
		var started []run.Request
		var startError error
		BeforeEach(func() {
			started, startError = nil, nil
			resolve, start := run.Resolve, run.Start
			DeferCleanup(func() { run.Resolve = resolve; run.Start = start })
			run.Resolve = func(ctx context.Context, req run.Request) (*run.Prepared, error) {
				persisted, err := req.Provider.Get(ctx, req.Todo.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(persisted.MarkdownBody).To(ContainSubstring("Persist this body"))
				return &run.Prepared{Resolution: &lifecycle.Resolution{}}, nil
			}
			run.Start = func(req run.Request) (run.StartResult, error) {
				started = append(started, req)
				return run.StartResult{Status: "started", SessionID: "approved-session"}, startError
			}
		})
		post := func(extra string) (*httptest.ResponseRecorder, map[string]any) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/todos/new?dir="+url.QueryEscape(workDir),
				strings.NewReader(`{"title":"New task","body":"Persist this body"`+extra+`}`))
			request.Header.Set("Content-Type", "application/json")
			server.Handler().ServeHTTP(recorder, request)
			var response map[string]any
			Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
			return recorder, response
		}

		It("starts the named step on the saved todo with approvals on the dashboard host", func() {
			recorder, response := post(`,"step":"plan"`)
			Expect(recorder.Code).To(Equal(http.StatusCreated), recorder.Body.String())
			Expect(started).To(HaveLen(1))
			Expect(started[0].Options.Step).To(Equal("plan"))
			Expect(started[0].Options.Host).To(Equal(lifecycle.HostDashboard))
			Expect(started[0].Approvals).To(BeTrue())
			Expect(response["run"]).To(Equal(map[string]any{"status": "started", "sessionId": "approved-session"}))
		})

		It("starts nothing when no step is named", func() {
			recorder, response := post("")
			Expect(recorder.Code).To(Equal(http.StatusCreated), recorder.Body.String())
			Expect(started).To(BeEmpty())
			Expect(response).NotTo(HaveKey("run"))
		})

		It("rejects a step that does not apply to the new todo's status, naming the applicable ones", func() {
			recorder, response := post(`,"status":"verified","step":"plan"`)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
			Expect(response["error"]).To(ContainSubstring("triage, triage.new"))
			Expect(started).To(BeEmpty())
		})

		It("returns the saved todo and the run failure when admission fails", func() {
			startError = fmt.Errorf("provider unavailable")
			recorder, response := post(`,"step":"plan"`)
			Expect(recorder.Code).To(Equal(http.StatusCreated), recorder.Body.String())
			Expect(response["todo"]).NotTo(BeNil())
			Expect(response["run"]).To(HaveKeyWithValue("error", "provider unavailable"))
		})
	})
})
