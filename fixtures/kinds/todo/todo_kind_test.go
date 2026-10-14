package todo_test

import (
	"context"
	"encoding/json"
	"net/http"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/kinds/todo"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo kind", func() {
	var (
		stub    *stubGavel
		kind    fixtures.Kind
		ctx     context.Context
		triage  = uuid.MustParse("0a7d3c1e-5b2f-4e8a-9c6d-1f2e3a4b5c6d")
		session = uuid.MustParse("1b8e4d2f-6c3a-4f9b-8d7e-2a3b4c5d6e7f")
	)

	BeforeEach(func() {
		stub = newStubGavel()
		kind = lookupKind("todo")
		ctx = context.Background()
	})
	AfterEach(func() { stub.Close() })

	It("publishes a JSON schema for its config", func() {
		var schema map[string]any
		Expect(json.Unmarshal(kind.Schema(), &schema)).To(Succeed())
		Expect(schema).To(HaveKeyWithValue("type", "object"))
		Expect(schema["properties"]).To(HaveKey("create"))
	})

	It("creates through /api/todos/new and exports TODO_ID and TODO_URL", func() {
		stub.set(func(s *stubGavel) {
			s.createResponse["triage"] = map[string]any{"status": "started", "sessionId": session.String(), "promptRunId": triage.String()}
			s.attempts = []attemptDetail{newAttempt(triage, session, "triage.new", captaindb.PromptRunStatusRunning, captaindb.PromptRunStateRunning)}
		})
		step := newStep("todo", "todo", []string{"GAVEL_URL=" + stub.URL()},
			`{"create":{"body":"add a greeting helper","labels":["enhancement"],"triage":true}}`)

		outcome, err := kind.Execute(ctx, step)
		Expect(err).NotTo(HaveOccurred())

		created := stub.recorded(http.MethodPost, "/api/todos/new")
		Expect(created).To(HaveLen(1))
		Expect(created[0].Body).To(Equal(map[string]any{
			"body": "add a greeting helper", "labels": []any{"enhancement"}, "triage": true,
		}))
		Expect(outcome.Exports).To(Equal(map[string]string{
			"TODO_ID":  stubTodoID,
			"TODO_URL": stub.URL() + "/todos/" + stubTodoID,
		}))
		Expect(outcome.Outputs).To(HaveKeyWithValue("id", stubTodoID))
		Expect(outcome.Outputs["todo"]).To(HaveKeyWithValue("title", "Add greeting helper"))
		Expect(outcome.Outputs["triage"]).To(Equal(map[string]any{
			"found": true, "state": "running", "runState": "running", "error": "",
			"step": "triage.new", "promptRunId": triage.String(), "sessionId": session.String(),
		}))
		Expect(outcome.Outputs["attempts"]).To(HaveLen(1))
	})

	It("fails the create when the triage run did not start", func() {
		stub.set(func(s *stubGavel) {
			s.createResponse["triage"] = map[string]any{"status": "failed", "error": "no model configured for triage.new"}
		})
		step := newStep("todo", "todo", []string{"GAVEL_URL=" + stub.URL()}, `{"create":{"body":"x","triage":true}}`)

		_, err := kind.Execute(ctx, step)
		Expect(err).To(MatchError(ContainSubstring("no model configured for triage.new")))
	})

	It("observes the triage attempt reaching a terminal state", func() {
		stub.set(func(s *stubGavel) {
			s.createResponse["triage"] = map[string]any{"status": "started", "sessionId": session.String(), "promptRunId": triage.String()}
			s.attempts = []attemptDetail{newAttempt(triage, session, "triage.new", captaindb.PromptRunStatusRunning, captaindb.PromptRunStateRunning)}
		})
		step := newStep("todo", "todo", []string{"GAVEL_URL=" + stub.URL()}, `{"create":{"body":"x","triage":true}}`)
		outcome, err := kind.Execute(ctx, step)
		Expect(err).NotTo(HaveOccurred())

		stub.set(func(s *stubGavel) {
			s.attempts[0].Status = captaindb.PromptRunStatusCompleted
			s.attempts[0].State = captaindb.PromptRunStateSucceeded
			s.item = map[string]any{"id": stubTodoID, "ref": stubTodoID, "cwd": stubWorkspace,
				"title": "Add Greeting helper to greet package", "status": string(types.StatusPending), "labels": []string{"enhancement"}}
		})
		observed, err := kind.(fixtures.Observer).Observe(ctx, step, outcome.Handle)
		Expect(err).NotTo(HaveOccurred())
		Expect(observed["triage"]).To(HaveKeyWithValue("state", "completed"))
		Expect(observed["todo"]).To(HaveKeyWithValue("labels", []any{"enhancement"}))
		Expect(observed["todo"]).To(HaveKeyWithValue("status", "pending"))
	})

	It("defaults to waiting for triage only when create asks for it", func() {
		waiter := kind.(fixtures.DefaultWaiter)
		Expect(waiter.DefaultWait(fixtures.StepSpec{Kind: "todo", Config: json.RawMessage(`{"create":{"body":"x","triage":true}}`)})).
			To(Equal(&fixtures.WaitPolicy{
				Until:    "triage.state in ['completed', 'failed', 'cancelled']",
				FailWhen: "triage.state in ['failed', 'cancelled']",
			}))
		Expect(waiter.DefaultWait(fixtures.StepSpec{Kind: "todo", Config: json.RawMessage(`{"create":{"body":"x"}}`)})).To(BeNil())
		Expect(waiter.DefaultWait(fixtures.StepSpec{Kind: "todo", Config: json.RawMessage(`{"approve":true}`)})).To(BeNil())
	})

	It("gets $TODO_ID by default and normalises absent labels to an empty list", func() {
		step := newStep("todo", "triaged", []string{"GAVEL_URL=" + stub.URL(), "TODO_ID=" + stubTodoID}, ``)

		outcome, err := kind.Execute(ctx, step)
		Expect(err).NotTo(HaveOccurred())
		Expect(stub.recorded(http.MethodGet, "/api/todos/item")[0].Query).To(Equal(map[string]string{"ref": stubTodoID}))
		Expect(outcome.Exports).To(BeEmpty())
		Expect(outcome.Outputs["todo"]).To(HaveKeyWithValue("labels", []any{}))
		Expect(outcome.Outputs["triage"]).To(HaveKeyWithValue("found", false))
	})

	It("approves the plan of the todo named by $TODO_ID", func() {
		stub.set(func(s *stubGavel) {
			s.item["status"] = string(types.StatusPending)
			s.reviewBody = map[string]any{"todo": s.item}
		})
		step := newStep("todo", "approve", []string{"GAVEL_URL=" + stub.URL(), "TODO_ID=" + stubTodoID}, `{"approve":true}`)

		outcome, err := kind.Execute(ctx, step)
		Expect(err).NotTo(HaveOccurred())
		Expect(stub.recorded(http.MethodPost, "/api/todos/plan/approve")[0].Body).To(Equal(map[string]any{"ref": stubTodoID}))
		Expect(outcome.Outputs["todo"]).To(HaveKeyWithValue("status", "pending"))
	})

	It("surfaces the dashboard's refusal of an approval", func() {
		stub.set(func(s *stubGavel) {
			s.reviewStatus = http.StatusConflict
			s.reviewBody = map[string]any{"error": "todo is not awaiting plan review (status: draft)"}
		})
		step := newStep("todo", "approve", []string{"GAVEL_URL=" + stub.URL(), "TODO_ID=" + stubTodoID}, `{"approve":true}`)

		_, err := kind.Execute(ctx, step)
		Expect(err).To(MatchError(And(ContainSubstring("409"), ContainSubstring("not awaiting plan review"))))
	})

	It("rejects through /api/todos/plan/reject with an explicit ref", func() {
		stub.set(func(s *stubGavel) { s.reviewBody = map[string]any{"todo": s.item} })
		step := newStep("todo", "reject", []string{"GAVEL_URL=" + stub.URL()}, `{"reject":true,"ref":"`+stubTodoID+`"}`)

		_, err := kind.Execute(ctx, step)
		Expect(err).NotTo(HaveOccurred())
		Expect(stub.recorded(http.MethodPost, "/api/todos/plan/reject")).To(HaveLen(1))
	})

	It("fails fast when neither server nor GAVEL_URL is set", func() {
		step := newStep("todo", "todo", nil, `{"create":{"body":"x"}}`)

		_, err := kind.Execute(ctx, step)
		Expect(err).To(MatchError(ContainSubstring("GAVEL_URL")))
	})

	It("fails fast when no ref is given and TODO_ID is unset", func() {
		step := newStep("todo", "todo", []string{"GAVEL_URL=" + stub.URL()}, `{"get":true}`)

		_, err := kind.Execute(ctx, step)
		Expect(err).To(MatchError(ContainSubstring("TODO_ID")))
	})

	It("rejects a step that names more than one action", func() {
		step := newStep("todo", "todo", []string{"GAVEL_URL=" + stub.URL(), "TODO_ID=" + stubTodoID}, `{"approve":true,"reject":true}`)

		_, err := kind.Execute(ctx, step)
		Expect(err).To(MatchError(ContainSubstring("exactly one of")))
	})
})
