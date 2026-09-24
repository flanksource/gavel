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
	"github.com/google/uuid"
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
		Expect(rendered.Budget).To(Equal(api.Budget{Timeout: "45m0s", MaxTurns: 12, MaxTokens: 4096}))
		Expect(rendered.Mode).To(Equal(api.ModeAgent))
		Expect(rendered.Name).To(Equal("gpt-5.5"))
	})

	It("returns the admitted Captain session id before reporting started", func(ctx SpecContext) {
		workDir := GinkgoT().TempDir()
		configureAutomaticPlanToolPolicies(GinkgoTB(), workDir)
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

	It("streams the dispatched spec before the admitted prompt run", func(ctx SpecContext) {
		workDir := GinkgoT().TempDir()
		configureAutomaticPlanToolPolicies(GinkgoTB(), workDir)
		server := &Server{ghOpts: github.Options{WorkDir: workDir}}
		created, err := uiTestProviderFor(workDir).Create(ctx, todos.CreateRequest{Title: "Show launch progress", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())

		promptRunID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
		previousStart := run.Start
		DeferCleanup(func() { run.Start = previousStart })
		var dispatched *run.Prepared
		run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
			dispatched = req.Prepared
			return todoRunStartResult{Status: "started", SessionID: "11111111-1111-4111-8111-111111111111", PromptRunID: promptRunID}, nil
		}
		body, err := json.Marshal(todoRunPayload{Ref: todos.TODOReference(created), Spec: specPayload("gpt-5.5", "high")})
		Expect(err).NotTo(HaveOccurred())
		request := httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body)))
		request.Header.Set("Accept", "text/event-stream")
		recorder := httptest.NewRecorder()
		server.handleTodoRun(recorder, request)
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		Expect(recorder.Header().Get("Content-Type")).To(Equal("text/event-stream"))
		Expect(dispatched).NotTo(BeNil())
		frames := strings.Split(strings.TrimSpace(recorder.Body.String()), "\n\n")
		Expect(frames).To(HaveLen(2))
		Expect(frames[0]).To(HavePrefix("event: resolved\ndata: "))
		var resolved struct {
			Spec     api.Spec `json:"spec"`
			SpecYAML string   `json:"specYaml"`
		}
		Expect(json.Unmarshal([]byte(strings.TrimPrefix(frames[0], "event: resolved\ndata: ")), &resolved)).To(Succeed())
		dispatchedJSON, err := json.Marshal(dispatched.Resolution.Spec)
		Expect(err).NotTo(HaveOccurred())
		resolvedJSON, err := json.Marshal(resolved.Spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(resolvedJSON).To(MatchJSON(dispatchedJSON))
		var rendered api.Spec
		Expect(yaml.Unmarshal([]byte(resolved.SpecYAML), &rendered)).To(Succeed())
		Expect(rendered.Name).To(Equal(dispatched.Resolution.Spec.Name))
		Expect(rendered.Prompt.User).To(Equal(dispatched.Resolution.Spec.Prompt.User))
		Expect(frames[1]).To(HavePrefix("event: admitted\ndata: "))
		var admitted todoRunResponse
		Expect(json.Unmarshal([]byte(strings.TrimPrefix(frames[1], "event: admitted\ndata: ")), &admitted)).To(Succeed())
		Expect(admitted.PromptRunID).To(Equal(promptRunID.String()))
		Expect(admitted.SessionID).To(Equal("11111111-1111-4111-8111-111111111111"))
	})

	It("streams a conflict after resolution so the client can offer force", func(ctx SpecContext) {
		workDir := GinkgoT().TempDir()
		configureAutomaticPlanToolPolicies(GinkgoTB(), workDir)
		server := &Server{ghOpts: github.Options{WorkDir: workDir}}
		created, err := uiTestProviderFor(workDir).Create(ctx, todos.CreateRequest{Title: "Confirm parallel run", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		previousStart := run.Start
		DeferCleanup(func() { run.Start = previousStart })
		run.Start = func(todoRunRequest) (todoRunStartResult, error) {
			return todoRunStartResult{}, &todos.ErrRunOwnedElsewhere{IssueID: uuid.Nil, PromptRunID: uuid.MustParse("22222222-2222-4222-8222-222222222222"), StepKind: "run", Owner: "another process"}
		}
		body, err := json.Marshal(todoRunPayload{Ref: todos.TODOReference(created), Spec: specPayload("gpt-5.5", "high")})
		Expect(err).NotTo(HaveOccurred())
		request := httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body)))
		request.Header.Set("Accept", "text/event-stream")
		recorder := httptest.NewRecorder()
		server.handleTodoRun(recorder, request)
		Expect(recorder.Body.String()).To(ContainSubstring("event: resolved"))
		Expect(recorder.Body.String()).To(ContainSubstring("event: failed\ndata: "))
		Expect(recorder.Body.String()).To(ContainSubstring("\"status\":409"))
		Expect(recorder.Body.String()).NotTo(ContainSubstring("event: admitted"))
	})

	It("streams approval continuation with the same prompt run identity", func(ctx SpecContext) {
		workDir := GinkgoT().TempDir()
		configureAutomaticPlanToolPolicies(GinkgoTB(), workDir)
		server := &Server{ghOpts: github.Options{WorkDir: workDir}}
		provider := uiTestProviderFor(workDir)
		created, err := provider.Create(ctx, todos.CreateRequest{Title: "Approve and run", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		review := types.StatusReview
		Expect(provider.UpdateState(ctx, created, todos.StateUpdate{Status: &review})).To(Succeed())
		previousStart := run.Start
		DeferCleanup(func() { run.Start = previousStart })
		promptRunID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
		run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
			Expect(req.Prepared).NotTo(BeNil())
			return todoRunStartResult{Status: "started", SessionID: "11111111-1111-4111-8111-111111111111", PromptRunID: promptRunID}, nil
		}
		body, err := json.Marshal(todoApprovePayload{Ref: todos.TODOReference(created), Run: true, Options: &todoRunPayload{Spec: specPayload("gpt-5.5", "high")}})
		Expect(err).NotTo(HaveOccurred())
		request := httptest.NewRequest(http.MethodPost, "/api/todos/plan/approve", strings.NewReader(string(body)))
		request.Header.Set("Accept", "text/event-stream")
		recorder := httptest.NewRecorder()
		server.handleTodoPlanApprove(recorder, request)
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		Expect(recorder.Body.String()).To(ContainSubstring("event: resolved"))
		Expect(recorder.Body.String()).To(ContainSubstring("event: admitted"))
		Expect(recorder.Body.String()).To(ContainSubstring("\"promptRunId\":\"" + promptRunID.String() + "\""))
	})

	It("streams a plan revision before admitting its continuation", func(ctx SpecContext) {
		workDir := GinkgoT().TempDir()
		server := &Server{ghOpts: github.Options{WorkDir: workDir}}
		provider := uiTestProviderFor(workDir)
		created, err := provider.Create(ctx, todos.CreateRequest{Title: "Revise the plan", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		review := types.StatusReview
		Expect(provider.UpdateState(ctx, created, todos.StateUpdate{Status: &review})).To(Succeed())
		previousStart := run.Start
		DeferCleanup(func() { run.Start = previousStart })
		promptRunID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
		run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
			Expect(req.Prepared).NotTo(BeNil())
			return todoRunStartResult{Status: "started", SessionID: "11111111-1111-4111-8111-111111111111", PromptRunID: promptRunID}, nil
		}
		body, err := json.Marshal(todoRevisePayload{Ref: todos.TODOReference(created), Feedback: "Keep the queue bounded", Options: &todoRunPayload{Spec: api.Spec{Model: api.Model{Name: "claude", Effort: "medium"}}}})
		Expect(err).NotTo(HaveOccurred())
		request := httptest.NewRequest(http.MethodPost, "/api/todos/plan/revise", strings.NewReader(string(body)))
		request.Header.Set("Accept", "text/event-stream")
		recorder := httptest.NewRecorder()
		server.handleTodoPlanRevise(recorder, request)
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		Expect(recorder.Body.String()).To(ContainSubstring("event: resolved"))
		Expect(recorder.Body.String()).To(ContainSubstring("event: admitted"))
		Expect(recorder.Body.String()).To(ContainSubstring("\"promptRunId\":\"" + promptRunID.String() + "\""))
	})

	It("streams an answered session's next prompt run", func(ctx SpecContext) {
		workDir := GinkgoT().TempDir()
		server := &Server{ghOpts: github.Options{WorkDir: workDir}}
		provider := uiTestProviderFor(workDir)
		created, err := provider.Create(ctx, todos.CreateRequest{Title: "Answer the agent", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		ask := types.StatusAsk
		Expect(provider.UpdateState(ctx, created, todos.StateUpdate{Status: &ask})).To(Succeed())
		sessionID := "sess-answer-1"
		Expect(provider.UpdateState(ctx, created, todos.StateUpdate{SessionID: &sessionID})).To(Succeed())
		seedAskingAttempt(provider, sessionID, types.PlanPhase)
		previousStart := run.Start
		DeferCleanup(func() { run.Start = previousStart })
		promptRunID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
		run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
			Expect(req.Prepared).NotTo(BeNil())
			return todoRunStartResult{Status: "started", SessionID: sessionID, PromptRunID: promptRunID}, nil
		}
		body, err := json.Marshal(todoAnswerPayload{Ref: todos.TODOReference(created), SessionID: sessionID, Answer: "Use PostgreSQL", Options: &todoRunPayload{Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeAgent, Effort: "medium"}}}})
		Expect(err).NotTo(HaveOccurred())
		request := httptest.NewRequest(http.MethodPost, "/api/todos/answer", strings.NewReader(string(body)))
		request.Header.Set("Accept", "text/event-stream")
		recorder := httptest.NewRecorder()
		server.handleTodoAnswer(recorder, request)
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		Expect(recorder.Body.String()).To(ContainSubstring("event: resolved"))
		Expect(recorder.Body.String()).To(ContainSubstring("event: admitted"))
		Expect(recorder.Body.String()).To(ContainSubstring("\"promptRunId\":\"" + promptRunID.String() + "\""))
	})
})
