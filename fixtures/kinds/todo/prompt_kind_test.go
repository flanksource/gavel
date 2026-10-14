package todo_test

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo_prompt kind", func() {
	var (
		stub    *stubGavel
		kind    fixtures.Kind
		ctx     context.Context
		env     []string
		planRun = uuid.MustParse("2c9f5e3a-7d4b-4a0c-9e8f-3b4c5d6e7f80")
		session = uuid.MustParse("3da06f4b-8e5c-4b1d-8f90-4c5d6e7f8091")
		prior   = uuid.MustParse("4eb1704c-9f6d-4c2e-9a01-5d6e7f8091a2")
	)

	admitted := func(step string) map[string]any {
		// pr/ui todoRunResponse as the admitted frame carries it.
		return map[string]any{
			"status": "started", "ref": stubTodoID, "count": 1, "dir": stubWorkspace, "step": step,
			"sessionId": session.String(), "promptRunId": planRun.String(), "timeout": "30m0s",
			"commit": false, "message": "Started 1 todo",
		}
	}
	resolved := map[string]any{"spec": map[string]any{}, "specYaml": "{}\n", "step": "plan"}

	BeforeEach(func() {
		stub = newStubGavel()
		kind = lookupKind("todo_prompt")
		ctx = context.Background()
		env = []string{"GAVEL_URL=" + stub.URL(), "TODO_ID=" + stubTodoID}
	})
	AfterEach(func() { stub.Close() })

	It("launches over SSE and reports the admitted attempt, its plan and session", func() {
		stub.set(func(s *stubGavel) {
			s.launchBody = sseFrames([2]any{"resolved", resolved}, [2]any{"admitted", admitted("plan")})
			s.attempts = []attemptDetail{
				newAttempt(planRun, session, "plan", captaindb.PromptRunStatusPlanning, captaindb.PromptRunStateRunning),
				newAttempt(prior, prior, "triage.new", captaindb.PromptRunStatusCompleted, captaindb.PromptRunStateSucceeded),
			}
			s.stats.SessionID, s.stats.State, s.stats.Found = session.String(), "working", true
			s.plan = map[string]any{"found": true, "content": "edit greet/greet.go", "ref": stubTodoID, "version": 3}
			s.item["planStatus"] = string(types.PlanNew)
		})
		step := newStep("todo_prompt", "plan", env, `{"step":"plan","force":true}`)

		outcome, err := kind.Execute(ctx, step)
		Expect(err).NotTo(HaveOccurred())

		launched := stub.recorded(http.MethodPost, "/api/todos/run")
		Expect(launched).To(HaveLen(1))
		Expect(launched[0].Accept).To(Equal("text/event-stream"))
		Expect(launched[0].Body).To(Equal(map[string]any{"ref": stubTodoID, "step": "plan", "force": true}))

		Expect(outcome.Outputs["launch"]).To(And(
			HaveKeyWithValue("failed", false),
			HaveKeyWithValue("error", ""),
			HaveKeyWithValue("sessionId", session.String()),
			HaveKeyWithValue("promptRunId", planRun.String()),
		))
		Expect(outcome.Outputs["session"]).To(And(
			HaveKeyWithValue("id", session.String()),
			HaveKeyWithValue("state", "planning"),
			HaveKeyWithValue("errors", []any{}),
		))
		Expect(outcome.Outputs["plan"]).To(Equal(map[string]any{
			"found": true, "content": "edit greet/greet.go", "status": "new",
		}))
		Expect(outcome.Outputs["verify"]).To(HaveKeyWithValue("ran", false))
	})

	It("fails loudly on a failed frame and keeps launch for retry.when", func() {
		stub.set(func(s *stubGavel) {
			s.launchBody = sseFrames([2]any{"resolved", resolved},
				[2]any{"failed", map[string]any{"error": "admission refused: run owned elsewhere", "status": http.StatusConflict}})
		})
		step := newStep("todo_prompt", "run", env, `{"step":"run"}`)

		outcome, err := kind.Execute(ctx, step)
		Expect(err).To(MatchError(ContainSubstring("admission refused")))
		Expect(outcome.Outputs["launch"]).To(And(
			HaveKeyWithValue("failed", true),
			HaveKeyWithValue("error", "admission refused: run owned elsewhere"),
			HaveKeyWithValue("status", http.StatusConflict),
		))
	})

	It("reports an unsent launch (no server) as not failed, so retry.when can evaluate and declines", func() {
		step := newStep("todo_prompt", "run", []string{"TODO_ID=" + stubTodoID}, `{"step":"run"}`)

		outcome, err := kind.Execute(ctx, step)
		Expect(err).To(MatchError(ContainSubstring("GAVEL_URL is not set")))
		Expect(outcome.Outputs["launch"]).To(And(
			HaveKeyWithValue("failed", false),
			HaveKeyWithValue("status", 0),
			HaveKeyWithValue("error", ContainSubstring("GAVEL_URL is not set")),
		))
		Expect(stub.recorded(http.MethodPost, "/api/todos/run")).To(BeEmpty())
	})

	It("fails loudly when the run is refused before the stream starts", func() {
		stub.set(func(s *stubGavel) {
			s.launchStatus, s.launchType = http.StatusBadRequest, "application/json"
			s.launchBody = `{"error":"invalid effort \"turbo\""}`
		})
		step := newStep("todo_prompt", "run", env, `{"step":"run","spec":{"effort":"turbo"}}`)

		outcome, err := kind.Execute(ctx, step)
		Expect(err).To(MatchError(ContainSubstring(`invalid effort "turbo"`)))
		Expect(stub.recorded(http.MethodPost, "/api/todos/run")[0].Body).To(HaveKeyWithValue("spec", map[string]any{"effort": "turbo"}))
		Expect(outcome.Outputs["launch"]).To(And(HaveKeyWithValue("failed", true), HaveKeyWithValue("status", http.StatusBadRequest)))
	})

	It("fails when the stream ends without admitting or failing", func() {
		stub.set(func(s *stubGavel) { s.launchBody = sseFrames([2]any{"resolved", resolved}) })
		step := newStep("todo_prompt", "run", env, `{"step":"run"}`)

		_, err := kind.Execute(ctx, step)
		Expect(err).To(MatchError(ContainSubstring("without an admitted or failed frame")))
	})

	It("observes the launched attempt's terminal state, session errors and verification", func() {
		stub.set(func(s *stubGavel) {
			s.launchBody = sseFrames([2]any{"resolved", resolved}, [2]any{"admitted", admitted("verify")})
			s.attempts = []attemptDetail{newAttempt(planRun, session, "verify", captaindb.PromptRunStatusRunning, captaindb.PromptRunStateRunning)}
		})
		step := newStep("todo_prompt", "verify", env, `{"step":"verify"}`)
		outcome, err := kind.Execute(ctx, step)
		Expect(err).NotTo(HaveOccurred())

		stub.set(func(s *stubGavel) {
			s.attempts[0].Status = captaindb.PromptRunStatusCompleted
			s.attempts[0].State = captaindb.PromptRunStateSucceeded
			s.attempts[0].ResultJSON = map[string]any{"endStatus": "completed"}
			s.attempts[0].Verification = &api.VerifyReport{Kind: "fixture", Ran: true, Passed: true, State: api.VerifyStatePassed,
				Summary: api.VerifySummary{Total: 1, Passed: 1}}
			s.stats.State, s.stats.Error = "error", "API Error: overloaded"
		})
		observed, err := kind.(fixtures.Observer).Observe(ctx, step, outcome.Handle)
		Expect(err).NotTo(HaveOccurred())
		Expect(observed["session"]).To(And(
			HaveKeyWithValue("state", "completed"),
			HaveKeyWithValue("runState", "succeeded"),
			HaveKeyWithValue("errors", []any{"API Error: overloaded"}),
		))
		Expect(observed["verify"]).To(And(HaveKeyWithValue("ran", true), HaveKeyWithValue("passed", true), HaveKeyWithValue("state", "passed")))
		Expect(observed["envelope"]).To(Equal(map[string]any{"endStatus": "completed"}))
		Expect(observed["launch"]).To(HaveKeyWithValue("failed", false))
	})

	It("fails an observation whose admitted run is not linked to the todo", func() {
		stub.set(func(s *stubGavel) {
			s.launchBody = sseFrames([2]any{"resolved", resolved}, [2]any{"admitted", admitted("plan")})
		})
		step := newStep("todo_prompt", "plan", env, `{"step":"plan"}`)

		_, err := kind.Execute(ctx, step)
		Expect(err).To(MatchError(ContainSubstring(planRun.String())))
	})

	It("waits by default for the launched attempt to be terminal", func() {
		Expect(kind.(fixtures.DefaultWaiter).DefaultWait(fixtures.StepSpec{Kind: "todo_prompt", Config: json.RawMessage(`{"step":"run"}`)})).
			To(Equal(&fixtures.WaitPolicy{
				Until:    "session.state in ['completed', 'failed', 'cancelled']",
				FailWhen: "launch.failed || size(session.errors) > 0",
			}))
	})

	It("publishes a JSON schema that requires step", func() {
		var schema map[string]any
		Expect(json.Unmarshal(kind.Schema(), &schema)).To(Succeed())
		Expect(schema).To(HaveKeyWithValue("required", []any{"step"}))
	})

	It("parses CRLF-delimited frames and ignores comment lines", func() {
		stub.set(func(s *stubGavel) {
			s.launchBody = ": keep-alive\r\nevent: failed\r\ndata: {\"error\":\"boom\",\r\ndata: \"status\":500}\r\n\r\n"
		})
		step := newStep("todo_prompt", "run", env, `{"step":"run"}`)

		outcome, err := kind.Execute(ctx, step)
		Expect(err).To(MatchError(ContainSubstring("boom")))
		Expect(outcome.Outputs["launch"]).To(HaveKeyWithValue("status", http.StatusInternalServerError))
	})

	It("requires a step", func() {
		step := newStep("todo_prompt", "run", env, `{}`)

		_, err := kind.Execute(ctx, step)
		Expect(err).To(MatchError(ContainSubstring("step is required")))
	})
})
